package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// passwordResetTTL is how long a reset token stays valid.
const passwordResetTTL = time.Hour

// AuthService implements the authentication use cases: register, login, token
// issuing/validation and password reset. It depends only on domain ports, so
// it can be unit-tested with fakes.
type AuthService struct {
	users      domain.UserRepository
	resets     domain.PasswordResetRepository
	mailer     domain.Mailer
	jwtSecret  []byte
	jwtExpiry  time.Duration
	appBaseURL string // base URL for links in emails (e.g. the reset link)
}

// NewAuthService builds an AuthService. appBaseURL is the public base URL used
// to build links in transactional emails.
func NewAuthService(users domain.UserRepository, resets domain.PasswordResetRepository, mailer domain.Mailer, jwtSecret string, jwtExpiry time.Duration, appBaseURL string) *AuthService {
	return &AuthService{
		users:      users,
		resets:     resets,
		mailer:     mailer,
		jwtSecret:  []byte(jwtSecret),
		jwtExpiry:  jwtExpiry,
		appBaseURL: strings.TrimRight(appBaseURL, "/"),
	}
}

// ValidationError marks bad client input (as opposed to a domain rule
// violation). Handlers map it to HTTP 400.
type ValidationError struct{ Msg string }

func (e ValidationError) Error() string { return e.Msg }

// Claims is the JWT payload carried in the bearer token.
type Claims struct {
	UserID int    `json:"uid"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// RegisterInput is the data needed to create an account.
type RegisterInput struct {
	Username    string
	Email       string
	Password    string
	PhoneNumber string
	Role        string
}

// AuthResult is returned on successful register/login.
type AuthResult struct {
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      *domain.User `json:"user"`
}

// Register validates input, ensures email/username are free, hashes the
// password and persists the user, then issues a token.
func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*AuthResult, error) {
	in.Username = strings.TrimSpace(in.Username)
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))

	if err := validateRegister(in); err != nil {
		return nil, err
	}

	if _, err := s.users.FindByEmail(ctx, in.Email); err == nil {
		return nil, domain.ErrEmailAlreadyExists
	} else if err != domain.ErrUserNotFound {
		return nil, err
	}

	if _, err := s.users.FindByUsername(ctx, in.Username); err == nil {
		return nil, domain.ErrUsernameAlreadyExists
	} else if err != domain.ErrUserNotFound {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := domain.NewUser(in.Username, in.Email, string(hash))
	if in.Role != "" {
		user.Role = in.Role
	}
	if in.PhoneNumber != "" {
		user.PhoneNumber = &in.PhoneNumber
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	return s.issue(user)
}

// Login verifies credentials and issues a token.
func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.users.FindByEmail(ctx, email)
	if err == domain.ErrUserNotFound {
		return nil, domain.ErrInvalidCredentials // do not leak which part failed
	} else if err != nil {
		return nil, err
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, domain.ErrInvalidCredentials
	}
	if !user.IsActive {
		return nil, domain.ErrAccountInactive
	}

	return s.issue(user)
}

// Profile returns the account for an authenticated user id.
func (s *AuthService) Profile(ctx context.Context, userID int) (*domain.User, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = ""
	return user, nil
}

// RequestPasswordReset issues a single-use reset token for the account with the
// given email and emails the reset link. To avoid leaking which emails are
// registered it is a silent no-op when no account matches — the caller responds
// identically either way, and no token is ever returned to the client.
func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.users.FindByEmail(ctx, email)
	if err == domain.ErrUserNotFound {
		return nil
	} else if err != nil {
		return err
	}

	raw, err := randomToken()
	if err != nil {
		return err
	}
	token := &domain.PasswordResetToken{
		UserID:    user.ID,
		TokenHash: hashToken(raw),
		ExpiresAt: time.Now().Add(passwordResetTTL),
	}
	if err := s.resets.Create(ctx, token); err != nil {
		return err
	}

	return s.mailer.Send(ctx, domain.Email{
		To:      user.Email,
		Subject: "Reset your password",
		Body:    s.resetEmailBody(user.Username, raw),
	})
}

// resetEmailBody renders the plain-text password-reset email.
func (s *AuthService) resetEmailBody(username, rawToken string) string {
	link := fmt.Sprintf("%s/reset-password?token=%s", s.appBaseURL, rawToken)
	return fmt.Sprintf(
		"Hi %s,\n\nWe received a request to reset your password. "+
			"Use the link below within %s to choose a new one:\n\n%s\n\n"+
			"If you didn't request this, you can safely ignore this email.\n",
		username, passwordResetTTL, link)
}

// ResetPassword redeems a reset token and sets a new password. The token must
// be unknown-free, unexpired and unused; it is consumed on success.
func (s *AuthService) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	if len(newPassword) < 6 {
		return ValidationError{Msg: "password must be at least 6 characters"}
	}

	token, err := s.resets.FindByHash(ctx, hashToken(rawToken))
	if err == domain.ErrResetTokenNotFound {
		return domain.ErrInvalidResetToken
	} else if err != nil {
		return err
	}
	if !token.Usable(time.Now()) {
		return domain.ErrInvalidResetToken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	if err := s.users.UpdatePassword(ctx, token.UserID, string(hash)); err != nil {
		return err
	}
	return s.resets.MarkUsed(ctx, token.ID)
}

// randomToken returns a 32-byte random token, hex-encoded.
func randomToken() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return hex.EncodeToString(b[:]), nil
}

// hashToken returns the sha256 hex digest stored for a raw token.
func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// ParseToken validates a signed token and returns its claims.
func (s *AuthService) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

// issue mints a JWT for the user and clears the hash before returning.
func (s *AuthService) issue(user *domain.User) (*AuthResult, error) {
	expiresAt := time.Now().Add(s.jwtExpiry)

	jti, err := randomToken()
	if err != nil {
		return nil, err
	}
	claims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti, // enables per-token revocation (see TokenDenylist)
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "food-delivery-api",
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("sign token: %w", err)
	}

	user.PasswordHash = ""
	return &AuthResult{Token: signed, ExpiresAt: expiresAt, User: user}, nil
}

func validateRegister(in RegisterInput) error {
	switch {
	case len(in.Username) < 3:
		return ValidationError{Msg: "username must be at least 3 characters"}
	case !strings.Contains(in.Email, "@"):
		return ValidationError{Msg: "a valid email is required"}
	case len(in.Password) < 6:
		return ValidationError{Msg: "password must be at least 6 characters"}
	case in.Role != "" && !domain.ValidRole(in.Role):
		return ValidationError{Msg: "invalid role: must be customer or chef"}
	case in.Role == domain.RoleAdmin:
		// Privileged roles must never be self-assigned at registration.
		return ValidationError{Msg: "the admin role cannot be self-assigned"}
	}
	return nil
}
