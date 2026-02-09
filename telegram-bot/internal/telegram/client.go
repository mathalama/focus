package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	httpClient *http.Client
	apiURL     string
}

func NewClient(token string, pollTimeout time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: pollTimeout + 15*time.Second,
		},
		apiURL: fmt.Sprintf("https://api.telegram.org/bot%s", token),
	}
}

func (c *Client) GetUpdates(ctx context.Context, offset int, pollTimeout time.Duration) ([]Update, error) {
	updatesURL, err := url.Parse(c.apiURL + "/getUpdates")
	if err != nil {
		return nil, err
	}

	query := updatesURL.Query()
	query.Set("offset", strconv.Itoa(offset))
	query.Set("timeout", strconv.Itoa(int(pollTimeout.Seconds())))
	updatesURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, updatesURL.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload APIResponse[[]Update]
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if !payload.OK {
		return nil, fmt.Errorf("telegram getUpdates failed: %s", payload.Description)
	}

	return payload.Result, nil
}

func (c *Client) SendMessage(ctx context.Context, chatID int64, text string, markup *InlineKeyboardMarkup) error {
	payload, err := json.Marshal(SendMessageRequest{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: markup,
	})
	if err != nil {
		return fmt.Errorf("marshal sendMessage payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL+"/sendMessage", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("build sendMessage request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sendMessage request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("telegram sendMessage failed (status=%d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return nil
}

func (c *Client) AnswerCallbackQuery(ctx context.Context, queryID, text string) {
	if strings.TrimSpace(queryID) == "" {
		return
	}

	payload, err := json.Marshal(AnswerCallbackQueryRequest{
		CallbackQueryID: queryID,
		Text:            strings.TrimSpace(text),
	})
	if err != nil {
		log.Printf("marshal answerCallbackQuery payload: %v", err)
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL+"/answerCallbackQuery", bytes.NewReader(payload))
	if err != nil {
		log.Printf("build answerCallbackQuery request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("answerCallbackQuery request failed: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		log.Printf("answerCallbackQuery failed (status=%d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
}
