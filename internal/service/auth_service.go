package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// passwordResetTTL is how long a reset token stays valid.
const passwordResetTTL = time.Hour

// emailVerificationTTL is how long an email-verification token stays valid.
// Longer than a reset token: the user may not check their inbox immediately.
const emailVerificationTTL = 24 * time.Hour

// AuthService implements the authentication use cases: register, login, token
// issuing/validation and password reset. It depends only on domain ports, so
// it can be unit-tested with fakes.
type AuthService struct {
	users         domain.UserRepository
	resets        domain.PasswordResetRepository
	verifications domain.EmailVerificationRepository
	mailer        domain.Mailer
	jwtSecret     []byte
	jwtExpiry     time.Duration
	appBaseURL    string // base URL for links in emails (e.g. the reset link)
}

// SetEmailVerification enables the email-verification flow (nil disables it, in
// which case registration issues no verification token). Wired from the
// composition root; kept a setter so existing constructor call sites (and their
// tests) are unaffected.
func (s *AuthService) SetEmailVerification(repo domain.EmailVerificationRepository) {
	s.verifications = repo
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

	// Email a verification link. A delivery failure must not abort registration
	// — the account exists and the user can request a fresh link later — so we
	// only log it.
	if err := s.sendVerification(ctx, user); err != nil {
		slog.Error("send verification email", "user_id", user.ID, "error", err)
	}

	return s.issue(user)
}

// sendVerification issues a single-use verification token for the user and
// emails the link. It is a no-op when the verification repository is not wired.
func (s *AuthService) sendVerification(ctx context.Context, user *domain.User) error {
	if s.verifications == nil {
		return nil
	}
	raw, err := randomToken()
	if err != nil {
		return err
	}
	token := &domain.EmailVerificationToken{
		UserID:    user.ID,
		TokenHash: hashToken(raw),
		ExpiresAt: time.Now().Add(emailVerificationTTL),
	}
	if err := s.verifications.Create(ctx, token); err != nil {
		return err
	}
	return s.mailer.Send(ctx, domain.Email{
		To:      user.Email,
		Subject: "Verify your email address",
		Body:    s.verifyEmailBody(user.Username, raw),
	})
}

// verifyEmailBody renders the plain-text email-verification message.
func (s *AuthService) verifyEmailBody(username, rawToken string) string {
	link := fmt.Sprintf("%s/verify-email?token=%s", s.appBaseURL, rawToken)
	return fmt.Sprintf(
		"Hi %s,\n\nWelcome! Please confirm your email address by opening the "+
			"link below within %s:\n\n%s\n\n"+
			"If you didn't create an account, you can safely ignore this email.\n",
		username, emailVerificationTTL, link)
}

// VerifyEmail redeems a verification token and marks the account verified. The
// token must be unknown-free, unexpired and unused; it is consumed on success.
func (s *AuthService) VerifyEmail(ctx context.Context, rawToken string) error {
	if s.verifications == nil {
		return domain.ErrInvalidVerificationToken
	}
	token, err := s.verifications.FindByHash(ctx, hashToken(rawToken))
	if err == domain.ErrVerificationTokenNotFound {
		return domain.ErrInvalidVerificationToken
	} else if err != nil {
		return err
	}
	if !token.Usable(time.Now()) {
		return domain.ErrInvalidVerificationToken
	}
	if err := s.users.MarkVerified(ctx, token.UserID); err != nil {
		return err
	}
	return s.verifications.MarkUsed(ctx, token.ID)
}

// ResendVerification issues a fresh verification link for a logged-in user. It
// is a no-op (nil) once the account is already verified so it can't be used to
// spam a confirmed address.
func (s *AuthService) ResendVerification(ctx context.Context, userID int) error {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.IsVerified {
		return domain.ErrAlreadyVerified
	}
	return s.sendVerification(ctx, user)
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

// ChangePassword sets a new password for a logged-in user after verifying the
// current one. Unlike the reset flow it requires knowledge of the old
// password, so a stolen session alone cannot lock the owner out silently.
func (s *AuthService) ChangePassword(ctx context.Context, userID int, current, newPassword string) error {
	if len(newPassword) < 6 {
		return ValidationError{Msg: "password must be at least 6 characters"}
	}
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(current)) != nil {
		return domain.ErrInvalidCredentials
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	return s.users.UpdatePassword(ctx, userID, string(hash))
}

// UpdateProfileInput is the editable slice of a user account: contact and
// default delivery location. Email, username and role are deliberately not
// editable here (identity/auth surface).
type UpdateProfileInput struct {
	PhoneNumber string
	Address     string
	City        string
	State       string
	ZipCode     string
	Latitude    *float64
	Longitude   *float64
	// EmailNotifications toggles order notification emails (#71); nil keeps
	// the current preference. Password-reset email is unaffected.
	EmailNotifications *bool
}

// UpdateProfile updates the caller's own contact/location fields and returns
// the fresh profile (password hash cleared).
func (s *AuthService) UpdateProfile(ctx context.Context, userID int, in UpdateProfileInput) (*domain.User, error) {
	if (in.Latitude == nil) != (in.Longitude == nil) {
		return nil, ValidationError{Msg: "latitude and longitude must be provided together"}
	}
	in.PhoneNumber = strings.TrimSpace(in.PhoneNumber)
	if len(in.PhoneNumber) > 20 {
		return nil, ValidationError{Msg: "phone_number must be at most 20 characters"}
	}
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	user.PhoneNumber = optional(in.PhoneNumber)
	user.Address = optional(in.Address)
	user.City = optional(in.City)
	user.State = optional(in.State)
	user.ZipCode = optional(in.ZipCode)
	user.Latitude = in.Latitude
	user.Longitude = in.Longitude
	if in.EmailNotifications != nil {
		user.EmailNotifications = *in.EmailNotifications
	}
	if err := s.users.UpdateProfile(ctx, user); err != nil {
		return nil, err
	}
	user.PasswordHash = ""
	return user, nil
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
