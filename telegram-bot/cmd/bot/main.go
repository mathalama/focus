package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	defaultBackendURL = "http://localhost:8080"
	defaultPollWait   = 30 * time.Second
)

type config struct {
	token       string
	backendURL  string
	botAuth     string
	pollTimeout time.Duration
}

type app struct {
	cfg    config
	client *http.Client
	apiURL string
}

type telegramAPIResponse[T any] struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
	Result      T      `json:"result"`
}

type telegramUpdate struct {
	UpdateID int              `json:"update_id"`
	Message  *telegramMessage `json:"message"`
}

type telegramMessage struct {
	Chat telegramChat `json:"chat"`
	From telegramUser `json:"from"`
	Text string       `json:"text"`
}

type telegramChat struct {
	ID int64 `json:"id"`
}

type telegramUser struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type sendMessageRequest struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}

type backendErrorResponse struct {
	Error string `json:"error"`
}

type backendTelegramLinkRequest struct {
	Code             string `json:"code"`
	TelegramUserID   int64  `json:"telegram_user_id"`
	TelegramUsername string `json:"telegram_username"`
	TelegramFirst    string `json:"telegram_first_name"`
	TelegramLast     string `json:"telegram_last_name"`
}

func main() {
	if err := loadDotEnvIfPresent(".env"); err != nil {
		log.Fatalf("config error: failed to load .env: %v", err)
	}

	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app := &app{
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.pollTimeout + 15*time.Second,
		},
		apiURL: fmt.Sprintf("https://api.telegram.org/bot%s", cfg.token),
	}

	log.Printf("telegram bot started (backend=%s)", cfg.backendURL)
	if err := app.run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("bot stopped with error: %v", err)
	}
}

func loadConfig() (config, error) {
	cfg := config{
		token:       strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN")),
		backendURL:  strings.TrimRight(strings.TrimSpace(envOrDefault("BACKEND_URL", defaultBackendURL)), "/"),
		botAuth:     strings.TrimSpace(os.Getenv("TELEGRAM_BOT_AUTH_TOKEN")),
		pollTimeout: parseDurationSeconds(os.Getenv("TELEGRAM_POLL_TIMEOUT_SECONDS"), defaultPollWait),
	}

	if cfg.token == "" {
		return config{}, errors.New("TELEGRAM_BOT_TOKEN is required")
	}
	if cfg.botAuth == "" {
		return config{}, errors.New("TELEGRAM_BOT_AUTH_TOKEN is required")
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

func (a *app) run(ctx context.Context) error {
	offset := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		updates, err := a.getUpdates(ctx, offset)
		if err != nil {
			log.Printf("getUpdates error: %v", err)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(2 * time.Second):
			}
			continue
		}

		for _, update := range updates {
			if update.UpdateID >= offset {
				offset = update.UpdateID + 1
			}
			if update.Message == nil || strings.TrimSpace(update.Message.Text) == "" {
				continue
			}

			a.handleMessage(ctx, *update.Message)
		}
	}
}

func (a *app) getUpdates(ctx context.Context, offset int) ([]telegramUpdate, error) {
	updatesURL, err := url.Parse(a.apiURL + "/getUpdates")
	if err != nil {
		return nil, err
	}

	query := updatesURL.Query()
	query.Set("offset", strconv.Itoa(offset))
	query.Set("timeout", strconv.Itoa(int(a.cfg.pollTimeout.Seconds())))
	updatesURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, updatesURL.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload telegramAPIResponse[[]telegramUpdate]
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if !payload.OK {
		return nil, fmt.Errorf("telegram getUpdates failed: %s", payload.Description)
	}

	return payload.Result, nil
}

func (a *app) handleMessage(ctx context.Context, msg telegramMessage) {
	text := strings.TrimSpace(msg.Text)
	if text == "" {
		return
	}

	command, arg := parseCommand(text)
	switch command {
	case "/start":
		code := parseStartCode(arg)
		if code == "" {
			a.sendMessage(ctx, msg.Chat.ID, "Привет! Для привязки аккаунта отправь команду:\n/link <CODE>\n\nКод получаешь в приложении через endpoint /api/v1/auth/telegram/link-code.")
			return
		}
		a.tryLink(ctx, msg, code)
	case "/link":
		if arg == "" {
			a.sendMessage(ctx, msg.Chat.ID, "Использование: /link <CODE>")
			return
		}
		a.tryLink(ctx, msg, arg)
	default:
		a.sendMessage(ctx, msg.Chat.ID, "Поддерживаемые команды:\n/start\n/link <CODE>")
	}
}

func parseCommand(text string) (string, string) {
	parts := strings.Fields(text)
	if len(parts) == 0 {
		return "", ""
	}

	command := strings.ToLower(parts[0])
	if at := strings.Index(command, "@"); at > 0 {
		command = command[:at]
	}
	if len(parts) == 1 {
		return command, ""
	}
	return command, strings.TrimSpace(strings.Join(parts[1:], " "))
}

func parseStartCode(arg string) string {
	value := strings.TrimSpace(arg)
	if value == "" {
		return ""
	}

	if strings.HasPrefix(strings.ToLower(value), "link_") {
		value = value[5:]
	}
	return strings.TrimSpace(value)
}

func (a *app) tryLink(ctx context.Context, msg telegramMessage, rawCode string) {
	code := strings.ToUpper(strings.TrimSpace(rawCode))
	if code == "" {
		a.sendMessage(ctx, msg.Chat.ID, "Код не найден. Использование: /link <CODE>")
		return
	}

	statusCode, backendErr := a.linkTelegram(ctx, backendTelegramLinkRequest{
		Code:             code,
		TelegramUserID:   msg.From.ID,
		TelegramUsername: msg.From.Username,
		TelegramFirst:    msg.From.FirstName,
		TelegramLast:     msg.From.LastName,
	})
	switch statusCode {
	case http.StatusOK:
		a.sendMessage(ctx, msg.Chat.ID, "Готово. Telegram аккаунт успешно привязан.")
	case http.StatusBadRequest:
		a.sendMessage(ctx, msg.Chat.ID, "Код недействителен или истек. Запроси новый код в приложении.")
	case http.StatusConflict:
		a.sendMessage(ctx, msg.Chat.ID, "Этот Telegram уже привязан к другому пользователю.")
	default:
		log.Printf("telegram link failed (status=%d): %s", statusCode, backendErr)
		a.sendMessage(ctx, msg.Chat.ID, "Не удалось привязать аккаунт. Попробуй позже.")
	}
}

func (a *app) linkTelegram(ctx context.Context, payload backendTelegramLinkRequest) (int, string) {
	body, err := json.Marshal(payload)
	if err != nil {
		return 0, err.Error()
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		a.cfg.backendURL+"/api/v1/integrations/telegram/link",
		bytes.NewReader(body),
	)
	if err != nil {
		return 0, err.Error()
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Telegram-Bot-Auth", a.cfg.botAuth)

	resp, err := a.client.Do(req)
	if err != nil {
		return 0, err.Error()
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return resp.StatusCode, ""
	}

	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	var parsed backendErrorResponse
	if err := json.Unmarshal(raw, &parsed); err == nil && parsed.Error != "" {
		return resp.StatusCode, parsed.Error
	}
	return resp.StatusCode, strings.TrimSpace(string(raw))
}

func (a *app) sendMessage(ctx context.Context, chatID int64, text string) {
	payload, err := json.Marshal(sendMessageRequest{
		ChatID: chatID,
		Text:   text,
	})
	if err != nil {
		log.Printf("marshal sendMessage payload: %v", err)
		return
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		a.apiURL+"/sendMessage",
		bytes.NewReader(payload),
	)
	if err != nil {
		log.Printf("build sendMessage request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		log.Printf("sendMessage request failed: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		log.Printf("sendMessage failed (status=%d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
}
