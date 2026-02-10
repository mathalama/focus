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
	overview, err := uc.analyticsRepo.AnalyticsOverview(ctx, userID)
	if err != nil {
		return domain.AnalyticsOverview{}, err
	}
	overview.PrimaryAction = pickPrimaryAction(overview)
	return overview, nil
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

func pickPrimaryAction(overview domain.AnalyticsOverview) string {
	if overview.SessionsTotal == 0 {
		return "start_sessions"
	}

	completionRate := float64(overview.CompletedSessions) / float64(maxInt(overview.SessionsTotal, 1))
	if completionRate < 0.6 {
		return "complete_more_sessions"
	}

	if overview.TotalPauses > 0 && overview.CalmScorePausePenalty >= overview.CalmScoreInterruptionPenalty {
		return "reduce_pauses"
	}

	if overview.TotalInterruptions > 0 {
		return "reduce_interruptions"
	}

	if overview.FocusStability < 75 || overview.PausedSessions*2 > maxInt(overview.CompletedSessions, 1) {
		return "stabilize_schedule"
	}

	return "keep_momentum"
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
