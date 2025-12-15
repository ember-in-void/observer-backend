// internal/shared/config/config.go
package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type JWTConfig struct {
	Secret string
	TTL    time.Duration
}

type Config struct {
	HTTPAddr string
	Google   GoogleOAuthConfig
	Database string
	JWT      JWTConfig
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		HTTPAddr: getEnv("HTTP_ADDR", ":8080"),
		Google: GoogleOAuthConfig{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		},
		Database: os.Getenv("DATABASE_URL"),
		JWT: JWTConfig{
			Secret: os.Getenv("JWT_SECRET"),
			TTL:    time.Duration(getEnvAsInt("JWT_TTL_SECONDS", 3600)) * time.Second,
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(name string, defaultVal int) int {
	if valueStr := os.Getenv(name); valueStr != "" {
		var value int
		_, err := fmt.Sscanf(valueStr, "%d", &value)
		if err == nil {
			return value
		}
	}
	return defaultVal
}
