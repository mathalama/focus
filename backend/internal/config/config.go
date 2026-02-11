package config

import (
	"fmt"
	"os"
	"regexp"
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
	RefreshSessionTTLHours     int
	MaxActiveAuthSessions      int
	AuthSessionBindClient      bool
	MaxSessionPauses           int
	TelegramBotAuthToken       string
	TelegramLinkCodeTTLMinutes int
	ResendAPIKey               string
	ResendFromEmail            string
	EmailOutboxPollSeconds     int
	EmailOutboxMaxAttempts     int
	EmailVerifyURLBase         string
	EmailVerifySuccessRedirect string
	EmailVerifyFailRedirect    string
	EmailVerificationTTLMin    int
	EmailResetPasswordURLBase  string
}

var postgresURLPattern = regexp.MustCompile(`(?i)postgres(?:ql)?://[^\s'"]+`)

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		Port:                       envOrDefault("PORT", "8080"),
		DatabaseURL:                normalizeDatabaseURL(os.Getenv("DATABASE_URL")),
		CorsOrigin:                 envOrDefault("CORS_ORIGIN", "http://localhost:5173"),
		JWTSecret:                  envOrDefault("JWT_SECRET", "dev-secret-change-me"),
		EnableDevLogin:             boolOrDefault("ENABLE_DEV_LOGIN", true),
		RefreshSessionTTLHours:     intOrDefault("REFRESH_SESSION_TTL_HOURS", 24*30),
		MaxActiveAuthSessions:      intOrDefault("MAX_ACTIVE_AUTH_SESSIONS", 5),
		AuthSessionBindClient:      boolOrDefault("AUTH_SESSION_BIND_CLIENT", true),
		MaxSessionPauses:           intOrDefault("MAX_SESSION_PAUSES", 3),
		TelegramBotAuthToken:       envOrDefault("TELEGRAM_BOT_AUTH_TOKEN", "dev-telegram-bot-auth-change-me"),
		TelegramLinkCodeTTLMinutes: intOrDefault("TELEGRAM_LINK_CODE_TTL_MINUTES", 10),
		ResendAPIKey:               strings.TrimSpace(os.Getenv("RESEND_API_KEY")),
		ResendFromEmail:            strings.TrimSpace(os.Getenv("RESEND_FROM_EMAIL")),
		EmailOutboxPollSeconds:     intOrDefault("EMAIL_OUTBOX_POLL_SECONDS", 2),
		EmailOutboxMaxAttempts:     intOrDefault("EMAIL_OUTBOX_MAX_ATTEMPTS", 5),
		EmailVerifyURLBase:         strings.TrimSpace(envOrDefault("EMAIL_VERIFY_URL_BASE", "http://localhost:8080/api/v1/auth/verify-email")),
		EmailVerifySuccessRedirect: strings.TrimSpace(envOrDefault("EMAIL_VERIFY_SUCCESS_REDIRECT", "http://localhost:5173/login?verified=1")),
		EmailVerifyFailRedirect:    strings.TrimSpace(envOrDefault("EMAIL_VERIFY_FAIL_REDIRECT", "http://localhost:5173/login?verified=0")),
		EmailVerificationTTLMin:    intOrDefault("EMAIL_VERIFICATION_TTL_MINUTES", 60),
		EmailResetPasswordURLBase:  strings.TrimSpace(envOrDefault("EMAIL_RESET_PASSWORD_URL_BASE", "http://localhost:5173/reset-password")),
	}

	if cfg.TelegramLinkCodeTTLMinutes <= 0 {
		cfg.TelegramLinkCodeTTLMinutes = 10
	}
	if cfg.EmailVerificationTTLMin <= 0 {
		cfg.EmailVerificationTTLMin = 60
	}
	if cfg.EmailOutboxPollSeconds <= 0 {
		cfg.EmailOutboxPollSeconds = 2
	}
	if cfg.EmailOutboxMaxAttempts <= 0 {
		cfg.EmailOutboxMaxAttempts = 5
	}
	if cfg.RefreshSessionTTLHours <= 0 {
		cfg.RefreshSessionTTLHours = 24 * 30
	}
	if cfg.MaxActiveAuthSessions <= 0 {
		cfg.MaxActiveAuthSessions = 5
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

// normalizeDatabaseURL accepts plain PostgreSQL URLs and pasted psql commands.
func normalizeDatabaseURL(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}

	if match := postgresURLPattern.FindString(value); match != "" {
		return strings.TrimRight(strings.TrimSpace(match), ";")
	}

	return strings.Trim(value, `"'`)
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
