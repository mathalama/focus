package config

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultBackendURL = "http://localhost:8080"
	DefaultAppURL     = "http://localhost:5173"
	DefaultPollWait   = 30 * time.Second
	DefaultAPIAddr    = ":8091"
	BotAuthHeader     = "X-Telegram-Bot-Auth"
)

type Config struct {
	Token           string
	BackendURL      string
	AppURL          string
	BotAuth         string
	PollTimeout     time.Duration
	InternalAPIAddr string
}

func Load() (Config, error) {
	if err := loadDotEnvIfPresent(".env"); err != nil {
		return Config{}, err
	}

	cfg := Config{
		Token:           strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN")),
		BackendURL:      strings.TrimRight(strings.TrimSpace(envOrDefault("BACKEND_URL", DefaultBackendURL)), "/"),
		AppURL:          strings.TrimRight(strings.TrimSpace(envOrDefault("APP_URL", DefaultAppURL)), "/"),
		BotAuth:         strings.TrimSpace(os.Getenv("TELEGRAM_BOT_AUTH_TOKEN")),
		PollTimeout:     parseDurationSeconds(os.Getenv("TELEGRAM_POLL_TIMEOUT_SECONDS"), DefaultPollWait),
		InternalAPIAddr: strings.TrimSpace(envOrDefault("BOT_INTERNAL_API_ADDR", DefaultAPIAddr)),
	}

	if cfg.Token == "" {
		return Config{}, errors.New("TELEGRAM_BOT_TOKEN is required")
	}
	if cfg.BotAuth == "" {
		return Config{}, errors.New("TELEGRAM_BOT_AUTH_TOKEN is required")
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func loadDotEnvIfPresent(path string) error {
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}

		key, rawValue, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}

		if _, exists := os.LookupEnv(key); exists {
			continue
		}

		value := normalizeEnvValue(rawValue)
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}

	return scanner.Err()
}

func normalizeEnvValue(raw string) string {
	value := strings.TrimSpace(raw)
	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
			return value[1 : len(value)-1]
		}
	}
	return value
}

func parseDurationSeconds(raw string, fallback time.Duration) time.Duration {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback
	}
	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds <= 0 {
		return fallback
	}
	return time.Duration(seconds) * time.Second
}
