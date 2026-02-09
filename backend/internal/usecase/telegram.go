package usecase

import (
	"context"
	"time"

	"mathalama-focus/backend/internal/domain"
)

// TelegramUseCase orchestrates telegram-integration business logic.
type TelegramUseCase struct {
	telegramRepo TelegramRepository
	linkCodeTTL  time.Duration
}

func NewTelegramUseCase(telegramRepo TelegramRepository, linkCodeTTL time.Duration) *TelegramUseCase {
	if linkCodeTTL <= 0 {
		linkCodeTTL = 10 * time.Minute
	}
	return &TelegramUseCase{
		telegramRepo: telegramRepo,
		linkCodeTTL:  linkCodeTTL,
	}
}

func (uc *TelegramUseCase) CreateLinkCode(ctx context.Context, userID string) (string, time.Time, error) {
	return uc.telegramRepo.CreateTelegramLinkCode(ctx, userID, uc.linkCodeTTL)
}

func (uc *TelegramUseCase) GetIdentity(ctx context.Context, userID string) (domain.TelegramIdentity, error) {
	return uc.telegramRepo.GetTelegramIdentity(ctx, userID)
}

func (uc *TelegramUseCase) Unlink(ctx context.Context, userID string) error {
	return uc.telegramRepo.UnlinkTelegram(ctx, userID)
}

func (uc *TelegramUseCase) LinkByCode(ctx context.Context, input TelegramLinkInput) (domain.User, error) {
	return uc.telegramRepo.LinkTelegramByCode(ctx, input)
}

func (uc *TelegramUseCase) GetStatus(ctx context.Context, telegramUserID int64) (domain.User, bool, error) {
	return uc.telegramRepo.GetUserByTelegramUserID(ctx, telegramUserID)
}

func (uc *TelegramUseCase) SetNotifications(ctx context.Context, telegramUserID int64, enabled bool) error {
	return uc.telegramRepo.SetTelegramNotifications(ctx, telegramUserID, enabled)
}
