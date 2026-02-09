package postgres

import (
	"context"
	"errors"
	"fmt"

	"mathalama-focus/backend/internal/domain"
	"mathalama-focus/backend/internal/usecase"

	"github.com/jackc/pgx/v5"
)

func (r *Repository) UpsertReflection(ctx context.Context, userID, sessionID string, input usecase.ReflectionInput) (domain.Reflection, error) {
	if err := r.pool.QueryRow(ctx,
		`SELECT 1 FROM focus_sessions WHERE id = $1 AND user_id = $2`,
		sessionID, userID,
	).Scan(new(int)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Reflection{}, domain.ErrNotFound
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
		&reflection.ID, &reflection.SessionID,
		&reflection.WhatLearned, &reflection.WhatWasHard,
		&reflection.NextAction, &reflection.CreatedAt,
	); err != nil {
		return domain.Reflection{}, fmt.Errorf("upsert reflection: %w", err)
	}

	return reflection, nil
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
		var r domain.Reflection
		if err := rows.Scan(
			&r.ID, &r.SessionID, &r.WhatLearned,
			&r.WhatWasHard, &r.NextAction, &r.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan reflection: %w", err)
		}
		reflections = append(reflections, r)
	}

	return reflections, nil
}
