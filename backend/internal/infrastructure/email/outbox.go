package email

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"mathalama-focus/backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type queuedEmail struct {
	ID         string
	ToEmail    string
	ToName     string
	VerifyLink string
	ResetLink  string
	EmailType  string
	Attempts   int
}

// OutboxService stores outgoing emails in DB and delivers them asynchronously.
type OutboxService struct {
	pool   *pgxpool.Pool
	sender interface {
		SendVerificationEmail(context.Context, string, string, string) error
		SendPasswordResetEmail(context.Context, string, string, string) error
	}
	pollEvery   time.Duration
	maxAttempts int
	startOnce   sync.Once
}

func NewOutboxService(
	pool *pgxpool.Pool,
	sender interface {
		SendVerificationEmail(context.Context, string, string, string) error
		SendPasswordResetEmail(context.Context, string, string, string) error
	},
	pollEvery time.Duration,
	maxAttempts int,
) *OutboxService {
	if pollEvery <= 0 {
		pollEvery = 2 * time.Second
	}
	if maxAttempts <= 0 {
		maxAttempts = 5
	}
	return &OutboxService{
		pool:        pool,
		sender:      sender,
		pollEvery:   pollEvery,
		maxAttempts: maxAttempts,
	}
}

func (s *OutboxService) SendVerificationEmail(ctx context.Context, toEmail, toName, verifyLink string) error {
	const query = `
		INSERT INTO email_outbox (to_email, to_name, verify_link, reset_link, email_type, status, next_attempt_at)
		VALUES ($1, $2, $3, '', 'verification', 'pending', NOW())
	`
	_, err := s.pool.Exec(ctx, query, strings.TrimSpace(toEmail), strings.TrimSpace(toName), strings.TrimSpace(verifyLink))
	if err != nil {
		return fmt.Errorf("enqueue verification email: %w", err)
	}
	return nil
}

func (s *OutboxService) SendPasswordResetEmail(ctx context.Context, toEmail, toName, resetLink string) error {
	const query = `
		INSERT INTO email_outbox (to_email, to_name, verify_link, reset_link, email_type, status, next_attempt_at)
		VALUES ($1, $2, '', $3, 'password_reset', 'pending', NOW())
	`
	_, err := s.pool.Exec(ctx, query, strings.TrimSpace(toEmail), strings.TrimSpace(toName), strings.TrimSpace(resetLink))
	if err != nil {
		return fmt.Errorf("enqueue password reset email: %w", err)
	}
	return nil
}

func (s *OutboxService) Start(ctx context.Context) {
	s.startOnce.Do(func() {
		go s.loop(ctx)
	})
}

func (s *OutboxService) loop(ctx context.Context) {
	ticker := time.NewTicker(s.pollEvery)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for i := 0; i < 50; i++ {
				processed, err := s.processOne(ctx)
				if err != nil {
					log.Printf("email outbox process error: %v", err)
					break
				}
				if !processed {
					break
				}
			}
		}
	}
}

func (s *OutboxService) processOne(ctx context.Context) (bool, error) {
	job, err := s.claimNext(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("claim next email: %w", err)
	}

	sendCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	var sendErr error
	if job.EmailType == "password_reset" {
		sendErr = s.sender.SendPasswordResetEmail(sendCtx, job.ToEmail, job.ToName, job.ResetLink)
	} else {
		sendErr = s.sender.SendVerificationEmail(sendCtx, job.ToEmail, job.ToName, job.VerifyLink)
	}
	cancel()

	if sendErr == nil {
		if err := s.markSent(ctx, job.ID); err != nil {
			return true, err
		}
		return true, nil
	}

	if job.Attempts >= s.maxAttempts {
		if err := s.markFailed(ctx, job.ID, sendErr.Error()); err != nil {
			return true, err
		}
		return true, nil
	}

	if err := s.markRetry(ctx, job.ID, job.Attempts, sendErr.Error()); err != nil {
		return true, err
	}
	return true, nil
}

func (s *OutboxService) claimNext(ctx context.Context) (queuedEmail, error) {
	const query = `
		WITH candidate AS (
			SELECT id
			FROM email_outbox
			WHERE status IN ('pending', 'retry') AND next_attempt_at <= NOW()
			ORDER BY created_at
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE email_outbox e
		SET status = 'processing',
		    attempts = e.attempts + 1,
		    updated_at = NOW()
		FROM candidate
		WHERE e.id = candidate.id
		RETURNING e.id, e.to_email, e.to_name, e.verify_link, COALESCE(e.reset_link, ''), COALESCE(e.email_type, 'verification'), e.attempts
	`

	var job queuedEmail
	err := s.pool.QueryRow(ctx, query).Scan(
		&job.ID, &job.ToEmail, &job.ToName, &job.VerifyLink, &job.ResetLink, &job.EmailType, &job.Attempts,
	)
	return job, err
}

func (s *OutboxService) markSent(ctx context.Context, jobID string) error {
	const query = `
		UPDATE email_outbox
		SET status = 'sent',
		    sent_at = NOW(),
		    last_error = NULL,
		    next_attempt_at = NOW(),
		    updated_at = NOW()
		WHERE id = $1
	`
	_, err := s.pool.Exec(ctx, query, jobID)
	if err != nil {
		return fmt.Errorf("mark email sent: %w", err)
	}
	return nil
}

func (s *OutboxService) markFailed(ctx context.Context, jobID, reason string) error {
	const query = `
		UPDATE email_outbox
		SET status = 'failed',
		    last_error = $2,
		    updated_at = NOW()
		WHERE id = $1
	`
	_, err := s.pool.Exec(ctx, query, jobID, truncateError(reason))
	if err != nil {
		return fmt.Errorf("mark email failed: %w", err)
	}
	return nil
}

func (s *OutboxService) markRetry(ctx context.Context, jobID string, attempts int, reason string) error {
	delaySeconds := 1
	for i := 1; i < attempts; i++ {
		delaySeconds *= 2
		if delaySeconds >= 300 {
			delaySeconds = 300
			break
		}
	}
	const query = `
		UPDATE email_outbox
		SET status = 'retry',
		    last_error = $2,
		    next_attempt_at = NOW() + ($3 * INTERVAL '1 second'),
		    updated_at = NOW()
		WHERE id = $1
	`
	_, err := s.pool.Exec(ctx, query, jobID, truncateError(reason), delaySeconds)
	if err != nil {
		return fmt.Errorf("mark email retry: %w", err)
	}
	return nil
}

func (s *OutboxService) ListRecentDeliveries(ctx context.Context, limit int) ([]domain.EmailDelivery, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	const query = `
		SELECT id, to_email, status, attempts, COALESCE(last_error, ''), next_attempt_at, created_at, updated_at, sent_at
		FROM email_outbox
		ORDER BY created_at DESC
		LIMIT $1
	`

	rows, err := s.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("list recent email deliveries: %w", err)
	}
	defer rows.Close()

	deliveries := make([]domain.EmailDelivery, 0, limit)
	for rows.Next() {
		var d domain.EmailDelivery
		var nextAttemptAt time.Time
		if err := rows.Scan(&d.ID, &d.ToEmail, &d.Status, &d.Attempts, &d.LastError, &nextAttemptAt, &d.CreatedAt, &d.UpdatedAt, &d.SentAt); err != nil {
			return nil, fmt.Errorf("scan email delivery: %w", err)
		}
		d.NextAttemptAt = &nextAttemptAt
		deliveries = append(deliveries, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate email deliveries: %w", err)
	}

	return deliveries, nil
}

func truncateError(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 1024 {
		return value
	}
	return value[:1024]
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
