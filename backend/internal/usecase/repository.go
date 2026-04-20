package usecase

import (
	"context"
	"time"

	"mathalama-focus/backend/internal/domain"
)

// ---------- repository ports ----------

// UserRepository handles user persistence and email verification tokens.
type UserRepository interface {
	GetUser(ctx context.Context, userID string) (domain.User, error)
	DevLogin(ctx context.Context, email, name string) (domain.User, error)
	RegisterUser(ctx context.Context, email, name, passwordHash string) (domain.User, error)
	GetAuthUserByEmail(ctx context.Context, email string) (domain.AuthUser, error)
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
	CreateEmailVerificationToken(ctx context.Context, userID string, ttl time.Duration) (string, time.Time, error)
	VerifyEmailByToken(ctx context.Context, rawToken string) (domain.User, error)
	CreatePasswordResetToken(ctx context.Context, userID string, ttl time.Duration) (string, time.Time, error)
	ResetPasswordByToken(ctx context.Context, rawToken, passwordHash string) (domain.User, error)
	CreateRefreshSession(ctx context.Context, userID, tokenHash, userAgent, ipAddress string, ttl time.Duration) (domain.AuthSession, error)
	RotateRefreshSession(ctx context.Context, currentTokenHash, newTokenHash, userAgent, ipAddress string, ttl time.Duration) (domain.AuthSession, error)
	RevokeRefreshSessionByTokenHash(ctx context.Context, tokenHash string) error
	RevokeAllRefreshSessions(ctx context.Context, userID string) error
	ListActiveRefreshSessions(ctx context.Context, userID string) ([]domain.AuthSession, error)
}

// GoalRepository handles goal persistence.
type GoalRepository interface {
	CreateGoal(ctx context.Context, userID string, input CreateGoalInput) (domain.Goal, error)
	ListGoals(ctx context.Context, userID string) ([]domain.Goal, error)
	ListGoalHistory(ctx context.Context, userID string) ([]domain.Goal, error)
	GetGoal(ctx context.Context, userID, goalID string) (domain.Goal, error)
	UpdateGoal(ctx context.Context, userID, goalID string, input UpdateGoalInput) (domain.Goal, error)
	DeleteGoal(ctx context.Context, userID, goalID string) error
}

// SessionRepository handles focus-session persistence.
type SessionRepository interface {
	StartSession(ctx context.Context, userID string, input StartSessionInput) (domain.FocusSession, error)
	PauseSession(ctx context.Context, userID, sessionID string) (domain.FocusSession, error)
	ResumeSession(ctx context.Context, userID, sessionID string) (domain.FocusSession, error)
	ResetSession(ctx context.Context, userID, sessionID string) (domain.FocusSession, error)
	AbandonSession(ctx context.Context, userID, sessionID string) (domain.FocusSession, error)
	CompleteSession(ctx context.Context, userID, sessionID string) (domain.FocusSession, error)
	GetActiveSession(ctx context.Context, userID string) (domain.FocusSession, error)
	GetSession(ctx context.Context, userID, sessionID string) (domain.FocusSession, error)
	ListSessionHistory(ctx context.Context, userID string, filter SessionHistoryFilter) ([]domain.SessionHistoryEntry, domain.SessionHistorySummary, error)
	AddInterruption(ctx context.Context, userID, sessionID, reason string) (domain.Interruption, error)
	UpsertReflection(ctx context.Context, userID, sessionID string, input ReflectionInput) (domain.Reflection, error)
	GetRecentReflections(ctx context.Context, userID string, limit int) ([]domain.Reflection, error)
	DeleteSession(ctx context.Context, userID, sessionID string) error
	DeleteReflection(ctx context.Context, userID, sessionID string) error
}

// AnalyticsRepository handles analytics queries.
type AnalyticsRepository interface {
	AnalyticsOverview(ctx context.Context, userID string) (domain.AnalyticsOverview, error)
	GetDailyActivity(ctx context.Context, userID, timezone string) ([]domain.DailyActivity, error)
	GetDailyContributions(ctx context.Context, userID, timezone, date string) ([]domain.DailyContribution, error)
}

// ShopRepository handles shop and leaderboard persistence.
type ShopRepository interface {
	ListItems(ctx context.Context) ([]domain.Item, error)
	BuyItem(ctx context.Context, userID, itemID string) (domain.UserItem, error)
	GetLeaderboard(ctx context.Context) ([]domain.LeaderboardEntry, error)
}


// NotificationRepository handles notification sound preferences.
type NotificationRepository interface {
	GetNotificationSound(ctx context.Context, userID string) (domain.NotificationSound, error)
	UpdateNotificationSound(ctx context.Context, userID string, sound domain.NotificationSound) (domain.NotificationSound, error)
}

// SessionPreferencesRepository handles user session customization.
type SessionPreferencesRepository interface {
	GetSessionPreferences(ctx context.Context, userID string) (domain.UserSessionPreferences, error)
	UpdateSessionPreferences(ctx context.Context, userID string, prefs domain.UserSessionPreferences) (domain.UserSessionPreferences, error)
}

// NotificationScheduleRepository handles notification reminders.
type NotificationScheduleRepository interface {
	GetNotificationSchedules(ctx context.Context, userID string) ([]domain.NotificationSchedule, error)
	CreateNotificationSchedule(ctx context.Context, userID string, schedule domain.NotificationSchedule) (domain.NotificationSchedule, error)
	UpdateNotificationSchedule(ctx context.Context, userID, scheduleID string, schedule domain.NotificationSchedule) (domain.NotificationSchedule, error)
	DeleteNotificationSchedule(ctx context.Context, userID, scheduleID string) error
	GetSchedulesReadyToSend(ctx context.Context) ([]domain.NotificationSchedule, error)
	UpdateLastSentAt(ctx context.Context, scheduleID string) error
}

// ---------- service ports ----------

// TokenService generates and validates JWT tokens.
type TokenService interface {
	GenerateToken(userID string) (string, error)
	ValidateToken(tokenString string) (string, error)
}

// EmailService sends transactional emails.
type EmailService interface {
	SendVerificationEmail(ctx context.Context, toEmail, toName, verifyLink string) error
	SendPasswordResetEmail(ctx context.Context, toEmail, toName, resetLink string) error
}
