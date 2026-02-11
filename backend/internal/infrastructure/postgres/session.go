package postgres

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"mathalama-focus/backend/internal/domain"
	"mathalama-focus/backend/internal/usecase"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repository) StartSession(ctx context.Context, userID string, input usecase.StartSessionInput) (domain.FocusSession, error) {
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
		&session.ID, &session.UserID, &session.GoalID,
		&session.RecommendedMinutes, &session.IsStrict, &session.Status,
		&session.PauseCount, &session.StartedAt, &session.PausedAt, &session.CompletedAt,
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
		&session.ID, &session.UserID, &session.GoalID,
		&session.RecommendedMinutes, &session.IsStrict, &session.Status,
		&session.PauseCount, &session.StartedAt, &session.PausedAt, &session.CompletedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			var isCompleted bool
			if stateErr := r.pool.QueryRow(ctx,
				`SELECT completed_at IS NOT NULL FROM goals WHERE id = $1 AND user_id = $2`,
				input.GoalID, userID,
			).Scan(&isCompleted); stateErr != nil {
				if errors.Is(stateErr, pgx.ErrNoRows) {
					return domain.FocusSession{}, domain.ErrNotFound
				}
				return domain.FocusSession{}, fmt.Errorf("check goal state: %w", stateErr)
			}
			if isCompleted {
				return domain.FocusSession{}, domain.ErrGoalCompleted
			}
			return domain.FocusSession{}, domain.ErrInvalidState
		}
		return domain.FocusSession{}, fmt.Errorf("start session: %w", err)
	}

	normalizeSessionStatus(&session)
	return session, nil
}

func (r *Repository) PauseSession(ctx context.Context, userID, sessionID string) (domain.FocusSession, error) {
	var isStrict bool
	if err := r.pool.QueryRow(ctx,
		`SELECT is_strict FROM focus_sessions WHERE id = $1 AND user_id = $2`,
		sessionID, userID,
	).Scan(&isStrict); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.FocusSession{}, domain.ErrNotFound
		}
		return domain.FocusSession{}, fmt.Errorf("check strict mode: %w", err)
	}

	if isStrict {
		return domain.FocusSession{}, domain.ErrInvalidState
	}

	const query = `
		UPDATE focus_sessions
		SET status = 'paused', paused_at = NOW(), pause_count = pause_count + 1
		WHERE id = $1 AND user_id = $2 AND status = 'active' AND pause_count < $3
		RETURNING id, user_id, goal_id, recommended_minutes, is_strict, status, pause_count, started_at, paused_at, completed_at
	`

	var session domain.FocusSession
	if err := r.pool.QueryRow(ctx, query, sessionID, userID, r.maxSessionPauses).Scan(
		&session.ID, &session.UserID, &session.GoalID,
		&session.RecommendedMinutes, &session.IsStrict, &session.Status,
		&session.PauseCount, &session.StartedAt, &session.PausedAt, &session.CompletedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			var currentPauseCount int
			var status string
			lookupErr := r.pool.QueryRow(ctx,
				`SELECT pause_count, status FROM focus_sessions WHERE id = $1 AND user_id = $2`,
				sessionID, userID,
			).Scan(&currentPauseCount, &status)
			if errors.Is(lookupErr, pgx.ErrNoRows) {
				return domain.FocusSession{}, domain.ErrNotFound
			}
			if lookupErr == nil && currentPauseCount >= r.maxSessionPauses {
				return domain.FocusSession{}, domain.ErrPauseLimitReached
			}
			return domain.FocusSession{}, domain.ErrInvalidState
		}
		return domain.FocusSession{}, fmt.Errorf("pause session: %w", err)
	}

	if _, err := r.pool.Exec(ctx,
		`INSERT INTO interruptions (session_id, kind, reason) VALUES ($1, 'pause', 'manual pause')`,
		session.ID,
	); err != nil {
		return domain.FocusSession{}, fmt.Errorf("log pause interruption: %w", err)
	}

	normalizeSessionStatus(&session)
	return session, nil
}

func (r *Repository) ResumeSession(ctx context.Context, userID, sessionID string) (domain.FocusSession, error) {
	const query = `
		UPDATE focus_sessions
		SET status = 'active',
			started_at = CASE WHEN paused_at IS NOT NULL THEN started_at + (NOW() - paused_at) ELSE started_at END,
			paused_at = NULL
		WHERE id = $1 AND user_id = $2 AND status = 'paused'
		RETURNING id, user_id, goal_id, recommended_minutes, is_strict, status, pause_count, started_at, paused_at, completed_at
	`

	var session domain.FocusSession
	if err := r.pool.QueryRow(ctx, query, sessionID, userID).Scan(
		&session.ID, &session.UserID, &session.GoalID,
		&session.RecommendedMinutes, &session.IsStrict, &session.Status,
		&session.PauseCount, &session.StartedAt, &session.PausedAt, &session.CompletedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.FocusSession{}, domain.ErrInvalidState
		}
		return domain.FocusSession{}, fmt.Errorf("resume session: %w", err)
	}

	normalizeSessionStatus(&session)
	return session, nil
}

func (r *Repository) ResetSession(ctx context.Context, userID, sessionID string) (domain.FocusSession, error) {
	const query = `
		UPDATE focus_sessions
		SET status = 'active', started_at = NOW(), paused_at = NULL
		WHERE id = $1 AND user_id = $2 AND status IN ('active', 'paused')
		RETURNING id, user_id, goal_id, recommended_minutes, is_strict, status, pause_count, started_at, paused_at, completed_at
	`

	var session domain.FocusSession
	if err := r.pool.QueryRow(ctx, query, sessionID, userID).Scan(
		&session.ID, &session.UserID, &session.GoalID,
		&session.RecommendedMinutes, &session.IsStrict, &session.Status,
		&session.PauseCount, &session.StartedAt, &session.PausedAt, &session.CompletedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.FocusSession{}, domain.ErrInvalidState
		}
		return domain.FocusSession{}, fmt.Errorf("reset session: %w", err)
	}

	normalizeSessionStatus(&session)
	return session, nil
}

func (r *Repository) AbandonSession(ctx context.Context, userID, sessionID string) (domain.FocusSession, error) {
	const query = `
		UPDATE focus_sessions
		SET status = 'cancelled', completed_at = NOW(), paused_at = NULL
		WHERE id = $1 AND user_id = $2 AND status IN ('active', 'paused')
		RETURNING id, user_id, goal_id, recommended_minutes, is_strict, status, pause_count, started_at, paused_at, completed_at
	`

	var session domain.FocusSession
	if err := r.pool.QueryRow(ctx, query, sessionID, userID).Scan(
		&session.ID, &session.UserID, &session.GoalID,
		&session.RecommendedMinutes, &session.IsStrict, &session.Status,
		&session.PauseCount, &session.StartedAt, &session.PausedAt, &session.CompletedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.FocusSession{}, domain.ErrInvalidState
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
		WHERE user_id = $1 AND status IN ('active', 'paused')
		ORDER BY started_at DESC LIMIT 1
	`

	var session domain.FocusSession
	if err := r.pool.QueryRow(ctx, query, userID).Scan(
		&session.ID, &session.UserID, &session.GoalID,
		&session.RecommendedMinutes, &session.IsStrict, &session.Status,
		&session.PauseCount, &session.StartedAt, &session.PausedAt, &session.CompletedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.FocusSession{}, domain.ErrNotFound
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
		&session.ID, &session.UserID, &session.GoalID,
		&session.RecommendedMinutes, &session.IsStrict, &session.Status,
		&session.PauseCount, &session.StartedAt, &session.PausedAt, &session.CompletedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.FocusSession{}, domain.ErrNotFound
		}
		return domain.FocusSession{}, fmt.Errorf("get session: %w", err)
	}

	normalizeSessionStatus(&session)
	return session, nil
}

func (r *Repository) CompleteSession(ctx context.Context, userID, sessionID string) (domain.FocusSession, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.FocusSession{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	const query = `
		UPDATE focus_sessions
		SET status = 'completed', completed_at = NOW(), paused_at = NULL
		WHERE id = $1 AND user_id = $2 AND status IN ('active', 'paused')
		RETURNING id, user_id, goal_id, recommended_minutes, is_strict, status, pause_count, started_at, paused_at, completed_at
	`

	var session domain.FocusSession
	if err := tx.QueryRow(ctx, query, sessionID, userID).Scan(
		&session.ID, &session.UserID, &session.GoalID,
		&session.RecommendedMinutes, &session.IsStrict, &session.Status,
		&session.PauseCount, &session.StartedAt, &session.PausedAt, &session.CompletedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.FocusSession{}, domain.ErrInvalidState
		}
		return domain.FocusSession{}, fmt.Errorf("complete session: %w", err)
	}

	nectarEarned := session.RecommendedMinutes
	if nectarEarned > 0 {
		if _, err := tx.Exec(ctx,
			`UPDATE users SET nectar_balance = nectar_balance + $1, total_nectar_earned = total_nectar_earned + $1 WHERE id = $2`,
			nectarEarned, userID,
		); err != nil {
			return domain.FocusSession{}, fmt.Errorf("add nectar: %w", err)
		}
	}

	if _, err := tx.Exec(ctx,
		`UPDATE goals SET completed_at = NOW() WHERE id = $1 AND completed_at IS NULL`,
		session.GoalID,
	); err != nil {
		return domain.FocusSession{}, fmt.Errorf("mark goal as completed: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.FocusSession{}, fmt.Errorf("commit transaction: %w", err)
	}

	normalizeSessionStatus(&session)
	return session, nil
}

func (r *Repository) AddInterruption(ctx context.Context, userID, sessionID, reason string) (domain.Interruption, error) {
	if err := r.pool.QueryRow(ctx,
		`SELECT 1 FROM focus_sessions WHERE id = $1 AND user_id = $2`,
		sessionID, userID,
	).Scan(new(int)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Interruption{}, domain.ErrNotFound
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
		&interruption.ID, &interruption.SessionID,
		&interruption.Kind, &interruption.Reason, &interruption.CreatedAt,
	); err != nil {
		return domain.Interruption{}, fmt.Errorf("insert interruption: %w", err)
	}

	return interruption, nil
}

func (r *Repository) ListSessionHistory(ctx context.Context, userID string, filter usecase.SessionHistoryFilter) ([]domain.SessionHistoryEntry, domain.SessionHistorySummary, error) {
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
		SELECT fs.id, fs.goal_id,
			COALESCE(g.topic, 'Unknown Goal') AS topic,
			COALESCE(g.desired_result, '') AS desired_result,
			COALESCE(g.tags, '{}'::text[]) AS tags,
			fs.status, fs.recommended_minutes, fs.pause_count, fs.started_at, fs.completed_at
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
			&item.SessionID, &item.GoalID, &item.Topic, &item.DesiredResult,
			&item.Tags, &item.Status, &item.RecommendedMinutes,
			&item.PauseCount, &item.StartedAt, &item.CompletedAt,
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

func (r *Repository) DeleteSession(ctx context.Context, userID, sessionID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get the session to check if it was completed and get nectar amount
	const getSessionQuery = `
		SELECT status, recommended_minutes
		FROM focus_sessions
		WHERE id = $1 AND user_id = $2
	`

	var status string
	var minutes int

	if err := tx.QueryRow(ctx, getSessionQuery, sessionID, userID).Scan(&status, &minutes); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("get session: %w", err)
	}

	// Only allow deletion of non-active sessions
	if status == "active" || status == "paused" {
		return domain.ErrInvalidState
	}

	// If session was completed, deduct nectar from user
	if status == "completed" && minutes > 0 {
		if _, err := tx.Exec(ctx,
			`UPDATE users SET nectar_balance = GREATEST(0, nectar_balance - $1) WHERE id = $2`,
			minutes, userID,
		); err != nil {
			return fmt.Errorf("deduct nectar: %w", err)
		}
	}

	// Delete interruptions associated with the session
	if _, err := tx.Exec(ctx, `DELETE FROM interruptions WHERE session_id = $1`, sessionID); err != nil {
		return fmt.Errorf("delete interruptions: %w", err)
	}

	// Delete reflections associated with the session
	if _, err := tx.Exec(ctx, `DELETE FROM reflections WHERE session_id = $1`, sessionID); err != nil {
		return fmt.Errorf("delete reflections: %w", err)
	}

	// Delete the session itself
	const deleteSessionQuery = `DELETE FROM focus_sessions WHERE id = $1 AND user_id = $2`
	if _, err := tx.Exec(ctx, deleteSessionQuery, sessionID, userID); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (r *Repository) DeleteReflection(ctx context.Context, userID, sessionID string) error {
	// Verify session belongs to user
	if err := r.pool.QueryRow(ctx,
		`SELECT 1 FROM focus_sessions WHERE id = $1 AND user_id = $2`,
		sessionID, userID,
	).Scan(new(int)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("check session ownership: %w", err)
	}

	const query = `DELETE FROM reflections WHERE session_id = $1`
	if _, err := r.pool.Exec(ctx, query, sessionID); err != nil {
		return fmt.Errorf("delete reflection: %w", err)
	}

	return nil
}
