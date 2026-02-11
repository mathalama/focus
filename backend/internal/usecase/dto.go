package usecase

import (
	"time"

	"mathalama-focus/backend/internal/domain"
)

type CreateGoalInput struct {
	Topic              string
	DesiredResult      string
	RecommendedMinutes int
	Tags               []string
}

type UpdateGoalInput struct {
	Topic              string
	DesiredResult      string
	RecommendedMinutes int
	Tags               []string
}

type StartSessionInput struct {
	GoalID             string
	RecommendedMinutes int
	IsStrict           bool
}

type ReflectionInput struct {
	WhatLearned string
	WhatWasHard string
	NextAction  string
}

type TelegramLinkInput struct {
	Code             string
	TelegramUserID   int64
	TelegramUsername string
	TelegramFirst    string
	TelegramLast     string
}

type SessionHistoryFilter struct {
	Period     string
	Timezone   string
	Tags       []string
	MinMinutes int
	MaxMinutes int
	Status     string
	Limit      int
}

type RegisterOutput struct {
	User                  domain.User
	VerificationEmailSent bool
	VerificationExpiresAt time.Time
}

type ResendOutput struct {
	Resent          bool
	AlreadyVerified bool
	ExpiresAt       time.Time
}

type ForgotPasswordOutput struct {
	ResetEmailSent bool
	ExpiresAt      time.Time
}

type AuthTokens struct {
	AccessToken  string
	RefreshToken string
}
