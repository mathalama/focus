package postgres

import (
	"context"
	"errors"
	"fmt"

	"mathalama-focus/backend/internal/domain"

	"github.com/jackc/pgx/v5"
)

func (r *Repository) GetNotificationSound(ctx context.Context, userID string) (domain.NotificationSound, error) {
	const query = `
		SELECT id, user_id, session_complete_sound, break_end_sound, notification_sound, volume, sounds_enabled, created_at, updated_at
		FROM notification_sounds
		WHERE user_id = $1
	`

	var sound domain.NotificationSound
	if err := r.pool.QueryRow(ctx, query, userID).Scan(
		&sound.ID, &sound.UserID, &sound.SessionCompleteSound, &sound.BreakEndSound,
		&sound.NotificationSound, &sound.Volume, &sound.SoundsEnabled, &sound.CreatedAt, &sound.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Return default notification sounds
			return domain.NotificationSound{
				UserID:               userID,
				SessionCompleteSound: "/sounds/session-complete.mp3",
				BreakEndSound:        "/sounds/break-end.mp3",
				NotificationSound:    "/sounds/notification.mp3",
				Volume:               0.7,
				SoundsEnabled:        true,
			}, nil
		}
		return domain.NotificationSound{}, fmt.Errorf("get notification sound: %w", err)
	}

	return sound, nil
}

func (r *Repository) UpdateNotificationSound(ctx context.Context, userID string, sound domain.NotificationSound) (domain.NotificationSound, error) {
	const query = `
		INSERT INTO notification_sounds (user_id, session_complete_sound, break_end_sound, notification_sound, volume, sounds_enabled)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id) DO UPDATE SET
			session_complete_sound = EXCLUDED.session_complete_sound,
			break_end_sound = EXCLUDED.break_end_sound,
			notification_sound = EXCLUDED.notification_sound,
			volume = EXCLUDED.volume,
			sounds_enabled = EXCLUDED.sounds_enabled,
			updated_at = NOW()
		RETURNING id, user_id, session_complete_sound, break_end_sound, notification_sound, volume, sounds_enabled, created_at, updated_at
	`

	var result domain.NotificationSound
	if err := r.pool.QueryRow(ctx, query,
		userID,
		sound.SessionCompleteSound,
		sound.BreakEndSound,
		sound.NotificationSound,
		sound.Volume,
		sound.SoundsEnabled,
	).Scan(
		&result.ID, &result.UserID, &result.SessionCompleteSound, &result.BreakEndSound,
		&result.NotificationSound, &result.Volume, &result.SoundsEnabled, &result.CreatedAt, &result.UpdatedAt,
	); err != nil {
		return domain.NotificationSound{}, fmt.Errorf("update notification sound: %w", err)
	}

	return result, nil
}
