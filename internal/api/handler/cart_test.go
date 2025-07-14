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

// MockCartService - Cart service mock
type MockCartService struct {
	mock.Mock
}

func (m *MockCartService) GetCart(userID uint) (*model.Cart, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Cart), args.Error(1)
}

func (m *MockCartService) AddToCart(userID uint, req *model.AddToCartRequest) (*model.CartItem, error) {
	args := m.Called(userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.CartItem), args.Error(1)
}

func (m *MockCartService) UpdateCartItem(userID uint, itemID uint, quantity int) (*model.CartItem, error) {
	args := m.Called(userID, itemID, quantity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.CartItem), args.Error(1)
}

func (m *MockCartService) RemoveFromCart(userID uint, itemID uint) error {
	args := m.Called(userID, itemID)
	return args.Error(0)
}

func (m *MockCartService) ClearCart(userID uint) error {
	args := m.Called(userID)
	return args.Error(0)
}

func setupCartHandler() (*CartHandler, *MockCartService) {
	mockService := new(MockCartService)
	handler := NewCartHandler(mockService)
	return handler, mockService
}

// TestCartHandler_GetCart - Test getting user cart
func TestCartHandler_GetCart(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, mockService := setupCartHandler()

	cart := &model.Cart{
		ID:     1,
		UserID: 1,
		Items: []model.CartItem{
			{
				ID:       1,
				CartID:   1,
				MealID:   1,
				Quantity: 2,
				Price:    25.99,
				Subtotal: 51.98,
			},
		},
		Total: 51.98,
	}

	tests := []struct {
		name           string
		userID         uint
		setupMocks     func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name:   "Get cart successfully",
			userID: 1,
			setupMocks: func() {
				mockService.On("GetCart", uint(1)).Return(cart, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:   "Empty cart",
			userID: 1,
			setupMocks: func() {
				emptyCart := &model.Cart{
					ID:     1,
					UserID: 1,
					Items:  []model.CartItem{},
					Total:  0,
				}
				mockService.On("GetCart", uint(1)).Return(emptyCart, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:   "Service error",
			userID: 1,
			setupMocks: func() {
				mockService.On("GetCart", uint(1)).Return(nil, assert.AnError)
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
			req := httptest.NewRequest(http.MethodGet, "/api/cart", nil)

			// Setup router
			router := gin.New()
			router.GET("/api/cart", func(c *gin.Context) {
				c.Set("userID", tt.userID)
				handler.GetCart(c)
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

// TestCartHandler_AddToCart - Test adding items to cart
func TestCartHandler_AddToCart(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, mockService := setupCartHandler()

	addRequest := model.AddToCartRequest{
		MealID:   1,
		Quantity: 2,
	}

	cartItem := &model.CartItem{
		ID:       1,
		CartID:   1,
		MealID:   1,
		Quantity: 2,
		Price:    25.99,
		Subtotal: 51.98,
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
			name:        "Add to cart successfully",
			userID:      1,
			requestBody: addRequest,
			setupMocks: func() {
				mockService.On("AddToCart", uint(1), mock.AnythingOfType("*model.AddToCartRequest")).Return(cartItem, nil)
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
			name: "Missing meal ID",
			userID: 1,
			requestBody: model.AddToCartRequest{
				Quantity: 2,
			},
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "Invalid quantity",
			userID: 1,
			requestBody: model.AddToCartRequest{
				MealID:   1,
				Quantity: 0,
			},
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:        "Service error",
			userID:      1,
			requestBody: addRequest,
			setupMocks: func() {
				mockService.On("AddToCart", uint(1), mock.AnythingOfType("*model.AddToCartRequest")).Return(nil, assert.AnError)
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

			req := httptest.NewRequest(http.MethodPost, "/api/cart/items", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			// Setup router
			router := gin.New()
			router.POST("/api/cart/items", func(c *gin.Context) {
				c.Set("userID", tt.userID)
				handler.AddToCart(c)
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

// TestCartHandler_UpdateCartItem - Test updating cart item quantity
func TestCartHandler_UpdateCartItem(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, mockService := setupCartHandler()

	updateRequest := struct {
		Quantity int `json:"quantity"`
	}{
		Quantity: 3,
	}

	updatedItem := &model.CartItem{
		ID:       1,
		CartID:   1,
		MealID:   1,
		Quantity: 3,
		Price:    25.99,
		Subtotal: 77.97,
	}

	tests := []struct {
		name           string
		userID         uint
		itemID         string
		requestBody    interface{}
		setupMocks     func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name:        "Update cart item successfully",
			userID:      1,
			itemID:      "1",
			requestBody: updateRequest,
			setupMocks: func() {
				mockService.On("UpdateCartItem", uint(1), uint(1), 3).Return(updatedItem, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "Invalid item ID",
			userID:         1,
			itemID:         "invalid",
			requestBody:    updateRequest,
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:        "Invalid request body",
			userID:      1,
			itemID:      "1",
			requestBody: "invalid json",
			setupMocks:  func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:   "Invalid quantity",
			userID: 1,
			itemID: "1",
			requestBody: struct {
				Quantity int `json:"quantity"`
			}{
				Quantity: 0,
			},
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:        "Service error",
			userID:      1,
			itemID:      "1",
			requestBody: updateRequest,
			setupMocks: func() {
				mockService.On("UpdateCartItem", uint(1), uint(1), 3).Return(nil, assert.AnError)
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

			req := httptest.NewRequest(http.MethodPut, "/api/cart/items/"+tt.itemID, bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			// Setup router
			router := gin.New()
			router.PUT("/api/cart/items/:id", func(c *gin.Context) {
				c.Set("userID", tt.userID)
				handler.UpdateCartItem(c)
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

// TestCartHandler_RemoveFromCart - Test removing items from cart
func TestCartHandler_RemoveFromCart(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, mockService := setupCartHandler()

	tests := []struct {
		name           string
		userID         uint
		itemID         string
		setupMocks     func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name:   "Remove from cart successfully",
			userID: 1,
			itemID: "1",
			setupMocks: func() {
				mockService.On("RemoveFromCart", uint(1), uint(1)).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "Invalid item ID",
			userID:         1,
			itemID:         "invalid",
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:   "Service error",
			userID: 1,
			itemID: "1",
			setupMocks: func() {
				mockService.On("RemoveFromCart", uint(1), uint(1)).Return(assert.AnError)
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
			req := httptest.NewRequest(http.MethodDelete, "/api/cart/items/"+tt.itemID, nil)

			// Setup router
			router := gin.New()
			router.DELETE("/api/cart/items/:id", func(c *gin.Context) {
				c.Set("userID", tt.userID)
				handler.RemoveFromCart(c)
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

// TestCartHandler_ClearCart - Test clearing the entire cart
func TestCartHandler_ClearCart(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, mockService := setupCartHandler()

	tests := []struct {
		name           string
		userID         uint
		setupMocks     func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name:   "Clear cart successfully",
			userID: 1,
			setupMocks: func() {
				mockService.On("ClearCart", uint(1)).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:   "Service error",
			userID: 1,
			setupMocks: func() {
				mockService.On("ClearCart", uint(1)).Return(assert.AnError)
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
			req := httptest.NewRequest(http.MethodDelete, "/api/cart", nil)

			// Setup router
			router := gin.New()
			router.DELETE("/api/cart", func(c *gin.Context) {
				c.Set("userID", tt.userID)
				handler.ClearCart(c)
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

// TestCartHandler_CartValidation - Test cart validation scenarios
func TestCartHandler_CartValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, mockService := setupCartHandler()

	// Test negative quantity validation
	invalidQuantityRequest := model.AddToCartRequest{
		MealID:   1,
		Quantity: -1,
	}

	reqBody, _ := json.Marshal(invalidQuantityRequest)
	req := httptest.NewRequest(http.MethodPost, "/api/cart/items", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	router := gin.New()
	router.POST("/api/cart/items", func(c *gin.Context) {
		c.Set("userID", uint(1))
		handler.AddToCart(c)
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test zero meal ID validation
	zeroMealIDRequest := model.AddToCartRequest{
		MealID:   0,
		Quantity: 1,
	}

	reqBody, _ = json.Marshal(zeroMealIDRequest)
	req = httptest.NewRequest(http.MethodPost, "/api/cart/items", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestCartHandler_ErrorHandling - Test error handling scenarios
func TestCartHandler_ErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, mockService := setupCartHandler()

	// Test empty request body
	router := gin.New()
	router.POST("/api/cart/items", func(c *gin.Context) {
		c.Set("userID", uint(1))
		handler.AddToCart(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/cart/items", bytes.NewBuffer([]byte("")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test malformed JSON
	req = httptest.NewRequest(http.MethodPost, "/api/cart/items", bytes.NewBuffer([]byte("{")))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// BenchmarkCartHandler_GetCart - Benchmark getting cart
func BenchmarkCartHandler_GetCart(b *testing.B) {
	gin.SetMode(gin.TestMode)
	handler, mockService := setupCartHandler()

	cart := &model.Cart{
		ID:     1,
		UserID: 1,
		Items: []model.CartItem{
			{
				ID:       1,
				CartID:   1,
				MealID:   1,
				Quantity: 2,
				Price:    25.99,
				Subtotal: 51.98,
			},
		},
		Total: 51.98,
	}

	mockService.On("GetCart", uint(1)).Return(cart, nil)

	router := gin.New()
	router.GET("/api/cart", func(c *gin.Context) {
		c.Set("userID", uint(1))
		handler.GetCart(c)
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/cart", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkCartHandler_AddToCart - Benchmark adding to cart
func BenchmarkCartHandler_AddToCart(b *testing.B) {
	gin.SetMode(gin.TestMode)
	handler, mockService := setupCartHandler()

	cartItem := &model.CartItem{
		ID:       1,
		CartID:   1,
		MealID:   1,
		Quantity: 2,
		Price:    25.99,
		Subtotal: 51.98,
	}

	mockService.On("AddToCart", uint(1), mock.AnythingOfType("*model.AddToCartRequest")).Return(cartItem, nil)

	router := gin.New()
	router.POST("/api/cart/items", func(c *gin.Context) {
		c.Set("userID", uint(1))
		handler.AddToCart(c)
	})

	addRequest := model.AddToCartRequest{
		MealID:   1,
		Quantity: 2,
	}
	reqBody, _ := json.Marshal(addRequest)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/cart/items", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
