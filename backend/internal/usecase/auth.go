package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/mail"
	"net/url"
	"strings"
	"time"

	"mathalama-focus/backend/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

// AuthUseCase orchestrates authentication and user-management business logic.
type AuthUseCase struct {
	userRepo   UserRepository
	tokenSvc   TokenService
	emailSvc   EmailService
	verifyTTL  time.Duration
	verifyURL  string
	successURL string
	failureURL string
}

func NewAuthUseCase(
	userRepo UserRepository,
	tokenSvc TokenService,
	emailSvc EmailService,
	verifyTTL time.Duration,
	verifyURL, successURL, failureURL string,
) *AuthUseCase {
	if verifyTTL <= 0 {
		verifyTTL = 60 * time.Minute
	}
	return &AuthUseCase{
		userRepo:   userRepo,
		tokenSvc:   tokenSvc,
		emailSvc:   emailSvc,
		verifyTTL:  verifyTTL,
		verifyURL:  strings.TrimRight(strings.TrimSpace(verifyURL), "/"),
		successURL: strings.TrimSpace(successURL),
		failureURL: strings.TrimSpace(failureURL),
	}
}

func (uc *AuthUseCase) emailConfigured() bool {
	return uc.emailSvc != nil && strings.TrimSpace(uc.verifyURL) != ""
}

// SuccessRedirectURL returns the configured success redirect URL for email verification.
func (uc *AuthUseCase) SuccessRedirectURL() string { return uc.successURL }

// FailureRedirectURL returns the configured failure redirect URL for email verification.
func (uc *AuthUseCase) FailureRedirectURL() string { return uc.failureURL }

// GetUser returns the user profile by ID.
func (uc *AuthUseCase) GetUser(ctx context.Context, userID string) (domain.User, error) {
	return uc.userRepo.GetUser(ctx, userID)
}

// DevLogin performs a development-only upsert login (no password).
func (uc *AuthUseCase) DevLogin(ctx context.Context, email, name string) (domain.User, string, error) {
	email = strings.TrimSpace(email)
	name = strings.TrimSpace(name)

	if email == "" {
		return domain.User{}, "", domain.ErrInvalidEmail
	}
	if name == "" {
		name = "Focus Learner"
	}

	user, err := uc.userRepo.DevLogin(ctx, email, name)
	if err != nil {
		return domain.User{}, "", err
	}

	token, err := uc.tokenSvc.GenerateToken(user.ID)
	if err != nil {
		return domain.User{}, "", fmt.Errorf("generate token: %w", err)
	}

	return user, token, nil
}

// Register creates a new user with email/password, sends verification email.
func (uc *AuthUseCase) Register(ctx context.Context, rawEmail, name, password string) (RegisterOutput, error) {
	if !uc.emailConfigured() {
		return RegisterOutput{}, domain.ErrEmailNotConfigured
	}

	email, err := normalizeEmail(rawEmail)
	if err != nil {
		return RegisterOutput{}, domain.ErrInvalidEmail
	}

	if len(password) < 8 {
		return RegisterOutput{}, domain.ErrWeakPassword
	}
	if len(password) > 72 {
		return RegisterOutput{}, domain.ErrWeakPassword
	}

	name = strings.TrimSpace(name)
	if name == "" {
		name = "Focus Learner"
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return RegisterOutput{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := uc.userRepo.RegisterUser(ctx, email, name, string(passwordHash))
	if err != nil {
		return RegisterOutput{}, err
	}

	rawToken, expiresAt, err := uc.userRepo.CreateEmailVerificationToken(ctx, user.ID, uc.verifyTTL)
	if err != nil {
		return RegisterOutput{}, fmt.Errorf("create verification token: %w", err)
	}

	verifyLink, err := uc.buildVerifyLink(rawToken)
	if err != nil {
		return RegisterOutput{}, fmt.Errorf("build verify link: %w", err)
	}

	emailSent := true
	if err := uc.emailSvc.SendVerificationEmail(ctx, user.Email, user.Name, verifyLink); err != nil {
		emailSent = false
		log.Printf("register verification email send failed (user=%s): %v", user.ID, err)
	}

	return RegisterOutput{
		User:                  user,
		VerificationEmailSent: emailSent,
		VerificationExpiresAt: expiresAt.UTC(),
	}, nil
}

// Login authenticates a user by email and password.
func (uc *AuthUseCase) Login(ctx context.Context, rawEmail, password string) (domain.User, string, error) {
	email, err := normalizeEmail(rawEmail)
	if err != nil {
		return domain.User{}, "", domain.ErrInvalidEmail
	}

	authUser, err := uc.userRepo.GetAuthUserByEmail(ctx, email)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.User{}, "", domain.ErrInvalidCredentials
	}
	if err != nil {
		return domain.User{}, "", err
	}

	if strings.TrimSpace(authUser.PasswordHash) == "" ||
		bcrypt.CompareHashAndPassword([]byte(authUser.PasswordHash), []byte(password)) != nil {
		return domain.User{}, "", domain.ErrInvalidCredentials
	}

	if authUser.EmailVerifiedAt == nil {
		return domain.User{}, "", domain.ErrEmailNotVerified
	}

	token, err := uc.tokenSvc.GenerateToken(authUser.User.ID)
	if err != nil {
		return domain.User{}, "", fmt.Errorf("generate token: %w", err)
	}

	return authUser.User, token, nil
}

// ResendVerification re-sends the verification email.
func (uc *AuthUseCase) ResendVerification(ctx context.Context, rawEmail string) (ResendOutput, error) {
	if !uc.emailConfigured() {
		return ResendOutput{}, domain.ErrEmailNotConfigured
	}

	email, err := normalizeEmail(rawEmail)
	if err != nil {
		return ResendOutput{}, domain.ErrInvalidEmail
	}

	authUser, err := uc.userRepo.GetAuthUserByEmail(ctx, email)
	if errors.Is(err, domain.ErrNotFound) {
		return ResendOutput{Resent: true}, nil // don't reveal whether email exists
	}
	if err != nil {
		return ResendOutput{}, err
	}

	if authUser.EmailVerifiedAt != nil {
		return ResendOutput{Resent: false, AlreadyVerified: true}, nil
	}

	rawToken, expiresAt, err := uc.userRepo.CreateEmailVerificationToken(ctx, authUser.User.ID, uc.verifyTTL)
	if err != nil {
		return ResendOutput{}, err
	}

	verifyLink, err := uc.buildVerifyLink(rawToken)
	if err != nil {
		return ResendOutput{}, err
	}

	if err := uc.emailSvc.SendVerificationEmail(ctx, authUser.User.Email, authUser.User.Name, verifyLink); err != nil {
		log.Printf("resend verification email send failed (user=%s): %v", authUser.User.ID, err)
		return ResendOutput{}, fmt.Errorf("send verification email: %w", err)
	}

	return ResendOutput{Resent: true, ExpiresAt: expiresAt.UTC()}, nil
}

// VerifyEmail verifies the user's email by the one-time token.
func (uc *AuthUseCase) VerifyEmail(ctx context.Context, rawToken string) (domain.User, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return domain.User{}, domain.ErrEmailTokenInvalid
	}
	return uc.userRepo.VerifyEmailByToken(ctx, rawToken)
}

// ---------- helpers ----------

func (uc *AuthUseCase) buildVerifyLink(token string) (string, error) {
	base := strings.TrimSpace(uc.verifyURL)
	if base == "" {
		return "", errors.New("email verify url base is not configured")
	}

	parsed, err := url.Parse(base)
	if err != nil {
		return "", err
	}

	query := parsed.Query()
	query.Set("token", strings.TrimSpace(token))
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func normalizeEmail(raw string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(raw))
	if email == "" {
		return "", errors.New("email is required")
	}
	if strings.ContainsAny(email, " \t\r\n") {
		return "", errors.New("email must not contain spaces")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return "", err
	}
	return email, nil
}
