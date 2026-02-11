package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
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

	// Apply migrations automatically
	migrationsDir := strings.TrimSpace(os.Getenv("MIGRATIONS_DIR"))
	if migrationsDir == "" {
		// Try to find migrations directory relative to executable
		execPath, err := os.Executable()
		if err == nil {
			// Try /app/migrations (Docker) or ./migrations (local)
			if _, err := os.Stat("/app/migrations"); err == nil {
				migrationsDir = "/app/migrations"
			} else if _, err := os.Stat(filepath.Join(filepath.Dir(execPath), "..", "..", "migrations")); err == nil {
				migrationsDir = filepath.Join(filepath.Dir(execPath), "..", "..", "migrations")
			} else {
				migrationsDir = "migrations"
			}
		} else {
			migrationsDir = "migrations"
		}
	}
	log.Printf("applying migrations from: %s", migrationsDir)
	if err := postgres.ApplyMigrations(ctx, pool, migrationsDir); err != nil {
		log.Printf("warning: migration error: %v", err)
	} else {
		log.Println("migrations applied successfully")
	}

	repo := postgres.NewRepository(pool, cfg.MaxSessionPauses, cfg.MaxActiveAuthSessions, cfg.AuthSessionBindClient)
	jwtSvc := jwt.NewService(cfg.JWTSecret)

	var emailSvc usecase.EmailService
	var emailOutbox *email.OutboxService
	if cfg.ResendAPIKey != "" && cfg.ResendFromEmail != "" {
		resendSvc := email.NewResendService(cfg.ResendAPIKey, cfg.ResendFromEmail, cfg.EmailVerificationTTLMin)
		emailOutbox = email.NewOutboxService(
			pool,
			resendSvc,
			time.Duration(cfg.EmailOutboxPollSeconds)*time.Second,
			cfg.EmailOutboxMaxAttempts,
		)
		emailOutbox.Start(ctx)
		emailSvc = emailOutbox
	}

	// Use cases
	authUC := usecase.NewAuthUseCase(
		repo, jwtSvc, emailSvc,
		time.Duration(cfg.EmailVerificationTTLMin)*time.Minute,
		time.Duration(cfg.RefreshSessionTTLHours)*time.Hour,
		cfg.EmailVerifyURLBase,
		cfg.EmailResetPasswordURLBase,
		cfg.EmailVerifySuccessRedirect,
		cfg.EmailVerifyFailRedirect,
	)
	sessionUC := usecase.NewSessionUseCase(repo)
	goalUC := usecase.NewGoalUseCase(repo)
	analyticsUC := usecase.NewAnalyticsUseCase(repo, repo)
	shopUC := usecase.NewShopUseCase(repo)
	telegramUC := usecase.NewTelegramUseCase(repo, time.Duration(cfg.TelegramLinkCodeTTLMinutes)*time.Minute)
	notificationUC := usecase.NewNotificationUseCase(repo)
	preferencesUC := usecase.NewPreferencesUseCase(repo, repo)

	// Delivery
	handler := httpapi.NewHandler(authUC, sessionUC, goalUC, analyticsUC, shopUC, telegramUC, notificationUC, preferencesUC, repo, emailOutbox, repo, cfg.TelegramBotAuthToken)
	router := httpapi.NewRouter(handler, cfg.CorsOrigin, cfg.EnableDevLogin, jwtSvc.ValidateToken, cfg.JWTSecret)

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
