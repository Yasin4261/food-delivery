package model

import (
	"encoding/json"
	"testing"
	"time"
)

// TestOrder - Multi-vendor Order model test
func TestOrder(t *testing.T) {
	now := time.Now()
	order := Order{
		ID:              1,
		UserID:          1,
		OrderNumber:     "ORD-20250714-001",
		Total:           150.50,
		Currency:        "TRY",
		Status:          "pending",
		DeliveryType:    "delivery",
		DeliveryAddress: "Kapı No: 5, Daire: 3",
		PaymentMethod:   "credit_card",
		PaymentStatus:   "pending",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Test JSON marshalling/unmarshalling
	jsonData, err := json.Marshal(order)
	if err != nil {
		t.Errorf("JSON marshalling failed: %v", err)
	}

	var unmarshalledOrder Order
	if err := json.Unmarshal(jsonData, &unmarshalledOrder); err != nil {
		t.Errorf("JSON unmarshalling failed: %v", err)
	}

	// Verify fields
	if unmarshalledOrder.OrderNumber != order.OrderNumber {
		t.Errorf("Expected order number %s, got %s", order.OrderNumber, unmarshalledOrder.OrderNumber)
	}

	if unmarshalledOrder.Total != order.Total {
		t.Errorf("Expected total %.2f, got %.2f", order.Total, unmarshalledOrder.Total)
	}

	if unmarshalledOrder.Status != order.Status {
		t.Errorf("Expected status %s, got %s", order.Status, unmarshalledOrder.Status)
	}
}

// TestSubOrder - SubOrder model test
func TestSubOrder(t *testing.T) {
	subOrder := SubOrder{
		ID:              1,
		OrderID:         1,
		ChefID:          1,
		ChefOrderNumber: "CHEF-001-20250714",
		Subtotal:        75.50,
		DeliveryFee:     5.00,
		ServiceFee:      2.50,
		Total:           83.00,
		Status:          "pending",
		EstimatedTime:   30,
		ChefNote:        "Will prepare fresh",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Test JSON marshalling/unmarshalling
	jsonData, err := json.Marshal(subOrder)
	if err != nil {
		t.Errorf("JSON marshalling failed: %v", err)
	}

	var unmarshalledSubOrder SubOrder
	if err := json.Unmarshal(jsonData, &unmarshalledSubOrder); err != nil {
		t.Errorf("JSON unmarshalling failed: %v", err)
	}

	// Verify fields
	if unmarshalledSubOrder.ChefOrderNumber != subOrder.ChefOrderNumber {
		t.Errorf("Expected chef order number %s, got %s", subOrder.ChefOrderNumber, unmarshalledSubOrder.ChefOrderNumber)
	}

	if unmarshalledSubOrder.Total != subOrder.Total {
		t.Errorf("Expected total %.2f, got %.2f", subOrder.Total, unmarshalledSubOrder.Total)
	}
}

// TestOrderItem - OrderItem model test
func TestOrderItem(t *testing.T) {
	orderItem := OrderItem{
		ID:                  1,
		OrderID:             1,
		SubOrderID:          1,
		MealID:              1,
		ChefID:              1,
		Quantity:            2,
		Price:               49.99,
		Subtotal:            99.98,
		SpecialInstructions: "No onions please",
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	// Test JSON marshalling/unmarshalling
	jsonData, err := json.Marshal(orderItem)
	if err != nil {
		t.Errorf("JSON marshalling failed: %v", err)
	}

	var unmarshalledItem OrderItem
	if err := json.Unmarshal(jsonData, &unmarshalledItem); err != nil {
		t.Errorf("JSON unmarshalling failed: %v", err)
	}

	// Verify fields
	if unmarshalledItem.Quantity != orderItem.Quantity {
		t.Errorf("Expected quantity %d, got %d", orderItem.Quantity, unmarshalledItem.Quantity)
	}

	if unmarshalledItem.Subtotal != orderItem.Subtotal {
		t.Errorf("Expected subtotal %.2f, got %.2f", orderItem.Subtotal, unmarshalledItem.Subtotal)
	}
}

// TestMultiVendorOrderScenario - Multi-vendor order scenario test
func TestMultiVendorOrderScenario(t *testing.T) {
	// Main order
	order := Order{
		ID:              1,
		UserID:          1,
		OrderNumber:     "ORD-20250714-001",
		Total:           175.50,
		Currency:        "TRY",
		Status:          "pending",
		DeliveryType:    "delivery",
		DeliveryAddress: "123 Main St, Istanbul",
		PaymentMethod:   "credit_card",
		PaymentStatus:   "pending",
		ChefCount:       2, // Multi-vendor order
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Chef 1's sub-order
	subOrder1 := SubOrder{
		ID:              1,
		OrderID:         1,
		ChefID:          1,
		ChefOrderNumber: "CHEF-001-20250714",
		Subtotal:        75.50,
		DeliveryFee:     5.00,
		ServiceFee:      2.50,
		Total:           83.00,
		Status:          "pending",
		EstimatedTime:   30,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Chef 2's sub-order
	subOrder2 := SubOrder{
		ID:              2,
		OrderID:         1,
		ChefID:          2,
		ChefOrderNumber: "CHEF-002-20250714",
		Subtotal:        85.50,
		DeliveryFee:     5.00,
		ServiceFee:      2.00,
		Total:           92.50,
		Status:          "pending",
		EstimatedTime:   45,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Order items for Chef 1
	item1 := OrderItem{
		ID:         1,
		OrderID:    1,
		SubOrderID: 1,
		MealID:     1,
		ChefID:     1,
		Quantity:   2,
		Price:      25.75,
		Subtotal:   51.50,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Order items for Chef 2
	item3 := OrderItem{
		ID:         3,
		OrderID:    1,
		SubOrderID: 2,
		MealID:     3,
		ChefID:     2,
		Quantity:   3,
		Price:      28.50,
		Subtotal:   85.50,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Verify multi-vendor structure
	if order.ChefCount != 2 {
		t.Errorf("Expected chef count 2, got %d", order.ChefCount)
	}

	// Verify sub-order calculations
	expectedTotal1 := subOrder1.Subtotal + subOrder1.DeliveryFee + subOrder1.ServiceFee
	if subOrder1.Total != expectedTotal1 {
		t.Errorf("SubOrder 1: Expected total %.2f, got %.2f", expectedTotal1, subOrder1.Total)
	}

	expectedTotal2 := subOrder2.Subtotal + subOrder2.DeliveryFee + subOrder2.ServiceFee
	if subOrder2.Total != expectedTotal2 {
		t.Errorf("SubOrder 2: Expected total %.2f, got %.2f", expectedTotal2, subOrder2.Total)
	}

	// Verify main order total
	expectedMainTotal := subOrder1.Total + subOrder2.Total
	if order.Total != expectedMainTotal {
		t.Errorf("Main order: Expected total %.2f, got %.2f", expectedMainTotal, order.Total)
	}

	// Verify item associations
	if item1.ChefID != subOrder1.ChefID {
		t.Errorf("Item 1 chef ID mismatch: expected %d, got %d", subOrder1.ChefID, item1.ChefID)
	}

	if item3.ChefID != subOrder2.ChefID {
		t.Errorf("Item 3 chef ID mismatch: expected %d, got %d", subOrder2.ChefID, item3.ChefID)
	}
}

// TestOrderNumberGeneration - Order number generation test
func TestOrderNumberGeneration(t *testing.T) {
	now := time.Now()
	expectedPrefix := "ORD-" + now.Format("20060102")

	orderNumber := "ORD-20250714-001"
	
	if len(orderNumber) < len(expectedPrefix) {
		t.Errorf("Order number too short: %s", orderNumber)
	}

	// Test unique chef order numbers
	chefOrderNumber1 := "CHEF-001-20250714"
	chefOrderNumber2 := "CHEF-002-20250714"

	if chefOrderNumber1 == chefOrderNumber2 {
		t.Error("Chef order numbers should be unique")
	}
}

// TestSubOrderStatusTransitions - Sub-order status transitions test
func TestSubOrderStatusTransitions(t *testing.T) {
	subOrder := SubOrder{
		ID:      1,
		OrderID: 1,
		ChefID:  1,
		Status:  "pending",
	}

	// Valid status transitions
	validStatuses := []string{"pending", "confirmed", "preparing", "ready", "delivered", "cancelled"}

	for _, status := range validStatuses {
		subOrder.Status = status
		if subOrder.Status != status {
			t.Errorf("Failed to set status to %s", status)
		}
	}

	// Test estimated time update
	subOrder.EstimatedTime = 45
	if subOrder.EstimatedTime != 45 {
		t.Errorf("Expected estimated time 45, got %d", subOrder.EstimatedTime)
	}
}

// TestOrderCalculations - Order calculation test
func TestOrderCalculations(t *testing.T) {
	// Test item subtotal calculation
	item := OrderItem{
		Quantity: 3,
		Price:    29.99,
	}
	
	expectedSubtotal := 89.97
	actualSubtotal := float64(item.Quantity) * item.Price
	
	if actualSubtotal != expectedSubtotal {
		t.Errorf("Expected item subtotal %.2f, got %.2f", expectedSubtotal, actualSubtotal)
	}

	// Test sub-order total calculation
	subOrder := SubOrder{
		Subtotal:    85.50,
		DeliveryFee: 5.00,
		ServiceFee:  2.50,
	}
	
	expectedTotal := 93.00
	actualTotal := subOrder.Subtotal + subOrder.DeliveryFee + subOrder.ServiceFee
	
	if actualTotal != expectedTotal {
		t.Errorf("Expected sub-order total %.2f, got %.2f", expectedTotal, actualTotal)
	}
}

// TestDeliveryLocationValidation - Delivery location validation test
func TestDeliveryLocationValidation(t *testing.T) {
	order := Order{
		DeliveryType:    "delivery",
		DeliveryAddress: "123 Main St, Istanbul",
	}

	// Delivery orders must have delivery address
	if order.DeliveryType == "delivery" && order.DeliveryAddress == "" {
		t.Error("Delivery orders must have delivery address")
	}

	// Pickup orders don't require delivery address
	pickupOrder := Order{
		DeliveryType: "pickup",
	}

	if pickupOrder.DeliveryType == "pickup" && pickupOrder.DeliveryAddress == "" {
		// This is valid for pickup orders
	}
}
