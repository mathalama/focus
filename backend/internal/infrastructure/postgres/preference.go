package postgres

import (
	"context"
	"errors"
	"fmt"

	"mathalama-focus/backend/internal/domain"

	"github.com/jackc/pgx/v5"
)

func (r *Repository) GetSessionPreferences(ctx context.Context, userID string) (domain.UserSessionPreferences, error) {
	const query = `
		SELECT id, user_id, preset_durations, default_duration, short_break_duration, 
		       long_break_duration, sessions_before_long_break, default_is_strict, created_at, updated_at
		FROM user_session_preferences
		WHERE user_id = $1
	`

	var prefs domain.UserSessionPreferences
	if err := r.pool.QueryRow(ctx, query, userID).Scan(
		&prefs.ID, &prefs.UserID, &prefs.PresetDurations, &prefs.DefaultDuration,
		&prefs.ShortBreakDuration, &prefs.LongBreakDuration, &prefs.SessionsBeforeLongBreak,
		&prefs.DefaultIsStrict, &prefs.CreatedAt, &prefs.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Return default preferences
			return domain.UserSessionPreferences{
				UserID:                  userID,
				PresetDurations:         []int{25, 45, 90},
				DefaultDuration:         25,
				ShortBreakDuration:      5,
				LongBreakDuration:       15,
				SessionsBeforeLongBreak: 4,
				DefaultIsStrict:         false,
			}, nil
		}
		return domain.UserSessionPreferences{}, fmt.Errorf("get session preferences: %w", err)
	}

	return prefs, nil
}

func (r *Repository) UpdateSessionPreferences(ctx context.Context, userID string, prefs domain.UserSessionPreferences) (domain.UserSessionPreferences, error) {
	const query = `
		INSERT INTO user_session_preferences 
		(user_id, preset_durations, default_duration, short_break_duration, long_break_duration, 
		 sessions_before_long_break, default_is_strict)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id) DO UPDATE SET
			preset_durations = EXCLUDED.preset_durations,
			default_duration = EXCLUDED.default_duration,
			short_break_duration = EXCLUDED.short_break_duration,
			long_break_duration = EXCLUDED.long_break_duration,
			sessions_before_long_break = EXCLUDED.sessions_before_long_break,
			default_is_strict = EXCLUDED.default_is_strict,
			updated_at = NOW()
		RETURNING id, user_id, preset_durations, default_duration, short_break_duration,
		          long_break_duration, sessions_before_long_break, default_is_strict, created_at, updated_at
	`

	var updated domain.UserSessionPreferences
	if err := r.pool.QueryRow(ctx, query,
		userID, prefs.PresetDurations, prefs.DefaultDuration, prefs.ShortBreakDuration,
		prefs.LongBreakDuration, prefs.SessionsBeforeLongBreak, prefs.DefaultIsStrict,
	).Scan(
		&updated.ID, &updated.UserID, &updated.PresetDurations, &updated.DefaultDuration,
		&updated.ShortBreakDuration, &updated.LongBreakDuration, &updated.SessionsBeforeLongBreak,
		&updated.DefaultIsStrict, &updated.CreatedAt, &updated.UpdatedAt,
	); err != nil {
		return domain.UserSessionPreferences{}, fmt.Errorf("update session preferences: %w", err)
	}

	return updated, nil
}
