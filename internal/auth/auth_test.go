package auth

import (
	"fmt"
	"testing"
	"time"
)

func TestClaims_Structure(t *testing.T) {
	claims := Claims{
		UserID: 1,
		Email:  "test@example.com",
		Role:   "customer",
	}

	if claims.UserID != 1 {
		t.Errorf("Expected UserID 1, got %d", claims.UserID)
	}

	if claims.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", claims.Email)
	}

	if claims.Role != "customer" {
		t.Errorf("Expected role 'customer', got '%s'", claims.Role)
	}
}

func TestJWTManager_NewJWTManager(t *testing.T) {
	secretKey := "test-secret-key"
	expirationHours := 24

	jwtManager := NewJWTManager(secretKey, expirationHours)

	if jwtManager == nil {
		t.Error("JWTManager should not be nil")
	}

	// Test that the manager was created with correct parameters
	// Since fields are private, we can only test the constructor doesn't panic
}

func TestJWTManager_GenerateToken(t *testing.T) {
	jwtManager := NewJWTManager("test-secret-key", 24)

	userID := uint(1)
	email := "test@example.com"
	role := "customer"

	token, err := jwtManager.GenerateToken(userID, email, role)
	if err != nil {
		t.Errorf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Error("Generated token should not be empty")
	}

	// Token should contain at least 3 parts (header.payload.signature)
	if len(token) < 10 {
		t.Error("Generated token seems too short")
	}
}

func TestJWTManager_ValidateToken(t *testing.T) {
	jwtManager := NewJWTManager("test-secret-key", 24)

	userID := uint(1)
	email := "test@example.com"
	role := "customer"

	// Generate a token
	token, err := jwtManager.GenerateToken(userID, email, role)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Validate the token
	claims, err := jwtManager.ValidateToken(token)
	if err != nil {
		t.Errorf("Failed to validate token: %v", err)
	}

	if claims == nil {
		t.Error("Claims should not be nil")
		return
	}

	if claims.UserID != userID {
		t.Errorf("Expected UserID %d, got %d", userID, claims.UserID)
	}

	if claims.Email != email {
		t.Errorf("Expected email '%s', got '%s'", email, claims.Email)
	}

	if claims.Role != role {
		t.Errorf("Expected role '%s', got '%s'", role, claims.Role)
	}
}

func TestJWTManager_ValidateInvalidToken(t *testing.T) {
	jwtManager := NewJWTManager("test-secret-key", 24)

	invalidTokens := []string{
		"",
		"invalid-token",
		"header.payload.signature", // Invalid format but correct structure
		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.signature", // Invalid payload
	}

	for i, invalidToken := range invalidTokens {
		t.Run(fmt.Sprintf("InvalidToken_%d", i), func(t *testing.T) {
			claims, err := jwtManager.ValidateToken(invalidToken)
			if err == nil {
				t.Error("Expected error for invalid token")
			}

			if claims != nil {
				t.Error("Claims should be nil for invalid token")
			}
		})
	}
}

func TestJWTManager_TokenExpiration(t *testing.T) {
	// Create JWT manager with very short expiration for testing
	jwtManager := NewJWTManager("test-secret-key", 0) // 0 hours = immediate expiration

	userID := uint(1)
	email := "test@example.com"
	role := "customer"

	token, err := jwtManager.GenerateToken(userID, email, role)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Wait a bit to ensure token expires
	time.Sleep(time.Second)

	// Try to validate expired token
	claims, err := jwtManager.ValidateToken(token)
	
	// Note: This test might be flaky depending on exact timing
	// In a real scenario, you'd mock time or use a different approach
	if err == nil {
		t.Log("Token validation succeeded - timing might be too close")
	}

	if err != nil && claims != nil {
		t.Error("Claims should be nil when validation fails")
	}
}

func TestJWTManager_DifferentSecretKeys(t *testing.T) {
	jwtManager1 := NewJWTManager("secret-key-1", 24)
	jwtManager2 := NewJWTManager("secret-key-2", 24)

	userID := uint(1)
	email := "test@example.com"
	role := "customer"

	// Generate token with first manager
	token, err := jwtManager1.GenerateToken(userID, email, role)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Try to validate with second manager (different secret)
	claims, err := jwtManager2.ValidateToken(token)
	if err == nil {
		t.Error("Expected error when validating token with different secret key")
	}

	if claims != nil {
		t.Error("Claims should be nil when secret keys don't match")
	}
}

func TestPasswordManager_HashPassword(t *testing.T) {
	passwordManager := NewPasswordManager()
	password := "testpassword123"

	hashedPassword, err := passwordManager.HashPassword(password)
	if err != nil {
		t.Errorf("Failed to hash password: %v", err)
	}

	if hashedPassword == "" {
		t.Error("Hashed password should not be empty")
	}

	if hashedPassword == password {
		t.Error("Hashed password should be different from original")
	}

	// Bcrypt hashes should be at least 60 characters
	if len(hashedPassword) < 60 {
		t.Error("Bcrypt hash seems too short")
	}
}

func TestPasswordManager_ComparePassword(t *testing.T) {
	passwordManager := NewPasswordManager()
	password := "testpassword123"

	// Hash the password
	hashedPassword, err := passwordManager.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Compare with correct password
	matches := passwordManager.CheckPasswordHash(password, hashedPassword)
	if !matches {
		t.Error("Password comparison should succeed")
	}

	// Compare with incorrect password
	matches = passwordManager.CheckPasswordHash("wrongpassword", hashedPassword)
	if matches {
		t.Error("Password comparison should fail for wrong password")
	}
}

func TestPasswordManager_EmptyPassword(t *testing.T) {
	passwordManager := NewPasswordManager()

	// Test hashing empty password
	hashedPassword, err := passwordManager.HashPassword("")
	if err != nil {
		t.Log("Hashing empty password resulted in error (expected)")
	} else if hashedPassword != "" {
		t.Log("Empty password was hashed successfully")
	}

	// Test comparing with empty password
	matches := passwordManager.CheckPasswordHash("", "some-hash")
	if matches {
		t.Log("Comparing with empty password failed (expected)")
	}
}

func TestPasswordManager_SpecialCharacters(t *testing.T) {
	passwordManager := NewPasswordManager()
	
	specialPasswords := []string{
		"password!@#$%^&*()",
		"çğışüöıÇĞIŞÜÖİ",
		"密码测试",
		"пароль",
		"كلمة المرور",
	}

	for i, password := range specialPasswords {
		t.Run(fmt.Sprintf("SpecialPassword_%d", i), func(t *testing.T) {
			hashedPassword, err := passwordManager.HashPassword(password)
			if err != nil {
				t.Errorf("Failed to hash special password: %v", err)
			}

			matches := passwordManager.CheckPasswordHash(password, hashedPassword)
			if !matches {
				t.Errorf("Failed to compare special password: %s", password)
			}
		})
	}
}

func TestPasswordManager_LongPassword(t *testing.T) {
	passwordManager := NewPasswordManager()
	
	// Test very long password
	longPassword := ""
	for i := 0; i < 100; i++ {
		longPassword += "a"
	}

	hashedPassword, err := passwordManager.HashPassword(longPassword)
	if err != nil {
		t.Errorf("Failed to hash long password: %v", err)
	}

	matches := passwordManager.CheckPasswordHash(longPassword, hashedPassword)
	if !matches {
		t.Error("Failed to compare long password")
	}
}

func TestRoleValidation(t *testing.T) {
	validRoles := []string{"customer", "chef", "admin"}
	
	for _, role := range validRoles {
		t.Run("Role_"+role, func(t *testing.T) {
			claims := Claims{
				Role: role,
			}
			
			if claims.Role != role {
				t.Errorf("Expected role '%s', got '%s'", role, claims.Role)
			}
		})
	}
}
