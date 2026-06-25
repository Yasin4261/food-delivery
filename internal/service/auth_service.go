package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// AuthService implements the authentication use cases: register, login and
// token issuing/validation. It depends only on the domain.UserRepository port,
// so it can be unit-tested with a fake repository.
type AuthService struct {
	users     domain.UserRepository
	jwtSecret []byte
	jwtExpiry time.Duration
}

// NewAuthService builds an AuthService.
func NewAuthService(users domain.UserRepository, jwtSecret string, jwtExpiry time.Duration) *AuthService {
	return &AuthService{
		users:     users,
		jwtSecret: []byte(jwtSecret),
		jwtExpiry: jwtExpiry,
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

	claims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
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
		return ValidationError{Msg: "invalid role: must be customer, chef or admin"}
	}
	return nil
}
