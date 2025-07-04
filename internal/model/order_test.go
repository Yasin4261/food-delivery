package model

import (
	"encoding/json"
	"testing"
	"time"
)

// TestOrder - Order model test
func TestOrder(t *testing.T) {
	now := time.Now()
	order := Order{
		ID:            1,
		UserID:        1,
		Total:         99.97,
		Status:        "pending",
		PaymentMethod: "credit_card",
		PaymentStatus: "pending",
		ShippingAddress: "123 Main St, City, State 12345",
		Notes:         "Please deliver after 6 PM",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	jsonData, err := json.Marshal(order)
	if err != nil {
		t.Errorf("JSON marshalling failed: %v", err)
	}

	var unmarshalledOrder Order
	if err := json.Unmarshal(jsonData, &unmarshalledOrder); err != nil {
		t.Errorf("JSON unmarshalling failed: %v", err)
	}

	if unmarshalledOrder.UserID != order.UserID {
		t.Errorf("Expected user ID %d, got %d", order.UserID, unmarshalledOrder.UserID)
	}
	if unmarshalledOrder.Total != order.Total {
		t.Errorf("Expected total %.2f, got %.2f", order.Total, unmarshalledOrder.Total)
	}
	if unmarshalledOrder.Status != order.Status {
		t.Errorf("Expected status %s, got %s", order.Status, unmarshalledOrder.Status)
	}
}

// TestOrderItem - Order item model test
func TestOrderItem(t *testing.T) {
	now := time.Now()
	orderItem := OrderItem{
		ID:        1,
		OrderID:   1,
		ProductID: 1,
		Quantity:  2,
		Price:     49.99,
		CreatedAt: now,
		UpdatedAt: now,
	}

	jsonData, err := json.Marshal(orderItem)
	if err != nil {
		t.Errorf("JSON marshalling failed: %v", err)
	}

	var unmarshalledItem OrderItem
	if err := json.Unmarshal(jsonData, &unmarshalledItem); err != nil {
		t.Errorf("JSON unmarshalling failed: %v", err)
	}

	if unmarshalledItem.OrderID != orderItem.OrderID {
		t.Errorf("Expected order ID %d, got %d", orderItem.OrderID, unmarshalledItem.OrderID)
	}
	if unmarshalledItem.ProductID != orderItem.ProductID {
		t.Errorf("Expected product ID %d, got %d", orderItem.ProductID, unmarshalledItem.ProductID)
	}
	if unmarshalledItem.Quantity != orderItem.Quantity {
		t.Errorf("Expected quantity %d, got %d", orderItem.Quantity, unmarshalledItem.Quantity)
	}
}

// TestCreateOrderRequest - Create order request test
func TestCreateOrderRequest(t *testing.T) {
	tests := []struct {
		name    string
		request CreateOrderRequest
		isValid bool
	}{
		{
			name: "Valid order request",
			request: CreateOrderRequest{
				PaymentMethod:   "credit_card",
				ShippingAddress: "123 Main St, City, State 12345",
				Notes:           "Handle with care",
			},
			isValid: true,
		},
		{
			name: "Empty payment method",
			request: CreateOrderRequest{
				PaymentMethod:   "",
				ShippingAddress: "123 Main St, City, State 12345",
				Notes:           "Handle with care",
			},
			isValid: false,
		},
		{
			name: "Empty shipping address",
			request: CreateOrderRequest{
				PaymentMethod:   "credit_card",
				ShippingAddress: "",
				Notes:           "Handle with care",
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.request)
			if err != nil {
				t.Errorf("JSON marshalling failed: %v", err)
			}

			var unmarshalledRequest CreateOrderRequest
			if err := json.Unmarshal(jsonData, &unmarshalledRequest); err != nil {
				t.Errorf("JSON unmarshalling failed: %v", err)
			}

			if unmarshalledRequest.PaymentMethod != tt.request.PaymentMethod {
				t.Errorf("Expected payment method %s, got %s", tt.request.PaymentMethod, unmarshalledRequest.PaymentMethod)
			}
			if unmarshalledRequest.ShippingAddress != tt.request.ShippingAddress {
				t.Errorf("Expected shipping address %s, got %s", tt.request.ShippingAddress, unmarshalledRequest.ShippingAddress)
			}
		})
	}
}

// TestUpdateOrderStatusRequest - Update order status request test
func TestUpdateOrderStatusRequest(t *testing.T) {
	validStatuses := []string{"pending", "confirmed", "preparing", "shipped", "delivered", "cancelled"}
	
	for _, status := range validStatuses {
		t.Run("Valid status: "+status, func(t *testing.T) {
			request := UpdateOrderStatusRequest{
				Status: status,
			}

			jsonData, err := json.Marshal(request)
			if err != nil {
				t.Errorf("JSON marshalling failed: %v", err)
			}

			var unmarshalledRequest UpdateOrderStatusRequest
			if err := json.Unmarshal(jsonData, &unmarshalledRequest); err != nil {
				t.Errorf("JSON unmarshalling failed: %v", err)
			}

			if unmarshalledRequest.Status != request.Status {
				t.Errorf("Expected status %s, got %s", request.Status, unmarshalledRequest.Status)
			}
		})
	}
}

// TestOrderItemWithProduct - Order item with product test
func TestOrderItemWithProduct(t *testing.T) {
	orderItem := OrderItemWithProduct{
		ID:           1,
		OrderID:      1,
		ProductID:    1,
		ProductName:  "Test Product",
		ProductImage: "https://example.com/image.jpg",
		Quantity:     2,
		Price:        49.99,
		Subtotal:     99.98,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	jsonData, err := json.Marshal(orderItem)
	if err != nil {
		t.Errorf("JSON marshalling failed: %v", err)
	}

	var unmarshalledItem OrderItemWithProduct
	if err := json.Unmarshal(jsonData, &unmarshalledItem); err != nil {
		t.Errorf("JSON unmarshalling failed: %v", err)
	}

	if unmarshalledItem.ProductName != orderItem.ProductName {
		t.Errorf("Expected product name %s, got %s", orderItem.ProductName, unmarshalledItem.ProductName)
	}
	if unmarshalledItem.Subtotal != orderItem.Subtotal {
		t.Errorf("Expected subtotal %.2f, got %.2f", orderItem.Subtotal, unmarshalledItem.Subtotal)
	}
}

// TestOrderWithItems - Order with items test
func TestOrderWithItems(t *testing.T) {
	items := []OrderItemWithProduct{
		{
			ID:           1,
			ProductName:  "Product 1",
			ProductImage: "https://example.com/image1.jpg",
			Quantity:     2,
			Price:        29.99,
			Subtotal:     59.98,
		},
		{
			ID:           2,
			ProductName:  "Product 2",
			ProductImage: "https://example.com/image2.jpg",
			Quantity:     1,
			Price:        39.99,
			Subtotal:     39.99,
		},
	}

	order := OrderWithItems{
		ID:              1,
		UserID:          1,
		Total:           99.97,
		Status:          "pending",
		PaymentMethod:   "credit_card",
		PaymentStatus:   "pending",
		ShippingAddress: "123 Main St, City, State 12345",
		Notes:           "Please deliver after 6 PM",
		Items:           items,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	jsonData, err := json.Marshal(order)
	if err != nil {
		t.Errorf("JSON marshalling failed: %v", err)
	}

	var unmarshalledOrder OrderWithItems
	if err := json.Unmarshal(jsonData, &unmarshalledOrder); err != nil {
		t.Errorf("JSON unmarshalling failed: %v", err)
	}

	if unmarshalledOrder.Total != order.Total {
		t.Errorf("Expected total %.2f, got %.2f", order.Total, unmarshalledOrder.Total)
	}
	if len(unmarshalledOrder.Items) != len(order.Items) {
		t.Errorf("Expected %d items, got %d", len(order.Items), len(unmarshalledOrder.Items))
	}
	if unmarshalledOrder.Status != order.Status {
		t.Errorf("Expected status %s, got %s", order.Status, unmarshalledOrder.Status)
	}
}

// TestOrderResponse - Order response test
func TestOrderResponse(t *testing.T) {
	orders := []OrderWithItems{
		{
			ID:              1,
			UserID:          1,
			Total:           99.97,
			Status:          "pending",
			PaymentMethod:   "credit_card",
			PaymentStatus:   "pending",
			ShippingAddress: "123 Main St",
			Items:           []OrderItemWithProduct{},
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              2,
			UserID:          1,
			Total:           149.95,
			Status:          "confirmed",
			PaymentMethod:   "paypal",
			PaymentStatus:   "paid",
			ShippingAddress: "456 Oak Ave",
			Items:           []OrderItemWithProduct{},
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}

	response := OrderResponse{
		Orders: orders,
		Total:  2,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Errorf("JSON marshalling failed: %v", err)
	}

	var unmarshalledResponse OrderResponse
	if err := json.Unmarshal(jsonData, &unmarshalledResponse); err != nil {
		t.Errorf("JSON unmarshalling failed: %v", err)
	}

	if unmarshalledResponse.Total != response.Total {
		t.Errorf("Expected total %d, got %d", response.Total, unmarshalledResponse.Total)
	}
	if len(unmarshalledResponse.Orders) != len(response.Orders) {
		t.Errorf("Expected %d orders, got %d", len(response.Orders), len(unmarshalledResponse.Orders))
	}
}

// TestOrderStatusValidation - Order status validation test
func TestOrderStatusValidation(t *testing.T) {
	validStatuses := []string{"pending", "confirmed", "preparing", "shipped", "delivered", "cancelled"}
	invalidStatuses := []string{"invalid", "unknown", "processing", "complete"}

	for _, status := range validStatuses {
		t.Run("Valid status: "+status, func(t *testing.T) {
			order := Order{Status: status}
			if order.Status != status {
				t.Errorf("Expected status %s, got %s", status, order.Status)
			}
		})
	}

	for _, status := range invalidStatuses {
		t.Run("Invalid status: "+status, func(t *testing.T) {
			// In a real application, you might want to validate status values
			// This test demonstrates how you might structure such validation
			order := Order{Status: status}
			if order.Status != status {
				t.Errorf("Expected status %s, got %s", status, order.Status)
			}
		})
	}
}

// TestOrderCalculations - Order calculations test
func TestOrderCalculations(t *testing.T) {
	// Test subtotal calculation for order item
	item := OrderItemWithProduct{
		Price:    29.99,
		Quantity: 3,
	}
	
	expectedSubtotal := 89.97
	actualSubtotal := item.Price * float64(item.Quantity)
	
	if actualSubtotal != expectedSubtotal {
		t.Errorf("Expected subtotal %.2f, got %.2f", expectedSubtotal, actualSubtotal)
	}
	
	// Test total calculation for order
	items := []OrderItemWithProduct{
		{Price: 29.99, Quantity: 2, Subtotal: 59.98},
		{Price: 19.99, Quantity: 1, Subtotal: 19.99},
		{Price: 39.99, Quantity: 1, Subtotal: 39.99},
	}
	
	expectedTotal := 119.96
	actualTotal := 0.0
	for _, item := range items {
		actualTotal += item.Subtotal
	}
	
	if actualTotal != expectedTotal {
		t.Errorf("Expected total %.2f, got %.2f", expectedTotal, actualTotal)
	}
}
