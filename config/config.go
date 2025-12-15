package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	// Server
	Port string
	Env  string

	// Database
	DatabaseURL string

	// JWT
	JWTSecret     string
	JWTExpiration time.Duration

	// CORS
	AllowedOrigins []string

	// Migration
	AutoMigrate bool
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists (development)
	if err := godotenv.Load(); err != nil {
		// .env file not found is okay (production might use real env vars)
		fmt.Println("No .env file found, using environment variables")
	}

	// Parse JWT expiration
	jwtExpStr := getEnv("JWT_EXPIRATION", "24h")
	jwtExpiration, err := time.ParseDuration(jwtExpStr)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRATION format: %w", err)
	}

	config := &Config{
		Port:           getEnv("PORT", "8080"),
		Env:            getEnv("ENV", "development"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://postgres:postgres123@localhost:5432/food_delivery?sslmode=disable"),
		JWTSecret:      getEnv("JWT_SECRET", ""),
		JWTExpiration:  jwtExpiration,
		AllowedOrigins: parseCommaSeparated(getEnv("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:8080")),
		AutoMigrate:    getEnvBool("AUTO_MIGRATE", true),
	}

	// Validate required fields
	if config.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET environment variable is required")
	}

	return config, nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool retrieves a boolean environment variable
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return defaultValue
		}
		return boolVal
	}
	return defaultValue
}

// parseCommaSeparated splits a comma-separated string into a slice
func parseCommaSeparated(s string) []string {
	if s == "" {
		return []string{}
	}
	
	var result []string
	for _, item := range splitByComma(s) {
		if trimmed := trimSpace(item); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// splitByComma splits a string by comma
func splitByComma(s string) []string {
	var result []string
	var current string
	
	for _, ch := range s {
		if ch == ',' {
			result = append(result, current)
			current = ""
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	
	return result
}

// trimSpace removes leading and trailing spaces
func trimSpace(s string) string {
	start := 0
	end := len(s)
	
	for start < end && s[start] == ' ' {
		start++
	}
	for end > start && s[end-1] == ' ' {
		end--
	}
	
	return s[start:end]
}
