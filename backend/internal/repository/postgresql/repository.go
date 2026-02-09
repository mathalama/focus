package postgresql

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"mathalama-focus/backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound              = errors.New("resource not found")
	ErrPauseLimitReached     = errors.New("pause limit reached")
	ErrInvalidState          = errors.New("session is in invalid state for this action")
	ErrGoalCompleted         = errors.New("goal already completed")
	ErrTelegramCodeInvalid   = errors.New("telegram link code invalid or expired")
	ErrTelegramAlreadyLinked = errors.New("telegram account already linked")
)

type Repository struct {
	pool             *pgxpool.Pool
	maxSessionPauses int
}

type CreateGoalInput struct {
	Topic              string
	DesiredResult      string
	RecommendedMinutes int
	Tags               []string
}

type StartSessionInput struct {
	GoalID             string
	RecommendedMinutes int
	IsStrict           bool
}

type ReflectionInput struct {
	WhatLearned string
	WhatWasHard string
	NextAction  string
}

type TelegramLinkInput struct {
	Code             string
	TelegramUserID   int64
	TelegramUsername string
	TelegramFirst    string
	TelegramLast     string
}

type SessionHistoryFilter struct {
	Period     string
	Timezone   string
	Tags       []string
	MinMinutes int
	MaxMinutes int
	Status     string
	Limit      int
}

func New(pool *pgxpool.Pool, maxSessionPauses int) *Repository {
	return &Repository{pool: pool, maxSessionPauses: maxSessionPauses}
}

func (r *Repository) GetUser(ctx context.Context, userID string) (domain.User, error) {
	const query = `
		SELECT id, email, name, nectar_balance, total_nectar_earned, created_at
		FROM users
		WHERE id = $1
	`

	var user domain.User
	if err := r.pool.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.NectarBalance,
		&user.TotalNectarEarned,
		&user.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, ErrNotFound
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
		&user.ID,
		&user.Email,
		&user.Name,
		&user.NectarBalance,
		&user.TotalNectarEarned,
		&user.CreatedAt,
	); err != nil {
		return domain.User{}, fmt.Errorf("upsert user: %w", err)
	}

	return user, nil
}

func (r *Repository) CreateTelegramLinkCode(ctx context.Context, userID string, ttl time.Duration) (string, time.Time, error) {
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}

	expiresAt := time.Now().UTC().Add(ttl)

	if _, err := r.pool.Exec(ctx, `DELETE FROM telegram_link_codes WHERE user_id = $1 OR expires_at <= NOW() OR used_at IS NOT NULL`, userID); err != nil {
		return "", time.Time{}, fmt.Errorf("cleanup telegram link codes: %w", err)
	}

	const insertQuery = `
		INSERT INTO telegram_link_codes (code, user_id, expires_at)
		VALUES ($1, $2, $3)
	`

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

func (r *Repository) LinkTelegramByCode(ctx context.Context, input TelegramLinkInput) (domain.User, error) {
	input.Code = strings.ToUpper(strings.TrimSpace(input.Code))
	input.TelegramUsername = strings.TrimSpace(input.TelegramUsername)
	input.TelegramFirst = strings.TrimSpace(input.TelegramFirst)
	input.TelegramLast = strings.TrimSpace(input.TelegramLast)

	if input.Code == "" || input.TelegramUserID <= 0 {
		return domain.User{}, ErrTelegramCodeInvalid
	}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.User{}, fmt.Errorf("begin telegram link tx: %w", err)
	}
	defer tx.Rollback(ctx)

	const codeQuery = `
		SELECT user_id
		FROM telegram_link_codes
		WHERE code = $1
		  AND used_at IS NULL
		  AND expires_at > NOW()
		FOR UPDATE
	`

	var userID string
	if err := tx.QueryRow(ctx, codeQuery, input.Code).Scan(&userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, ErrTelegramCodeInvalid
		}
		return domain.User{}, fmt.Errorf("find telegram link code: %w", err)
	}

	const upsertIdentityQuery = `
		INSERT INTO telegram_identities (user_id, telegram_user_id, telegram_username, telegram_first_name, telegram_last_name)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE SET
			telegram_user_id = EXCLUDED.telegram_user_id,
			telegram_username = EXCLUDED.telegram_username,
			telegram_first_name = EXCLUDED.telegram_first_name,
			telegram_last_name = EXCLUDED.telegram_last_name,
			linked_at = NOW()
	`

	if _, err := tx.Exec(ctx, upsertIdentityQuery, userID, input.TelegramUserID, input.TelegramUsername, input.TelegramFirst, input.TelegramLast); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.User{}, ErrTelegramAlreadyLinked
		}
		return domain.User{}, fmt.Errorf("upsert telegram identity: %w", err)
	}

	if _, err := tx.Exec(ctx, `UPDATE telegram_link_codes SET used_at = NOW() WHERE code = $1`, input.Code); err != nil {
		return domain.User{}, fmt.Errorf("mark telegram link code as used: %w", err)
	}

	const userQuery = `
		SELECT id, email, name, nectar_balance, total_nectar_earned, created_at
		FROM users
		WHERE id = $1
	`

	var user domain.User
	if err := tx.QueryRow(ctx, userQuery, userID).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.NectarBalance,
		&user.TotalNectarEarned,
		&user.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, ErrNotFound
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
		SELECT telegram_user_id, telegram_username, telegram_first_name, telegram_last_name, linked_at
		FROM telegram_identities
		WHERE user_id = $1
	`

	var identity domain.TelegramIdentity
	if err := r.pool.QueryRow(ctx, query, userID).Scan(
		&identity.TelegramUserID,
		&identity.TelegramUsername,
		&identity.TelegramFirst,
		&identity.TelegramLast,
		&identity.LinkedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.TelegramIdentity{}, ErrNotFound
		}
		return domain.TelegramIdentity{}, fmt.Errorf("get telegram identity: %w", err)
	}

	return identity, nil
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

func (r *Repository) CreateGoal(ctx context.Context, userID string, input CreateGoalInput) (domain.Goal, error) {
	const query = `
		INSERT INTO goals (user_id, topic, desired_result, recommended_minutes, tags)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, topic, desired_result, recommended_minutes, tags, completed_at, created_at
	`

	minutes := input.RecommendedMinutes
	if minutes == 0 {
		minutes = 25
	}
	tags := input.Tags
	if tags == nil {
		tags = []string{}
	}

	var goal domain.Goal
	if err := r.pool.QueryRow(ctx, query, userID, input.Topic, input.DesiredResult, minutes, tags).Scan(
		&goal.ID,
		&goal.UserID,
		&goal.Topic,
		&goal.DesiredResult,
		&goal.RecommendedMinutes,
		&goal.Tags,
		&goal.CompletedAt,
		&goal.CreatedAt,
	); err != nil {
		return domain.Goal{}, fmt.Errorf("create goal: %w", err)
	}

	return goal, nil
}

func (r *Repository) ListGoals(ctx context.Context, userID string) ([]domain.Goal, error) {
	const query = `
		SELECT id, user_id, topic, desired_result, recommended_minutes, tags, completed_at, created_at
		FROM goals
		WHERE user_id = $1
		  AND completed_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list goals: %w", err)
	}
	defer rows.Close()

	goals := make([]domain.Goal, 0)
	for rows.Next() {
		var goal domain.Goal
		if err := rows.Scan(
			&goal.ID,
			&goal.UserID,
			&goal.Topic,
			&goal.DesiredResult,
			&goal.RecommendedMinutes,
			&goal.Tags,
			&goal.CompletedAt,
			&goal.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan goal: %w", err)
		}
		goals = append(goals, goal)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("iterate goals: %w", rows.Err())
	}

	return goals, nil
}

func (r *Repository) ListGoalHistory(ctx context.Context, userID string) ([]domain.Goal, error) {
	const query = `
		SELECT id, user_id, topic, desired_result, recommended_minutes, tags, completed_at, created_at
		FROM goals
		WHERE user_id = $1
		  AND completed_at IS NOT NULL
		ORDER BY completed_at DESC
		LIMIT 100
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list goal history: %w", err)
	}
	defer rows.Close()

	goals := make([]domain.Goal, 0)
	for rows.Next() {
		var goal domain.Goal
		if err := rows.Scan(
			&goal.ID,
			&goal.UserID,
			&goal.Topic,
			&goal.DesiredResult,
			&goal.RecommendedMinutes,
			&goal.Tags,
			&goal.CompletedAt,
			&goal.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan goal history: %w", err)
		}
		goals = append(goals, goal)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("iterate goal history: %w", rows.Err())
	}

	return goals, nil
}

func (r *Repository) StartSession(ctx context.Context, userID string, input StartSessionInput) (domain.FocusSession, error) {
	if _, err := uuid.Parse(input.GoalID); err != nil {
		return domain.FocusSession{}, fmt.Errorf("invalid goal id: %w", err)
	}

	const existingQuery = `
		SELECT id, user_id, goal_id, recommended_minutes, is_strict, status, pause_count, started_at, paused_at, completed_at
		FROM focus_sessions
		WHERE user_id = $1
		  AND goal_id = $2
		  AND status IN ('active', 'paused')
		ORDER BY started_at DESC
		LIMIT 1
	`

	minutes := input.RecommendedMinutes
	if minutes == 0 {
		minutes = 25
	}

	var session domain.FocusSession
	if err := r.pool.QueryRow(ctx, existingQuery, userID, input.GoalID).Scan(
		&session.ID,
		&session.UserID,
		&session.GoalID,
		&session.RecommendedMinutes,
		&session.IsStrict,
		&session.Status,
		&session.PauseCount,
		&session.StartedAt,
		&session.PausedAt,
		&session.CompletedAt,
	); err == nil {
		normalizeSessionStatus(&session)
		return session, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return domain.FocusSession{}, fmt.Errorf("find active session: %w", err)
	}

	const insertQuery = `
		INSERT INTO focus_sessions (user_id, goal_id, recommended_minutes, is_strict, status)
		SELECT $1, g.id, $3, $4, 'active'
		FROM goals g
		WHERE g.id = $2
		  AND g.user_id = $1
		  AND g.completed_at IS NULL
		RETURNING id, user_id, goal_id, recommended_minutes, is_strict, status, pause_count, started_at, paused_at, completed_at
	`

	if err := r.pool.QueryRow(ctx, insertQuery, userID, input.GoalID, minutes, input.IsStrict).Scan(
		&session.ID,
		&session.UserID,
		&session.GoalID,
		&session.RecommendedMinutes,
		&session.IsStrict,
		&session.Status,
		&session.PauseCount,
		&session.StartedAt,
		&session.PausedAt,
		&session.CompletedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			const goalStateQuery = `
				SELECT completed_at IS NOT NULL
				FROM goals
				WHERE id = $1
				  AND user_id = $2
			`

			var isCompleted bool
			if stateErr := r.pool.QueryRow(ctx, goalStateQuery, input.GoalID, userID).Scan(&isCompleted); stateErr != nil {
				if errors.Is(stateErr, pgx.ErrNoRows) {
					return domain.FocusSession{}, ErrNotFound
				}
				return domain.FocusSession{}, fmt.Errorf("check goal state: %w", stateErr)
			}
			if isCompleted {
				return domain.FocusSession{}, ErrGoalCompleted
			}
			return domain.FocusSession{}, ErrInvalidState
		}
		return domain.FocusSession{}, fmt.Errorf("start session: %w", err)
	}

	normalizeSessionStatus(&session)
	return session, nil
}

func (r *Repository) PauseSession(ctx context.Context, userID, sessionID string) (domain.FocusSession, error) {
	// First check if strict mode is enabled
	var isStrict bool
	checkQuery := `SELECT is_strict FROM focus_sessions WHERE id = $1 AND user_id = $2`
	if err := r.pool.QueryRow(ctx, checkQuery, sessionID, userID).Scan(&isStrict); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.FocusSession{}, ErrNotFound
		}
		return domain.FocusSession{}, fmt.Errorf("check strict mode: %w", err)
	}

	if isStrict {
		return domain.FocusSession{}, ErrInvalidState // Or a specific ErrStrictMode
	}

	const query = `
		UPDATE focus_sessions
		SET status = 'paused',
			paused_at = NOW(),
			pause_count = pause_count + 1
		WHERE id = $1
		  AND user_id = $2
		  AND status = 'active'
		  AND pause_count < $3
		RETURNING id, user_id, goal_id, recommended_minutes, is_strict, status, pause_count, started_at, paused_at, completed_at
	`

	var session domain.FocusSession
	if err := r.pool.QueryRow(ctx, query, sessionID, userID, r.maxSessionPauses).Scan(
		&session.ID,
		&session.UserID,
		&session.GoalID,
		&session.RecommendedMinutes,
		&session.IsStrict,
		&session.Status,
		&session.PauseCount,
		&session.StartedAt,
		&session.PausedAt,
		&session.CompletedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			var currentPauseCount int
			var status string
			lookupErr := r.pool.QueryRow(ctx, `SELECT pause_count, status FROM focus_sessions WHERE id = $1 AND user_id = $2`, sessionID, userID).Scan(&currentPauseCount, &status)
			if errors.Is(lookupErr, pgx.ErrNoRows) {
				return domain.FocusSession{}, ErrNotFound
			}
			if lookupErr == nil && currentPauseCount >= r.maxSessionPauses {
				return domain.FocusSession{}, ErrPauseLimitReached
			}
			return domain.FocusSession{}, ErrInvalidState
		}
		return domain.FocusSession{}, fmt.Errorf("pause session: %w", err)
	}

	const interruptionQuery = `
		INSERT INTO interruptions (session_id, kind, reason)
		VALUES ($1, 'pause', 'manual pause')
	`
	if _, err := r.pool.Exec(ctx, interruptionQuery, session.ID); err != nil {
		return domain.FocusSession{}, fmt.Errorf("log pause interruption: %w", err)
	}

	normalizeSessionStatus(&session)
	return session, nil
}

func (r *Repository) ResumeSession(ctx context.Context, userID, sessionID string) (domain.FocusSession, error) {
	const query = `
		UPDATE focus_sessions
		SET status = 'active',
			started_at = CASE
				WHEN paused_at IS NOT NULL THEN started_at + (NOW() - paused_at)
				ELSE started_at
			END,
			paused_at = NULL
		WHERE id = $1
		  AND user_id = $2
		  AND status = 'paused'
		RETURNING id, user_id, goal_id, recommended_minutes, is_strict, status, pause_count, started_at, paused_at, completed_at
	`

	var session domain.FocusSession
	if err := r.pool.QueryRow(ctx, query, sessionID, userID).Scan(
		&session.ID,
		&session.UserID,
		&session.GoalID,
		&session.RecommendedMinutes,
		&session.IsStrict,
		&session.Status,
		&session.PauseCount,
		&session.StartedAt,
		&session.PausedAt,
		&session.CompletedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.FocusSession{}, ErrInvalidState
		}
		return domain.FocusSession{}, fmt.Errorf("resume session: %w", err)
	}

	normalizeSessionStatus(&session)
	return session, nil
}

func (r *Repository) ResetSession(ctx context.Context, userID, sessionID string) (domain.FocusSession, error) {
	const query = `
		UPDATE focus_sessions
		SET status = 'active',
			started_at = NOW(),
			paused_at = NULL
		WHERE id = $1
		  AND user_id = $2
		  AND status IN ('active', 'paused')
		RETURNING id, user_id, goal_id, recommended_minutes, is_strict, status, pause_count, started_at, paused_at, completed_at
	`

	var session domain.FocusSession
	if err := r.pool.QueryRow(ctx, query, sessionID, userID).Scan(
		&session.ID,
		&session.UserID,
		&session.GoalID,
		&session.RecommendedMinutes,
		&session.IsStrict,
		&session.Status,
		&session.PauseCount,
		&session.StartedAt,
		&session.PausedAt,
		&session.CompletedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.FocusSession{}, ErrInvalidState
		}
		return domain.FocusSession{}, fmt.Errorf("reset session: %w", err)
	}

	normalizeSessionStatus(&session)
	return session, nil
}

func (r *Repository) AbandonSession(ctx context.Context, userID, sessionID string) (domain.FocusSession, error) {
	const query = `
		UPDATE focus_sessions
		SET status = 'cancelled',
			completed_at = NOW(),
			paused_at = NULL
		WHERE id = $1
		  AND user_id = $2
		  AND status IN ('active', 'paused')
		RETURNING id, user_id, goal_id, recommended_minutes, is_strict, status, pause_count, started_at, paused_at, completed_at
	`

	var session domain.FocusSession
	if err := r.pool.QueryRow(ctx, query, sessionID, userID).Scan(
		&session.ID,
		&session.UserID,
		&session.GoalID,
		&session.RecommendedMinutes,
		&session.IsStrict,
		&session.Status,
		&session.PauseCount,
		&session.StartedAt,
		&session.PausedAt,
		&session.CompletedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.FocusSession{}, ErrInvalidState
		}
		return domain.FocusSession{}, fmt.Errorf("abandon session: %w", err)
	}

	normalizeSessionStatus(&session)
	return session, nil
}

func (r *Repository) GetActiveSession(ctx context.Context, userID string) (domain.FocusSession, error) {
	const query = `
		SELECT id, user_id, goal_id, recommended_minutes, is_strict, status, pause_count, started_at, paused_at, completed_at
		FROM focus_sessions
		WHERE user_id = $1
		  AND status IN ('active', 'paused')
		ORDER BY started_at DESC
		LIMIT 1
	`

	var session domain.FocusSession
	if err := r.pool.QueryRow(ctx, query, userID).Scan(
		&session.ID,
		&session.UserID,
		&session.GoalID,
		&session.RecommendedMinutes,
		&session.IsStrict,
		&session.Status,
		&session.PauseCount,
		&session.StartedAt,
		&session.PausedAt,
		&session.CompletedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.FocusSession{}, ErrNotFound
		}
		return domain.FocusSession{}, fmt.Errorf("get active session: %w", err)
	}

	normalizeSessionStatus(&session)
	return session, nil
}

func (r *Repository) GetSession(ctx context.Context, userID, sessionID string) (domain.FocusSession, error) {
	const query = `
		SELECT id, user_id, goal_id, recommended_minutes, is_strict, status, pause_count, started_at, paused_at, completed_at
		FROM focus_sessions
		WHERE id = $1 AND user_id = $2
	`

	var session domain.FocusSession
	if err := r.pool.QueryRow(ctx, query, sessionID, userID).Scan(
		&session.ID,
		&session.UserID,
		&session.GoalID,
		&session.RecommendedMinutes,
		&session.IsStrict,
		&session.Status,
		&session.PauseCount,
		&session.StartedAt,
		&session.PausedAt,
		&session.CompletedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.FocusSession{}, ErrNotFound
		}
		return domain.FocusSession{}, fmt.Errorf("get session: %w", err)
	}

	normalizeSessionStatus(&session)
	return session, nil
}

func (r *Repository) AddInterruption(ctx context.Context, userID, sessionID, reason string) (domain.Interruption, error) {
	const accessQuery = `
		SELECT 1
		FROM focus_sessions
		WHERE id = $1
		  AND user_id = $2
	`
	if err := r.pool.QueryRow(ctx, accessQuery, sessionID, userID).Scan(new(int)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Interruption{}, ErrNotFound
		}
		return domain.Interruption{}, fmt.Errorf("check session ownership: %w", err)
	}

	const query = `
		INSERT INTO interruptions (session_id, kind, reason)
		VALUES ($1, 'interruption', $2)
		RETURNING id, session_id, kind, reason, created_at
	`

	var interruption domain.Interruption
	if err := r.pool.QueryRow(ctx, query, sessionID, reason).Scan(
		&interruption.ID,
		&interruption.SessionID,
		&interruption.Kind,
		&interruption.Reason,
		&interruption.CreatedAt,
	); err != nil {
		return domain.Interruption{}, fmt.Errorf("insert interruption: %w", err)
	}

	return interruption, nil
}

func (r *Repository) CompleteSession(ctx context.Context, userID, sessionID string) (domain.FocusSession, error) {
	// Start a transaction
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.FocusSession{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	const query = `
		UPDATE focus_sessions
		SET status = 'completed',
			completed_at = NOW(),
			paused_at = NULL
		WHERE id = $1
		  AND user_id = $2
		  AND status IN ('active', 'paused')
		RETURNING id, user_id, goal_id, recommended_minutes, is_strict, status, pause_count, started_at, paused_at, completed_at
	`

	var session domain.FocusSession
	if err := tx.QueryRow(ctx, query, sessionID, userID).Scan(
		&session.ID,
		&session.UserID,
		&session.GoalID,
		&session.RecommendedMinutes,
		&session.IsStrict,
		&session.Status,
		&session.PauseCount,
		&session.StartedAt,
		&session.PausedAt,
		&session.CompletedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.FocusSession{}, ErrInvalidState
		}
		return domain.FocusSession{}, fmt.Errorf("complete session: %w", err)
	}

	// Calculate Nectar (1 min = 1 Nectar)
	nectarEarned := session.RecommendedMinutes
	if nectarEarned > 0 {
		const nectarQuery = `
			UPDATE users
			SET nectar_balance = nectar_balance + $1,
				total_nectar_earned = total_nectar_earned + $1
			WHERE id = $2
		`
		if _, err := tx.Exec(ctx, nectarQuery, nectarEarned, userID); err != nil {
			return domain.FocusSession{}, fmt.Errorf("add nectar: %w", err)
		}
	}

	const markGoalQuery = `
		UPDATE goals
		SET completed_at = NOW()
		WHERE id = $1
		  AND completed_at IS NULL
	`
	if _, err := tx.Exec(ctx, markGoalQuery, session.GoalID); err != nil {
		return domain.FocusSession{}, fmt.Errorf("mark goal as completed: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.FocusSession{}, fmt.Errorf("commit transaction: %w", err)
	}

	normalizeSessionStatus(&session)
	return session, nil
}

func (r *Repository) ListSessionHistory(ctx context.Context, userID string, filter SessionHistoryFilter) ([]domain.SessionHistoryEntry, domain.SessionHistorySummary, error) {
	status := strings.ToLower(strings.TrimSpace(filter.Status))
	switch status {
	case "", "completed":
		status = "completed"
	case "abandoned", "all":
	default:
		status = "completed"
	}

	period := strings.ToLower(strings.TrimSpace(filter.Period))
	switch period {
	case "", "all", "today", "week", "month":
	default:
		period = "all"
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 200
	}
	if limit > 1000 {
		limit = 1000
	}

	timezone := strings.TrimSpace(filter.Timezone)
	if timezone == "" {
		timezone = "UTC"
	}

	var (
		args       = []any{userID}
		argPos     = 2
		conditions = []string{
			"fs.user_id = $1",
			"fs.completed_at IS NOT NULL",
		}
	)

	switch status {
	case "completed":
		conditions = append(conditions, "fs.status = 'completed'")
	case "abandoned":
		conditions = append(conditions, "fs.status = 'cancelled'")
	case "all":
		conditions = append(conditions, "fs.status IN ('completed', 'cancelled')")
	}

	if period != "all" {
		args = append(args, timezone)
		tzArg := argPos
		argPos++
		switch period {
		case "today":
			conditions = append(conditions, fmt.Sprintf("(fs.completed_at AT TIME ZONE $%d)::date = (NOW() AT TIME ZONE $%d)::date", tzArg, tzArg))
		case "week":
			conditions = append(conditions, fmt.Sprintf("(fs.completed_at AT TIME ZONE $%d)::date >= DATE_TRUNC('week', NOW() AT TIME ZONE $%d)::date", tzArg, tzArg))
		case "month":
			conditions = append(conditions, fmt.Sprintf("(fs.completed_at AT TIME ZONE $%d)::date >= DATE_TRUNC('month', NOW() AT TIME ZONE $%d)::date", tzArg, tzArg))
		}
	}

	if len(filter.Tags) > 0 {
		safeTags := make([]string, 0, len(filter.Tags))
		for _, tag := range filter.Tags {
			trimmed := strings.ToLower(strings.TrimSpace(tag))
			if trimmed != "" {
				safeTags = append(safeTags, trimmed)
			}
		}
		if len(safeTags) > 0 {
			args = append(args, safeTags)
			conditions = append(conditions, fmt.Sprintf("EXISTS (SELECT 1 FROM unnest(COALESCE(g.tags, '{}'::text[])) AS t(tag) WHERE lower(t.tag) = ANY($%d::text[]))", argPos))
			argPos++
		}
	}

	if filter.MinMinutes > 0 {
		args = append(args, filter.MinMinutes)
		conditions = append(conditions, fmt.Sprintf("fs.recommended_minutes >= $%d", argPos))
		argPos++
	}
	if filter.MaxMinutes > 0 {
		args = append(args, filter.MaxMinutes)
		conditions = append(conditions, fmt.Sprintf("fs.recommended_minutes <= $%d", argPos))
		argPos++
	}

	args = append(args, limit)
	limitArg := argPos

	query := fmt.Sprintf(`
		SELECT
			fs.id,
			fs.goal_id,
			COALESCE(g.topic, 'Unknown Goal') AS topic,
			COALESCE(g.desired_result, '') AS desired_result,
			COALESCE(g.tags, '{}'::text[]) AS tags,
			fs.status,
			fs.recommended_minutes,
			fs.pause_count,
			fs.started_at,
			fs.completed_at
		FROM focus_sessions fs
		LEFT JOIN goals g ON g.id = fs.goal_id
		WHERE %s
		ORDER BY fs.completed_at DESC
		LIMIT $%d
	`, strings.Join(conditions, "\n  AND "), limitArg)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, domain.SessionHistorySummary{}, fmt.Errorf("list session history: %w", err)
	}
	defer rows.Close()

	items := make([]domain.SessionHistoryEntry, 0)
	summary := domain.SessionHistorySummary{}

	for rows.Next() {
		var item domain.SessionHistoryEntry
		if err := rows.Scan(
			&item.SessionID,
			&item.GoalID,
			&item.Topic,
			&item.DesiredResult,
			&item.Tags,
			&item.Status,
			&item.RecommendedMinutes,
			&item.PauseCount,
			&item.StartedAt,
			&item.CompletedAt,
		); err != nil {
			return nil, domain.SessionHistorySummary{}, fmt.Errorf("scan session history: %w", err)
		}
		if item.Status == "cancelled" {
			item.Status = "abandoned"
		}
		summary.CompletedCount++
		summary.TotalMinutes += item.RecommendedMinutes
		items = append(items, item)
	}

	if rows.Err() != nil {
		return nil, domain.SessionHistorySummary{}, fmt.Errorf("iterate session history: %w", rows.Err())
	}

	if summary.CompletedCount > 0 {
		summary.AverageMinutes = math.Round((float64(summary.TotalMinutes)/float64(summary.CompletedCount))*10) / 10
	}

	return items, summary, nil
}

func (r *Repository) UpsertReflection(ctx context.Context, userID, sessionID string, input ReflectionInput) (domain.Reflection, error) {
	const ownershipQuery = `
		SELECT 1
		FROM focus_sessions
		WHERE id = $1
		  AND user_id = $2
	`
	if err := r.pool.QueryRow(ctx, ownershipQuery, sessionID, userID).Scan(new(int)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Reflection{}, ErrNotFound
		}
		return domain.Reflection{}, fmt.Errorf("check reflection ownership: %w", err)
	}

	const query = `
		INSERT INTO reflections (session_id, what_learned, what_was_hard, next_action)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (session_id) DO UPDATE SET
			what_learned = EXCLUDED.what_learned,
			what_was_hard = EXCLUDED.what_was_hard,
			next_action = EXCLUDED.next_action,
			updated_at = NOW()
		RETURNING id, session_id, what_learned, what_was_hard, next_action, created_at
	`

	var reflection domain.Reflection
	if err := r.pool.QueryRow(ctx, query, sessionID, input.WhatLearned, input.WhatWasHard, input.NextAction).Scan(
		&reflection.ID,
		&reflection.SessionID,
		&reflection.WhatLearned,
		&reflection.WhatWasHard,
		&reflection.NextAction,
		&reflection.CreatedAt,
	); err != nil {
		return domain.Reflection{}, fmt.Errorf("upsert reflection: %w", err)
	}

	return reflection, nil
}

func (r *Repository) AnalyticsOverview(ctx context.Context, userID string) (domain.AnalyticsOverview, error) {
	overview := domain.AnalyticsOverview{BestHourOfDayUTC: -1}

	const completedGoalsQuery = `
		SELECT COUNT(*)::INT
		FROM goals
		WHERE user_id = $1
		  AND completed_at IS NOT NULL
	`
	if err := r.pool.QueryRow(ctx, completedGoalsQuery, userID).Scan(&overview.CompletedGoals); err != nil {
		return domain.AnalyticsOverview{}, fmt.Errorf("load completed goals: %w", err)
	}

	const scoreQuery = `
		SELECT
			COALESCE(AVG(
				CASE
					WHEN fs.status = 'completed' THEN 100 - (fs.pause_count * 9 + COALESCE(i.interruptions, 0) * 12)
					ELSE 55 - (fs.pause_count * 6 + COALESCE(i.interruptions, 0) * 8)
				END
			), 0) AS calm_score,
			COALESCE(AVG(
				CASE
					WHEN fs.status = 'completed' THEN 100
					WHEN fs.status = 'paused' THEN 70
					ELSE 50
				END
			), 0) AS focus_stability
		FROM focus_sessions fs
		LEFT JOIN (
			SELECT session_id, COUNT(*)::INT AS interruptions
			FROM interruptions
			WHERE kind = 'interruption'
			GROUP BY session_id
		) i ON i.session_id = fs.id
		WHERE fs.user_id = $1
	`
	if err := r.pool.QueryRow(ctx, scoreQuery, userID).Scan(&overview.CalmScore, &overview.FocusStability); err != nil {
		return domain.AnalyticsOverview{}, fmt.Errorf("load overview scores: %w", err)
	}

	const bestHourQuery = `
		SELECT EXTRACT(HOUR FROM started_at AT TIME ZONE 'UTC')::INT AS hour_utc
		FROM focus_sessions
		WHERE user_id = $1
		GROUP BY hour_utc
		ORDER BY COUNT(*) DESC, hour_utc ASC
		LIMIT 1
	`
	if err := r.pool.QueryRow(ctx, bestHourQuery, userID).Scan(&overview.BestHourOfDayUTC); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return domain.AnalyticsOverview{}, fmt.Errorf("load best hour: %w", err)
	}

	const nectarQuery = `
		SELECT total_nectar_earned
		FROM users
		WHERE id = $1
	`
	if err := r.pool.QueryRow(ctx, nectarQuery, userID).Scan(&overview.TotalNectarEarned); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.AnalyticsOverview{}, ErrNotFound
		}
		return domain.AnalyticsOverview{}, fmt.Errorf("load total nectar: %w", err)
	}

	overview.CalmScore = clampScore(overview.CalmScore)
	overview.FocusStability = clampScore(overview.FocusStability)

	return overview, nil
}

func (r *Repository) GetLeaderboard(ctx context.Context) ([]domain.LeaderboardEntry, error) {
	const query = `
		SELECT id, name, total_nectar_earned,
		RANK() OVER (ORDER BY total_nectar_earned DESC)::INT as rank
		FROM users
		ORDER BY total_nectar_earned DESC
		LIMIT 10
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get leaderboard: %w", err)
	}
	defer rows.Close()

	entries := make([]domain.LeaderboardEntry, 0)
	for rows.Next() {
		var entry domain.LeaderboardEntry
		if err := rows.Scan(&entry.UserID, &entry.Name, &entry.TotalNectarEarned, &entry.Rank); err != nil {
			return nil, fmt.Errorf("scan leaderboard entry: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func (r *Repository) ListItems(ctx context.Context) ([]domain.Item, error) {
	const query = `
		SELECT id, name, description, cost, type
		FROM items
		ORDER BY cost ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	defer rows.Close()

	items := make([]domain.Item, 0)
	for rows.Next() {
		var item domain.Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Cost, &item.Type); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *Repository) BuyItem(ctx context.Context, userID, itemID string) (domain.UserItem, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.UserItem{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get Item Cost
	var cost int
	if err := tx.QueryRow(ctx, `SELECT cost FROM items WHERE id = $1`, itemID).Scan(&cost); err != nil {
		return domain.UserItem{}, fmt.Errorf("get item cost: %w", err)
	}

	// Check User Balance
	var balance int
	if err := tx.QueryRow(ctx, `SELECT nectar_balance FROM users WHERE id = $1`, userID).Scan(&balance); err != nil {
		return domain.UserItem{}, fmt.Errorf("get user balance: %w", err)
	}

	if balance < cost {
		return domain.UserItem{}, fmt.Errorf("insufficient nectar balance")
	}

	// Deduct Balance
	if _, err := tx.Exec(ctx, `UPDATE users SET nectar_balance = nectar_balance - $1 WHERE id = $2`, cost, userID); err != nil {
		return domain.UserItem{}, fmt.Errorf("deduct balance: %w", err)
	}

	// Add Item to User
	var userItem domain.UserItem
	const insertQuery = `
		INSERT INTO user_items (user_id, item_id)
		VALUES ($1, $2)
		RETURNING id, user_id, item_id, purchased_at
	`
	if err := tx.QueryRow(ctx, insertQuery, userID, itemID).Scan(
		&userItem.ID,
		&userItem.UserID,
		&userItem.ItemID,
		&userItem.PurchasedAt,
	); err != nil {
		return domain.UserItem{}, fmt.Errorf("add item to user: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.UserItem{}, fmt.Errorf("commit transaction: %w", err)
	}

	return userItem, nil
}

func (r *Repository) GetDailyActivity(ctx context.Context, userID string, timezone string) ([]domain.DailyActivity, error) {
	if timezone == "" {
		timezone = "UTC"
	}

	// Basic SQL injection protection for timezone could be added, but pg handles invalid timezones gracefully usually
	const query = `
		SELECT
			TO_CHAR(started_at AT TIME ZONE $2, 'YYYY-MM-DD') as date,
			COUNT(*)::INT as session_count,
			COALESCE(SUM(recommended_minutes), 0)::INT as total_minutes
		FROM focus_sessions
		WHERE user_id = $1
		  AND started_at > NOW() - INTERVAL '365 days'
		  AND status = 'completed'
		GROUP BY date
		ORDER BY date ASC
	`

	rows, err := r.pool.Query(ctx, query, userID, timezone)
	if err != nil {
		return nil, fmt.Errorf("get daily activity: %w", err)
	}
	defer rows.Close()

	activities := make([]domain.DailyActivity, 0)
	for rows.Next() {
		var activity domain.DailyActivity
		if err := rows.Scan(&activity.Date, &activity.SessionCount, &activity.TotalMinutes); err != nil {
			return nil, fmt.Errorf("scan daily activity: %w", err)
		}
		activities = append(activities, activity)
	}

	return activities, nil
}

func (r *Repository) GetDailyContributions(ctx context.Context, userID string, timezone string, date string) ([]domain.DailyContribution, error) {
	if timezone == "" {
		timezone = "UTC"
	}

	const query = `
		SELECT
			fs.id,
			fs.goal_id,
			COALESCE(g.topic, 'Unknown Goal') AS topic,
			fs.recommended_minutes,
			fs.started_at,
			fs.completed_at
		FROM focus_sessions fs
		LEFT JOIN goals g ON g.id = fs.goal_id
		WHERE fs.user_id = $1
		  AND fs.status = 'completed'
		  AND TO_CHAR(fs.started_at AT TIME ZONE $2, 'YYYY-MM-DD') = $3
		ORDER BY fs.started_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID, timezone, date)
	if err != nil {
		return nil, fmt.Errorf("get daily contributions: %w", err)
	}
	defer rows.Close()

	contributions := make([]domain.DailyContribution, 0)
	for rows.Next() {
		var item domain.DailyContribution
		if err := rows.Scan(
			&item.SessionID,
			&item.GoalID,
			&item.Topic,
			&item.Minutes,
			&item.StartedAt,
			&item.CompletedAt,
		); err != nil {
			return nil, fmt.Errorf("scan daily contribution: %w", err)
		}
		contributions = append(contributions, item)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("iterate daily contributions: %w", rows.Err())
	}

	return contributions, nil
}

func (r *Repository) GetRecentReflections(ctx context.Context, userID string, limit int) ([]domain.Reflection, error) {
	const query = `
		SELECT r.id, r.session_id, r.what_learned, r.what_was_hard, r.next_action, r.created_at
		FROM reflections r
		JOIN focus_sessions s ON r.session_id = s.id
		WHERE s.user_id = $1
		ORDER BY r.created_at DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("get recent reflections: %w", err)
	}
	defer rows.Close()

	reflections := make([]domain.Reflection, 0)
	for rows.Next() {
		var reflection domain.Reflection
		if err := rows.Scan(
			&reflection.ID,
			&reflection.SessionID,
			&reflection.WhatLearned,
			&reflection.WhatWasHard,
			&reflection.NextAction,
			&reflection.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan reflection: %w", err)
		}
		reflections = append(reflections, reflection)
	}

	return reflections, nil
}

func clampScore(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return math.Round(value*10) / 10
}

func normalizeSessionStatus(session *domain.FocusSession) {
	if session == nil {
		return
	}
	if session.Status == "cancelled" {
		session.Status = "abandoned"
	}
}

func generateTelegramLinkCode(length int) (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	if length <= 0 {
		return "", errors.New("invalid code length")
	}

	buf := make([]byte, length)
	randBytes := make([]byte, length)
	if _, err := rand.Read(randBytes); err != nil {
		return "", err
	}

	for i, b := range randBytes {
		buf[i] = alphabet[int(b)%len(alphabet)]
	}

	return string(buf), nil
}
