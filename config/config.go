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

	// Mail. When SMTPHost is empty the app uses the dev logging mailer.
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	MailFrom     string
	AppBaseURL   string // public base URL for links in emails and gateway callbacks

	// Payments (iyzico). When IyzicoAPIKey is empty the app uses the dev mock
	// gateway; outside development real credentials are required.
	IyzicoAPIKey    string
	IyzicoSecretKey string
	IyzicoBaseURL   string
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

		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		MailFrom:     getEnv("MAIL_FROM", ""),
		AppBaseURL:   getEnv("APP_BASE_URL", "http://localhost:8080"),

		IyzicoAPIKey:    getEnv("IYZICO_API_KEY", ""),
		IyzicoSecretKey: getEnv("IYZICO_SECRET_KEY", ""),
		IyzicoBaseURL:   getEnv("IYZICO_BASE_URL", "https://sandbox-api.iyzipay.com"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	if err := validateJWTSecret(cfg); err != nil {
		return nil, err
	}
	if err := validateMail(cfg); err != nil {
		return nil, err
	}
	if err := validatePayments(cfg); err != nil {
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

// validateMail fails fast on a misconfigured mailer outside development. In
// development an empty SMTP_HOST means "use the dev logging mailer"; in any
// other environment real SMTP delivery is required (host + from address).
func validateMail(cfg *Config) error {
	if cfg.Env == "development" {
		return nil
	}
	if cfg.SMTPHost == "" {
		return fmt.Errorf("SMTP_HOST is required in env=%q (no dev logging mailer outside development)", cfg.Env)
	}
	if cfg.MailFrom == "" {
		return fmt.Errorf("MAIL_FROM is required when SMTP is configured")
	}
	return nil
}

// validatePayments fails fast outside development when the payment gateway is
// not configured — the dev mock gateway must never run in production.
func validatePayments(cfg *Config) error {
	if cfg.Env == "development" {
		return nil
	}
	if cfg.IyzicoAPIKey == "" || cfg.IyzicoSecretKey == "" {
		return fmt.Errorf("IYZICO_API_KEY and IYZICO_SECRET_KEY are required in env=%q (no dev mock gateway outside development)", cfg.Env)
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
