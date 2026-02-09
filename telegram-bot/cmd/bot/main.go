package main

import (
	"context"
	"errors"
	"log"
	"os/signal"
	"syscall"

	"mathalama-focus/telegram-bot/internal/backend"
	"mathalama-focus/telegram-bot/internal/bot"
	"mathalama-focus/telegram-bot/internal/config"
	"mathalama-focus/telegram-bot/internal/telegram"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	tg := telegram.NewClient(cfg.Token, cfg.PollTimeout)
	be := backend.NewClient(cfg)
	b := bot.New(cfg, tg, be)

	log.Printf("telegram bot started (backend=%s)", cfg.BackendURL)
	if err := b.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("bot stopped with error: %v", err)
	}
}
