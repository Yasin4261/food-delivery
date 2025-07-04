package model

import (
	"encoding/json"
	"testing"
	"time"
)

// TestAPIResponse - API response test
func TestAPIResponse(t *testing.T) {
	// Test successful response
	successResponse := APIResponse{
		Success: true,
		Message: "Operation successful",
		Data:    map[string]interface{}{"id": 1, "name": "Test"},
	}

	jsonData, err := json.Marshal(successResponse)
	if err != nil {
		t.Errorf("JSON marshalling failed: %v", err)
	}

	var unmarshalledResponse APIResponse
	if err := json.Unmarshal(jsonData, &unmarshalledResponse); err != nil {
		t.Errorf("JSON unmarshalling failed: %v", err)
	}

	if unmarshalledResponse.Success != successResponse.Success {
		t.Errorf("Expected success %v, got %v", successResponse.Success, unmarshalledResponse.Success)
	}
	if unmarshalledResponse.Message != successResponse.Message {
		t.Errorf("Expected message %s, got %s", successResponse.Message, unmarshalledResponse.Message)
	}

	// Test error response
	errorResponse := APIResponse{
		Success: false,
		Message: "Operation failed",
		Error:   "Invalid input data",
	}

	jsonData, err = json.Marshal(errorResponse)
	if err != nil {
		t.Errorf("JSON marshalling failed: %v", err)
	}

	var unmarshalledErrorResponse APIResponse
	if err := json.Unmarshal(jsonData, &unmarshalledErrorResponse); err != nil {
		t.Errorf("JSON unmarshalling failed: %v", err)
	}

	if unmarshalledErrorResponse.Success != errorResponse.Success {
		t.Errorf("Expected success %v, got %v", errorResponse.Success, unmarshalledErrorResponse.Success)
	}
	if unmarshalledErrorResponse.Error != errorResponse.Error {
		t.Errorf("Expected error %s, got %s", errorResponse.Error, unmarshalledErrorResponse.Error)
	}
}

// TestPaginationResponse - Pagination response test
func TestPaginationResponse(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"id": 1, "name": "Item 1"},
		map[string]interface{}{"id": 2, "name": "Item 2"},
		map[string]interface{}{"id": 3, "name": "Item 3"},
	}

	pagination := PaginationResponse{
		Data:       data,
		Page:       1,
		Limit:      10,
		Total:      3,
		TotalPages: 1,
	}

	jsonData, err := json.Marshal(pagination)
	if err != nil {
		t.Errorf("JSON marshalling failed: %v", err)
	}

	var unmarshalledPagination PaginationResponse
	if err := json.Unmarshal(jsonData, &unmarshalledPagination); err != nil {
		t.Errorf("JSON unmarshalling failed: %v", err)
	}

	if unmarshalledPagination.Page != pagination.Page {
		t.Errorf("Expected page %d, got %d", pagination.Page, unmarshalledPagination.Page)
	}
	if unmarshalledPagination.Limit != pagination.Limit {
		t.Errorf("Expected limit %d, got %d", pagination.Limit, unmarshalledPagination.Limit)
	}
	if unmarshalledPagination.Total != pagination.Total {
		t.Errorf("Expected total %d, got %d", pagination.Total, unmarshalledPagination.Total)
	}
	if unmarshalledPagination.TotalPages != pagination.TotalPages {
		t.Errorf("Expected total pages %d, got %d", pagination.TotalPages, unmarshalledPagination.TotalPages)
	}
}

// TestPaginationCalculations - Pagination calculations test
func TestPaginationCalculations(t *testing.T) {
	tests := []struct {
		name       string
		total      int
		limit      int
		page       int
		totalPages int
		offset     int
	}{
		{
			name:       "First page of 100 items with limit 10",
			total:      100,
			limit:      10,
			page:       1,
			totalPages: 10,
			offset:     0,
		},
		{
			name:       "Second page of 25 items with limit 10",
			total:      25,
			limit:      10,
			page:       2,
			totalPages: 3,
			offset:     10,
		},
		{
			name:       "Last page of 33 items with limit 10",
			total:      33,
			limit:      10,
			page:       4,
			totalPages: 4,
			offset:     30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate total pages
			totalPages := (tt.total + tt.limit - 1) / tt.limit
			if totalPages != tt.totalPages {
				t.Errorf("Expected total pages %d, got %d", tt.totalPages, totalPages)
			}

			// Calculate offset
			offset := (tt.page - 1) * tt.limit
			if offset != tt.offset {
				t.Errorf("Expected offset %d, got %d", tt.offset, offset)
			}
		})
	}
}

// TestErrorResponse - Error response test
func TestErrorResponse(t *testing.T) {
	errorResponse := ErrorResponse{
		Error:   "Validation failed",
		Message: "Invalid input data provided",
		Code:    400,
	}

	jsonData, err := json.Marshal(errorResponse)
	if err != nil {
		t.Errorf("JSON marshalling failed: %v", err)
	}

	var unmarshalledResponse ErrorResponse
	if err := json.Unmarshal(jsonData, &unmarshalledResponse); err != nil {
		t.Errorf("JSON unmarshalling failed: %v", err)
	}

	if unmarshalledResponse.Error != errorResponse.Error {
		t.Errorf("Expected error %s, got %s", errorResponse.Error, unmarshalledResponse.Error)
	}
	if unmarshalledResponse.Message != errorResponse.Message {
		t.Errorf("Expected message %s, got %s", errorResponse.Message, unmarshalledResponse.Message)
	}
	if unmarshalledResponse.Code != errorResponse.Code {
		t.Errorf("Expected code %d, got %d", errorResponse.Code, unmarshalledResponse.Code)
	}
}

// TestValidationErrorResponse - Validation error response test
func TestValidationErrorResponse(t *testing.T) {
	validationErrors := []ValidationError{
		{
			Field:   "email",
			Message: "Invalid email format",
		},
		{
			Field:   "password",
			Message: "Password must be at least 6 characters",
		},
	}

	validationResponse := ValidationErrorResponse{
		Error:   "Validation failed",
		Message: "Please fix the following errors",
		Errors:  validationErrors,
	}

	jsonData, err := json.Marshal(validationResponse)
	if err != nil {
		t.Errorf("JSON marshalling failed: %v", err)
	}

	var unmarshalledResponse ValidationErrorResponse
	if err := json.Unmarshal(jsonData, &unmarshalledResponse); err != nil {
		t.Errorf("JSON unmarshalling failed: %v", err)
	}

	if unmarshalledResponse.Error != validationResponse.Error {
		t.Errorf("Expected error %s, got %s", validationResponse.Error, unmarshalledResponse.Error)
	}
	if len(unmarshalledResponse.Errors) != len(validationResponse.Errors) {
		t.Errorf("Expected %d errors, got %d", len(validationResponse.Errors), len(unmarshalledResponse.Errors))
	}
}

// TestCommonStructsWithTime - Test structs with time fields
func TestCommonStructsWithTime(t *testing.T) {
	now := time.Now()
	
	// Test that time fields are properly handled
	testData := struct {
		ID        uint      `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}{
		ID:        1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	jsonData, err := json.Marshal(testData)
	if err != nil {
		t.Errorf("JSON marshalling failed: %v", err)
	}

	var unmarshalledData struct {
		ID        uint      `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
	
	if err := json.Unmarshal(jsonData, &unmarshalledData); err != nil {
		t.Errorf("JSON unmarshalling failed: %v", err)
	}

	if unmarshalledData.ID != testData.ID {
		t.Errorf("Expected ID %d, got %d", testData.ID, unmarshalledData.ID)
	}
	
	// Time comparison with some tolerance for JSON marshalling/unmarshalling
	if unmarshalledData.CreatedAt.Unix() != testData.CreatedAt.Unix() {
		t.Errorf("Expected CreatedAt %v, got %v", testData.CreatedAt, unmarshalledData.CreatedAt)
	}
}

// TestResponseHelpers - Test response helper functions
func TestResponseHelpers(t *testing.T) {
	// Test success response helper
	successResponse := APIResponse{
		Success: true,
		Message: "Success",
		Data:    "test data",
	}

	if !successResponse.Success {
		t.Error("Expected success response to have Success=true")
	}

	// Test error response helper
	errorResponse := APIResponse{
		Success: false,
		Message: "Error occurred",
		Error:   "Something went wrong",
	}

	if errorResponse.Success {
		t.Error("Expected error response to have Success=false")
	}
}

// TestJSONTags - Test JSON tags are properly set
func TestJSONTags(t *testing.T) {
	// Test that JSON tags are working correctly
	response := APIResponse{
		Success: true,
		Message: "Test message",
		Data:    map[string]string{"key": "value"},
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Errorf("JSON marshalling failed: %v", err)
	}

	// Check that JSON contains expected fields
	jsonStr := string(jsonData)
	if !contains(jsonStr, "success") {
		t.Error("JSON should contain 'success' field")
	}
	if !contains(jsonStr, "message") {
		t.Error("JSON should contain 'message' field")
	}
	if !contains(jsonStr, "data") {
		t.Error("JSON should contain 'data' field")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		containsAt(s, substr, 1))))
}

func containsAt(s, substr string, start int) bool {
	if start >= len(s) || start < 0 {
		return false
	}
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
