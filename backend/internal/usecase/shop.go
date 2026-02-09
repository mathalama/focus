package usecase

import (
	"context"

	"mathalama-focus/backend/internal/domain"
)

// ShopUseCase orchestrates shop and leaderboard business logic.
type ShopUseCase struct {
	shopRepo ShopRepository
}

func NewShopUseCase(shopRepo ShopRepository) *ShopUseCase {
	return &ShopUseCase{shopRepo: shopRepo}
}

func (uc *ShopUseCase) ListItems(ctx context.Context) ([]domain.Item, error) {
	return uc.shopRepo.ListItems(ctx)
}

func (uc *ShopUseCase) BuyItem(ctx context.Context, userID, itemID string) (domain.UserItem, error) {
	return uc.shopRepo.BuyItem(ctx, userID, itemID)
}

func (uc *ShopUseCase) Leaderboard(ctx context.Context) ([]domain.LeaderboardEntry, error) {
	return uc.shopRepo.GetLeaderboard(ctx)
}
