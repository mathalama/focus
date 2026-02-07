package postgresql

import (
	"context"
	"errors"
	"fmt"
	"math"

	"mathalama-focus/backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound          = errors.New("resource not found")
	ErrPauseLimitReached = errors.New("pause limit reached")
	ErrInvalidState      = errors.New("session is in invalid state for this action")
)

type Repository struct {
	pool             *pgxpool.Pool
	maxSessionPauses int
}

type CreateGoalInput struct {
	Topic              string
	DesiredResult      string
	RecommendedMinutes int
}

type StartSessionInput struct {
	GoalID             string
	RecommendedMinutes int
}

type ReflectionInput struct {
	WhatLearned string
	WhatWasHard string
	NextAction  string
}

func New(pool *pgxpool.Pool, maxSessionPauses int) *Repository {
	return &Repository{pool: pool, maxSessionPauses: maxSessionPauses}
}

func (r *Repository) DevLogin(ctx context.Context, email, name string) (domain.User, error) {
	const query = `
		INSERT INTO users (email, name)
		VALUES ($1, $2)
		ON CONFLICT (email) DO UPDATE SET
			name = EXCLUDED.name
		RETURNING id, email, name, created_at
	`

	var user domain.User
	if err := r.pool.QueryRow(ctx, query, email, name).Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt); err != nil {
		return domain.User{}, fmt.Errorf("upsert user: %w", err)
	}

	return user, nil
}

func (r *Repository) CreateGoal(ctx context.Context, userID string, input CreateGoalInput) (domain.Goal, error) {
	const query = `
		INSERT INTO goals (user_id, topic, desired_result, recommended_minutes)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, topic, desired_result, recommended_minutes, created_at
	`

	minutes := input.RecommendedMinutes
	if minutes == 0 {
		minutes = 25
	}

	var goal domain.Goal
	if err := r.pool.QueryRow(ctx, query, userID, input.Topic, input.DesiredResult, minutes).Scan(
		&goal.ID,
		&goal.UserID,
		&goal.Topic,
		&goal.DesiredResult,
		&goal.RecommendedMinutes,
		&goal.CreatedAt,
	); err != nil {
		return domain.Goal{}, fmt.Errorf("create goal: %w", err)
	}

	return goal, nil
}

func (r *Repository) ListGoals(ctx context.Context, userID string) ([]domain.Goal, error) {
	const query = `
		SELECT id, user_id, topic, desired_result, recommended_minutes, created_at
		FROM goals
		WHERE user_id = $1
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

func (r *Repository) StartSession(ctx context.Context, userID string, input StartSessionInput) (domain.FocusSession, error) {
	if _, err := uuid.Parse(input.GoalID); err != nil {
		return domain.FocusSession{}, fmt.Errorf("invalid goal id: %w", err)
	}

	const query = `
		INSERT INTO focus_sessions (user_id, goal_id, recommended_minutes, status)
		SELECT $1, g.id, $3, 'active'
		FROM goals g
		WHERE g.id = $2 AND g.user_id = $1
		RETURNING id, user_id, goal_id, recommended_minutes, status, pause_count, started_at, paused_at, completed_at
	`

	minutes := input.RecommendedMinutes
	if minutes == 0 {
		minutes = 25
	}

	var session domain.FocusSession
	if err := r.pool.QueryRow(ctx, query, userID, input.GoalID, minutes).Scan(
		&session.ID,
		&session.UserID,
		&session.GoalID,
		&session.RecommendedMinutes,
		&session.Status,
		&session.PauseCount,
		&session.StartedAt,
		&session.PausedAt,
		&session.CompletedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.FocusSession{}, ErrNotFound
		}
		return domain.FocusSession{}, fmt.Errorf("start session: %w", err)
	}

	return session, nil
}

func (r *Repository) PauseSession(ctx context.Context, userID, sessionID string) (domain.FocusSession, error) {
	const query = `
		UPDATE focus_sessions
		SET status = 'paused',
			paused_at = NOW(),
			pause_count = pause_count + 1
		WHERE id = $1
		  AND user_id = $2
		  AND status = 'active'
		  AND pause_count < $3
		RETURNING id, user_id, goal_id, recommended_minutes, status, pause_count, started_at, paused_at, completed_at
	`

	var session domain.FocusSession
	if err := r.pool.QueryRow(ctx, query, sessionID, userID, r.maxSessionPauses).Scan(
		&session.ID,
		&session.UserID,
		&session.GoalID,
		&session.RecommendedMinutes,
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

	return session, nil
}

func (r *Repository) ResumeSession(ctx context.Context, userID, sessionID string) (domain.FocusSession, error) {
	const query = `
		UPDATE focus_sessions
		SET status = 'active',
			paused_at = NULL
		WHERE id = $1
		  AND user_id = $2
		  AND status = 'paused'
		RETURNING id, user_id, goal_id, recommended_minutes, status, pause_count, started_at, paused_at, completed_at
	`

	var session domain.FocusSession
	if err := r.pool.QueryRow(ctx, query, sessionID, userID).Scan(
		&session.ID,
		&session.UserID,
		&session.GoalID,
		&session.RecommendedMinutes,
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
	const query = `
		UPDATE focus_sessions
		SET status = 'completed',
			completed_at = NOW(),
			paused_at = NULL
		WHERE id = $1
		  AND user_id = $2
		  AND status IN ('active', 'paused')
		RETURNING id, user_id, goal_id, recommended_minutes, status, pause_count, started_at, paused_at, completed_at
	`

	var session domain.FocusSession
	if err := r.pool.QueryRow(ctx, query, sessionID, userID).Scan(
		&session.ID,
		&session.UserID,
		&session.GoalID,
		&session.RecommendedMinutes,
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

	const markGoalQuery = `
		UPDATE goals
		SET completed_at = NOW()
		WHERE id = $1
		  AND completed_at IS NULL
	`
	if _, err := r.pool.Exec(ctx, markGoalQuery, session.GoalID); err != nil {
		return domain.FocusSession{}, fmt.Errorf("mark goal as completed: %w", err)
	}

	return session, nil
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
		SELECT COUNT(*)
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
			SELECT session_id, COUNT(*) AS interruptions
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

	overview.CalmScore = clampScore(overview.CalmScore)
	overview.FocusStability = clampScore(overview.FocusStability)

	return overview, nil
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
