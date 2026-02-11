package postgres

import (
	"context"
	"errors"
	"fmt"

	"mathalama-focus/backend/internal/domain"

	"github.com/jackc/pgx/v5"
)

func (r *Repository) AnalyticsOverview(ctx context.Context, userID string) (domain.AnalyticsOverview, error) {
	overview := domain.AnalyticsOverview{BestHourOfDayUTC: -1}

	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*)::INT FROM goals WHERE user_id = $1 AND completed_at IS NOT NULL`,
		userID,
	).Scan(&overview.CompletedGoals); err != nil {
		return domain.AnalyticsOverview{}, fmt.Errorf("load completed goals: %w", err)
	}

	const scoreQuery = `
		WITH interruption_counts AS (
			SELECT session_id, COUNT(*)::INT AS interruptions
			FROM interruptions
			WHERE kind = 'interruption'
			GROUP BY session_id
		),
		session_stats AS (
			SELECT
				fs.status,
				fs.pause_count,
				COALESCE(i.interruptions, 0) AS interruptions
			FROM focus_sessions fs
			LEFT JOIN interruption_counts i ON i.session_id = fs.id
			WHERE fs.user_id = $1
		)
		SELECT
			COALESCE(AVG(
				CASE
					WHEN status = 'completed' THEN 100 - (pause_count * 9 + interruptions * 12)
					ELSE 55 - (pause_count * 6 + interruptions * 8)
				END
			), 0) AS calm_score,
			COALESCE(AVG(
				CASE
					WHEN status = 'completed' THEN 100
					WHEN status = 'paused' THEN 70
					ELSE 50
				END
			), 0) AS focus_stability,
			COALESCE(AVG(
				CASE
					WHEN status = 'completed' THEN 100
					ELSE 55
				END
			), 0) AS calm_score_base,
			COALESCE(AVG(
				CASE
					WHEN status = 'completed' THEN pause_count * 9
					ELSE pause_count * 6
				END
			), 0) AS calm_score_pause_penalty,
			COALESCE(AVG(
				CASE
					WHEN status = 'completed' THEN interruptions * 12
					ELSE interruptions * 8
				END
			), 0) AS calm_score_interruption_penalty,
			COUNT(*)::INT AS sessions_total,
			COUNT(*) FILTER (WHERE status = 'completed')::INT AS completed_sessions,
			COUNT(*) FILTER (WHERE status = 'paused')::INT AS paused_sessions,
			COUNT(*) FILTER (WHERE status NOT IN ('completed', 'paused'))::INT AS other_sessions,
			COALESCE(SUM(pause_count), 0)::INT AS total_pauses,
			COALESCE(SUM(interruptions), 0)::INT AS total_interruptions
		FROM session_stats
	`
	if err := r.pool.QueryRow(ctx, scoreQuery, userID).Scan(
		&overview.CalmScore,
		&overview.FocusStability,
		&overview.CalmScoreBase,
		&overview.CalmScorePausePenalty,
		&overview.CalmScoreInterruptionPenalty,
		&overview.SessionsTotal,
		&overview.CompletedSessions,
		&overview.PausedSessions,
		&overview.OtherSessions,
		&overview.TotalPauses,
		&overview.TotalInterruptions,
	); err != nil {
		return domain.AnalyticsOverview{}, fmt.Errorf("load overview scores: %w", err)
	}

	if err := r.pool.QueryRow(ctx,
		`SELECT EXTRACT(HOUR FROM started_at AT TIME ZONE 'UTC')::INT AS hour_utc
		 FROM focus_sessions WHERE user_id = $1
		 GROUP BY hour_utc ORDER BY COUNT(*) DESC, hour_utc ASC LIMIT 1`,
		userID,
	).Scan(&overview.BestHourOfDayUTC); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return domain.AnalyticsOverview{}, fmt.Errorf("load best hour: %w", err)
	}

	if err := r.pool.QueryRow(ctx,
		`SELECT total_nectar_earned FROM users WHERE id = $1`, userID,
	).Scan(&overview.TotalNectarEarned); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.AnalyticsOverview{}, domain.ErrNotFound
		}
		return domain.AnalyticsOverview{}, fmt.Errorf("load total nectar: %w", err)
	}

	overview.CalmScore = clampScore(overview.CalmScore)
	overview.FocusStability = clampScore(overview.FocusStability)
	overview.CalmScoreBase = roundToTenth(overview.CalmScoreBase)
	overview.CalmScorePausePenalty = roundToTenth(overview.CalmScorePausePenalty)
	overview.CalmScoreInterruptionPenalty = roundToTenth(overview.CalmScoreInterruptionPenalty)

	return overview, nil
}

func (r *Repository) GetDailyActivity(ctx context.Context, userID, timezone string) ([]domain.DailyActivity, error) {
	if timezone == "" {
		timezone = "UTC"
	}

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
		var a domain.DailyActivity
		if err := rows.Scan(&a.Date, &a.SessionCount, &a.TotalMinutes); err != nil {
			return nil, fmt.Errorf("scan daily activity: %w", err)
		}
		activities = append(activities, a)
	}

	return activities, nil
}

func (r *Repository) GetDailyContributions(ctx context.Context, userID, timezone, date string) ([]domain.DailyContribution, error) {
	if timezone == "" {
		timezone = "UTC"
	}

	const query = `
		SELECT fs.id, fs.goal_id,
			COALESCE(g.topic, 'Unknown Goal') AS topic,
			fs.recommended_minutes, fs.started_at, fs.completed_at
		FROM focus_sessions fs
		LEFT JOIN goals g ON g.id = fs.goal_id
		WHERE fs.user_id = $1 AND fs.status = 'completed'
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
			&item.SessionID, &item.GoalID, &item.Topic,
			&item.Minutes, &item.StartedAt, &item.CompletedAt,
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
