package domain

import "errors"

var (
	ErrNotFound              = errors.New("resource not found")
	ErrPauseLimitReached     = errors.New("pause limit reached")
	ErrInvalidState          = errors.New("session is in invalid state for this action")
	ErrGoalCompleted         = errors.New("goal already completed")
	ErrTelegramCodeInvalid   = errors.New("telegram link code invalid or expired")
	ErrTelegramAlreadyLinked = errors.New("telegram account already linked")
	ErrAlreadyExists         = errors.New("resource already exists")
	ErrEmailTokenInvalid     = errors.New("email verification token invalid or expired")
	ErrInvalidEmail          = errors.New("invalid email address")
	ErrWeakPassword          = errors.New("password must be between 8 and 72 characters")
	ErrEmailNotVerified      = errors.New("email is not verified")
	ErrInvalidCredentials    = errors.New("invalid email or password")
	ErrEmailNotConfigured    = errors.New("email verification is not configured")
	ErrRefreshTokenInvalid   = errors.New("refresh token is invalid or expired")
	ErrAuthUnavailable       = errors.New("authentication service is temporarily unavailable")
	ErrForbidden             = errors.New("forbidden")
	ErrInsufficientBalance   = errors.New("insufficient nectar balance")
)
