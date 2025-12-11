// internal/shared/config/config.go
package config

import (
	"os"

	"github.com/joho/godotenv"
)

type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type Config struct {
	HTTPAddr string
	Google   GoogleOAuthConfig
	Database string
}

func Load() *Config {
	// Загружаем .env файл (игнорируем ошибку если файла нет)
	_ = godotenv.Load()

	return &Config{
		HTTPAddr: getEnv("HTTP_ADDR", ":8080"),
		Google: GoogleOAuthConfig{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		},
		Database: os.Getenv("DATABASE_URL"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
