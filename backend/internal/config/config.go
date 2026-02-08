package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                       string
	DatabaseURL                string
	CorsOrigin                 string
	JWTSecret                  string
	MaxSessionPauses           int
	TelegramBotAuthToken       string
	TelegramLinkCodeTTLMinutes int
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		Port:                       envOrDefault("PORT", "8080"),
		DatabaseURL:                os.Getenv("DATABASE_URL"),
		CorsOrigin:                 envOrDefault("CORS_ORIGIN", "http://localhost:5173"),
		JWTSecret:                  envOrDefault("JWT_SECRET", "dev-secret-change-me"),
		MaxSessionPauses:           intOrDefault("MAX_SESSION_PAUSES", 3),
		TelegramBotAuthToken:       envOrDefault("TELEGRAM_BOT_AUTH_TOKEN", "dev-telegram-bot-auth-change-me"),
		TelegramLinkCodeTTLMinutes: intOrDefault("TELEGRAM_LINK_CODE_TTL_MINUTES", 10),
	}

	if cfg.TelegramLinkCodeTTLMinutes <= 0 {
		cfg.TelegramLinkCodeTTLMinutes = 10
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func intOrDefault(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}
