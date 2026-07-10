package config_test

import (
	"strings"
	"testing"
	"time"

	"github.com/Yasin4261/food-delivery/config"
)

// setBase clears every config variable (t.Setenv restores them afterwards) and
// sets the minimal valid environment; individual tests override from there.
func setBase(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		"PORT", "ENV", "DATABASE_URL", "JWT_SECRET", "JWT_EXPIRATION", "AUTO_MIGRATE",
		"ALLOWED_ORIGINS", "SMTP_HOST", "SMTP_PORT", "SMTP_USERNAME", "SMTP_PASSWORD",
		"MAIL_FROM", "APP_BASE_URL", "IYZICO_API_KEY", "IYZICO_SECRET_KEY", "IYZICO_BASE_URL",
	} {
		t.Setenv(k, "")
	}
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost:5432/db?sslmode=disable")
	t.Setenv("JWT_SECRET", "a-sufficiently-strong-test-secret")
}

func TestLoadConfig_Defaults(t *testing.T) {
	setBase(t)

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Port != "8080" || cfg.Env != "development" {
		t.Errorf("port/env = %q/%q, want 8080/development", cfg.Port, cfg.Env)
	}
	if cfg.JWTExpiration != 24*time.Hour {
		t.Errorf("jwt expiration = %v, want 24h", cfg.JWTExpiration)
	}
	if cfg.AutoMigrate {
		t.Error("auto migrate should default to false")
	}
	if len(cfg.AllowedOrigins) != 0 {
		t.Errorf("allowed origins = %v, want empty", cfg.AllowedOrigins)
	}
	if cfg.SMTPPort != "587" {
		t.Errorf("smtp port default = %q, want 587", cfg.SMTPPort)
	}
}

func TestLoadConfig_RequiresDatabaseURL(t *testing.T) {
	setBase(t)
	t.Setenv("DATABASE_URL", "")

	if _, err := config.LoadConfig(); err == nil || !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Errorf("err = %v, want DATABASE_URL required", err)
	}
}

func TestLoadConfig_RequiresJWTSecret(t *testing.T) {
	setBase(t)
	t.Setenv("JWT_SECRET", "")

	if _, err := config.LoadConfig(); err == nil || !strings.Contains(err.Error(), "JWT_SECRET") {
		t.Errorf("err = %v, want JWT_SECRET required", err)
	}
}

// setProductionExtras satisfies the non-dev requirements (mail + payments) so
// tests can isolate a single validation rule.
func setProductionExtras(t *testing.T) {
	t.Helper()
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("MAIL_FROM", "no-reply@example.com")
	t.Setenv("IYZICO_API_KEY", "iyzi-key")
	t.Setenv("IYZICO_SECRET_KEY", "iyzi-secret")
}

// TestLoadConfig_PlaceholderSecret is the OWASP A05 fail-fast: the known weak
// placeholder must be rejected everywhere except development.
func TestLoadConfig_PlaceholderSecret(t *testing.T) {
	setBase(t)
	setProductionExtras(t)
	t.Setenv("JWT_SECRET", "change-me-in-production")

	t.Setenv("ENV", "development")
	if _, err := config.LoadConfig(); err != nil {
		t.Errorf("placeholder in development should load, got %v", err)
	}

	for _, env := range []string{"production", "staging"} {
		t.Setenv("ENV", env)
		if _, err := config.LoadConfig(); err == nil || !strings.Contains(err.Error(), "placeholder") {
			t.Errorf("env=%s: err = %v, want placeholder rejection", env, err)
		}
	}
}

// TestLoadConfig_MailValidation: outside development SMTP_HOST and MAIL_FROM
// are required (the dev logging mailer is development-only).
func TestLoadConfig_MailValidation(t *testing.T) {
	setBase(t)
	t.Setenv("ENV", "production")
	// Payments configured so only the mail rules are under test.
	t.Setenv("IYZICO_API_KEY", "iyzi-key")
	t.Setenv("IYZICO_SECRET_KEY", "iyzi-secret")

	// No SMTP host -> refuse to boot.
	if _, err := config.LoadConfig(); err == nil || !strings.Contains(err.Error(), "SMTP_HOST") {
		t.Errorf("err = %v, want SMTP_HOST required in production", err)
	}

	// Host without a From address -> refuse.
	t.Setenv("SMTP_HOST", "smtp.example.com")
	if _, err := config.LoadConfig(); err == nil || !strings.Contains(err.Error(), "MAIL_FROM") {
		t.Errorf("err = %v, want MAIL_FROM required", err)
	}

	// Fully configured -> loads.
	t.Setenv("MAIL_FROM", "no-reply@example.com")
	if _, err := config.LoadConfig(); err != nil {
		t.Errorf("fully configured production mail should load, got %v", err)
	}

	// In development an empty SMTP_HOST is fine (logging mailer).
	t.Setenv("ENV", "development")
	t.Setenv("SMTP_HOST", "")
	t.Setenv("MAIL_FROM", "")
	if _, err := config.LoadConfig(); err != nil {
		t.Errorf("development without SMTP should load, got %v", err)
	}
}

// TestLoadConfig_PaymentValidation: outside development the iyzico credentials
// are required (the dev mock gateway is development-only).
func TestLoadConfig_PaymentValidation(t *testing.T) {
	setBase(t)
	t.Setenv("ENV", "production")
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("MAIL_FROM", "no-reply@example.com")

	if _, err := config.LoadConfig(); err == nil || !strings.Contains(err.Error(), "IYZICO") {
		t.Errorf("err = %v, want iyzico credentials required in production", err)
	}

	t.Setenv("IYZICO_API_KEY", "iyzi-key")
	t.Setenv("IYZICO_SECRET_KEY", "iyzi-secret")
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("fully configured production should load, got %v", err)
	}
	if cfg.IyzicoBaseURL != "https://sandbox-api.iyzipay.com" {
		t.Errorf("iyzico base url default = %q, want the sandbox", cfg.IyzicoBaseURL)
	}

	// Development runs happily on the mock gateway.
	t.Setenv("ENV", "development")
	t.Setenv("IYZICO_API_KEY", "")
	t.Setenv("IYZICO_SECRET_KEY", "")
	if _, err := config.LoadConfig(); err != nil {
		t.Errorf("development without iyzico should load, got %v", err)
	}
}

func TestLoadConfig_AllowedOriginsParsing(t *testing.T) {
	setBase(t)
	t.Setenv("ALLOWED_ORIGINS", " https://app.example.com, http://localhost:5173 ,,https://x.dev ")

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	want := []string{"https://app.example.com", "http://localhost:5173", "https://x.dev"}
	if len(cfg.AllowedOrigins) != len(want) {
		t.Fatalf("origins = %v, want %v", cfg.AllowedOrigins, want)
	}
	for i, o := range want {
		if cfg.AllowedOrigins[i] != o {
			t.Errorf("origin[%d] = %q, want %q (must be trimmed, blanks dropped)", i, cfg.AllowedOrigins[i], o)
		}
	}
}

func TestLoadConfig_InvalidJWTExpiration(t *testing.T) {
	setBase(t)
	t.Setenv("JWT_EXPIRATION", "not-a-duration")

	if _, err := config.LoadConfig(); err == nil || !strings.Contains(err.Error(), "JWT_EXPIRATION") {
		t.Errorf("err = %v, want invalid JWT_EXPIRATION", err)
	}

	t.Setenv("JWT_EXPIRATION", "90m")
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.JWTExpiration != 90*time.Minute {
		t.Errorf("expiration = %v, want 90m", cfg.JWTExpiration)
	}
}
