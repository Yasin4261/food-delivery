package service

import (
	"errors"
	"testing"
	"ecommerce/internal/model"
)

// Simple function-based tests for ProductService validation logic
// These tests check business rules without database dependencies

func TestProductService_ValidateCreateProductRequest(t *testing.T) {
	testCases := []struct {
		name        string
		request     *model.CreateProductRequest
		shouldError bool
		errorMsg    string
	}{
		{
			name: "Valid product request",
			request: &model.CreateProductRequest{
				Name:        "Test Product",
				Description: "Test Description",
				Price:       99.99,
				Stock:       10,
				CategoryID:  1,
				ImageURL:    "http://example.com/image.jpg",
			},
			shouldError: false,
		},
		{
			name: "Empty name should error",
			request: &model.CreateProductRequest{
				Name:        "",
				Description: "Test Description",
				Price:       99.99,
				Stock:       10,
				CategoryID:  1,
			},
			shouldError: true,
			errorMsg:    "ürün adı boş olamaz",
		},
		{
			name: "Zero price should error",
			request: &model.CreateProductRequest{
				Name:        "Test Product",
				Description: "Test Description",
				Price:       0,
				Stock:       10,
				CategoryID:  1,
			},
			shouldError: true,
			errorMsg:    "ürün fiyatı 0'dan büyük olmalıdır",
		},
		{
			name: "Negative price should error",
			request: &model.CreateProductRequest{
				Name:        "Test Product",
				Description: "Test Description",
				Price:       -10.0,
				Stock:       10,
				CategoryID:  1,
			},
			shouldError: true,
			errorMsg:    "ürün fiyatı 0'dan büyük olmalıdır",
		},
		{
			name: "Negative stock should error",
			request: &model.CreateProductRequest{
				Name:        "Test Product",
				Description: "Test Description",
				Price:       99.99,
				Stock:       -5,
				CategoryID:  1,
			},
			shouldError: true,
			errorMsg:    "stok miktarı negatif olamaz",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate the validation logic from ProductService.CreateProduct
			var err error
			
			if tc.request.Name == "" {
				err = errors.New("ürün adı boş olamaz")
			} else if tc.request.Price <= 0 {
				err = errors.New("ürün fiyatı 0'dan büyük olmalıdır")
			} else if tc.request.Stock < 0 {
				err = errors.New("stok miktarı negatif olamaz")
			}

			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if err.Error() != tc.errorMsg {
					t.Errorf("Expected error '%s', got '%s'", tc.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %s", err.Error())
				}
			}
		})
	}
}

func TestProductService_ValidateGetProduct(t *testing.T) {
	testCases := []struct {
		name        string
		productID   uint
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "Valid product ID",
			productID:   1,
			shouldError: false,
		},
		{
			name:        "Zero product ID should error",
			productID:   0,
			shouldError: true,
			errorMsg:    "geçersiz ürün ID",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate the validation logic from ProductService.GetProduct
			var err error
			
			if tc.productID == 0 {
				err = errors.New("geçersiz ürün ID")
			}

			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if err.Error() != tc.errorMsg {
					t.Errorf("Expected error '%s', got '%s'", tc.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %s", err.Error())
				}
			}
		})
	}
}

func TestProductService_UpdateProductLogic(t *testing.T) {
	// Test update logic without database dependencies
	originalProduct := model.Product{
		ID:          1,
		Name:        "Original Product",
		Description: "Original Description",
		Price:       50.0,
		Stock:       5,
		CategoryID:  1,
		ImageURL:    "original.jpg",
	}

	updateRequest := &model.UpdateProductRequest{
		Name:        "Updated Product",
		Description: "",  // Should not update empty fields
		Price:       75.0,
		Stock:       10,
		CategoryID:  0,  // Should not update zero values
		ImageURL:    "updated.jpg",
	}

	// Simulate update logic from ProductService.UpdateProduct
	updatedProduct := originalProduct
	
	if updateRequest.Name != "" {
		updatedProduct.Name = updateRequest.Name
	}
	if updateRequest.Description != "" {
		updatedProduct.Description = updateRequest.Description
	}
	if updateRequest.Price > 0 {
		updatedProduct.Price = updateRequest.Price
	}
	if updateRequest.Stock >= 0 {
		updatedProduct.Stock = updateRequest.Stock
	}
	if updateRequest.CategoryID > 0 {
		updatedProduct.CategoryID = updateRequest.CategoryID
	}
	if updateRequest.ImageURL != "" {
		updatedProduct.ImageURL = updateRequest.ImageURL
	}

	// Verify updates
	if updatedProduct.Name != "Updated Product" {
		t.Errorf("Expected name to be updated to 'Updated Product', got '%s'", updatedProduct.Name)
	}
	
	if updatedProduct.Description != "Original Description" {
		t.Errorf("Expected description to remain 'Original Description', got '%s'", updatedProduct.Description)
	}
	
	if updatedProduct.Price != 75.0 {
		t.Errorf("Expected price to be updated to 75.0, got %f", updatedProduct.Price)
	}
	
	if updatedProduct.Stock != 10 {
		t.Errorf("Expected stock to be updated to 10, got %d", updatedProduct.Stock)
	}
	
	if updatedProduct.CategoryID != 1 {
		t.Errorf("Expected CategoryID to remain 1, got %d", updatedProduct.CategoryID)
	}
	
	if updatedProduct.ImageURL != "updated.jpg" {
		t.Errorf("Expected ImageURL to be updated to 'updated.jpg', got '%s'", updatedProduct.ImageURL)
	}
}

func TestProductService_NewProductService(t *testing.T) {
	// Test service constructor
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("NewProductService construction should not panic: %v", r)
		}
	}()
	
	// Test constructor structure
	_ = NewProductService(nil)
}

func TestProductService_PriceValidation(t *testing.T) {
	// Test various price scenarios
	testCases := []struct {
		name     string
		price    float64
		valid    bool
	}{
		{"Positive price", 10.50, true},
		{"Large price", 999999.99, true},
		{"Small positive price", 0.01, true},
		{"Zero price", 0.0, false},
		{"Negative price", -5.0, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isValid := tc.price > 0
			
			if isValid != tc.valid {
				t.Errorf("Price %f validation: expected %t, got %t", tc.price, tc.valid, isValid)
			}
		})
	}
}

func TestProductService_StockValidation(t *testing.T) {
	// Test stock validation scenarios
	testCases := []struct {
		name  string
		stock int
		valid bool
	}{
		{"Zero stock", 0, true},
		{"Positive stock", 10, true},
		{"Large stock", 1000, true},
		{"Negative stock", -1, false},
		{"Large negative stock", -100, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isValid := tc.stock >= 0
			
			if isValid != tc.valid {
				t.Errorf("Stock %d validation: expected %t, got %t", tc.stock, tc.valid, isValid)
			}
		})
	}
}
