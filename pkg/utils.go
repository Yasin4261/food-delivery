package pkg

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateRandomString rastgele string oluşturur
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// StringPtr string pointer döndürür
func StringPtr(s string) *string {
	return &s
}

// UintPtr uint pointer döndürür
func UintPtr(u uint) *uint {
	return &u
}

// BoolPtr bool pointer döndürür
func BoolPtr(b bool) *bool {
	return &b
}
