package postgres

import (
	"context"
	"errors"
	"fmt"

	"mathalama-focus/backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *Repository) GetUser(ctx context.Context, userID string) (domain.User, error) {
	const query = `
		SELECT id, email, name, nectar_balance, total_nectar_earned, created_at
		FROM users
		WHERE id = $1
	`

	var user domain.User
	if err := r.pool.QueryRow(ctx, query, userID).Scan(
		&user.ID, &user.Email, &user.Name,
		&user.NectarBalance, &user.TotalNectarEarned, &user.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, fmt.Errorf("get user: %w", err)
	}

	return user, nil
}

func (r *Repository) DevLogin(ctx context.Context, email, name string) (domain.User, error) {
	const query = `
		INSERT INTO users (email, name)
		VALUES ($1, $2)
		ON CONFLICT (email) DO UPDATE SET
			name = EXCLUDED.name
		RETURNING id, email, name, nectar_balance, total_nectar_earned, created_at
	`

	var user domain.User
	if err := r.pool.QueryRow(ctx, query, email, name).Scan(
		&user.ID, &user.Email, &user.Name,
		&user.NectarBalance, &user.TotalNectarEarned, &user.CreatedAt,
	); err != nil {
		return domain.User{}, fmt.Errorf("upsert user: %w", err)
	}

	return user, nil
}

func (r *Repository) RegisterUser(ctx context.Context, email, name, passwordHash string) (domain.User, error) {
	const query = `
		INSERT INTO users (email, name, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, email, name, nectar_balance, total_nectar_earned, created_at
	`

	var user domain.User
	if err := r.pool.QueryRow(ctx, query, email, name, passwordHash).Scan(
		&user.ID, &user.Email, &user.Name,
		&user.NectarBalance, &user.TotalNectarEarned, &user.CreatedAt,
	); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.User{}, domain.ErrAlreadyExists
		}
		return domain.User{}, fmt.Errorf("register user: %w", err)
	}

	return user, nil
}

func (r *Repository) GetAuthUserByEmail(ctx context.Context, email string) (domain.AuthUser, error) {
	const query = `
		SELECT id, email, name, nectar_balance, total_nectar_earned, created_at, password_hash, email_verified_at
		FROM users
		WHERE email = $1
	`

	var result domain.AuthUser
	if err := r.pool.QueryRow(ctx, query, email).Scan(
		&result.User.ID, &result.User.Email, &result.User.Name,
		&result.User.NectarBalance, &result.User.TotalNectarEarned, &result.User.CreatedAt,
		&result.PasswordHash, &result.EmailVerifiedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.AuthUser{}, domain.ErrNotFound
		}
		return domain.AuthUser{}, fmt.Errorf("get auth user by email: %w", err)
	}

	return result, nil
}
