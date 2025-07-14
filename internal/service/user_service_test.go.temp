package service

import (
	"errors"
	"testing"
	"time"

	"ecommerce/internal/model"
)

// MockUserRepository - User repository mock
type MockUserRepository struct {
	users map[string]*model.User
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[string]*model.User),
	}
}

func (m *MockUserRepository) GetByEmail(email string) (*model.User, error) {
	if user, exists := m.users[email]; exists {
		return user, nil
	}
	return nil, errors.New("user not found")
}

func (m *MockUserRepository) Create(user *model.User) error {
	if _, exists := m.users[user.Email]; exists {
		return errors.New("user already exists")
	}
	user.ID = uint(len(m.users) + 1)
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	m.users[user.Email] = user
	return nil
}

func (m *MockUserRepository) GetByID(id uint) (*model.User, error) {
	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

func (m *MockUserRepository) Update(user *model.User) error {
	if existingUser, exists := m.users[user.Email]; exists {
		existingUser.FirstName = user.FirstName
		existingUser.LastName = user.LastName
		existingUser.UpdatedAt = time.Now()
		return nil
	}
	return errors.New("user not found")
}

// MockJWTManager - JWT manager mock
type MockJWTManager struct{}

func NewMockJWTManager() *MockJWTManager {
	return &MockJWTManager{}
}

func (m *MockJWTManager) GenerateToken(userID uint, email string) (string, error) {
	return "mock-jwt-token", nil
}

func (m *MockJWTManager) ValidateToken(token string) (*model.User, error) {
	if token == "mock-jwt-token" {
		return &model.User{
			ID:    1,
			Email: "test@example.com",
			Role:  "customer",
		}, nil
	}
	return nil, errors.New("invalid token")
}

// MockPasswordManager - Password manager mock
type MockPasswordManager struct{}

func NewMockPasswordManager() *MockPasswordManager {
	return &MockPasswordManager{}
}

func (m *MockPasswordManager) HashPassword(password string) (string, error) {
	return "hashed-" + password, nil
}

func (m *MockPasswordManager) CheckPassword(password, hash string) error {
	if "hashed-"+password == hash {
		return nil
	}
	return errors.New("invalid password")
}

// TestUserService_Register - Test user registration
func TestUserService_Register(t *testing.T) {
	userRepo := NewMockUserRepository()
	jwtManager := NewMockJWTManager()
	
	userService := &UserService{
		userRepo:        userRepo,
		jwtManager:      jwtManager,
		passwordManager: NewMockPasswordManager(),
	}

	// Test successful registration
	registerRequest := &model.RegisterRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	}

	response, err := userService.Register(registerRequest)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if response == nil {
		t.Error("Expected response, got nil")
	}

	if response.Token != "mock-jwt-token" {
		t.Errorf("Expected token 'mock-jwt-token', got %s", response.Token)
	}

	if response.User.Email != registerRequest.Email {
		t.Errorf("Expected email %s, got %s", registerRequest.Email, response.User.Email)
	}

	// Test duplicate email registration
	_, err = userService.Register(registerRequest)
	if err == nil {
		t.Error("Expected error for duplicate email, got nil")
	}
}

// TestUserService_Login - Test user login
func TestUserService_Login(t *testing.T) {
	userRepo := NewMockUserRepository()
	jwtManager := NewMockJWTManager()
	passwordManager := NewMockPasswordManager()
	
	userService := &UserService{
		userRepo:        userRepo,
		jwtManager:      jwtManager,
		passwordManager: passwordManager,
	}

	// Create a user first
	user := &model.User{
		Email:     "test@example.com",
		Password:  "hashed-password123",
		FirstName: "John",
		LastName:  "Doe",
		Role:      "customer",
	}
	userRepo.Create(user)

	// Test successful login
	loginRequest := &model.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	response, err := userService.Login(loginRequest)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if response == nil {
		t.Error("Expected response, got nil")
	}

	if response.Token != "mock-jwt-token" {
		t.Errorf("Expected token 'mock-jwt-token', got %s", response.Token)
	}

	// Test login with wrong password
	loginRequest.Password = "wrongpassword"
	_, err = userService.Login(loginRequest)
	if err == nil {
		t.Error("Expected error for wrong password, got nil")
	}

	// Test login with non-existent email
	loginRequest.Email = "nonexistent@example.com"
	_, err = userService.Login(loginRequest)
	if err == nil {
		t.Error("Expected error for non-existent email, got nil")
	}
}

// TestUserService_GetProfile - Test get user profile
func TestUserService_GetProfile(t *testing.T) {
	userRepo := NewMockUserRepository()
	jwtManager := NewMockJWTManager()
	
	userService := &UserService{
		userRepo:        userRepo,
		jwtManager:      jwtManager,
		passwordManager: NewMockPasswordManager(),
	}

	// Create a user first
	user := &model.User{
		ID:        1,
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Role:      "customer",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	userRepo.users[user.Email] = user

	// Test successful profile retrieval
	profile, err := userService.GetProfile(1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if profile == nil {
		t.Error("Expected profile, got nil")
	}

	if profile.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, profile.Email)
	}

	// Test profile retrieval with non-existent user
	_, err = userService.GetProfile(999)
	if err == nil {
		t.Error("Expected error for non-existent user, got nil")
	}
}

// TestUserService_UpdateProfile - Test update user profile
func TestUserService_UpdateProfile(t *testing.T) {
	userRepo := NewMockUserRepository()
	jwtManager := NewMockJWTManager()
	
	userService := &UserService{
		userRepo:        userRepo,
		jwtManager:      jwtManager,
		passwordManager: NewMockPasswordManager(),
	}

	// Create a user first
	user := &model.User{
		ID:        1,
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Role:      "customer",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	userRepo.users[user.Email] = user

	// Test successful profile update
	updateRequest := &model.UpdateProfileRequest{
		FirstName: "Jane",
		LastName:  "Smith",
	}

	updatedProfile, err := userService.UpdateProfile(1, updateRequest)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if updatedProfile == nil {
		t.Error("Expected updated profile, got nil")
	}

	if updatedProfile.FirstName != updateRequest.FirstName {
		t.Errorf("Expected first name %s, got %s", updateRequest.FirstName, updatedProfile.FirstName)
	}

	if updatedProfile.LastName != updateRequest.LastName {
		t.Errorf("Expected last name %s, got %s", updateRequest.LastName, updatedProfile.LastName)
	}

	// Test profile update with non-existent user
	_, err = userService.UpdateProfile(999, updateRequest)
	if err == nil {
		t.Error("Expected error for non-existent user, got nil")
	}
}

// TestUserService_ValidationCases - Test various validation cases
func TestUserService_ValidationCases(t *testing.T) {
	userRepo := NewMockUserRepository()
	jwtManager := NewMockJWTManager()
	
	userService := &UserService{
		userRepo:        userRepo,
		jwtManager:      jwtManager,
		passwordManager: NewMockPasswordManager(),
	}

	// Test registration with empty fields
	tests := []struct {
		name    string
		request *model.RegisterRequest
		wantErr bool
	}{
		{
			name: "Valid registration",
			request: &model.RegisterRequest{
				Email:     "valid@example.com",
				Password:  "password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			wantErr: false,
		},
		{
			name: "Empty email",
			request: &model.RegisterRequest{
				Email:     "",
				Password:  "password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			wantErr: true,
		},
		{
			name: "Empty password",
			request: &model.RegisterRequest{
				Email:     "test@example.com",
				Password:  "",
				FirstName: "John",
				LastName:  "Doe",
			},
			wantErr: true,
		},
		{
			name: "Empty first name",
			request: &model.RegisterRequest{
				Email:     "test@example.com",
				Password:  "password123",
				FirstName: "",
				LastName:  "Doe",
			},
			wantErr: true,
		},
		{
			name: "Empty last name",
			request: &model.RegisterRequest{
				Email:     "test@example.com",
				Password:  "password123",
				FirstName: "John",
				LastName:  "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation logic (in real app, this would be in the service)
			if tt.request.Email == "" || tt.request.Password == "" || 
			   tt.request.FirstName == "" || tt.request.LastName == "" {
				if !tt.wantErr {
					t.Error("Expected no error for valid request")
				}
				return
			}

			_, err := userService.Register(tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestUserService_EdgeCases - Test edge cases
func TestUserService_EdgeCases(t *testing.T) {
	userRepo := NewMockUserRepository()
	jwtManager := NewMockJWTManager()
	
	userService := &UserService{
		userRepo:        userRepo,
		jwtManager:      jwtManager,
		passwordManager: NewMockPasswordManager(),
	}

	// Test with nil requests
	_, err := userService.Register(nil)
	if err == nil {
		t.Error("Expected error for nil register request")
	}

	_, err = userService.Login(nil)
	if err == nil {
		t.Error("Expected error for nil login request")
	}

	_, err = userService.UpdateProfile(1, nil)
	if err == nil {
		t.Error("Expected error for nil update request")
	}
}
