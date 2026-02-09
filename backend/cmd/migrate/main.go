package main

import (
	"context"
	"log"
	"os"
	"strings"

	"mathalama-focus/backend/internal/config"
	"mathalama-focus/backend/internal/infrastructure/postgres"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database error: %v", err)
	}
	defer pool.Close()

	migrationsDir := strings.TrimSpace(os.Getenv("MIGRATIONS_DIR"))
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}

	if err := postgres.ApplyMigrations(ctx, pool, migrationsDir); err != nil {
		log.Fatalf("migration error: %v", err)
	}

	log.Printf("migrations applied successfully from %s", migrationsDir)
}
