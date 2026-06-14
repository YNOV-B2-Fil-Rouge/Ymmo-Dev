// Package config loads and validates application configuration from
// environment variables (12-factor app). Centralizing this here keeps
// secrets out of the code (DRY) and makes the app easy to deploy.
package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds every runtime setting the application needs.
type Config struct {
	AppEnv     string // "development" | "production"
	AppPort    string // HTTP port the API listens on
	JWTSecret  string // signing key for JWT (used by the auth module later)
	AIBaseURL  string // internal URL of the Python AI service (proxied)
	UploadDir  string // on-disk directory where uploaded photos are stored
	PublicURL  string // public base URL of this API (to build photo URLs)

	DB DBConfig
}

// DBConfig holds the MariaDB connection settings.
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

// DSN builds the MariaDB/MySQL connection string used by GORM.
func (d DBConfig) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.Name,
	)
}

// Load reads the .env file (if present) then the environment, applies
// sane defaults, and fails fast if a required secret is missing.
func Load() (*Config, error) {
	// .env is optional: in production, variables come from the environment.
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:    getEnv("APP_ENV", "development"),
		AppPort:   getEnv("APP_PORT", "8080"),
		JWTSecret: os.Getenv("JWT_SECRET"),
		AIBaseURL: getEnv("AI_BASE_URL", "http://ai:8000"),
		UploadDir: getEnv("UPLOAD_DIR", "/app/uploads"),
		PublicURL: getEnv("PUBLIC_API_URL", "http://localhost:8080"),
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "127.0.0.1"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     getEnv("DB_NAME", "ymmo"),
		},
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}

// getEnv returns the variable's value or a fallback when it is unset.
func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
