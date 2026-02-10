package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository implements all repository interfaces using PostgreSQL.
type Repository struct {
	pool                  *pgxpool.Pool
	maxSessionPauses      int
	maxActiveAuthSessions int
	bindAuthSessionClient bool
}

// NewRepository creates a new postgres-backed repository.
func NewRepository(
	pool *pgxpool.Pool,
	maxSessionPauses int,
	maxActiveAuthSessions int,
	bindAuthSessionClient bool,
) *Repository {
	return &Repository{
		pool:                  pool,
		maxSessionPauses:      maxSessionPauses,
		maxActiveAuthSessions: maxActiveAuthSessions,
		bindAuthSessionClient: bindAuthSessionClient,
	}
}

// KeepAlive runs a minimal query to verify database responsiveness.
func (r *Repository) KeepAlive(ctx context.Context) error {
	var one int
	return r.pool.QueryRow(ctx, "SELECT 1").Scan(&one)
}
