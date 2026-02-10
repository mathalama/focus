package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"mathalama-focus/telegram-bot/internal/config"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	botAuth    string
}

func NewClient(cfg config.Config) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		baseURL:    cfg.BackendURL,
		botAuth:    cfg.BotAuth,
	}
}

func (c *Client) GetTelegramStatus(ctx context.Context, telegramUserID int64) (TelegramStatusResponse, int, string) {
	if telegramUserID <= 0 {
		return TelegramStatusResponse{}, http.StatusBadRequest, "telegram_user_id is invalid"
	}

	statusURL := fmt.Sprintf("%s/api/v1/integrations/telegram/status?telegram_user_id=%d", c.baseURL, telegramUserID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, statusURL, nil)
	if err != nil {
		return TelegramStatusResponse{}, 0, err.Error()
	}
	req.Header.Set(config.BotAuthHeader, c.botAuth)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return TelegramStatusResponse{}, 0, err.Error()
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var payload TelegramStatusResponse
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			return TelegramStatusResponse{}, resp.StatusCode, err.Error()
		}
		return payload, resp.StatusCode, ""
	}

	return TelegramStatusResponse{}, resp.StatusCode, readErrorBody(resp.Body)
}

func (c *Client) LinkTelegram(ctx context.Context, payload TelegramLinkRequest) (int, string) {
	body, err := json.Marshal(payload)
	if err != nil {
		return 0, err.Error()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/integrations/telegram/link", bytes.NewReader(body))
	if err != nil {
		return 0, err.Error()
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(config.BotAuthHeader, c.botAuth)
	req.Header.Set("Idempotency-Key", "tg-link-"+strconv.FormatInt(payload.TelegramUserID, 10)+"-"+strconv.FormatInt(time.Now().UnixNano(), 10))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err.Error()
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return resp.StatusCode, ""
	}

	return resp.StatusCode, readErrorBody(resp.Body)
}

func (c *Client) SetTelegramNotifications(ctx context.Context, telegramUserID int64, enabled bool) (int, string) {
	body, err := json.Marshal(TelegramNotificationsRequest{
		TelegramUserID: telegramUserID,
		Enabled:        enabled,
	})
	if err != nil {
		return 0, err.Error()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.baseURL+"/api/v1/integrations/telegram/notifications", bytes.NewReader(body))
	if err != nil {
		return 0, err.Error()
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(config.BotAuthHeader, c.botAuth)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err.Error()
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return resp.StatusCode, ""
	}

	return resp.StatusCode, readErrorBody(resp.Body)
}

func readErrorBody(body io.Reader) string {
	raw, _ := io.ReadAll(io.LimitReader(body, 4096))
	var parsed ErrorResponse
	if err := json.Unmarshal(raw, &parsed); err == nil && parsed.Error != "" {
		return parsed.Error
	}
	return strings.TrimSpace(string(raw))
}
