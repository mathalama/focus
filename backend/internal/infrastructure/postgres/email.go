package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"mathalama-focus/backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *Repository) CreateEmailVerificationToken(ctx context.Context, userID string, ttl time.Duration) (string, time.Time, error) {
	if ttl <= 0 {
		ttl = 60 * time.Minute
	}

	expiresAt := time.Now().UTC().Add(ttl)

	if _, err := r.pool.Exec(ctx,
		`DELETE FROM email_verification_tokens WHERE user_id = $1 OR expires_at <= NOW() OR used_at IS NOT NULL`,
		userID,
	); err != nil {
		return "", time.Time{}, fmt.Errorf("cleanup email verification tokens: %w", err)
	}

	const insertQuery = `INSERT INTO email_verification_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`

	for attempt := 0; attempt < 5; attempt++ {
		rawToken, err := generateRandomToken(32)
		if err != nil {
			return "", time.Time{}, fmt.Errorf("generate email verification token: %w", err)
		}
		tokenHash := hashToken(rawToken)

		if _, err := r.pool.Exec(ctx, insertQuery, userID, tokenHash, expiresAt); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				continue
			}
			return "", time.Time{}, fmt.Errorf("insert email verification token: %w", err)
		}

		return rawToken, expiresAt, nil
	}

	return "", time.Time{}, errors.New("failed to allocate unique email verification token")
}

func (r *Repository) VerifyEmailByToken(ctx context.Context, rawToken string) (domain.User, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return domain.User{}, domain.ErrEmailTokenInvalid
	}

	tokenHash := hashToken(rawToken)

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.User{}, fmt.Errorf("begin verify email tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var userID string
	if err := tx.QueryRow(ctx,
		`SELECT user_id FROM email_verification_tokens WHERE token_hash = $1 AND used_at IS NULL AND expires_at > NOW() FOR UPDATE`,
		tokenHash,
	).Scan(&userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.Warn("verify token not found or already used", "token_hash", tokenHash)
			return domain.User{}, domain.ErrEmailTokenInvalid
		}
		return domain.User{}, fmt.Errorf("find email verification token: %w", err)
	}

	var user domain.User
	const updateQuery = `
		UPDATE users 
		SET email_verified_at = COALESCE(email_verified_at, NOW()) 
		WHERE id = $1
		RETURNING id, email, name, role, nectar_balance, total_nectar_earned, created_at
	`
	if err := tx.QueryRow(ctx, updateQuery, userID).Scan(
		&user.ID, &user.Email, &user.Name, &user.Role, &user.NectarBalance, &user.TotalNectarEarned, &user.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, fmt.Errorf("mark user email verified: %w", err)
	}

	if _, err := tx.Exec(ctx, `UPDATE email_verification_tokens SET used_at = NOW() WHERE token_hash = $1`, tokenHash); err != nil {
		return domain.User{}, fmt.Errorf("mark verification token used: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.User{}, fmt.Errorf("commit verify email tx: %w", err)
	}

	slog.Info("successfully verified email", "user_id", user.ID, "email", user.Email)
	return user, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	const query = `
		SELECT id, email, name, role, nectar_balance, total_nectar_earned, created_at
		FROM users
		WHERE email = $1
	`

	var user domain.User
	if err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Name, &user.Role,
		&user.NectarBalance, &user.TotalNectarEarned, &user.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, fmt.Errorf("get user by email: %w", err)
	}

	return user, nil
}

func (r *Repository) CreatePasswordResetToken(ctx context.Context, userID string, ttl time.Duration) (string, time.Time, error) {
	if ttl <= 0 {
		ttl = 30 * time.Minute
	}

	expiresAt := time.Now().UTC().Add(ttl)

	// Invalidate old reset tokens
	if _, err := r.pool.Exec(ctx,
		`DELETE FROM password_reset_tokens WHERE user_id = $1 OR expires_at <= NOW() OR used_at IS NOT NULL`,
		userID,
	); err != nil {
		return "", time.Time{}, fmt.Errorf("cleanup password reset tokens: %w", err)
	}

	const insertQuery = `INSERT INTO password_reset_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`

	for attempt := 0; attempt < 5; attempt++ {
		rawToken, err := generateRandomToken(32)
		if err != nil {
			return "", time.Time{}, fmt.Errorf("generate password reset token: %w", err)
		}
		tokenHash := hashToken(rawToken)

		if _, err := r.pool.Exec(ctx, insertQuery, userID, tokenHash, expiresAt); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				continue
			}
			return "", time.Time{}, fmt.Errorf("insert password reset token: %w", err)
		}

		return rawToken, expiresAt, nil
	}

	return "", time.Time{}, errors.New("failed to allocate unique password reset token")
}

func (r *Repository) ResetPasswordByToken(ctx context.Context, rawToken, passwordHash string) (domain.User, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return domain.User{}, domain.ErrInvalidToken
	}

	tokenHash := hashToken(rawToken)

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.User{}, fmt.Errorf("begin reset password tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var userID string
	if err := tx.QueryRow(ctx,
		`SELECT user_id FROM password_reset_tokens WHERE token_hash = $1 AND used_at IS NULL AND expires_at > NOW() FOR UPDATE`,
		tokenHash,
	).Scan(&userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrInvalidToken
		}
		return domain.User{}, fmt.Errorf("find password reset token: %w", err)
	}

	var user domain.User
	if err := tx.QueryRow(ctx,
		`UPDATE users SET password_hash = $1 WHERE id = $2
		 RETURNING id, email, name, role, nectar_balance, total_nectar_earned, created_at`,
		passwordHash, userID,
	).Scan(&user.ID, &user.Email, &user.Name, &user.Role, &user.NectarBalance, &user.TotalNectarEarned, &user.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, fmt.Errorf("update password: %w", err)
	}

	if _, err := tx.Exec(ctx, `UPDATE password_reset_tokens SET used_at = NOW() WHERE token_hash = $1`, tokenHash); err != nil {
		return domain.User{}, fmt.Errorf("mark reset token used: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.User{}, fmt.Errorf("commit reset password tx: %w", err)
	}

	return user, nil
}
