package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"ecommerce/internal/model"
	"ecommerce/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockOrderService - Order service mock
type MockOrderService struct {
	mock.Mock
}

func (m *MockOrderService) CreateOrder(userID uint, req *model.CreateOrderRequest) (*model.Order, error) {
	args := m.Called(userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

func (m *MockOrderService) GetOrderByID(orderID uint) (*model.Order, error) {
	args := m.Called(orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

func (m *MockOrderService) GetOrdersByUser(userID uint) ([]model.Order, error) {
	args := m.Called(userID)
	return args.Get(0).([]model.Order), args.Error(1)
}

func (m *MockOrderService) UpdateOrderStatus(orderID uint, status string) error {
	args := m.Called(orderID, status)
	return args.Error(0)
}

func (m *MockOrderService) UpdateSubOrderStatus(subOrderID uint, status string) error {
	args := m.Called(subOrderID, status)
	return args.Error(0)
}

func (m *MockOrderService) CancelOrder(orderID uint, userID uint) error {
	args := m.Called(orderID, userID)
	return args.Error(0)
}

// SetupOrderTestRouter - Test router setup
func SetupOrderTestRouter(orderService service.OrderService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	orderHandler := NewOrderHandler(orderService)

	api := router.Group("/api/v1")
	{
		orders := api.Group("/orders")
		{
			orders.POST("", orderHandler.CreateOrder)
			orders.GET("/:id", orderHandler.GetOrder)
			orders.GET("/user", orderHandler.GetUserOrders)
			orders.PUT("/:id/status", orderHandler.UpdateOrderStatus)
			orders.DELETE("/:id", orderHandler.CancelOrder)
			orders.PUT("/sub/:id/status", orderHandler.UpdateSubOrderStatus)
		}
	}

	return router
}

// TestOrderHandler_CreateOrder - Sipariş oluşturma handler testi
func TestOrderHandler_CreateOrder(t *testing.T) {
	mockService := new(MockOrderService)
	router := SetupOrderTestRouter(mockService)

	tests := []struct {
		name           string
		userID         uint
		requestBody    model.CreateOrderRequest
		mockReturn     *model.Order
		mockError      error
		expectedStatus int
	}{
		{
			name:   "Successful single chef order creation",
			userID: 1,
			requestBody: model.CreateOrderRequest{
				DeliveryType:    "delivery",
				DeliveryAddress: "123 Test St",
				PaymentMethod:   "credit_card",
				Items: []model.OrderItemInput{
					{MealID: 1, Quantity: 2},
				},
			},
			mockReturn: &model.Order{
				ID:          1,
				UserID:      1,
				OrderNumber: "ORD-20250714-001",
				Total:       51.00,
				Status:      "pending",
				ChefCount:   1,
				CreatedAt:   time.Now(),
			},
			mockError:      nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "Successful multi-vendor order creation",
			userID: 1,
			requestBody: model.CreateOrderRequest{
				DeliveryType:    "delivery",
				DeliveryAddress: "456 Test Ave",
				PaymentMethod:   "cash",
				Items: []model.OrderItemInput{
					{MealID: 1, Quantity: 1}, // Chef 1
					{MealID: 3, Quantity: 2}, // Chef 2
				},
			},
			mockReturn: &model.Order{
				ID:          2,
				UserID:      1,
				OrderNumber: "ORD-20250714-002",
				Total:       85.50,
				Status:      "pending",
				ChefCount:   2,
				CreatedAt:   time.Now(),
				SubOrders: []model.SubOrder{
					{
						ID:     1,
						ChefID: 1,
						Total:  25.50,
						Status: "pending",
					},
					{
						ID:     2,
						ChefID: 2,
						Total:  60.00,
						Status: "pending",
					},
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "Invalid request - missing delivery address",
			userID: 1,
			requestBody: model.CreateOrderRequest{
				DeliveryType:  "delivery",
				PaymentMethod: "credit_card",
				Items: []model.OrderItemInput{
					{MealID: 1, Quantity: 1},
				},
			},
			mockReturn:     nil,
			mockError:      assert.AnError,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Service error",
			userID: 1,
			requestBody: model.CreateOrderRequest{
				DeliveryType:    "delivery",
				DeliveryAddress: "123 Test St",
				PaymentMethod:   "credit_card",
				Items: []model.OrderItemInput{
					{MealID: 999, Quantity: 1}, // Non-existent meal
				},
			},
			mockReturn:     nil,
			mockError:      assert.AnError,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			mockService.On("CreateOrder", tt.userID, &tt.requestBody).Return(tt.mockReturn, tt.mockError).Once()

			// Prepare request
			reqBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/api/v1/orders", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("User-ID", strconv.Itoa(int(tt.userID))) // Mock auth middleware

			// Execute request
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, resp.Code)

			if tt.expectedStatus == http.StatusCreated {
				var response struct {
					Order model.Order `json:"order"`
				}
				err := json.Unmarshal(resp.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockReturn.OrderNumber, response.Order.OrderNumber)
				assert.Equal(t, tt.mockReturn.ChefCount, response.Order.ChefCount)
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

// TestOrderHandler_GetOrder - Sipariş detayı getirme testi
func TestOrderHandler_GetOrder(t *testing.T) {
	mockService := new(MockOrderService)
	router := SetupOrderTestRouter(mockService)

	tests := []struct {
		name           string
		orderID        string
		mockReturn     *model.Order
		mockError      error
		expectedStatus int
	}{
		{
			name:    "Successful order retrieval",
			orderID: "1",
			mockReturn: &model.Order{
				ID:          1,
				UserID:      1,
				OrderNumber: "ORD-20250714-001",
				Total:       100.50,
				Status:      "confirmed",
				ChefCount:   2,
				SubOrders: []model.SubOrder{
					{
						ID:              1,
						OrderID:         1,
						ChefID:          1,
						ChefOrderNumber: "ORD-20250714-001-CHEF1",
						Total:           50.25,
						Status:          "preparing",
					},
					{
						ID:              2,
						OrderID:         1,
						ChefID:          2,
						ChefOrderNumber: "ORD-20250714-001-CHEF2",
						Total:           50.25,
						Status:          "confirmed",
					},
				},
				Items: []model.OrderItem{
					{
						ID:         1,
						OrderID:    1,
						SubOrderID: 1,
						MealID:     1,
						ChefID:     1,
						Quantity:   2,
						Price:      25.00,
						Subtotal:   50.00,
					},
					{
						ID:         2,
						OrderID:    1,
						SubOrderID: 2,
						MealID:     3,
						ChefID:     2,
						Quantity:   1,
						Price:      30.00,
						Subtotal:   30.00,
					},
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Order not found",
			orderID:        "999",
			mockReturn:     nil,
			mockError:      assert.AnError,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Invalid order ID",
			orderID:        "invalid",
			mockReturn:     nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations (only for valid IDs)
			if orderID, err := strconv.Atoi(tt.orderID); err == nil {
				mockService.On("GetOrderByID", uint(orderID)).Return(tt.mockReturn, tt.mockError).Once()
			}

			// Prepare request
			req := httptest.NewRequest("GET", "/api/v1/orders/"+tt.orderID, nil)

			// Execute request
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, resp.Code)

			if tt.expectedStatus == http.StatusOK {
				var response struct {
					Order model.Order `json:"order"`
				}
				err := json.Unmarshal(resp.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockReturn.ID, response.Order.ID)
				assert.Equal(t, tt.mockReturn.OrderNumber, response.Order.OrderNumber)
				assert.Equal(t, len(tt.mockReturn.SubOrders), len(response.Order.SubOrders))
				assert.Equal(t, len(tt.mockReturn.Items), len(response.Order.Items))
			}

			// Verify mock expectations
			if orderID, err := strconv.Atoi(tt.orderID); err == nil {
				mockService.AssertExpectations(t)
			}
		})
	}
}

// TestOrderHandler_UpdateSubOrderStatus - Alt sipariş durumu güncelleme testi
func TestOrderHandler_UpdateSubOrderStatus(t *testing.T) {
	mockService := new(MockOrderService)
	router := SetupOrderTestRouter(mockService)

	tests := []struct {
		name           string
		subOrderID     string
		newStatus      string
		mockError      error
		expectedStatus int
	}{
		{
			name:           "Successful status update to confirmed",
			subOrderID:     "1",
			newStatus:      "confirmed",
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Successful status update to preparing",
			subOrderID:     "1",
			newStatus:      "preparing",
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Successful status update to ready",
			subOrderID:     "1",
			newStatus:      "ready",
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Successful status update to delivered",
			subOrderID:     "1",
			newStatus:      "delivered",
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid sub order ID",
			subOrderID:     "invalid",
			newStatus:      "confirmed",
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Service error - sub order not found",
			subOrderID:     "999",
			newStatus:      "confirmed",
			mockError:      assert.AnError,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Invalid status transition",
			subOrderID:     "1",
			newStatus:      "invalid_status",
			mockError:      assert.AnError,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations (only for valid IDs)
			if subOrderID, err := strconv.Atoi(tt.subOrderID); err == nil {
				mockService.On("UpdateSubOrderStatus", uint(subOrderID), tt.newStatus).Return(tt.mockError).Once()
			}

			// Prepare request body
			requestBody := struct {
				Status string `json:"status"`
			}{
				Status: tt.newStatus,
			}
			reqBody, _ := json.Marshal(requestBody)

			// Prepare request
			req := httptest.NewRequest("PUT", "/api/v1/orders/sub/"+tt.subOrderID+"/status", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			// Execute request
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, resp.Code)

			if tt.expectedStatus == http.StatusOK {
				var response struct {
					Message string `json:"message"`
				}
				err := json.Unmarshal(resp.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response.Message, "başarıyla güncellendi")
			}

			// Verify mock expectations
			if subOrderID, err := strconv.Atoi(tt.subOrderID); err == nil {
				mockService.AssertExpectations(t)
			}
		})
	}
}

// TestOrderHandler_GetUserOrders - Kullanıcı siparişleri getirme testi
func TestOrderHandler_GetUserOrders(t *testing.T) {
	mockService := new(MockOrderService)
	router := SetupOrderTestRouter(mockService)

	tests := []struct {
		name           string
		userID         uint
		mockReturn     []model.Order
		mockError      error
		expectedStatus int
		expectedCount  int
	}{
		{
			name:   "Successful retrieval with multiple orders",
			userID: 1,
			mockReturn: []model.Order{
				{
					ID:          1,
					UserID:      1,
					OrderNumber: "ORD-20250714-001",
					Total:       100.50,
					Status:      "delivered",
					ChefCount:   1,
					CreatedAt:   time.Now().AddDate(0, 0, -2),
				},
				{
					ID:          2,
					UserID:      1,
					OrderNumber: "ORD-20250714-002",
					Total:       75.25,
					Status:      "preparing",
					ChefCount:   2,
					CreatedAt:   time.Now().AddDate(0, 0, -1),
				},
				{
					ID:          3,
					UserID:      1,
					OrderNumber: "ORD-20250714-003",
					Total:       150.00,
					Status:      "pending",
					ChefCount:   3,
					CreatedAt:   time.Now(),
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
		{
			name:           "User with no orders",
			userID:         2,
			mockReturn:     []model.Order{},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name:           "Service error",
			userID:         1,
			mockReturn:     nil,
			mockError:      assert.AnError,
			expectedStatus: http.StatusInternalServerError,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			mockService.On("GetOrdersByUser", tt.userID).Return(tt.mockReturn, tt.mockError).Once()

			// Prepare request
			req := httptest.NewRequest("GET", "/api/v1/orders/user", nil)
			req.Header.Set("User-ID", strconv.Itoa(int(tt.userID))) // Mock auth middleware

			// Execute request
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, resp.Code)

			if tt.expectedStatus == http.StatusOK {
				var response struct {
					Orders []model.Order `json:"orders"`
				}
				err := json.Unmarshal(resp.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, len(response.Orders))

				// Verify orders are sorted by creation date (newest first)
				if len(response.Orders) > 1 {
					for i := 1; i < len(response.Orders); i++ {
						assert.True(t, response.Orders[i-1].CreatedAt.After(response.Orders[i].CreatedAt) || 
									response.Orders[i-1].CreatedAt.Equal(response.Orders[i].CreatedAt))
					}
				}
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

// TestOrderHandler_CancelOrder - Sipariş iptali testi
func TestOrderHandler_CancelOrder(t *testing.T) {
	mockService := new(MockOrderService)
	router := SetupOrderTestRouter(mockService)

	tests := []struct {
		name           string
		orderID        string
		userID         uint
		mockError      error
		expectedStatus int
	}{
		{
			name:           "Successful order cancellation",
			orderID:        "1",
			userID:         1,
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Order not found",
			orderID:        "999",
			userID:         1,
			mockError:      assert.AnError,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Invalid order ID",
			orderID:        "invalid",
			userID:         1,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Order cannot be cancelled (already delivered)",
			orderID:        "1",
			userID:         1,
			mockError:      assert.AnError,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations (only for valid IDs)
			if orderID, err := strconv.Atoi(tt.orderID); err == nil {
				mockService.On("CancelOrder", uint(orderID), tt.userID).Return(tt.mockError).Once()
			}

			// Prepare request
			req := httptest.NewRequest("DELETE", "/api/v1/orders/"+tt.orderID, nil)
			req.Header.Set("User-ID", strconv.Itoa(int(tt.userID))) // Mock auth middleware

			// Execute request
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, resp.Code)

			if tt.expectedStatus == http.StatusOK {
				var response struct {
					Message string `json:"message"`
				}
				err := json.Unmarshal(resp.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response.Message, "iptal edildi")
			}

			// Verify mock expectations
			if orderID, err := strconv.Atoi(tt.orderID); err == nil {
				mockService.AssertExpectations(t)
			}
		})
	}
}

// TestMultiVendorOrderResponse - Multi-vendor sipariş yanıt formatı testi
func TestMultiVendorOrderResponse(t *testing.T) {
	// Test order with multiple sub-orders
	order := &model.Order{
		ID:          1,
		UserID:      1,
		OrderNumber: "ORD-20250714-001",
		Total:       150.75,
		Status:      "confirmed",
		ChefCount:   3,
		SubOrders: []model.SubOrder{
			{
				ID:              1,
				OrderID:         1,
				ChefID:          1,
				ChefOrderNumber: "ORD-20250714-001-CHEF1",
				Subtotal:        50.25,
				Total:           55.25,
				Status:          "preparing",
				EstimatedTime:   30,
			},
			{
				ID:              2,
				OrderID:         1,
				ChefID:          2,
				ChefOrderNumber: "ORD-20250714-001-CHEF2",
				Subtotal:        45.00,
				Total:           50.00,
				Status:          "confirmed",
				EstimatedTime:   45,
			},
			{
				ID:              3,
				OrderID:         1,
				ChefID:          3,
				ChefOrderNumber: "ORD-20250714-001-CHEF3",
				Subtotal:        40.50,
				Total:           45.50,
				Status:          "ready",
				EstimatedTime:   15,
			},
		},
		Items: []model.OrderItem{
			{ID: 1, OrderID: 1, SubOrderID: 1, MealID: 1, ChefID: 1, Quantity: 2, Price: 25.00, Subtotal: 50.00},
			{ID: 2, OrderID: 1, SubOrderID: 2, MealID: 3, ChefID: 2, Quantity: 1, Price: 30.00, Subtotal: 30.00},
			{ID: 3, OrderID: 1, SubOrderID: 2, MealID: 4, ChefID: 2, Quantity: 1, Price: 15.00, Subtotal: 15.00},
			{ID: 4, OrderID: 1, SubOrderID: 3, MealID: 5, ChefID: 3, Quantity: 2, Price: 20.00, Subtotal: 40.00},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(order)
	assert.NoError(t, err)

	// Deserialize back
	var deserializedOrder model.Order
	err = json.Unmarshal(jsonData, &deserializedOrder)
	assert.NoError(t, err)

	// Verify structure
	assert.Equal(t, order.ID, deserializedOrder.ID)
	assert.Equal(t, order.ChefCount, deserializedOrder.ChefCount)
	assert.Equal(t, len(order.SubOrders), len(deserializedOrder.SubOrders))
	assert.Equal(t, len(order.Items), len(deserializedOrder.Items))

	// Verify sub-orders
	for i, subOrder := range deserializedOrder.SubOrders {
		assert.Equal(t, order.SubOrders[i].ChefID, subOrder.ChefID)
		assert.Equal(t, order.SubOrders[i].Status, subOrder.Status)
		assert.Equal(t, order.SubOrders[i].Total, subOrder.Total)
	}

	// Verify order items
	for i, item := range deserializedOrder.Items {
		assert.Equal(t, order.Items[i].SubOrderID, item.SubOrderID)
		assert.Equal(t, order.Items[i].ChefID, item.ChefID)
		assert.Equal(t, order.Items[i].Subtotal, item.Subtotal)
	}
}
