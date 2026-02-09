package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                       string
	DatabaseURL                string
	CorsOrigin                 string
	JWTSecret                  string
	EnableDevLogin             bool
	MaxSessionPauses           int
	TelegramBotAuthToken       string
	TelegramLinkCodeTTLMinutes int
	ResendAPIKey               string
	ResendFromEmail            string
	EmailVerifyURLBase         string
	EmailVerifySuccessRedirect string
	EmailVerifyFailRedirect    string
	EmailVerificationTTLMin    int
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		Port:                       envOrDefault("PORT", "8080"),
		DatabaseURL:                os.Getenv("DATABASE_URL"),
		CorsOrigin:                 envOrDefault("CORS_ORIGIN", "http://localhost:5173"),
		JWTSecret:                  envOrDefault("JWT_SECRET", "dev-secret-change-me"),
		EnableDevLogin:             boolOrDefault("ENABLE_DEV_LOGIN", true),
		MaxSessionPauses:           intOrDefault("MAX_SESSION_PAUSES", 3),
		TelegramBotAuthToken:       envOrDefault("TELEGRAM_BOT_AUTH_TOKEN", "dev-telegram-bot-auth-change-me"),
		TelegramLinkCodeTTLMinutes: intOrDefault("TELEGRAM_LINK_CODE_TTL_MINUTES", 10),
		ResendAPIKey:               strings.TrimSpace(os.Getenv("RESEND_API_KEY")),
		ResendFromEmail:            strings.TrimSpace(os.Getenv("RESEND_FROM_EMAIL")),
		EmailVerifyURLBase:         strings.TrimSpace(envOrDefault("EMAIL_VERIFY_URL_BASE", "http://localhost:8080/api/v1/auth/verify-email")),
		EmailVerifySuccessRedirect: strings.TrimSpace(envOrDefault("EMAIL_VERIFY_SUCCESS_REDIRECT", "http://localhost:5173/login?verified=1")),
		EmailVerifyFailRedirect:    strings.TrimSpace(envOrDefault("EMAIL_VERIFY_FAIL_REDIRECT", "http://localhost:5173/login?verified=0")),
		EmailVerificationTTLMin:    intOrDefault("EMAIL_VERIFICATION_TTL_MINUTES", 60),
	}

	if cfg.TelegramLinkCodeTTLMinutes <= 0 {
		cfg.TelegramLinkCodeTTLMinutes = 10
	}
	if cfg.EmailVerificationTTLMin <= 0 {
		cfg.EmailVerificationTTLMin = 60
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

func boolOrDefault(key string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	switch value {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}
