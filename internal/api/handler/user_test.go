package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ecommerce/internal/model"
	"github.com/gin-gonic/gin"
)

// MockUserService - User service mock
type MockUserService struct {
	users map[uint]*model.User
}

func NewMockUserService() *MockUserService {
	return &MockUserService{
		users: make(map[uint]*model.User),
	}
}

func (m *MockUserService) Register(req *model.RegisterRequest) (*model.AuthResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}
	
	user := &model.User{
		ID:        uint(len(m.users) + 1),
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      "customer",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	m.users[user.ID] = user
	
	return &model.AuthResponse{
		Token: "mock-jwt-token",
		User:  *user,
	}, nil
}

func (m *MockUserService) Login(req *model.LoginRequest) (*model.AuthResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}
	
	// Find user by email
	for _, user := range m.users {
		if user.Email == req.Email {
			return &model.AuthResponse{
				Token: "mock-jwt-token",
				User:  *user,
			}, nil
		}
	}
	
	return nil, errors.New("user not found")
}

func (m *MockUserService) GetProfile(userID uint) (*model.UserProfileResponse, error) {
	user, exists := m.users[userID]
	if !exists {
		return nil, errors.New("user not found")
	}
	
	return &model.UserProfileResponse{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (m *MockUserService) UpdateProfile(userID uint, req *model.UpdateProfileRequest) (*model.UserProfileResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}
	
	user, exists := m.users[userID]
	if !exists {
		return nil, errors.New("user not found")
	}
	
	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.UpdatedAt = time.Now()
	
	return &model.UserProfileResponse{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// Setup test router
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

// TestUserHandler_Register - Test user registration handler
func TestUserHandler_Register(t *testing.T) {
	userService := NewMockUserService()
	userHandler := NewUserHandler(userService)
	
	router := setupTestRouter()
	router.POST("/register", userHandler.Register)

	// Test successful registration
	registerRequest := model.RegisterRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	}

	body, _ := json.Marshal(registerRequest)
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	var response model.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success response")
	}

	// Test invalid JSON
	req, _ = http.NewRequest("POST", "/register", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	// Test empty body
	req, _ = http.NewRequest("POST", "/register", bytes.NewBuffer([]byte("")))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestUserHandler_Login - Test user login handler
func TestUserHandler_Login(t *testing.T) {
	userService := NewMockUserService()
	userHandler := NewUserHandler(userService)
	
	// Create a user first
	userService.Register(&model.RegisterRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	})
	
	router := setupTestRouter()
	router.POST("/login", userHandler.Login)

	// Test successful login
	loginRequest := model.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	body, _ := json.Marshal(loginRequest)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response model.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success response")
	}

	// Test login with non-existent user
	loginRequest.Email = "nonexistent@example.com"
	body, _ = json.Marshal(loginRequest)
	req, _ = http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestUserHandler_GetProfile - Test get profile handler
func TestUserHandler_GetProfile(t *testing.T) {
	userService := NewMockUserService()
	userHandler := NewUserHandler(userService)
	
	// Create a user first
	userService.Register(&model.RegisterRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	})
	
	router := setupTestRouter()
	router.GET("/profile", func(c *gin.Context) {
		// Mock authentication middleware
		c.Set("user_id", uint(1))
		userHandler.GetProfile(c)
	})

	// Test successful profile retrieval
	req, _ := http.NewRequest("GET", "/profile", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response model.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success response")
	}
}

// TestUserHandler_UpdateProfile - Test update profile handler
func TestUserHandler_UpdateProfile(t *testing.T) {
	userService := NewMockUserService()
	userHandler := NewUserHandler(userService)
	
	// Create a user first
	userService.Register(&model.RegisterRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	})
	
	router := setupTestRouter()
	router.PUT("/profile", func(c *gin.Context) {
		// Mock authentication middleware
		c.Set("user_id", uint(1))
		userHandler.UpdateProfile(c)
	})

	// Test successful profile update
	updateRequest := model.UpdateProfileRequest{
		FirstName: "Jane",
		LastName:  "Smith",
	}

	body, _ := json.Marshal(updateRequest)
	req, _ := http.NewRequest("PUT", "/profile", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response model.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success response")
	}
}

// TestUserHandler_ValidationErrors - Test validation error handling
func TestUserHandler_ValidationErrors(t *testing.T) {
	userService := NewMockUserService()
	userHandler := NewUserHandler(userService)
	
	router := setupTestRouter()
	router.POST("/register", userHandler.Register)

	// Test registration with invalid email
	registerRequest := model.RegisterRequest{
		Email:     "invalid-email",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	}

	body, _ := json.Marshal(registerRequest)
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	// Test registration with short password
	registerRequest.Email = "test@example.com"
	registerRequest.Password = "123"

	body, _ = json.Marshal(registerRequest)
	req, _ = http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestUserHandler_HTTPMethods - Test HTTP method validation
func TestUserHandler_HTTPMethods(t *testing.T) {
	userService := NewMockUserService()
	userHandler := NewUserHandler(userService)
	
	router := setupTestRouter()
	router.POST("/register", userHandler.Register)

	// Test GET request on POST endpoint
	req, _ := http.NewRequest("GET", "/register", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}

	// Test PUT request on POST endpoint
	req, _ = http.NewRequest("PUT", "/register", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

// TestUserHandler_ContentType - Test content type validation
func TestUserHandler_ContentType(t *testing.T) {
	userService := NewMockUserService()
	userHandler := NewUserHandler(userService)
	
	router := setupTestRouter()
	router.POST("/register", userHandler.Register)

	// Test without Content-Type header
	registerRequest := model.RegisterRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	}

	body, _ := json.Marshal(registerRequest)
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	// No Content-Type header set

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should still work as Gin is flexible with content types
	if w.Code != http.StatusCreated && w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d or %d, got %d", http.StatusCreated, http.StatusBadRequest, w.Code)
	}
}

// BenchmarkUserHandler_Register - Benchmark registration handler
func BenchmarkUserHandler_Register(b *testing.B) {
	userService := NewMockUserService()
	userHandler := NewUserHandler(userService)
	
	router := setupTestRouter()
	router.POST("/register", userHandler.Register)

	registerRequest := model.RegisterRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	}

	body, _ := json.Marshal(registerRequest)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// Test helper functions
func assertStatusCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected status code %d, got %d", expected, actual)
	}
}

func assertSuccessResponse(t *testing.T, body []byte) {
	var response model.APIResponse
	err := json.Unmarshal(body, &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	
	if !response.Success {
		t.Error("Expected success response")
	}
}

func assertErrorResponse(t *testing.T, body []byte) {
	var response model.APIResponse
	err := json.Unmarshal(body, &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	
	if response.Success {
		t.Error("Expected error response")
	}
}
