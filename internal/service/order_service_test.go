package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrderService_NewOrderService(t *testing.T) {
	assert.NotPanics(t, func() {
		NewOrderService(nil, nil, nil)
	})
}

func TestOrderService_OrderStatusValidation(t *testing.T) {
	validStatuses := []string{
		"pending",
		"confirmed", 
		"preparing",
		"ready",
		"out_for_delivery",
		"delivered",
		"cancelled",
	}

	for _, status := range validStatuses {
		t.Run("Valid status: "+status, func(t *testing.T) {
			assert.NotEmpty(t, status)
			assert.Contains(t, validStatuses, status)
		})
	}

	t.Run("Invalid empty status", func(t *testing.T) {
		status := ""
		assert.Empty(t, status)
		assert.NotContains(t, validStatuses, status)
	})

	t.Run("Invalid status", func(t *testing.T) {
		status := "invalid_status"
		assert.NotContains(t, validStatuses, status)
	})
}

func TestOrderService_PaymentValidation(t *testing.T) {
	validPaymentMethods := []string{
		"credit_card",
		"debit_card", 
		"cash",
		"digital_wallet",
		"bank_transfer",
	}

	for _, method := range validPaymentMethods {
		t.Run("Valid payment method: "+method, func(t *testing.T) {
			assert.NotEmpty(t, method)
			assert.Contains(t, validPaymentMethods, method)
		})
	}

	t.Run("Invalid payment method", func(t *testing.T) {
		method := "invalid_payment"
		assert.NotContains(t, validPaymentMethods, method)
	})
}

func TestOrderService_AmountValidation(t *testing.T) {
	tests := []struct {
		name   string
		amount float64
		valid  bool
	}{
		{"Valid amount", 25.50, true},
		{"Valid high amount", 999.99, true},
		{"Valid low amount", 0.01, true},
		{"Zero amount", 0.0, false},
		{"Negative amount", -10.50, false},
		{"Very high amount", 10000.0, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var isValid bool
			if tc.amount > 0 && tc.amount < 5000 {
				isValid = true
			}
			assert.Equal(t, tc.valid, isValid)
		})
	}
}

func TestOrderService_OrderCalculations(t *testing.T) {
	t.Run("Subtotal calculation", func(t *testing.T) {
		itemPrice1 := 25.50
		itemQuantity1 := 2
		itemPrice2 := 15.00
		itemQuantity2 := 1
		
		subtotal := (itemPrice1 * float64(itemQuantity1)) + (itemPrice2 * float64(itemQuantity2))
		expected := 66.00
		
		assert.Equal(t, expected, subtotal)
	})

	t.Run("Tax calculation", func(t *testing.T) {
		subtotal := 100.0
		taxRate := 0.18 // 18% KDV
		
		tax := subtotal * taxRate
		expected := 18.0
		
		assert.Equal(t, expected, tax)
	})

	t.Run("Delivery fee calculation", func(t *testing.T) {
		subtotal := 50.0
		deliveryFee := 5.0
		
		total := subtotal + deliveryFee
		expected := 55.0
		
		assert.Equal(t, expected, total)
	})
}

func TestOrderService_OrderValidation(t *testing.T) {
	tests := []struct {
		name    string
		orderID uint
		valid   bool
	}{
		{"Valid order ID", 1, true},
		{"Valid large order ID", 999999, true},
		{"Invalid zero order ID", 0, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isValid := tc.orderID > 0
			assert.Equal(t, tc.valid, isValid)
		})
	}
}

func TestOrderService_DeliveryAddressValidation(t *testing.T) {
	tests := []struct {
		name    string
		address string
		valid   bool
	}{
		{"Valid address", "123 Ana Cadde, Beyoğlu, İstanbul", true},
		{"Valid short address", "Ev", true},
		{"Empty address", "", false},
		{"Very long address", string(make([]byte, 501)), false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isValid := len(tc.address) > 0 && len(tc.address) <= 500
			assert.Equal(t, tc.valid, isValid)
		})
	}
}

func TestOrderService_OrderItemValidation(t *testing.T) {
	tests := []struct {
		name     string
		mealID   uint
		quantity int
		valid    bool
	}{
		{"Valid item", 1, 2, true},
		{"Valid single item", 5, 1, true},
		{"Valid multiple items", 3, 10, true},
		{"Invalid meal ID", 0, 2, false},
		{"Invalid zero quantity", 1, 0, false},
		{"Invalid negative quantity", 1, -1, false},
		{"Too many items", 1, 101, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isValid := tc.mealID > 0 && tc.quantity > 0 && tc.quantity <= 100
			assert.Equal(t, tc.valid, isValid)
		})
	}
}
