package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"ecommerce/internal/model"
	"ecommerce/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMealService - Meal service mock
type MockMealService struct {
	mock.Mock
}

func (m *MockMealService) CreateMeal(chefID uint, req *model.MealRequest) (*model.Meal, error) {
	args := m.Called(chefID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Meal), args.Error(1)
}

func (m *MockMealService) GetMealByID(mealID uint) (*model.Meal, error) {
	args := m.Called(mealID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Meal), args.Error(1)
}

func (m *MockMealService) GetMealsByChef(chefID uint, page, limit int) ([]model.Meal, int, error) {
	args := m.Called(chefID, page, limit)
	return args.Get(0).([]model.Meal), args.Int(1), args.Error(2)
}

func (m *MockMealService) UpdateMeal(mealID uint, req *model.MealRequest) (*model.Meal, error) {
	args := m.Called(mealID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Meal), args.Error(1)
}

func (m *MockMealService) DeleteMeal(mealID uint, chefID uint) error {
	args := m.Called(mealID, chefID)
	return args.Error(0)
}

func (m *MockMealService) GetAllMeals(page, limit int) ([]model.Meal, int, error) {
	args := m.Called(page, limit)
	return args.Get(0).([]model.Meal), args.Int(1), args.Error(2)
}

func (m *MockMealService) SearchMeals(query string, page, limit int) ([]model.Meal, int, error) {
	args := m.Called(query, page, limit)
	return args.Get(0).([]model.Meal), args.Int(1), args.Error(2)
}

func (m *MockMealService) ToggleAvailability(mealID uint, chefID uint) error {
	args := m.Called(mealID, chefID)
	return args.Error(0)
}

func setupMealHandler() (*MealHandler, *MockMealService) {
	mockService := new(MockMealService)
	handler := NewMealHandler(mockService)
	return handler, mockService
}

// TestMealHandler_CreateMeal - Test meal creation
func TestMealHandler_CreateMeal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, mockService := setupMealHandler()

	mealRequest := model.MealRequest{
		Name:        "Turkish Kebab",
		Description: "Delicious traditional kebab",
		Price:       25.99,
		CookingTime: 30,
		Category:    "Main Course",
		IsVegetarian: false,
		IsVegan:     false,
		IsGlutenFree: false,
		Ingredients: []string{"Beef", "Onions", "Spices"},
		Allergens:   []string{"None"},
	}

	createdMeal := &model.Meal{
		ID:          1,
		ChefID:      1,
		Name:        mealRequest.Name,
		Description: mealRequest.Description,
		Price:       mealRequest.Price,
		CookingTime: mealRequest.CookingTime,
		Category:    mealRequest.Category,
		IsVegetarian: mealRequest.IsVegetarian,
		IsVegan:     mealRequest.IsVegan,
		IsGlutenFree: mealRequest.IsGlutenFree,
		Ingredients: mealRequest.Ingredients,
		Allergens:   mealRequest.Allergens,
		IsActive:    true,
	}

	tests := []struct {
		name           string
		chefID         uint
		requestBody    interface{}
		setupMocks     func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name:   "Create meal successfully",
			chefID: 1,
			requestBody: mealRequest,
			setupMocks: func() {
				mockService.On("CreateMeal", uint(1), mock.AnythingOfType("*model.MealRequest")).Return(createdMeal, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedError:  false,
		},
		{
			name:   "Invalid request body",
			chefID: 1,
			requestBody: "invalid json",
			setupMocks: func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:   "Missing required fields",
			chefID: 1,
			requestBody: model.MealRequest{
				Name: "Incomplete Meal",
			},
			setupMocks: func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:   "Service error",
			chefID: 1,
			requestBody: mealRequest,
			setupMocks: func() {
				mockService.On("CreateMeal", uint(1), mock.AnythingOfType("*model.MealRequest")).Return(nil, assert.AnError)
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
			var reqBody []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				reqBody = []byte(str)
			} else {
				reqBody, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/meals", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			// Setup router
			router := gin.New()
			router.POST("/api/meals", func(c *gin.Context) {
				c.Set("chefID", tt.chefID)
				handler.CreateMeal(c)
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

// TestMealHandler_GetMeal - Test get meal by ID
func TestMealHandler_GetMeal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, mockService := setupMealHandler()

	meal := &model.Meal{
		ID:          1,
		ChefID:      1,
		Name:        "Turkish Kebab",
		Description: "Delicious traditional kebab",
		Price:       25.99,
		CookingTime: 30,
		Category:    "Main Course",
		IsActive:    true,
	}

	tests := []struct {
		name           string
		mealID         string
		setupMocks     func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name:   "Get meal successfully",
			mealID: "1",
			setupMocks: func() {
				mockService.On("GetMealByID", uint(1)).Return(meal, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:   "Meal not found",
			mealID: "999",
			setupMocks: func() {
				mockService.On("GetMealByID", uint(999)).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  true,
		},
		{
			name:           "Invalid meal ID",
			mealID:         "invalid",
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
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
			req := httptest.NewRequest(http.MethodGet, "/api/meals/"+tt.mealID, nil)

			// Setup router
			router := gin.New()
			router.GET("/api/meals/:id", handler.GetMeal)

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

// TestMealHandler_GetAllMeals - Test get all meals with pagination
func TestMealHandler_GetAllMeals(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, mockService := setupMealHandler()

	meals := []model.Meal{
		{ID: 1, Name: "Meal 1", Price: 15.99, IsActive: true},
		{ID: 2, Name: "Meal 2", Price: 20.99, IsActive: true},
	}

	tests := []struct {
		name           string
		queryParams    string
		setupMocks     func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name:        "Get meals with pagination",
			queryParams: "?page=1&limit=10",
			setupMocks: func() {
				mockService.On("GetAllMeals", 1, 10).Return(meals, 2, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:        "Default pagination",
			queryParams: "",
			setupMocks: func() {
				mockService.On("GetAllMeals", 1, 10).Return(meals, 2, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:        "Service error",
			queryParams: "?page=1&limit=10",
			setupMocks: func() {
				mockService.On("GetAllMeals", 1, 10).Return([]model.Meal{}, 0, assert.AnError)
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
			req := httptest.NewRequest(http.MethodGet, "/api/meals"+tt.queryParams, nil)

			// Setup router
			router := gin.New()
			router.GET("/api/meals", handler.GetAllMeals)

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
				assert.NotNil(t, data["meals"])
				assert.Equal(t, float64(2), data["total"])
			}

			mockService.AssertExpectations(t)
		})
	}
}

// TestMealHandler_UpdateMeal - Test meal update
func TestMealHandler_UpdateMeal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, mockService := setupMealHandler()

	updateRequest := model.MealRequest{
		Name:        "Updated Turkish Kebab",
		Description: "Even more delicious kebab",
		Price:       30.99,
		CookingTime: 35,
		Category:    "Main Course",
	}

	updatedMeal := &model.Meal{
		ID:          1,
		ChefID:      1,
		Name:        updateRequest.Name,
		Description: updateRequest.Description,
		Price:       updateRequest.Price,
		CookingTime: updateRequest.CookingTime,
		Category:    updateRequest.Category,
		IsActive:    true,
	}

	tests := []struct {
		name           string
		mealID         string
		chefID         uint
		requestBody    interface{}
		setupMocks     func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name:        "Update meal successfully",
			mealID:      "1",
			chefID:      1,
			requestBody: updateRequest,
			setupMocks: func() {
				mockService.On("UpdateMeal", uint(1), mock.AnythingOfType("*model.MealRequest")).Return(updatedMeal, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "Invalid meal ID",
			mealID:         "invalid",
			chefID:         1,
			requestBody:    updateRequest,
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:        "Invalid request body",
			mealID:      "1",
			chefID:      1,
			requestBody: "invalid json",
			setupMocks:  func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:        "Service error",
			mealID:      "1",
			chefID:      1,
			requestBody: updateRequest,
			setupMocks: func() {
				mockService.On("UpdateMeal", uint(1), mock.AnythingOfType("*model.MealRequest")).Return(nil, assert.AnError)
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
			var reqBody []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				reqBody = []byte(str)
			} else {
				reqBody, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPut, "/api/meals/"+tt.mealID, bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			// Setup router
			router := gin.New()
			router.PUT("/api/meals/:id", func(c *gin.Context) {
				c.Set("chefID", tt.chefID)
				handler.UpdateMeal(c)
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

// TestMealHandler_DeleteMeal - Test meal deletion
func TestMealHandler_DeleteMeal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, mockService := setupMealHandler()

	tests := []struct {
		name           string
		mealID         string
		chefID         uint
		setupMocks     func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name:   "Delete meal successfully",
			mealID: "1",
			chefID: 1,
			setupMocks: func() {
				mockService.On("DeleteMeal", uint(1), uint(1)).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "Invalid meal ID",
			mealID:         "invalid",
			chefID:         1,
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:   "Service error",
			mealID: "1",
			chefID: 1,
			setupMocks: func() {
				mockService.On("DeleteMeal", uint(1), uint(1)).Return(assert.AnError)
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
			req := httptest.NewRequest(http.MethodDelete, "/api/meals/"+tt.mealID, nil)

			// Setup router
			router := gin.New()
			router.DELETE("/api/meals/:id", func(c *gin.Context) {
				c.Set("chefID", tt.chefID)
				handler.DeleteMeal(c)
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
			}

			mockService.AssertExpectations(t)
		})
	}
}

// TestMealHandler_SearchMeals - Test meal search
func TestMealHandler_SearchMeals(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, mockService := setupMealHandler()

	meals := []model.Meal{
		{ID: 1, Name: "Turkish Kebab", Category: "Main Course", IsActive: true},
	}

	tests := []struct {
		name           string
		queryParams    string
		setupMocks     func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name:        "Search meals successfully",
			queryParams: "?q=kebab&page=1&limit=10",
			setupMocks: func() {
				mockService.On("SearchMeals", "kebab", 1, 10).Return(meals, 1, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:        "Search with no query",
			queryParams: "?page=1&limit=10",
			setupMocks: func() {
				mockService.On("SearchMeals", "", 1, 10).Return([]model.Meal{}, 0, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:        "Service error",
			queryParams: "?q=kebab&page=1&limit=10",
			setupMocks: func() {
				mockService.On("SearchMeals", "kebab", 1, 10).Return([]model.Meal{}, 0, assert.AnError)
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
			req := httptest.NewRequest(http.MethodGet, "/api/meals/search"+tt.queryParams, nil)

			// Setup router
			router := gin.New()
			router.GET("/api/meals/search", handler.SearchMeals)

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
				assert.NotNil(t, data["meals"])
			}

			mockService.AssertExpectations(t)
		})
	}
}

// TestMealHandler_ToggleAvailability - Test toggle meal availability
func TestMealHandler_ToggleAvailability(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, mockService := setupMealHandler()

	tests := []struct {
		name           string
		mealID         string
		chefID         uint
		setupMocks     func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name:   "Toggle availability successfully",
			mealID: "1",
			chefID: 1,
			setupMocks: func() {
				mockService.On("ToggleAvailability", uint(1), uint(1)).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "Invalid meal ID",
			mealID:         "invalid",
			chefID:         1,
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:   "Service error",
			mealID: "1",
			chefID: 1,
			setupMocks: func() {
				mockService.On("ToggleAvailability", uint(1), uint(1)).Return(assert.AnError)
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
			req := httptest.NewRequest(http.MethodPut, "/api/meals/"+tt.mealID+"/toggle", nil)

			// Setup router
			router := gin.New()
			router.PUT("/api/meals/:id/toggle", func(c *gin.Context) {
				c.Set("chefID", tt.chefID)
				handler.ToggleAvailability(c)
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
			}

			mockService.AssertExpectations(t)
		})
	}
}

// BenchmarkMealHandler_GetAllMeals - Benchmark get all meals
func BenchmarkMealHandler_GetAllMeals(b *testing.B) {
	gin.SetMode(gin.TestMode)
	handler, mockService := setupMealHandler()

	meals := []model.Meal{
		{ID: 1, Name: "Meal 1", Price: 15.99, IsActive: true},
		{ID: 2, Name: "Meal 2", Price: 20.99, IsActive: true},
	}

	mockService.On("GetAllMeals", 1, 10).Return(meals, 2, nil)

	router := gin.New()
	router.GET("/api/meals", handler.GetAllMeals)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/meals?page=1&limit=10", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
