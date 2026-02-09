package postgres

import "github.com/jackc/pgx/v5/pgxpool"

// Repository implements all repository interfaces using PostgreSQL.
type Repository struct {
	pool             *pgxpool.Pool
	maxSessionPauses int
}

// NewRepository creates a new postgres-backed repository.
func NewRepository(pool *pgxpool.Pool, maxSessionPauses int) *Repository {
	return &Repository{pool: pool, maxSessionPauses: maxSessionPauses}
}
