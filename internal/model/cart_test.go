package model

import (
	"encoding/json"
	"testing"
	"time"
)

// TestCart - Cart model test
func TestCart(t *testing.T) {
	now := time.Now()
	cart := Cart{
		ID:        1,
		UserID:    1,
		Total:     59.98,
		CreatedAt: now,
		UpdatedAt: now,
	}

	jsonData, err := json.Marshal(cart)
	if err != nil {
		t.Errorf("JSON marshalling failed: %v", err)
	}

	var unmarshalledCart Cart
	if err := json.Unmarshal(jsonData, &unmarshalledCart); err != nil {
		t.Errorf("JSON unmarshalling failed: %v", err)
	}

	if unmarshalledCart.UserID != cart.UserID {
		t.Errorf("Expected user ID %d, got %d", cart.UserID, unmarshalledCart.UserID)
	}
	if unmarshalledCart.Total != cart.Total {
		t.Errorf("Expected total %.2f, got %.2f", cart.Total, unmarshalledCart.Total)
	}
}

// TestCartItem - Cart item model test
func TestCartItem(t *testing.T) {
	now := time.Now()
	cartItem := CartItem{
		ID:        1,
		CartID:    1,
		ProductID: 1,
		Quantity:  2,
		Price:     29.99,
		CreatedAt: now,
		UpdatedAt: now,
	}

	jsonData, err := json.Marshal(cartItem)
	if err != nil {
		t.Errorf("JSON marshalling failed: %v", err)
	}

	var unmarshalledItem CartItem
	if err := json.Unmarshal(jsonData, &unmarshalledItem); err != nil {
		t.Errorf("JSON unmarshalling failed: %v", err)
	}

	if unmarshalledItem.ProductID != cartItem.ProductID {
		t.Errorf("Expected product ID %d, got %d", cartItem.ProductID, unmarshalledItem.ProductID)
	}
	if unmarshalledItem.Quantity != cartItem.Quantity {
		t.Errorf("Expected quantity %d, got %d", cartItem.Quantity, unmarshalledItem.Quantity)
	}
	if unmarshalledItem.Price != cartItem.Price {
		t.Errorf("Expected price %.2f, got %.2f", cartItem.Price, unmarshalledItem.Price)
	}
}

// TestAddToCartRequest - Add to cart request test
func TestAddToCartRequest(t *testing.T) {
	tests := []struct {
		name    string
		request AddToCartRequest
		isValid bool
	}{
		{
			name: "Valid add to cart request",
			request: AddToCartRequest{
				ProductID: 1,
				Quantity:  2,
			},
			isValid: true,
		},
		{
			name: "Invalid quantity (zero)",
			request: AddToCartRequest{
				ProductID: 1,
				Quantity:  0,
			},
			isValid: false,
		},
		{
			name: "Invalid quantity (negative)",
			request: AddToCartRequest{
				ProductID: 1,
				Quantity:  -1,
			},
			isValid: false,
		},
		{
			name: "Invalid product ID (zero)",
			request: AddToCartRequest{
				ProductID: 0,
				Quantity:  2,
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

			var unmarshalledRequest AddToCartRequest
			if err := json.Unmarshal(jsonData, &unmarshalledRequest); err != nil {
				t.Errorf("JSON unmarshalling failed: %v", err)
			}

			if unmarshalledRequest.ProductID != tt.request.ProductID {
				t.Errorf("Expected product ID %d, got %d", tt.request.ProductID, unmarshalledRequest.ProductID)
			}
			if unmarshalledRequest.Quantity != tt.request.Quantity {
				t.Errorf("Expected quantity %d, got %d", tt.request.Quantity, unmarshalledRequest.Quantity)
			}
		})
	}
}

// TestUpdateCartItemRequest - Update cart item request test
func TestUpdateCartItemRequest(t *testing.T) {
	updateRequest := UpdateCartItemRequest{
		Quantity: 3,
	}

	jsonData, err := json.Marshal(updateRequest)
	if err != nil {
		t.Errorf("JSON marshalling failed: %v", err)
	}

	var unmarshalledRequest UpdateCartItemRequest
	if err := json.Unmarshal(jsonData, &unmarshalledRequest); err != nil {
		t.Errorf("JSON unmarshalling failed: %v", err)
	}

	if unmarshalledRequest.Quantity != updateRequest.Quantity {
		t.Errorf("Expected quantity %d, got %d", updateRequest.Quantity, unmarshalledRequest.Quantity)
	}
}

// TestCartItemWithProduct - Cart item with product test
func TestCartItemWithProduct(t *testing.T) {
	cartItem := CartItemWithProduct{
		ID:           1,
		CartID:       1,
		ProductID:    1,
		ProductName:  "Test Product",
		ProductPrice: 29.99,
		ProductImage: "https://example.com/image.jpg",
		Quantity:     2,
		Price:        29.99,
		Subtotal:     59.98,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	jsonData, err := json.Marshal(cartItem)
	if err != nil {
		t.Errorf("JSON marshalling failed: %v", err)
	}

	var unmarshalledItem CartItemWithProduct
	if err := json.Unmarshal(jsonData, &unmarshalledItem); err != nil {
		t.Errorf("JSON unmarshalling failed: %v", err)
	}

	if unmarshalledItem.ProductName != cartItem.ProductName {
		t.Errorf("Expected product name %s, got %s", cartItem.ProductName, unmarshalledItem.ProductName)
	}
	if unmarshalledItem.Subtotal != cartItem.Subtotal {
		t.Errorf("Expected subtotal %.2f, got %.2f", cartItem.Subtotal, unmarshalledItem.Subtotal)
	}
}

// TestCartResponse - Cart response test
func TestCartResponse(t *testing.T) {
	items := []CartItemWithProduct{
		{
			ID:           1,
			ProductName:  "Product 1",
			ProductPrice: 29.99,
			Quantity:     2,
			Subtotal:     59.98,
		},
		{
			ID:           2,
			ProductName:  "Product 2",
			ProductPrice: 19.99,
			Quantity:     1,
			Subtotal:     19.99,
		},
	}

	response := CartResponse{
		Items: items,
		Total: 79.97,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Errorf("JSON marshalling failed: %v", err)
	}

	var unmarshalledResponse CartResponse
	if err := json.Unmarshal(jsonData, &unmarshalledResponse); err != nil {
		t.Errorf("JSON unmarshalling failed: %v", err)
	}

	if unmarshalledResponse.Total != response.Total {
		t.Errorf("Expected total %.2f, got %.2f", response.Total, unmarshalledResponse.Total)
	}
	if len(unmarshalledResponse.Items) != len(response.Items) {
		t.Errorf("Expected %d items, got %d", len(response.Items), len(unmarshalledResponse.Items))
	}
}

// TestCartCalculations - Cart calculations test
func TestCartCalculations(t *testing.T) {
	// Test subtotal calculation
	item := CartItemWithProduct{
		ProductPrice: 29.99,
		Quantity:     3,
	}
	
	expectedSubtotal := 89.97
	actualSubtotal := item.ProductPrice * float64(item.Quantity)
	
	if actualSubtotal != expectedSubtotal {
		t.Errorf("Expected subtotal %.2f, got %.2f", expectedSubtotal, actualSubtotal)
	}
	
	// Test total calculation
	items := []CartItemWithProduct{
		{ProductPrice: 29.99, Quantity: 2, Subtotal: 59.98},
		{ProductPrice: 19.99, Quantity: 1, Subtotal: 19.99},
		{ProductPrice: 9.99, Quantity: 3, Subtotal: 29.97},
	}
	
	expectedTotal := 109.94
	actualTotal := 0.0
	for _, item := range items {
		actualTotal += item.Subtotal
	}
	
	if actualTotal != expectedTotal {
		t.Errorf("Expected total %.2f, got %.2f", expectedTotal, actualTotal)
	}
}
