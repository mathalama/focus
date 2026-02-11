package postgres

import (
	"context"
	"errors"
	"fmt"

	"mathalama-focus/backend/internal/domain"
	"github.com/jackc/pgx/v5"
)

func (r *Repository) GetNotificationSchedules(ctx context.Context, userID string) ([]domain.NotificationSchedule, error) {
	const query = `
		SELECT id, user_id, enabled, timezone, days_of_week, reminder_time, notification_type, last_sent_at, created_at, updated_at
		FROM notification_schedules
		WHERE user_id = $1
		ORDER BY created_at
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query notification schedules: %w", err)
	}
	defer rows.Close()

	var schedules []domain.NotificationSchedule
	for rows.Next() {
		var schedule domain.NotificationSchedule
		if err := rows.Scan(
			&schedule.ID, &schedule.UserID, &schedule.Enabled, &schedule.Timezone,
			&schedule.DaysOfWeek, &schedule.ReminderTime, &schedule.NotificationType,
			&schedule.LastSentAt, &schedule.CreatedAt, &schedule.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan notification schedule: %w", err)
		}
		schedules = append(schedules, schedule)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return schedules, nil
}

func (r *Repository) CreateNotificationSchedule(ctx context.Context, userID string, schedule domain.NotificationSchedule) (domain.NotificationSchedule, error) {
	const query = `
		INSERT INTO notification_schedules 
		(user_id, enabled, timezone, days_of_week, reminder_time, notification_type)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, enabled, timezone, days_of_week, reminder_time, notification_type, last_sent_at, created_at, updated_at
	`

	var created domain.NotificationSchedule
	if err := r.pool.QueryRow(ctx, query,
		userID, schedule.Enabled, schedule.Timezone, schedule.DaysOfWeek,
		schedule.ReminderTime, schedule.NotificationType,
	).Scan(
		&created.ID, &created.UserID, &created.Enabled, &created.Timezone,
		&created.DaysOfWeek, &created.ReminderTime, &created.NotificationType,
		&created.LastSentAt, &created.CreatedAt, &created.UpdatedAt,
	); err != nil {
		return domain.NotificationSchedule{}, fmt.Errorf("create notification schedule: %w", err)
	}

	return created, nil
}

func (r *Repository) UpdateNotificationSchedule(ctx context.Context, userID, scheduleID string, schedule domain.NotificationSchedule) (domain.NotificationSchedule, error) {
	const query = `
		UPDATE notification_schedules
		SET enabled = $1, timezone = $2, days_of_week = $3, reminder_time = $4, notification_type = $5, updated_at = NOW()
		WHERE id = $6 AND user_id = $7
		RETURNING id, user_id, enabled, timezone, days_of_week, reminder_time, notification_type, last_sent_at, created_at, updated_at
	`

	var updated domain.NotificationSchedule
	if err := r.pool.QueryRow(ctx, query,
		schedule.Enabled, schedule.Timezone, schedule.DaysOfWeek,
		schedule.ReminderTime, schedule.NotificationType, scheduleID, userID,
	).Scan(
		&updated.ID, &updated.UserID, &updated.Enabled, &updated.Timezone,
		&updated.DaysOfWeek, &updated.ReminderTime, &updated.NotificationType,
		&updated.LastSentAt, &updated.CreatedAt, &updated.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.NotificationSchedule{}, fmt.Errorf("notification schedule not found")
		}
		return domain.NotificationSchedule{}, fmt.Errorf("update notification schedule: %w", err)
	}

	return updated, nil
}

func (r *Repository) DeleteNotificationSchedule(ctx context.Context, userID, scheduleID string) error {
	const query = `DELETE FROM notification_schedules WHERE id = $1 AND user_id = $2`
	
	result, err := r.pool.Exec(ctx, query, scheduleID, userID)
	if err != nil {
		return fmt.Errorf("delete notification schedule: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("notification schedule not found")
	}

	return nil
}

// GetSchedulesReadyToSend gets all schedules that need sending now (for background job)
func (r *Repository) GetSchedulesReadyToSend(ctx context.Context) ([]domain.NotificationSchedule, error) {
	const query = `
		SELECT id, user_id, enabled, timezone, days_of_week, reminder_time, notification_type, last_sent_at, created_at, updated_at
		FROM notification_schedules
		WHERE enabled = true
		AND (last_sent_at IS NULL OR last_sent_at < NOW() - INTERVAL '1 day')
		ORDER BY created_at
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query schedules to send: %w", err)
	}
	defer rows.Close()

	var schedules []domain.NotificationSchedule
	for rows.Next() {
		var schedule domain.NotificationSchedule
		if err := rows.Scan(
			&schedule.ID, &schedule.UserID, &schedule.Enabled, &schedule.Timezone,
			&schedule.DaysOfWeek, &schedule.ReminderTime, &schedule.NotificationType,
			&schedule.LastSentAt, &schedule.CreatedAt, &schedule.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan schedule: %w", err)
		}
		schedules = append(schedules, schedule)
	}

	return schedules, rows.Err()
}

// UpdateLastSentAt updates the last_sent_at timestamp
func (r *Repository) UpdateLastSentAt(ctx context.Context, scheduleID string) error {
	const query = `UPDATE notification_schedules SET last_sent_at = NOW() WHERE id = $1`
	
	_, err := r.pool.Exec(ctx, query, scheduleID)
	return err
}
