package usecase

import (
	"context"

	"mathalama-focus/backend/internal/domain"
)

type NotificationUseCase struct {
	notificationRepo NotificationRepository
}

func NewNotificationUseCase(notificationRepo NotificationRepository) *NotificationUseCase {
	return &NotificationUseCase{
		notificationRepo: notificationRepo,
	}
}

func (uc *NotificationUseCase) GetSoundPreferences(ctx context.Context, userID string) (domain.NotificationSound, error) {
	return uc.notificationRepo.GetNotificationSound(ctx, userID)
}

type UpdateSoundPreferencesInput struct {
	SessionCompleteSound string  `json:"session_complete_sound"`
	BreakEndSound        string  `json:"break_end_sound"`
	NotificationSound    string  `json:"notification_sound"`
	Volume               float64 `json:"volume"`
	SoundsEnabled        bool    `json:"sounds_enabled"`
}

func (uc *NotificationUseCase) UpdateSoundPreferences(ctx context.Context, userID string, input UpdateSoundPreferencesInput) (domain.NotificationSound, error) {
	// Validate volume
	if input.Volume < 0 || input.Volume > 1 {
		input.Volume = 0.7
	}

	sound := domain.NotificationSound{
		UserID:               userID,
		SessionCompleteSound: input.SessionCompleteSound,
		BreakEndSound:        input.BreakEndSound,
		NotificationSound:    input.NotificationSound,
		Volume:               input.Volume,
		SoundsEnabled:        input.SoundsEnabled,
	}

	return uc.notificationRepo.UpdateNotificationSound(ctx, userID, sound)
}
