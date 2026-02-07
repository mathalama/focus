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
	"mathalama-focus/backend/internal/db"
	httpapi "mathalama-focus/backend/internal/http"
	"mathalama-focus/backend/internal/http/handlers"
	"mathalama-focus/backend/internal/repository/postgresql"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database error: %v", err)
	}
	defer pool.Close()

	repo := postgresql.New(pool, cfg.MaxSessionPauses)
	handler := handlers.New(repo)
	router := httpapi.NewRouter(handler, cfg.CorsOrigin)

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
