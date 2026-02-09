package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"mathalama-focus/backend/internal/domain"
	"mathalama-focus/backend/internal/usecase"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *Repository) CreateTelegramLinkCode(ctx context.Context, userID string, ttl time.Duration) (string, time.Time, error) {
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}

	expiresAt := time.Now().UTC().Add(ttl)

	if _, err := r.pool.Exec(ctx,
		`DELETE FROM telegram_link_codes WHERE user_id = $1 OR expires_at <= NOW() OR used_at IS NOT NULL`,
		userID,
	); err != nil {
		return "", time.Time{}, fmt.Errorf("cleanup telegram link codes: %w", err)
	}

	const insertQuery = `INSERT INTO telegram_link_codes (code, user_id, expires_at) VALUES ($1, $2, $3)`

	for attempt := 0; attempt < 5; attempt++ {
		code, err := generateTelegramLinkCode(8)
		if err != nil {
			return "", time.Time{}, fmt.Errorf("generate telegram link code: %w", err)
		}

		if _, err := r.pool.Exec(ctx, insertQuery, code, userID, expiresAt); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				continue
			}
			return "", time.Time{}, fmt.Errorf("insert telegram link code: %w", err)
		}

		return code, expiresAt, nil
	}

	return "", time.Time{}, errors.New("failed to allocate unique telegram link code")
}

func (r *Repository) LinkTelegramByCode(ctx context.Context, input usecase.TelegramLinkInput) (domain.User, error) {
	input.Code = strings.ToUpper(strings.TrimSpace(input.Code))
	input.TelegramUsername = strings.TrimSpace(input.TelegramUsername)
	input.TelegramFirst = strings.TrimSpace(input.TelegramFirst)
	input.TelegramLast = strings.TrimSpace(input.TelegramLast)

	if input.Code == "" || input.TelegramUserID <= 0 {
		return domain.User{}, domain.ErrTelegramCodeInvalid
	}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.User{}, fmt.Errorf("begin telegram link tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var userID string
	if err := tx.QueryRow(ctx,
		`SELECT user_id FROM telegram_link_codes WHERE code = $1 AND used_at IS NULL AND expires_at > NOW() FOR UPDATE`,
		input.Code,
	).Scan(&userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrTelegramCodeInvalid
		}
		return domain.User{}, fmt.Errorf("find telegram link code: %w", err)
	}

	const upsertIdentityQuery = `
		INSERT INTO telegram_identities (user_id, telegram_user_id, telegram_username, telegram_first_name, telegram_last_name, notifications_enabled)
		VALUES ($1, $2, $3, $4, $5, FALSE)
		ON CONFLICT (user_id) DO UPDATE SET
			telegram_user_id = EXCLUDED.telegram_user_id,
			telegram_username = EXCLUDED.telegram_username,
			telegram_first_name = EXCLUDED.telegram_first_name,
			telegram_last_name = EXCLUDED.telegram_last_name,
			notifications_enabled = FALSE,
			linked_at = NOW()
	`

	if _, err := tx.Exec(ctx, upsertIdentityQuery,
		userID, input.TelegramUserID, input.TelegramUsername, input.TelegramFirst, input.TelegramLast,
	); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.User{}, domain.ErrTelegramAlreadyLinked
		}
		return domain.User{}, fmt.Errorf("upsert telegram identity: %w", err)
	}

	if _, err := tx.Exec(ctx, `UPDATE telegram_link_codes SET used_at = NOW() WHERE code = $1`, input.Code); err != nil {
		return domain.User{}, fmt.Errorf("mark telegram link code as used: %w", err)
	}

	var user domain.User
	if err := tx.QueryRow(ctx,
		`SELECT id, email, name, nectar_balance, total_nectar_earned, created_at FROM users WHERE id = $1`,
		userID,
	).Scan(&user.ID, &user.Email, &user.Name, &user.NectarBalance, &user.TotalNectarEarned, &user.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, fmt.Errorf("read linked user: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.User{}, fmt.Errorf("commit telegram link tx: %w", err)
	}

	return user, nil
}

func (r *Repository) GetTelegramIdentity(ctx context.Context, userID string) (domain.TelegramIdentity, error) {
	const query = `
		SELECT telegram_user_id, telegram_username, telegram_first_name, telegram_last_name, notifications_enabled, linked_at
		FROM telegram_identities WHERE user_id = $1
	`

	var identity domain.TelegramIdentity
	if err := r.pool.QueryRow(ctx, query, userID).Scan(
		&identity.TelegramUserID, &identity.TelegramUsername,
		&identity.TelegramFirst, &identity.TelegramLast,
		&identity.NotificationsOn, &identity.LinkedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.TelegramIdentity{}, domain.ErrNotFound
		}
		return domain.TelegramIdentity{}, fmt.Errorf("get telegram identity: %w", err)
	}

	return identity, nil
}

func (r *Repository) GetUserByTelegramUserID(ctx context.Context, telegramUserID int64) (domain.User, bool, error) {
	const query = `
		SELECT u.id, u.email, u.name, u.nectar_balance, u.total_nectar_earned, u.created_at, ti.notifications_enabled
		FROM telegram_identities ti
		JOIN users u ON u.id = ti.user_id
		WHERE ti.telegram_user_id = $1
	`

	var user domain.User
	var notificationsEnabled bool
	if err := r.pool.QueryRow(ctx, query, telegramUserID).Scan(
		&user.ID, &user.Email, &user.Name,
		&user.NectarBalance, &user.TotalNectarEarned, &user.CreatedAt,
		&notificationsEnabled,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, false, domain.ErrNotFound
		}
		return domain.User{}, false, fmt.Errorf("get user by telegram user id: %w", err)
	}

	return user, notificationsEnabled, nil
}

func (r *Repository) SetTelegramNotifications(ctx context.Context, telegramUserID int64, enabled bool) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE telegram_identities SET notifications_enabled = $2 WHERE telegram_user_id = $1`,
		telegramUserID, enabled,
	)
	if err != nil {
		return fmt.Errorf("set telegram notifications: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *Repository) UnlinkTelegram(ctx context.Context, userID string) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin unlink telegram tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM telegram_identities WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("delete telegram identity: %w", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM telegram_link_codes WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("delete telegram link codes: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit unlink telegram tx: %w", err)
	}
	return nil
}
