package usecase

import (
	"context"

	"mathalama-focus/backend/internal/domain"
)

type AnalyticsUseCase struct {
	analyticsRepo AnalyticsRepository
	sessionRepo   SessionRepository
}

func NewAnalyticsUseCase(analyticsRepo AnalyticsRepository, sessionRepo SessionRepository) *AnalyticsUseCase {
	return &AnalyticsUseCase{
		analyticsRepo: analyticsRepo,
		sessionRepo:   sessionRepo,
	}
}

func (uc *AnalyticsUseCase) Overview(ctx context.Context, userID string) (domain.AnalyticsOverview, error) {
	return uc.analyticsRepo.AnalyticsOverview(ctx, userID)
}

func (uc *AnalyticsUseCase) DailyActivity(ctx context.Context, userID, timezone string) ([]domain.DailyActivity, error) {
	return uc.analyticsRepo.GetDailyActivity(ctx, userID, timezone)
}

func (uc *AnalyticsUseCase) DailyContributions(ctx context.Context, userID, timezone, date string) ([]domain.DailyContribution, error) {
	return uc.analyticsRepo.GetDailyContributions(ctx, userID, timezone, date)
}

func (uc *AnalyticsUseCase) Insights(ctx context.Context, userID string) (domain.Insight, error) {
	reflections, err := uc.sessionRepo.GetRecentReflections(ctx, userID, 5)
	if err != nil {
		return domain.Insight{}, err
	}

	insight := domain.Insight{
		Title:   "Focus Optimizer",
		Content: "Keep up the great work! You're building a consistent focus habit.",
		Type:    "encouragement",
	}

	if len(reflections) > 0 {
		hardCount := 0
		for _, r := range reflections {
			if len(r.WhatWasHard) > 20 {
				hardCount++
			}
		}

		if hardCount >= 3 {
			insight = domain.Insight{
				Title:   "Burnout Alert",
				Content: "You've been reporting high difficulty lately. Try reducing your next session to 15 minutes to reset your mental energy.",
				Type:    "warning",
			}
		} else if len(reflections) >= 2 {
			insight = domain.Insight{
				Title:   "Deep Work Insight",
				Content: "You seem to be most productive when you define clear 'Next Actions'. Try to make your next objective even more specific.",
				Type:    "tip",
			}
		}
	}

	return insight, nil
}
