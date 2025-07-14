package model

import (
	"encoding/json"
	"testing"
	"time"
)

// TestProduct - Product model testleri
func TestProduct(t *testing.T) {
	now := time.Now()
	product := Product{
		ID:          1,
		Name:        "Test Product",
		Description: "Test product description",
		Price:       19.99,
		ImageURL:    "https://example.com/image.jpg",
		CategoryID:  1,
		Stock:       100,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Test JSON marshalling
	jsonData, err := json.Marshal(product)
	if err != nil {
		t.Errorf("JSON marshalling failed: %v", err)
	}

	// Test struct unmarshalling
	var unmarshalledProduct Product
	if err := json.Unmarshal(jsonData, &unmarshalledProduct); err != nil {
		t.Errorf("JSON unmarshalling to struct failed: %v", err)
	}

	if unmarshalledProduct.Name != product.Name {
		t.Errorf("Expected name %s, got %s", product.Name, unmarshalledProduct.Name)
	}
	if unmarshalledProduct.Price != product.Price {
		t.Errorf("Expected price %.2f, got %.2f", product.Price, unmarshalledProduct.Price)
	}
	if unmarshalledProduct.Stock != product.Stock {
		t.Errorf("Expected stock %d, got %d", product.Stock, unmarshalledProduct.Stock)
	}
}

// TestCreateProductRequest - Create product request test
func TestCreateProductRequest(t *testing.T) {
	tests := []struct {
		name    string
		request CreateProductRequest
		isValid bool
	}{
		{
			name: "Valid product request",
			request: CreateProductRequest{
				Name:        "Test Product",
				Description: "Test description",
				Price:       29.99,
				ImageURL:    "https://example.com/image.jpg",
				CategoryID:  1,
				Stock:       50,
			},
			isValid: true,
		},
		{
			name: "Invalid price (negative)",
			request: CreateProductRequest{
				Name:        "Test Product",
				Description: "Test description",
				Price:       -5.99,
				ImageURL:    "https://example.com/image.jpg",
				CategoryID:  1,
				Stock:       50,
			},
			isValid: false,
		},
		{
			name: "Invalid stock (negative)",
			request: CreateProductRequest{
				Name:        "Test Product",
				Description: "Test description",
				Price:       29.99,
				ImageURL:    "https://example.com/image.jpg",
				CategoryID:  1,
				Stock:       -10,
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

			var unmarshalledRequest CreateProductRequest
			if err := json.Unmarshal(jsonData, &unmarshalledRequest); err != nil {
				t.Errorf("JSON unmarshalling failed: %v", err)
			}

			if unmarshalledRequest.Name != tt.request.Name {
				t.Errorf("Expected name %s, got %s", tt.request.Name, unmarshalledRequest.Name)
			}
		})
	}
}

// TestUpdateProductRequest - Update product request test
func TestUpdateProductRequest(t *testing.T) {
	updateRequest := UpdateProductRequest{
		Name:        "Updated Product",
		Description: "Updated description",
		Price:       39.99,
		ImageURL:    "https://example.com/updated-image.jpg",
		CategoryID:  2,
		Stock:       75,
		IsActive:    false,
	}

	jsonData, err := json.Marshal(updateRequest)
	if err != nil {
		t.Errorf("JSON marshalling failed: %v", err)
	}

	var unmarshalledRequest UpdateProductRequest
	if err := json.Unmarshal(jsonData, &unmarshalledRequest); err != nil {
		t.Errorf("JSON unmarshalling failed: %v", err)
	}

	if unmarshalledRequest.Name != updateRequest.Name {
		t.Errorf("Expected name %s, got %s", updateRequest.Name, unmarshalledRequest.Name)
	}
	if unmarshalledRequest.Price != updateRequest.Price {
		t.Errorf("Expected price %.2f, got %.2f", updateRequest.Price, unmarshalledRequest.Price)
	}
	if unmarshalledRequest.IsActive != updateRequest.IsActive {
		t.Errorf("Expected active status %v, got %v", updateRequest.IsActive, unmarshalledRequest.IsActive)
	}
}

// TestCategory - Category model test
func TestCategory(t *testing.T) {
	category := Category{
		ID:          1,
		Name:        "Electronics",
		Description: "Electronic devices and accessories",
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	jsonData, err := json.Marshal(category)
	if err != nil {
		t.Errorf("JSON marshalling failed: %v", err)
	}

	var unmarshalledCategory Category
	if err := json.Unmarshal(jsonData, &unmarshalledCategory); err != nil {
		t.Errorf("JSON unmarshalling failed: %v", err)
	}

	if unmarshalledCategory.Name != category.Name {
		t.Errorf("Expected name %s, got %s", category.Name, unmarshalledCategory.Name)
	}
	if unmarshalledCategory.IsActive != category.IsActive {
		t.Errorf("Expected active status %v, got %v", category.IsActive, unmarshalledCategory.IsActive)
	}
}

// TestProductWithCategory - Product with category test
func TestProductWithCategory(t *testing.T) {
	product := ProductWithCategory{
		ID:           1,
		Name:         "Smartphone",
		Description:  "Latest smartphone",
		Price:        699.99,
		ImageURL:     "https://example.com/phone.jpg",
		CategoryID:   1,
		CategoryName: "Electronics",
		Stock:        25,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	jsonData, err := json.Marshal(product)
	if err != nil {
		t.Errorf("JSON marshalling failed: %v", err)
	}

	var unmarshalledProduct ProductWithCategory
	if err := json.Unmarshal(jsonData, &unmarshalledProduct); err != nil {
		t.Errorf("JSON unmarshalling failed: %v", err)
	}

	if unmarshalledProduct.Name != product.Name {
		t.Errorf("Expected name %s, got %s", product.Name, unmarshalledProduct.Name)
	}
	if unmarshalledProduct.CategoryName != product.CategoryName {
		t.Errorf("Expected category name %s, got %s", product.CategoryName, unmarshalledProduct.CategoryName)
	}
}

// TestProductResponse - Product response test
func TestProductResponse(t *testing.T) {
	products := []ProductWithCategory{
		{
			ID:           1,
			Name:         "Product 1",
			Description:  "Description 1",
			Price:        29.99,
			CategoryName: "Category 1",
			Stock:        10,
			IsActive:     true,
		},
		{
			ID:           2,
			Name:         "Product 2",
			Description:  "Description 2",
			Price:        39.99,
			CategoryName: "Category 2",
			Stock:        5,
			IsActive:     true,
		},
	}

	response := ProductResponse{
		Products: products,
		Total:    2,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Errorf("JSON marshalling failed: %v", err)
	}

	var unmarshalledResponse ProductResponse
	if err := json.Unmarshal(jsonData, &unmarshalledResponse); err != nil {
		t.Errorf("JSON unmarshalling failed: %v", err)
	}

	if unmarshalledResponse.Total != response.Total {
		t.Errorf("Expected total %d, got %d", response.Total, unmarshalledResponse.Total)
	}
	if len(unmarshalledResponse.Products) != len(response.Products) {
		t.Errorf("Expected %d products, got %d", len(response.Products), len(unmarshalledResponse.Products))
	}
}
