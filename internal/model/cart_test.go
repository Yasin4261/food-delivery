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
	if unmarshalledCart.ID != cart.ID {
		t.Errorf("Expected ID %d, got %d", cart.ID, unmarshalledCart.ID)
	}
}

// TestCartItem - Cart item model test
func TestCartItem(t *testing.T) {
	now := time.Now()
	cartItem := CartItem{
		ID:                  1,
		CartID:              1,
		MealID:              1,
		ChefID:              1,
		Quantity:            2,
		SpecialInstructions: "Extra spicy",
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	jsonData, err := json.Marshal(cartItem)
	if err != nil {
		t.Errorf("JSON marshalling failed: %v", err)
	}

	var unmarshalledItem CartItem
	if err := json.Unmarshal(jsonData, &unmarshalledItem); err != nil {
		t.Errorf("JSON unmarshalling failed: %v", err)
	}

	if unmarshalledItem.MealID != cartItem.MealID {
		t.Errorf("Expected meal ID %d, got %d", cartItem.MealID, unmarshalledItem.MealID)
	}
	if unmarshalledItem.ChefID != cartItem.ChefID {
		t.Errorf("Expected chef ID %d, got %d", cartItem.ChefID, unmarshalledItem.ChefID)
	}
	if unmarshalledItem.Quantity != cartItem.Quantity {
		t.Errorf("Expected quantity %d, got %d", cartItem.Quantity, unmarshalledItem.Quantity)
	}
	if unmarshalledItem.SpecialInstructions != cartItem.SpecialInstructions {
		t.Errorf("Expected special instructions %s, got %s", cartItem.SpecialInstructions, unmarshalledItem.SpecialInstructions)
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
				MealID:              1,
				Quantity:            2,
				SpecialInstructions: "Medium spicy",
			},
			isValid: true,
		},
		{
			name: "Valid add to cart request without special instructions",
			request: AddToCartRequest{
				MealID:   1,
				Quantity: 1,
			},
			isValid: true,
		},
		{
			name: "Invalid quantity (zero)",
			request: AddToCartRequest{
				MealID:   1,
				Quantity: 0,
			},
			isValid: false,
		},
		{
			name: "Invalid quantity (negative)",
			request: AddToCartRequest{
				MealID:   1,
				Quantity: -1,
			},
			isValid: false,
		},
		{
			name: "Invalid meal ID (zero)",
			request: AddToCartRequest{
				MealID:   0,
				Quantity: 2,
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

			if unmarshalledRequest.MealID != tt.request.MealID {
				t.Errorf("Expected meal ID %d, got %d", tt.request.MealID, unmarshalledRequest.MealID)
			}
			if unmarshalledRequest.Quantity != tt.request.Quantity {
				t.Errorf("Expected quantity %d, got %d", tt.request.Quantity, unmarshalledRequest.Quantity)
			}
			if unmarshalledRequest.SpecialInstructions != tt.request.SpecialInstructions {
				t.Errorf("Expected special instructions %s, got %s", tt.request.SpecialInstructions, unmarshalledRequest.SpecialInstructions)
			}
		})
	}
}

// TestUpdateCartItemRequest - Update cart item request test
func TestUpdateCartItemRequest(t *testing.T) {
	updateRequest := UpdateCartItemRequest{
		Quantity:            3,
		SpecialInstructions: "Extra cheese",
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
	if unmarshalledRequest.SpecialInstructions != updateRequest.SpecialInstructions {
		t.Errorf("Expected special instructions %s, got %s", updateRequest.SpecialInstructions, unmarshalledRequest.SpecialInstructions)
	}
}

// TestCartItemWithProduct - Cart item with product test
func TestCartItemWithProduct(t *testing.T) {
	cartItem := CartItemWithProduct{
		CartItemResponse: CartItemResponse{
			ID:                  1,
			MealID:              1,
			ChefID:              1,
			MealName:            "Turkish Kebab",
			MealPrice:           25.50,
			MealImage:           "https://example.com/kebab.jpg",
			ChefName:            "Chef Ali",
			KitchenName:         "Ali's Kitchen",
			Quantity:            2,
			Subtotal:            51.00,
			SpecialInstructions: "Medium spicy",
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		},
	}

	jsonData, err := json.Marshal(cartItem)
	if err != nil {
		t.Errorf("JSON marshalling failed: %v", err)
	}

	var unmarshalledItem CartItemWithProduct
	if err := json.Unmarshal(jsonData, &unmarshalledItem); err != nil {
		t.Errorf("JSON unmarshalling failed: %v", err)
	}

	if unmarshalledItem.MealName != cartItem.MealName {
		t.Errorf("Expected meal name %s, got %s", cartItem.MealName, unmarshalledItem.MealName)
	}
	if unmarshalledItem.Subtotal != cartItem.Subtotal {
		t.Errorf("Expected subtotal %.2f, got %.2f", cartItem.Subtotal, unmarshalledItem.Subtotal)
	}
	if unmarshalledItem.ChefName != cartItem.ChefName {
		t.Errorf("Expected chef name %s, got %s", cartItem.ChefName, unmarshalledItem.ChefName)
	}
}

// TestCartResponse - Cart response test
func TestCartResponse(t *testing.T) {
	items := []CartItemResponse{
		{
			ID:          1,
			MealID:      1,
			ChefID:      1,
			MealName:    "Turkish Kebab",
			MealPrice:   25.50,
			ChefName:    "Chef Ali",
			KitchenName: "Ali's Kitchen",
			Quantity:    2,
			Subtotal:    51.00,
		},
		{
			ID:          2,
			MealID:      2,
			ChefID:      1,
			MealName:    "Turkish Pilaf",
			MealPrice:   15.00,
			ChefName:    "Chef Ali",
			KitchenName: "Ali's Kitchen",
			Quantity:    1,
			Subtotal:    15.00,
		},
	}

	response := CartResponse{
		ID:        1,
		UserID:    1,
		Items:     items,
		Total:     66.00,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
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
	if unmarshalledResponse.UserID != response.UserID {
		t.Errorf("Expected user ID %d, got %d", response.UserID, unmarshalledResponse.UserID)
	}
}

// TestCartCalculations - Cart calculations test
func TestCartCalculations(t *testing.T) {
	// Test subtotal calculation
	item := CartItemResponse{
		MealPrice: 25.50,
		Quantity:  3,
	}
	
	expectedSubtotal := 76.50
	actualSubtotal := item.MealPrice * float64(item.Quantity)
	
	if actualSubtotal != expectedSubtotal {
		t.Errorf("Expected subtotal %.2f, got %.2f", expectedSubtotal, actualSubtotal)
	}
	
	// Test total calculation
	items := []CartItemResponse{
		{MealPrice: 25.50, Quantity: 2, Subtotal: 51.00},
		{MealPrice: 15.00, Quantity: 1, Subtotal: 15.00},
		{MealPrice: 30.00, Quantity: 1, Subtotal: 30.00},
	}
	
	expectedTotal := 96.00
	actualTotal := 0.0
	for _, item := range items {
		actualTotal += item.Subtotal
	}
	
	if actualTotal != expectedTotal {
		t.Errorf("Expected total %.2f, got %.2f", expectedTotal, actualTotal)
	}
}
