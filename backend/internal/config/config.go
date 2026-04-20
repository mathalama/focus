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

	appURL := strings.TrimRight(envOrDefault("APP_URL", "http://localhost:5173"), "/")
	backendURL := strings.TrimRight(envOrDefault("BACKEND_URL", "http://localhost:8080"), "/")

	cfg := Config{
		Port:                       envOrDefault("PORT", "8080"),
		DatabaseURL:                normalizeDatabaseURL(os.Getenv("DATABASE_URL")),
		CorsOrigin:                 envOrDefault("CORS_ORIGIN", appURL),
		JWTSecret:                  envOrDefault("JWT_SECRET", "dev-secret-change-me"),
		EnableDevLogin:             boolOrDefault("ENABLE_DEV_LOGIN", false),
		RefreshSessionTTLHours:     intOrDefault("REFRESH_SESSION_TTL_HOURS", 720),
		MaxActiveAuthSessions:      intOrDefault("MAX_ACTIVE_AUTH_SESSIONS", 5),
		AuthSessionBindClient:      boolOrDefault("AUTH_SESSION_BIND_CLIENT", true),
		MaxSessionPauses:           intOrDefault("MAX_SESSION_PAUSES", 3),
		ResendAPIKey:               strings.TrimSpace(os.Getenv("RESEND_API_KEY")),
		ResendFromEmail:            strings.TrimSpace(os.Getenv("RESEND_FROM_EMAIL")),
		EmailOutboxPollSeconds:     intOrDefault("EMAIL_OUTBOX_POLL_SECONDS", 2),
		EmailOutboxMaxAttempts:     intOrDefault("EMAIL_OUTBOX_MAX_ATTEMPTS", 5),
		EmailVerifyURLBase:         envOrDefault("EMAIL_VERIFY_URL_BASE", backendURL+"/api/v1/auth/verify-email"),
		EmailVerifySuccessRedirect: envOrDefault("EMAIL_VERIFY_SUCCESS_REDIRECT", appURL+"/login?verified=1"),
		EmailVerifyFailRedirect:    envOrDefault("EMAIL_VERIFY_FAIL_REDIRECT", appURL+"/login?verified=0"),
		EmailVerificationTTLMin:    intOrDefault("EMAIL_VERIFICATION_TTL_MINUTES", 60),
		EmailResetPasswordURLBase:  envOrDefault("EMAIL_RESET_PASSWORD_URL_BASE", appURL+"/reset-password"),
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
