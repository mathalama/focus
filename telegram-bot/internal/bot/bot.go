package bot

import (
	"context"
	"errors"
	"log"
	"net/http"

	"mathalama-focus/telegram-bot/internal/backend"
	"mathalama-focus/telegram-bot/internal/config"
	"mathalama-focus/telegram-bot/internal/telegram"
)

// Bot orchestrates the Telegram polling loop and internal API server.
type Bot struct {
	cfg     config.Config
	tg      *telegram.Client
	backend *backend.Client
}

// New creates a new Bot instance.
func New(cfg config.Config, tg *telegram.Client, be *backend.Client) *Bot {
	return &Bot{cfg: cfg, tg: tg, backend: be}
}

// Run starts the polling loop and internal API server, blocking until ctx is cancelled.
func (b *Bot) Run(ctx context.Context) error {
	errCh := make(chan error, 2)

	go func() {
		errCh <- b.runPolling(ctx)
	}()

	go func() {
		errCh <- b.runInternalAPI(ctx)
	}()

	var firstErr error
	for i := 0; i < cap(errCh); i++ {
		err := <-errCh
		if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, http.ErrServerClosed) {
			continue
		}
		if firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

// sendMessage sends a text message with the default action keyboard.
func (b *Bot) sendMessage(ctx context.Context, chatID int64, text string) {
	if err := b.tg.SendMessage(ctx, chatID, text, ActionKeyboard()); err != nil {
		log.Printf("sendMessage failed: %v", err)
	}
}
