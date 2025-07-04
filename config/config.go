package config

import (
	"fmt"
	"os"
	"gopkg.in/yaml.v3"
)

// Config yapısı - uygulama konfigürasyonu
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
	Logging  LoggingConfig  `yaml:"logging"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
	Host string `yaml:"host"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	SSLMode  string `yaml:"ssl_mode"`
}

type JWTConfig struct {
	Secret          string `yaml:"secret"`
	ExpirationHours int    `yaml:"expiration_hours"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// Load config dosyasını yükler
func Load() (*Config, error) {
	configPath := "config/config.yaml"
	
	// Docker ortamında farklı config kullan
	if _, err := os.Stat("config/config.docker.yaml"); err == nil {
		configPath = "config/config.docker.yaml"
	}
	
	// Dosya var mı kontrol et
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config dosyası bulunamadı: %s", configPath)
	}

	// Dosyayı oku
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("config dosyası okunamadı: %v", err)
	}

	// YAML'ı parse et
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("config dosyası parse edilemedi: %v", err)
	}

	// Environment variable'lardan override et
	config.overrideFromEnv()

	return &config, nil
}

// overrideFromEnv environment variable'lardan config'i override eder
func (c *Config) overrideFromEnv() {
	if host := os.Getenv("DB_HOST"); host != "" {
		c.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &c.Database.Port)
	}
	if user := os.Getenv("DB_USER"); user != "" {
		c.Database.Username = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		c.Database.Password = password
	}
	if dbname := os.Getenv("DB_NAME"); dbname != "" {
		c.Database.Database = dbname
	}
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		c.JWT.Secret = jwtSecret
	}
}

// GetDatabaseURL database connection string'i döndürür
func (c *Config) GetDatabaseURL() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.Username,
		c.Database.Password,
		c.Database.Database,
		c.Database.SSLMode,
	)
}
