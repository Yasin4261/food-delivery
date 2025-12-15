package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// UserService provides business logic for user operations
type UserService struct {
	userRepo  domain.UserRepository
	jwtSecret []byte
}

// NewUserService creates a new user service instance
func NewUserService(userRepo domain.UserRepository, jwtSecret string) *UserService {
	return &UserService{
		userRepo:  userRepo,
		jwtSecret: []byte(jwtSecret),
	}
}

// JWTClaims represents the JWT token claims
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// RegisterRequest represents the registration request data
type RegisterRequest struct {
	Username  string  `json:"username"`
	Email     string  `json:"email"`
	Password  string  `json:"password"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Phone     string  `json:"phone"`
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
	Role      string  `json:"role"` // customer, chef, admin
}

// LoginRequest represents the login request data
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      *domain.User `json:"user"`
}

// Register creates a new user account
func (s *UserService) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	// Validate input
	if err := s.validateRegisterRequest(req); err != nil {
		return nil, err
	}

	// Check if email already exists
	existingUser, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, domain.ErrEmailAlreadyExists
	}

	// Check if username already exists
	existingUser, err = s.userRepo.FindByUsername(ctx, req.Username)
	if err == nil && existingUser != nil {
		return nil, errors.New("username already exists")
	}

	// Hash password
	passwordHash, err := s.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user entity
	user := domain.NewUser(
		req.Username,
		req.Email,
		passwordHash,
		req.FirstName,
		req.LastName,
		req.Phone,
	)

	// Set role (default to customer if not specified)
	if req.Role == "" {
		req.Role = domain.RoleCustomer
	}
	user.Role = req.Role

	// Set location if provided
	if req.Latitude != 0 && req.Longitude != 0 {
		user.Latitude = &req.Latitude
		user.Longitude = &req.Longitude
	}

	// Save user to database
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate JWT token
	token, expiresAt, err := s.GenerateJWT(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Clear password hash before returning
	user.PasswordHash = ""

	return &AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user,
	}, nil
}

// Login authenticates a user and returns a JWT token
func (s *UserService) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	// Validate input
	if req.Email == "" {
		return nil, errors.New("email is required")
	}
	if req.Password == "" {
		return nil, errors.New("password is required")
	}

	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	// Verify password
	if err := s.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("user account is deactivated")
	}

	// Update last login timestamp
	now := time.Now()
	user.LastLoginAt = &now
	if err := s.userRepo.Update(ctx, user); err != nil {
		// Log error but don't fail login
		fmt.Printf("failed to update last login: %v\n", err)
	}

	// Generate JWT token
	token, expiresAt, err := s.GenerateJWT(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Clear password hash before returning
	user.PasswordHash = ""

	return &AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user,
	}, nil
}

// GetProfile retrieves user profile by ID
func (s *UserService) GetProfile(ctx context.Context, userID string) (*domain.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	// Clear password hash
	user.PasswordHash = ""

	return user, nil
}

// UpdateProfile updates user profile information
func (s *UserService) UpdateProfile(ctx context.Context, userID string, updates map[string]interface{}) (*domain.User, error) {
	// Get existing user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	// Apply updates
	if firstName, ok := updates["first_name"].(string); ok {
		user.FirstName = firstName
	}
	if lastName, ok := updates["last_name"].(string); ok {
		user.LastName = lastName
	}
	if phone, ok := updates["phone"].(string); ok {
		user.Phone = phone
	}
	if profileImage, ok := updates["profile_image"].(string); ok {
		user.ProfileImageURL = &profileImage
	}

	// Update in database
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	// Clear password hash
	user.PasswordHash = ""

	return user, nil
}

// UpdateLocation updates user's location coordinates
func (s *UserService) UpdateLocation(ctx context.Context, userID string, lat, lng float64) error {
	// Validate coordinates
	if lat < -90 || lat > 90 {
		return errors.New("invalid latitude: must be between -90 and 90")
	}
	if lng < -180 || lng > 180 {
		return errors.New("invalid longitude: must be between -180 and 180")
	}

	return s.userRepo.UpdateLocation(ctx, userID, lat, lng)
}

// ChangePassword changes user's password
func (s *UserService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return domain.ErrUserNotFound
	}

	// Verify current password
	if err := s.VerifyPassword(user.PasswordHash, currentPassword); err != nil {
		return errors.New("current password is incorrect")
	}

	// Validate new password
	if len(newPassword) < 6 {
		return errors.New("new password must be at least 6 characters long")
	}

	// Hash new password
	newPasswordHash, err := s.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update password
	user.PasswordHash = newPasswordHash
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// FindNearbyUsers finds users within a specified radius
func (s *UserService) FindNearbyUsers(ctx context.Context, lat, lng, radiusKm float64, limit, offset int) ([]*domain.User, error) {
	users, err := s.userRepo.FindNearby(ctx, lat, lng, radiusKm, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find nearby users: %w", err)
	}

	// Clear password hashes
	for _, user := range users {
		user.PasswordHash = ""
	}

	return users, nil
}

// ListUsers retrieves a paginated list of users
func (s *UserService) ListUsers(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*domain.User, error) {
	users, err := s.userRepo.List(ctx, filters, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// Clear password hashes
	for _, user := range users {
		user.PasswordHash = ""
	}

	return users, nil
}

// HashPassword hashes a plaintext password using bcrypt
func (s *UserService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword verifies a plaintext password against a hash
func (s *UserService) VerifyPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// GenerateJWT generates a JWT token for the user
func (s *UserService) GenerateJWT(user *domain.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(24 * time.Hour) // Token valid for 24 hours

	claims := JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "food-delivery-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// ValidateJWT validates a JWT token and returns the claims
func (s *UserService) ValidateJWT(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// validateRegisterRequest validates the registration request
func (s *UserService) validateRegisterRequest(req RegisterRequest) error {
	if req.Username == "" {
		return errors.New("username is required")
	}
	if len(req.Username) < 3 {
		return errors.New("username must be at least 3 characters long")
	}
	if req.Email == "" {
		return errors.New("email is required")
	}
	if req.Password == "" {
		return errors.New("password is required")
	}
	if len(req.Password) < 6 {
		return errors.New("password must be at least 6 characters long")
	}
	if req.FirstName == "" {
		return errors.New("first name is required")
	}
	if req.LastName == "" {
		return errors.New("last name is required")
	}
	if req.Phone == "" {
		return errors.New("phone is required")
	}
	if req.Role != "" && req.Role != domain.RoleCustomer && req.Role != domain.RoleChef && req.Role != domain.RoleAdmin {
		return fmt.Errorf("invalid role: must be one of %s, %s, %s", domain.RoleCustomer, domain.RoleChef, domain.RoleAdmin)
	}

	return nil
}
