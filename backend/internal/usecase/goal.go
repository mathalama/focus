package usecase

import (
	"context"

	"mathalama-focus/backend/internal/domain"
)

// GoalUseCase orchestrates goal business logic.
type GoalUseCase struct {
	goalRepo GoalRepository
}

func NewGoalUseCase(goalRepo GoalRepository) *GoalUseCase {
	return &GoalUseCase{goalRepo: goalRepo}
}

func (uc *GoalUseCase) Create(ctx context.Context, userID string, input CreateGoalInput) (domain.Goal, error) {
	return uc.goalRepo.CreateGoal(ctx, userID, input)
}

func (uc *GoalUseCase) List(ctx context.Context, userID string) ([]domain.Goal, error) {
	return uc.goalRepo.ListGoals(ctx, userID)
}

func (uc *GoalUseCase) ListHistory(ctx context.Context, userID string) ([]domain.Goal, error) {
	return uc.goalRepo.ListGoalHistory(ctx, userID)
}
