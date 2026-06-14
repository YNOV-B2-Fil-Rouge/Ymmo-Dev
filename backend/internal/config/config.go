// Package config loads and validates configuration from environment variables.
package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv     string
	AppPort    string
	JWTSecret  string
	AIBaseURL  string
	UploadDir  string
	PublicURL  string

	DB DBConfig
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

func (d DBConfig) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.Name,
	)
}

func Load() (*Config, error) {
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

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
