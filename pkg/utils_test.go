package pkg

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateRandomString(t *testing.T) {
	t.Run("Generate string with specified length", func(t *testing.T) {
		length := 16
		result, err := GenerateRandomString(length)
		require.NoError(t, err)
		
		// Hex encoding doubles the length
		assert.Equal(t, length*2, len(result))
		
		// Should be valid hex
		_, err = hex.DecodeString(result)
		assert.NoError(t, err)
	})

	t.Run("Generate different strings on multiple calls", func(t *testing.T) {
		result1, err1 := GenerateRandomString(8)
		result2, err2 := GenerateRandomString(8)
		
		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.NotEqual(t, result1, result2)
	})

	t.Run("Generate empty string for zero length", func(t *testing.T) {
		result, err := GenerateRandomString(0)
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})
}

func TestStringPtr(t *testing.T) {
	t.Run("Return pointer to string", func(t *testing.T) {
		testStr := "test string"
		ptr := StringPtr(testStr)
		
		assert.NotNil(t, ptr)
		assert.Equal(t, testStr, *ptr)
	})

	t.Run("Return pointer to empty string", func(t *testing.T) {
		ptr := StringPtr("")
		
		assert.NotNil(t, ptr)
		assert.Equal(t, "", *ptr)
	})
}

func TestUintPtr(t *testing.T) {
	t.Run("Return pointer to uint", func(t *testing.T) {
		testUint := uint(42)
		ptr := UintPtr(testUint)
		
		assert.NotNil(t, ptr)
		assert.Equal(t, testUint, *ptr)
	})

	t.Run("Return pointer to zero", func(t *testing.T) {
		ptr := UintPtr(0)
		
		assert.NotNil(t, ptr)
		assert.Equal(t, uint(0), *ptr)
	})
}

func TestBoolPtr(t *testing.T) {
	t.Run("Return pointer to true", func(t *testing.T) {
		ptr := BoolPtr(true)
		
		assert.NotNil(t, ptr)
		assert.Equal(t, true, *ptr)
	})

	t.Run("Return pointer to false", func(t *testing.T) {
		ptr := BoolPtr(false)
		
		assert.NotNil(t, ptr)
		assert.Equal(t, false, *ptr)
	})
}
