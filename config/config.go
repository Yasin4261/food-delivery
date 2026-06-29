package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration, loaded from environment variables.
type Config struct {
	Port           string
	Env            string
	DatabaseURL    string
	JWTSecret      string
	JWTExpiration  time.Duration
	AutoMigrate    bool
	AllowedOrigins []string
}

// LoadConfig reads configuration from the environment (and a local .env file
// if present). In Docker the variables come from the environment directly, so
// a missing .env file is not an error.
func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port:           getEnv("PORT", "8080"),
		Env:            getEnv("ENV", "development"),
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		JWTSecret:      getEnv("JWT_SECRET", ""),
		AutoMigrate:    getEnv("AUTO_MIGRATE", "false") == "true",
		AllowedOrigins: splitAndTrim(getEnv("ALLOWED_ORIGINS", "")),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	if err := validateJWTSecret(cfg); err != nil {
		return nil, err
	}

	exp, err := time.ParseDuration(getEnv("JWT_EXPIRATION", "24h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRATION: %w", err)
	}
	cfg.JWTExpiration = exp

	return cfg, nil
}

// placeholderJWTSecret is the value historically committed to compose files. It
// must never be used outside development.
const placeholderJWTSecret = "change-me-in-production"

// validateJWTSecret fails fast on a missing secret, and on the known weak
// placeholder in any non-development environment.
func validateJWTSecret(cfg *Config) error {
	if cfg.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if cfg.Env != "development" && cfg.JWTSecret == placeholderJWTSecret {
		return fmt.Errorf("JWT_SECRET is set to the insecure placeholder; set a strong secret in env=%q", cfg.Env)
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

// splitAndTrim parses a comma-separated list, dropping blanks.
func splitAndTrim(s string) []string {
	out := make([]string, 0)
	for _, part := range strings.Split(s, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}
