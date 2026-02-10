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
	CreateEmailVerificationToken(ctx context.Context, userID string, ttl time.Duration) (string, time.Time, error)
	VerifyEmailByToken(ctx context.Context, rawToken string) (domain.User, error)
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

// TelegramRepository handles telegram-identity persistence.
type TelegramRepository interface {
	CreateTelegramLinkCode(ctx context.Context, userID string, ttl time.Duration) (string, time.Time, error)
	LinkTelegramByCode(ctx context.Context, input TelegramLinkInput) (domain.User, error)
	GetTelegramIdentity(ctx context.Context, userID string) (domain.TelegramIdentity, error)
	GetUserByTelegramUserID(ctx context.Context, telegramUserID int64) (domain.User, bool, error)
	SetTelegramNotifications(ctx context.Context, telegramUserID int64, enabled bool) error
	UnlinkTelegram(ctx context.Context, userID string) error
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
}
