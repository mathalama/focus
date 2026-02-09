package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const schemaMigrationsDDL = `
CREATE TABLE IF NOT EXISTS schema_migrations (
	filename TEXT PRIMARY KEY,
	checksum TEXT NOT NULL,
	applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);`

// ApplyMigrations executes SQL files from migrationsDir in lexical order.
func ApplyMigrations(ctx context.Context, pool *pgxpool.Pool, migrationsDir string) error {
	if strings.TrimSpace(migrationsDir) == "" {
		return fmt.Errorf("migrations directory is empty")
	}

	if _, err := pool.Exec(ctx, schemaMigrationsDDL); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	files, err := listSQLFiles(migrationsDir)
	if err != nil {
		return err
	}

	for _, migrationPath := range files {
		filename := filepath.Base(migrationPath)
		sqlBytes, readErr := os.ReadFile(migrationPath)
		if readErr != nil {
			return fmt.Errorf("read migration %s: %w", filename, readErr)
		}

		checksum := hashSQL(sqlBytes)
		var appliedChecksum string
		queryErr := pool.QueryRow(ctx, "SELECT checksum FROM schema_migrations WHERE filename = $1", filename).Scan(&appliedChecksum)
		switch {
		case queryErr == nil:
			if appliedChecksum != checksum {
				return fmt.Errorf("migration %s checksum mismatch (applied=%s current=%s)", filename, appliedChecksum, checksum)
			}
			continue
		case !errors.Is(queryErr, pgx.ErrNoRows):
			return fmt.Errorf("check migration %s status: %w", filename, queryErr)
		}

		tx, beginErr := pool.Begin(ctx)
		if beginErr != nil {
			return fmt.Errorf("start tx for migration %s: %w", filename, beginErr)
		}

		if _, execErr := tx.Exec(ctx, string(sqlBytes)); execErr != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("apply migration %s: %w", filename, execErr)
		}

		if _, markErr := tx.Exec(ctx, "INSERT INTO schema_migrations (filename, checksum) VALUES ($1, $2)", filename, checksum); markErr != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("record migration %s: %w", filename, markErr)
		}

		if commitErr := tx.Commit(ctx); commitErr != nil {
			return fmt.Errorf("commit migration %s: %w", filename, commitErr)
		}
	}

	return nil
}

func listSQLFiles(migrationsDir string) ([]string, error) {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("read migrations directory: %w", err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		files = append(files, filepath.Join(migrationsDir, entry.Name()))
	}

	sort.Strings(files)
	return files, nil
}

func hashSQL(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
