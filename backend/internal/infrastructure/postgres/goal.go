package postgres

import (
	"context"
	"errors"
	"fmt"

	"mathalama-focus/backend/internal/domain"
	"mathalama-focus/backend/internal/usecase"

	"github.com/jackc/pgx/v5/pgconn"
)

func (r *Repository) CreateGoal(ctx context.Context, userID string, input usecase.CreateGoalInput) (domain.Goal, error) {
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
		&goal.ID, &goal.UserID, &goal.Topic, &goal.DesiredResult,
		&goal.RecommendedMinutes, &goal.Tags, &goal.CompletedAt, &goal.CreatedAt,
	); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return domain.Goal{}, domain.ErrNotFound
		}
		return domain.Goal{}, fmt.Errorf("create goal: %w", err)
	}

	return goal, nil
}

func (r *Repository) ListGoals(ctx context.Context, userID string) ([]domain.Goal, error) {
	const query = `
		SELECT id, user_id, topic, desired_result, recommended_minutes, tags, completed_at, created_at
		FROM goals
		WHERE user_id = $1 AND completed_at IS NULL
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
			&goal.ID, &goal.UserID, &goal.Topic, &goal.DesiredResult,
			&goal.RecommendedMinutes, &goal.Tags, &goal.CompletedAt, &goal.CreatedAt,
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
		WHERE user_id = $1 AND completed_at IS NOT NULL
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
			&goal.ID, &goal.UserID, &goal.Topic, &goal.DesiredResult,
			&goal.RecommendedMinutes, &goal.Tags, &goal.CompletedAt, &goal.CreatedAt,
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
