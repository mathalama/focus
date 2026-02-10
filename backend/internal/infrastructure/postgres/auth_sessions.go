package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"mathalama-focus/backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *Repository) CreateRefreshSession(
	ctx context.Context,
	userID, tokenHash, userAgent, ipAddress string,
	ttl time.Duration,
) (domain.AuthSession, error) {
	expiresAt := time.Now().UTC().Add(ttl)

	const query = `
		INSERT INTO auth_sessions (user_id, token_hash, user_agent, ip_address, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, user_agent, ip_address, expires_at, revoked_at, created_at, last_used_at
	`

	var session domain.AuthSession
	if err := r.pool.QueryRow(ctx, query, userID, tokenHash, userAgent, ipAddress, expiresAt).Scan(
		&session.ID, &session.UserID, &session.UserAgent, &session.IPAddress,
		&session.ExpiresAt, &session.RevokedAt, &session.CreatedAt, &session.LastUsedAt,
	); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && (pgErr.Code == "42P01" || pgErr.Code == "42703") {
			return domain.AuthSession{}, domain.ErrAuthUnavailable
		}
		return domain.AuthSession{}, fmt.Errorf("create refresh session: %w", err)
	}

	if r.maxActiveAuthSessions > 0 {
		if err := r.revokeOldestSessions(ctx, userID, r.maxActiveAuthSessions); err != nil {
			return domain.AuthSession{}, err
		}
	}

	return session, nil
}

func (r *Repository) RotateRefreshSession(
	ctx context.Context,
	currentTokenHash, newTokenHash, userAgent, ipAddress string,
	ttl time.Duration,
) (domain.AuthSession, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.AuthSession{}, fmt.Errorf("begin rotate refresh session tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var userID string
	var existingUserAgent string
	var existingIPAddress string
	const lockQuery = `
		SELECT user_id, COALESCE(user_agent, ''), COALESCE(ip_address, '')
		FROM auth_sessions
		WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > NOW()
		FOR UPDATE
	`
	if err := tx.QueryRow(ctx, lockQuery, currentTokenHash).Scan(&userID, &existingUserAgent, &existingIPAddress); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.AuthSession{}, domain.ErrRefreshTokenInvalid
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && (pgErr.Code == "42P01" || pgErr.Code == "42703") {
			return domain.AuthSession{}, domain.ErrAuthUnavailable
		}
		return domain.AuthSession{}, fmt.Errorf("lock refresh session: %w", err)
	}

	if r.bindAuthSessionClient {
		if !sameClient(existingUserAgent, userAgent, existingIPAddress, ipAddress) {
			return domain.AuthSession{}, domain.ErrRefreshTokenInvalid
		}
	}

	if _, err := tx.Exec(ctx, `UPDATE auth_sessions SET revoked_at = NOW() WHERE token_hash = $1`, currentTokenHash); err != nil {
		return domain.AuthSession{}, fmt.Errorf("revoke old refresh session: %w", err)
	}

	expiresAt := time.Now().UTC().Add(ttl)
	const insertQuery = `
		INSERT INTO auth_sessions (user_id, token_hash, user_agent, ip_address, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, user_agent, ip_address, expires_at, revoked_at, created_at, last_used_at
	`

	var session domain.AuthSession
	if err := tx.QueryRow(ctx, insertQuery, userID, newTokenHash, userAgent, ipAddress, expiresAt).Scan(
		&session.ID, &session.UserID, &session.UserAgent, &session.IPAddress,
		&session.ExpiresAt, &session.RevokedAt, &session.CreatedAt, &session.LastUsedAt,
	); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && (pgErr.Code == "42P01" || pgErr.Code == "42703") {
			return domain.AuthSession{}, domain.ErrAuthUnavailable
		}
		return domain.AuthSession{}, fmt.Errorf("insert rotated refresh session: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.AuthSession{}, fmt.Errorf("commit rotate refresh session tx: %w", err)
	}

	if r.maxActiveAuthSessions > 0 {
		if err := r.revokeOldestSessions(ctx, userID, r.maxActiveAuthSessions); err != nil {
			return domain.AuthSession{}, err
		}
	}

	return session, nil
}

func (r *Repository) RevokeRefreshSessionByTokenHash(ctx context.Context, tokenHash string) error {
	const query = `
		UPDATE auth_sessions
		SET revoked_at = NOW()
		WHERE token_hash = $1 AND revoked_at IS NULL
	`

	tag, err := r.pool.Exec(ctx, query, tokenHash)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && (pgErr.Code == "42P01" || pgErr.Code == "42703") {
			return domain.ErrAuthUnavailable
		}
		return fmt.Errorf("revoke refresh session by token hash: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrRefreshTokenInvalid
	}
	return nil
}

func (r *Repository) RevokeAllRefreshSessions(ctx context.Context, userID string) error {
	const query = `
		UPDATE auth_sessions
		SET revoked_at = NOW()
		WHERE user_id = $1 AND revoked_at IS NULL
	`

	if _, err := r.pool.Exec(ctx, query, userID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && (pgErr.Code == "42P01" || pgErr.Code == "42703") {
			return domain.ErrAuthUnavailable
		}
		return fmt.Errorf("revoke all refresh sessions: %w", err)
	}
	return nil
}

func (r *Repository) ListActiveRefreshSessions(ctx context.Context, userID string) ([]domain.AuthSession, error) {
	const query = `
		SELECT id, user_id, user_agent, ip_address, expires_at, revoked_at, created_at, last_used_at
		FROM auth_sessions
		WHERE user_id = $1 AND revoked_at IS NULL AND expires_at > NOW()
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && (pgErr.Code == "42P01" || pgErr.Code == "42703") {
			return nil, domain.ErrAuthUnavailable
		}
		return nil, fmt.Errorf("list active refresh sessions: %w", err)
	}
	defer rows.Close()

	sessions := make([]domain.AuthSession, 0)
	for rows.Next() {
		var session domain.AuthSession
		if err := rows.Scan(
			&session.ID, &session.UserID, &session.UserAgent, &session.IPAddress,
			&session.ExpiresAt, &session.RevokedAt, &session.CreatedAt, &session.LastUsedAt,
		); err != nil {
			return nil, fmt.Errorf("scan auth session: %w", err)
		}
		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate auth sessions: %w", err)
	}

	return sessions, nil
}

func (r *Repository) revokeOldestSessions(ctx context.Context, userID string, keep int) error {
	const query = `
		WITH ranked AS (
			SELECT id, ROW_NUMBER() OVER (ORDER BY created_at DESC) AS rn
			FROM auth_sessions
			WHERE user_id = $1 AND revoked_at IS NULL AND expires_at > NOW()
		)
		UPDATE auth_sessions
		SET revoked_at = NOW()
		WHERE id IN (SELECT id FROM ranked WHERE rn > $2)
	`
	if _, err := r.pool.Exec(ctx, query, userID, keep); err != nil {
		return fmt.Errorf("revoke oldest refresh sessions: %w", err)
	}
	return nil
}

func sameClient(storedUA, incomingUA, storedIP, incomingIP string) bool {
	storedUA = strings.TrimSpace(storedUA)
	incomingUA = strings.TrimSpace(incomingUA)
	storedIP = strings.TrimSpace(storedIP)
	incomingIP = strings.TrimSpace(incomingIP)

	uaOK := storedUA == "" || incomingUA == "" || storedUA == incomingUA
	ipOK := storedIP == "" || incomingIP == "" || storedIP == incomingIP
	return uaOK && ipOK
}
