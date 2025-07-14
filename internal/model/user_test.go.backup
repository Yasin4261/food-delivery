package model

import (
	"encoding/json"
	"testing"
	"time"
)

// TestUser - User model testleri
func TestUser(t *testing.T) {
	// Test data
	now := time.Now()
	user := User{
		ID:        1,
		Email:     "test@example.com",
		Password:  "hashedpassword123",
		FirstName: "John",
		LastName:  "Doe",
		Role:      "customer",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Test JSON marshalling
	jsonData, err := json.Marshal(user)
	if err != nil {
		t.Errorf("JSON marshalling failed: %v", err)
	}

	// Password field should not be included in JSON
	var jsonMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &jsonMap); err != nil {
		t.Errorf("JSON unmarshalling failed: %v", err)
	}

	if _, exists := jsonMap["password"]; exists {
		t.Error("Password field should not be included in JSON output")
	}

	// Test struct unmarshalling
	var unmarshalledUser User
	if err := json.Unmarshal(jsonData, &unmarshalledUser); err != nil {
		t.Errorf("JSON unmarshalling to struct failed: %v", err)
	}

	if unmarshalledUser.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, unmarshalledUser.Email)
	}
}

// TestLoginRequest - Login request validation test
func TestLoginRequest(t *testing.T) {
	tests := []struct {
		name    string
		request LoginRequest
		isValid bool
	}{
		{
			name: "Valid login request",
			request: LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			isValid: true,
		},
		{
			name: "Invalid email",
			request: LoginRequest{
				Email:    "invalid-email",
				Password: "password123",
			},
			isValid: false,
		},
		{
			name: "Short password",
			request: LoginRequest{
				Email:    "test@example.com",
				Password: "123",
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON marshalling
			jsonData, err := json.Marshal(tt.request)
			if err != nil {
				t.Errorf("JSON marshalling failed: %v", err)
			}

			var unmarshalledRequest LoginRequest
			if err := json.Unmarshal(jsonData, &unmarshalledRequest); err != nil {
				t.Errorf("JSON unmarshalling failed: %v", err)
			}

			if unmarshalledRequest.Email != tt.request.Email {
				t.Errorf("Expected email %s, got %s", tt.request.Email, unmarshalledRequest.Email)
			}
		})
	}
}

// TestRegisterRequest - Register request test
func TestRegisterRequest(t *testing.T) {
	validRequest := RegisterRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	}

	jsonData, err := json.Marshal(validRequest)
	if err != nil {
		t.Errorf("JSON marshalling failed: %v", err)
	}

	var unmarshalledRequest RegisterRequest
	if err := json.Unmarshal(jsonData, &unmarshalledRequest); err != nil {
		t.Errorf("JSON unmarshalling failed: %v", err)
	}

	if unmarshalledRequest.Email != validRequest.Email {
		t.Errorf("Expected email %s, got %s", validRequest.Email, unmarshalledRequest.Email)
	}
	if unmarshalledRequest.FirstName != validRequest.FirstName {
		t.Errorf("Expected first name %s, got %s", validRequest.FirstName, unmarshalledRequest.FirstName)
	}
}

// TestAuthResponse - Auth response test
func TestAuthResponse(t *testing.T) {
	user := User{
		ID:        1,
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Role:      "customer",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	authResponse := AuthResponse{
		Token: "jwt-token-123",
		User:  user,
	}

	jsonData, err := json.Marshal(authResponse)
	if err != nil {
		t.Errorf("JSON marshalling failed: %v", err)
	}

	var unmarshalledResponse AuthResponse
	if err := json.Unmarshal(jsonData, &unmarshalledResponse); err != nil {
		t.Errorf("JSON unmarshalling failed: %v", err)
	}

	if unmarshalledResponse.Token != authResponse.Token {
		t.Errorf("Expected token %s, got %s", authResponse.Token, unmarshalledResponse.Token)
	}
	if unmarshalledResponse.User.Email != authResponse.User.Email {
		t.Errorf("Expected user email %s, got %s", authResponse.User.Email, unmarshalledResponse.User.Email)
	}
}

// TestUserProfileResponse - User profile response test
func TestUserProfileResponse(t *testing.T) {
	profileResponse := UserProfileResponse{
		ID:        1,
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Role:      "customer",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	jsonData, err := json.Marshal(profileResponse)
	if err != nil {
		t.Errorf("JSON marshalling failed: %v", err)
	}

	var unmarshalledResponse UserProfileResponse
	if err := json.Unmarshal(jsonData, &unmarshalledResponse); err != nil {
		t.Errorf("JSON unmarshalling failed: %v", err)
	}

	if unmarshalledResponse.Email != profileResponse.Email {
		t.Errorf("Expected email %s, got %s", profileResponse.Email, unmarshalledResponse.Email)
	}
}
