package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ecommerce/internal/model"
	"ecommerce/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock ChefService
type MockChefService struct {
	mock.Mock
}

func (m *MockChefService) CreateProfile(userID uint, profile *model.ChefProfile) (*model.Chef, error) {
	args := m.Called(userID, profile)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Chef), args.Error(1)
}

func (m *MockChefService) GetProfile(userID uint) (*model.Chef, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Chef), args.Error(1)
}

func (m *MockChefService) UpdateProfile(userID uint, profile *model.ChefProfile) (*model.Chef, error) {
	args := m.Called(userID, profile)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Chef), args.Error(1)
}

func (m *MockChefService) GetAllChefs(page, limit int) ([]model.Chef, int, error) {
	args := m.Called(page, limit)
	return args.Get(0).([]model.Chef), args.Int(1), args.Error(2)
}

func (m *MockChefService) SearchChefs(query string, page, limit int) ([]model.Chef, int, error) {
	args := m.Called(query, page, limit)
	return args.Get(0).([]model.Chef), args.Int(1), args.Error(2)
}

func setupChefHandler() (*ChefHandler, *MockChefService) {
	mockService := new(MockChefService)
	handler := NewChefHandler(mockService)
	return handler, mockService
}

func TestChefHandler_CreateProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, mockService := setupChefHandler()

	chefProfile := &model.ChefProfile{
		KitchenName: "Test Kitchen",
		Description: "Best chef in town",
		Speciality:  "Turkish Cuisine",
		Experience:  5,
	}

	createdChef := &model.Chef{
		ID:          1,
		UserID:      1,
		KitchenName: "Test Kitchen",
		Description: "Best chef in town",
		Speciality:  "Turkish Cuisine",
		Experience:  5,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	tests := []struct {
		name           string
		userID         uint
		requestBody    interface{}
		setupMocks     func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name:        "Create chef profile successfully",
			userID:      1,
			requestBody: chefProfile,
			setupMocks: func() {
				mockService.On("CreateProfile", uint(1), mock.AnythingOfType("*model.ChefProfile")).Return(createdChef, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedError:  false,
		},
		{
			name:        "Invalid request body",
			userID:      1,
			requestBody: "invalid json",
			setupMocks:  func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:        "Service error",
			userID:      1,
			requestBody: chefProfile,
			setupMocks: func() {
				mockService.On("CreateProfile", uint(1), mock.AnythingOfType("*model.ChefProfile")).Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockService.ExpectedCalls = nil
			mockService.Calls = nil

			tt.setupMocks()

			// Prepare request
			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/chef/profile", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			
			// Setup router
			router := gin.New()
			router.POST("/api/chef/profile", func(c *gin.Context) {
				c.Set("userID", tt.userID)
				handler.CreateProfile(c)
			})

			// Execute request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectedError {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response["success"].(bool))
				assert.NotNil(t, response["data"])
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestChefHandler_GetProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, mockService := setupChefHandler()

	chef := &model.Chef{
		ID:          1,
		UserID:      1,
		KitchenName: "Test Kitchen",
		Description: "Best chef in town",
		IsActive:    true,
	}

	tests := []struct {
		name           string
		userID         uint
		setupMocks     func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name:   "Get chef profile successfully",
			userID: 1,
			setupMocks: func() {
				mockService.On("GetProfile", uint(1)).Return(chef, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:   "Chef profile not found",
			userID: 999,
			setupMocks: func() {
				mockService.On("GetProfile", uint(999)).Return(nil, errors.New("chef profile not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockService.ExpectedCalls = nil
			mockService.Calls = nil

			tt.setupMocks()

			// Prepare request
			req := httptest.NewRequest(http.MethodGet, "/api/chef/profile", nil)
			
			// Setup router
			router := gin.New()
			router.GET("/api/chef/profile", func(c *gin.Context) {
				c.Set("userID", tt.userID)
				handler.GetProfile(c)
			})

			// Execute request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectedError {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response["success"].(bool))
				assert.NotNil(t, response["data"])
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestChefHandler_UpdateProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, mockService := setupChefHandler()

	updateProfile := &model.ChefProfile{
		KitchenName: "Updated Kitchen",
		Description: "Updated description",
		Speciality:  "Italian Cuisine",
		Experience:  10,
	}

	updatedChef := &model.Chef{
		ID:          1,
		UserID:      1,
		KitchenName: "Updated Kitchen",
		Description: "Updated description",
		Speciality:  "Italian Cuisine",
		Experience:  10,
		IsActive:    true,
	}

	tests := []struct {
		name           string
		userID         uint
		requestBody    interface{}
		setupMocks     func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name:        "Update chef profile successfully",
			userID:      1,
			requestBody: updateProfile,
			setupMocks: func() {
				mockService.On("UpdateProfile", uint(1), mock.AnythingOfType("*model.ChefProfile")).Return(updatedChef, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:        "Invalid request body",
			userID:      1,
			requestBody: "invalid json",
			setupMocks:  func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:        "Service error",
			userID:      1,
			requestBody: updateProfile,
			setupMocks: func() {
				mockService.On("UpdateProfile", uint(1), mock.AnythingOfType("*model.ChefProfile")).Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockService.ExpectedCalls = nil
			mockService.Calls = nil

			tt.setupMocks()

			// Prepare request
			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPUT, "/api/chef/profile", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			
			// Setup router
			router := gin.New()
			router.PUT("/api/chef/profile", func(c *gin.Context) {
				c.Set("userID", tt.userID)
				handler.UpdateProfile(c)
			})

			// Execute request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectedError {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response["success"].(bool))
				assert.NotNil(t, response["data"])
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestChefHandler_GetAllChefs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, mockService := setupChefHandler()

	chefs := []model.Chef{
		{ID: 1, KitchenName: "Kitchen 1", IsActive: true},
		{ID: 2, KitchenName: "Kitchen 2", IsActive: true},
	}

	tests := []struct {
		name           string
		queryParams    string
		setupMocks     func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name:        "Get all chefs successfully",
			queryParams: "?page=1&limit=10",
			setupMocks: func() {
				mockService.On("GetAllChefs", 1, 10).Return(chefs, 2, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:        "Default pagination parameters",
			queryParams: "",
			setupMocks: func() {
				mockService.On("GetAllChefs", 1, 10).Return(chefs, 2, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:        "Service error",
			queryParams: "?page=1&limit=10",
			setupMocks: func() {
				mockService.On("GetAllChefs", 1, 10).Return([]model.Chef{}, 0, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockService.ExpectedCalls = nil
			mockService.Calls = nil

			tt.setupMocks()

			// Prepare request
			req := httptest.NewRequest(http.MethodGet, "/api/chefs"+tt.queryParams, nil)
			
			// Setup router
			router := gin.New()
			router.GET("/api/chefs", handler.GetAllChefs)

			// Execute request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectedError {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response["success"].(bool))
				
				data := response["data"].(map[string]interface{})
				assert.NotNil(t, data["chefs"])
				assert.Equal(t, float64(2), data["total"])
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestChefHandler_SearchChefs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, mockService := setupChefHandler()

	chefs := []model.Chef{
		{ID: 1, KitchenName: "Turkish Kitchen", Speciality: "Turkish Cuisine", IsActive: true},
	}

	tests := []struct {
		name           string
		queryParams    string
		setupMocks     func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name:        "Search chefs successfully",
			queryParams: "?q=turkish&page=1&limit=10",
			setupMocks: func() {
				mockService.On("SearchChefs", "turkish", 1, 10).Return(chefs, 1, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:        "Search with no query parameter",
			queryParams: "?page=1&limit=10",
			setupMocks: func() {
				mockService.On("SearchChefs", "", 1, 10).Return([]model.Chef{}, 0, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:        "Service error",
			queryParams: "?q=turkish&page=1&limit=10",
			setupMocks: func() {
				mockService.On("SearchChefs", "turkish", 1, 10).Return([]model.Chef{}, 0, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockService.ExpectedCalls = nil
			mockService.Calls = nil

			tt.setupMocks()

			// Prepare request
			req := httptest.NewRequest(http.MethodGet, "/api/chefs/search"+tt.queryParams, nil)
			
			// Setup router
			router := gin.New()
			router.GET("/api/chefs/search", handler.SearchChefs)

			// Execute request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectedError {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response["success"].(bool))
				
				data := response["data"].(map[string]interface{})
				assert.NotNil(t, data["chefs"])
			}

			mockService.AssertExpectations(t)
		})
	}
}
