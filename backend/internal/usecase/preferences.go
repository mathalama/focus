package usecase

import (
	"context"

	"mathalama-focus/backend/internal/domain"
)

type PreferencesUseCase struct {
	sessionPrefsRepo SessionPreferencesRepository
	scheduleRepo     NotificationScheduleRepository
}

func NewPreferencesUseCase(
	sessionPrefsRepo SessionPreferencesRepository,
	scheduleRepo NotificationScheduleRepository,
) *PreferencesUseCase {
	return &PreferencesUseCase{
		sessionPrefsRepo: sessionPrefsRepo,
		scheduleRepo:     scheduleRepo,
	}
}

// GetSessionPreferences retrieves user's session customization settings
func (uc *PreferencesUseCase) GetSessionPreferences(ctx context.Context, userID string) (domain.UserSessionPreferences, error) {
	return uc.sessionPrefsRepo.GetSessionPreferences(ctx, userID)
}

// UpdateSessionPreferences updates user's session customization settings
func (uc *PreferencesUseCase) UpdateSessionPreferences(ctx context.Context, userID string, prefs domain.UserSessionPreferences) (domain.UserSessionPreferences, error) {
	prefs.UserID = userID
	return uc.sessionPrefsRepo.UpdateSessionPreferences(ctx, userID, prefs)
}

// GetNotificationSchedules retrieves all notification schedules for user
func (uc *PreferencesUseCase) GetNotificationSchedules(ctx context.Context, userID string) ([]domain.NotificationSchedule, error) {
	return uc.scheduleRepo.GetNotificationSchedules(ctx, userID)
}

// CreateNotificationSchedule creates a new notification schedule
func (uc *PreferencesUseCase) CreateNotificationSchedule(ctx context.Context, userID string, schedule domain.NotificationSchedule) (domain.NotificationSchedule, error) {
	schedule.UserID = userID
	return uc.scheduleRepo.CreateNotificationSchedule(ctx, userID, schedule)
}

// UpdateNotificationSchedule updates existing notification schedule
func (uc *PreferencesUseCase) UpdateNotificationSchedule(ctx context.Context, userID, scheduleID string, schedule domain.NotificationSchedule) (domain.NotificationSchedule, error) {
	schedule.UserID = userID
	return uc.scheduleRepo.UpdateNotificationSchedule(ctx, userID, scheduleID, schedule)
}

// DeleteNotificationSchedule deletes notification schedule
func (uc *PreferencesUseCase) DeleteNotificationSchedule(ctx context.Context, userID, scheduleID string) error {
	return uc.scheduleRepo.DeleteNotificationSchedule(ctx, userID, scheduleID)
}
