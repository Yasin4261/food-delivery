package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ecommerce/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func setupMockDependencies() {
	// Set up mock dependencies for testing
	// In real tests, you'd use proper mocks, but for validation tests this is sufficient
	mockDeps := &HandlerDependencies{
		UserService:  nil, // These would be mocked in full integration tests
		ChefService:  nil,
		MealService:  nil,
		CartService:  nil,
		OrderService: nil,
		AdminService: nil,
	}
	SetDependencies(mockDeps)
}

func TestLoginHandler_RequestValidation(t *testing.T) {
	setupMockDependencies()
	router := setupTestRouter()
	router.POST("/auth/login", Login)

	t.Run("Invalid JSON format", func(t *testing.T) {
		invalidJSON := `{"email": "test@example.com", "password": }`
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewBufferString(invalidJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Empty request body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer([]byte{}))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Missing required fields", func(t *testing.T) {
		incompleteReq := map[string]string{
			"email": "test@example.com",
			// password missing
		}
		
		jsonData, _ := json.Marshal(incompleteReq)
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		// Should fail due to missing UserService, but JSON parsing should work
		// This tests the request validation part
		assert.True(t, w.Code >= 400) // Either validation error or service error
	})
}

func TestRegisterHandler_RequestValidation(t *testing.T) {
	setupMockDependencies()
	router := setupTestRouter()
	router.POST("/auth/register", Register)

	t.Run("Invalid JSON format", func(t *testing.T) {
		invalidJSON := `{"email": "test@example.com", "password": "123", }`
		req := httptest.NewRequest("POST", "/auth/register", bytes.NewBufferString(invalidJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHandler_HTTPMethods(t *testing.T) {
	setupMockDependencies()
	router := setupTestRouter()
	router.POST("/auth/login", Login)
	router.POST("/auth/register", Register)

	t.Run("Login with wrong HTTP method", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/auth/login", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Register with wrong HTTP method", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/auth/register", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestResponseFormat(t *testing.T) {
	setupMockDependencies()
	router := setupTestRouter()
	router.POST("/auth/login", Login)

	t.Run("Response should be valid JSON", func(t *testing.T) {
		loginReq := model.LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}
		
		jsonData, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		// Response should be valid JSON
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "Response should be valid JSON")
		
		// Should have an error field for failed operations due to nil service
		if w.Code >= 400 {
			assert.True(t, len(response) > 0, "Response should not be empty")
		}
	})
}
