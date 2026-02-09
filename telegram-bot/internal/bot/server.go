package bot

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"mathalama-focus/telegram-bot/internal/config"
)

// NotifyRequest is the payload for the internal /internal/notify endpoint.
type NotifyRequest struct {
	TelegramUserID int64  `json:"telegram_user_id"`
	Message        string `json:"message"`
	DisableButtons bool   `json:"disable_buttons"`
	Force          bool   `json:"force"`
}

func (b *Bot) runInternalAPI(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", b.handleHealth)
	mux.HandleFunc("/internal/notify", b.handleNotify)

	srv := &http.Server{
		Addr:              b.cfg.InternalAPIAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("internal api shutdown error: %v", err)
		}
	}()

	log.Printf("telegram internal api listening on %s", b.cfg.InternalAPIAddr)
	err := srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (b *Bot) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (b *Bot) handleNotify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if subtle.ConstantTimeCompare([]byte(strings.TrimSpace(r.Header.Get(config.BotAuthHeader))), []byte(b.cfg.BotAuth)) != 1 {
		writeJSONError(w, http.StatusUnauthorized, "invalid bot auth token")
		return
	}

	var req NotifyRequest
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Message = strings.TrimSpace(req.Message)
	if req.TelegramUserID <= 0 || req.Message == "" {
		writeJSONError(w, http.StatusBadRequest, "telegram_user_id and message are required")
		return
	}

	if !req.Force {
		status, statusCode, statusErr := b.backend.GetTelegramStatus(r.Context(), req.TelegramUserID)
		if statusCode != http.StatusOK {
			log.Printf("internal notify status check failed (status=%d): %s", statusCode, statusErr)
			writeJSONError(w, http.StatusBadGateway, "failed to resolve telegram link status")
			return
		}
		if !status.Linked {
			writeJSONError(w, http.StatusNotFound, "telegram account is not linked")
			return
		}
		if !status.NotificationsEnabled {
			writeJSONError(w, http.StatusConflict, "telegram notifications are disabled")
			return
		}
	}

	markup := ActionKeyboard()
	if req.DisableButtons {
		markup = nil
	}

	if err := b.tg.SendMessage(r.Context(), req.TelegramUserID, req.Message, markup); err != nil {
		log.Printf("internal notify send failed: %v", err)
		writeJSONError(w, http.StatusBadGateway, "failed to send telegram message")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"sent":                 true,
		"telegram_user_id":     req.TelegramUserID,
		"notifications_forced": req.Force,
	})
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("write json response failed: %v", err)
	}
}

func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]any{"error": strings.TrimSpace(message)})
}
