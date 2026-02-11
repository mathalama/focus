package usecase

import (
	"context"

	"mathalama-focus/backend/internal/domain"
)

type SessionUseCase struct {
	sessionRepo SessionRepository
}

func NewSessionUseCase(sessionRepo SessionRepository) *SessionUseCase {
	return &SessionUseCase{sessionRepo: sessionRepo}
}

func (uc *SessionUseCase) Start(ctx context.Context, userID string, input StartSessionInput) (domain.FocusSession, error) {
	return uc.sessionRepo.StartSession(ctx, userID, input)
}

func (uc *SessionUseCase) Pause(ctx context.Context, userID, sessionID string) (domain.FocusSession, error) {
	return uc.sessionRepo.PauseSession(ctx, userID, sessionID)
}

func (uc *SessionUseCase) Resume(ctx context.Context, userID, sessionID string) (domain.FocusSession, error) {
	return uc.sessionRepo.ResumeSession(ctx, userID, sessionID)
}

func (uc *SessionUseCase) Reset(ctx context.Context, userID, sessionID string) (domain.FocusSession, error) {
	return uc.sessionRepo.ResetSession(ctx, userID, sessionID)
}

func (uc *SessionUseCase) Abandon(ctx context.Context, userID, sessionID string) (domain.FocusSession, error) {
	return uc.sessionRepo.AbandonSession(ctx, userID, sessionID)
}

func (uc *SessionUseCase) Complete(ctx context.Context, userID, sessionID string) (domain.FocusSession, error) {
	return uc.sessionRepo.CompleteSession(ctx, userID, sessionID)
}

func (uc *SessionUseCase) GetActive(ctx context.Context, userID string) (domain.FocusSession, error) {
	return uc.sessionRepo.GetActiveSession(ctx, userID)
}

func (uc *SessionUseCase) Get(ctx context.Context, userID, sessionID string) (domain.FocusSession, error) {
	return uc.sessionRepo.GetSession(ctx, userID, sessionID)
}

func (uc *SessionUseCase) ListHistory(ctx context.Context, userID string, filter SessionHistoryFilter) ([]domain.SessionHistoryEntry, domain.SessionHistorySummary, error) {
	return uc.sessionRepo.ListSessionHistory(ctx, userID, filter)
}

func (uc *SessionUseCase) AddInterruption(ctx context.Context, userID, sessionID, reason string) (domain.Interruption, error) {
	return uc.sessionRepo.AddInterruption(ctx, userID, sessionID, reason)
}

func (uc *SessionUseCase) UpsertReflection(ctx context.Context, userID, sessionID string, input ReflectionInput) (domain.Reflection, error) {
	return uc.sessionRepo.UpsertReflection(ctx, userID, sessionID, input)
}

func (uc *SessionUseCase) DeleteSession(ctx context.Context, userID, sessionID string) error {
	return uc.sessionRepo.DeleteSession(ctx, userID, sessionID)
}

func (uc *SessionUseCase) DeleteReflection(ctx context.Context, userID, sessionID string) error {
	return uc.sessionRepo.DeleteReflection(ctx, userID, sessionID)
}
