package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"mathalama-focus/backend/internal/config"
	"mathalama-focus/backend/internal/delivery/httpapi"
	"mathalama-focus/backend/internal/infrastructure/email"
	"mathalama-focus/backend/internal/infrastructure/jwt"
	"mathalama-focus/backend/internal/infrastructure/postgres"
	"mathalama-focus/backend/internal/usecase"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	// Infrastructure
	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database error: %v", err)
	}
	defer pool.Close()

	repo := postgres.NewRepository(pool, cfg.MaxSessionPauses)
	jwtSvc := jwt.NewService(cfg.JWTSecret)

	var emailSvc usecase.EmailService
	if cfg.ResendAPIKey != "" && cfg.ResendFromEmail != "" {
		emailSvc = email.NewResendService(cfg.ResendAPIKey, cfg.ResendFromEmail, cfg.EmailVerificationTTLMin)
	}

	// Use cases
	authUC := usecase.NewAuthUseCase(
		repo, jwtSvc, emailSvc,
		time.Duration(cfg.EmailVerificationTTLMin)*time.Minute,
		cfg.EmailVerifyURLBase,
		cfg.EmailVerifySuccessRedirect,
		cfg.EmailVerifyFailRedirect,
	)
	sessionUC := usecase.NewSessionUseCase(repo)
	goalUC := usecase.NewGoalUseCase(repo)
	analyticsUC := usecase.NewAnalyticsUseCase(repo, repo)
	shopUC := usecase.NewShopUseCase(repo)
	telegramUC := usecase.NewTelegramUseCase(repo, time.Duration(cfg.TelegramLinkCodeTTLMinutes)*time.Minute)

	// Delivery
	handler := httpapi.NewHandler(authUC, sessionUC, goalUC, analyticsUC, shopUC, telegramUC, cfg.TelegramBotAuthToken)
	router := httpapi.NewRouter(handler, cfg.CorsOrigin, jwtSvc.ValidateToken)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.Port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()

		if shutdownErr := srv.Shutdown(shutdownCtx); shutdownErr != nil {
			log.Printf("server shutdown error: %v", shutdownErr)
		}
	}()

	log.Printf("Mathalama Focus backend listening on :%s", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
