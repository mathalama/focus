package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/mail"
	"net/url"
	"strings"
	"time"

	"mathalama-focus/backend/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

var trustedDomains = map[string]bool{
	// Global
	"gmail.com":      true,
	"googlemail.com": true,
	"outlook.com":    true,
	"hotmail.com":    true,
	"live.com":       true,
	"icloud.com":     true,
	"me.com":          true,
	"yahoo.com":      true,
	"proton.me":      true,
	"protonmail.com": true,
	"zoho.com":       true,
	"gmx.com":        true,
	"aol.com":        true,

	// CIS / SNG
	"mail.ru":    true,
	"yandex.ru":  true,
	"yandex.kz":  true,
	"yandex.by":  true,
	"yandex.com": true,
	"ya.ru":      true,
	"list.ru":    true,
	"bk.ru":      true,
	"inbox.ru":   true,
	"rambler.ru": true,
	"mathalama.dev": true,
}

// AuthUseCase orchestrates authentication and user-management business logic.
type AuthUseCase struct {
	userRepo   UserRepository
	tokenSvc   TokenService
	emailSvc   EmailService
	verifyTTL  time.Duration
	refreshTTL time.Duration
	verifyURL  string
	resetURL   string
	successURL string
	failureURL string
}

func NewAuthUseCase(
	userRepo UserRepository,
	tokenSvc TokenService,
	emailSvc EmailService,
	verifyTTL time.Duration,
	refreshTTL time.Duration,
	verifyURL, resetURL, successURL, failureURL string,
) *AuthUseCase {
	if verifyTTL <= 0 {
		verifyTTL = 60 * time.Minute
	}
	if refreshTTL <= 0 {
		refreshTTL = 30 * 24 * time.Hour
	}
	return &AuthUseCase{
		userRepo:   userRepo,
		tokenSvc:   tokenSvc,
		emailSvc:   emailSvc,
		verifyTTL:  verifyTTL,
		refreshTTL: refreshTTL,
		verifyURL:  strings.TrimRight(strings.TrimSpace(verifyURL), "/"),
		resetURL:   strings.TrimRight(strings.TrimSpace(resetURL), "/"),
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
func (uc *AuthUseCase) DevLogin(ctx context.Context, email, name, userAgent, ipAddress string) (domain.User, AuthTokens, error) {
	email = strings.TrimSpace(email)
	name = strings.TrimSpace(name)

	if email == "" {
		return domain.User{}, AuthTokens{}, domain.ErrInvalidEmail
	}
	if name == "" {
		name = "Focus Learner"
	}

	user, err := uc.userRepo.DevLogin(ctx, email, name)
	if err != nil {
		return domain.User{}, AuthTokens{}, err
	}

	tokens, err := uc.issueAuthTokens(ctx, user, userAgent, ipAddress)
	if err != nil {
		return domain.User{}, AuthTokens{}, err
	}

	return user, tokens, nil
}

// Register creates a new user with email/password.
// Email verification is optional (sent but not required to login).
func (uc *AuthUseCase) Register(ctx context.Context, rawEmail, name, password string) (RegisterOutput, error) {
	email, err := normalizeEmail(rawEmail)
	if err != nil {
		return RegisterOutput{}, err
	}

	if !isTrustedDomain(email) {
		return RegisterOutput{}, errors.New("only trusted email providers are allowed (Gmail, Mail.ru, Yandex, etc.)")
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

	// Try to send verification email, but don't fail registration if it doesn't work
	emailSent := false
	expiresAt := time.Time{}

	if uc.emailConfigured() {
		rawToken, tokenExpiresAt, err := uc.userRepo.CreateEmailVerificationToken(ctx, user.ID, uc.verifyTTL)
		if err == nil {
			verifyLink, err := uc.buildVerifyLink(rawToken)
			if err == nil {
				if err := uc.emailSvc.SendVerificationEmail(ctx, user.Email, user.Name, verifyLink); err == nil {
					emailSent = true
					expiresAt = tokenExpiresAt.UTC()
				} else {
					slog.Error("register verification email send failed", "user_id", user.ID, "error", err)
				}
			}
		}
	}

	return RegisterOutput{
		User:                  user,
		VerificationEmailSent: emailSent,
		VerificationExpiresAt: expiresAt,
	}, nil
}

// Login authenticates a user by email and password.
func (uc *AuthUseCase) Login(ctx context.Context, rawEmail, password, userAgent, ipAddress string) (domain.User, AuthTokens, error) {
	email, err := normalizeEmail(rawEmail)
	if err != nil {
		return domain.User{}, AuthTokens{}, domain.ErrInvalidEmail
	}

	authUser, err := uc.userRepo.GetAuthUserByEmail(ctx, email)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.User{}, AuthTokens{}, domain.ErrInvalidCredentials
	}
	if err != nil {
		return domain.User{}, AuthTokens{}, err
	}

	if strings.TrimSpace(authUser.PasswordHash) == "" ||
		bcrypt.CompareHashAndPassword([]byte(authUser.PasswordHash), []byte(password)) != nil {
		return domain.User{}, AuthTokens{}, domain.ErrInvalidCredentials
	}

	if authUser.EmailVerifiedAt == nil {
		return domain.User{}, AuthTokens{}, domain.ErrEmailNotVerified
	}

	tokens, err := uc.issueAuthTokens(ctx, authUser.User, userAgent, ipAddress)
	if err != nil {
		return domain.User{}, AuthTokens{}, err
	}

	return authUser.User, tokens, nil
}

// Refresh rotates the refresh token and issues a fresh access token.
func (uc *AuthUseCase) Refresh(ctx context.Context, rawRefreshToken, userAgent, ipAddress string) (domain.User, AuthTokens, error) {
	rawRefreshToken = strings.TrimSpace(rawRefreshToken)
	if rawRefreshToken == "" {
		return domain.User{}, AuthTokens{}, domain.ErrRefreshTokenInvalid
	}

	oldTokenHash := hashToken(rawRefreshToken)
	newRawRefreshToken, err := generateRawToken(32)
	if err != nil {
		return domain.User{}, AuthTokens{}, fmt.Errorf("generate refresh token: %w", err)
	}
	newTokenHash := hashToken(newRawRefreshToken)

	session, err := uc.userRepo.RotateRefreshSession(ctx, oldTokenHash, newTokenHash, userAgent, ipAddress, uc.refreshTTL)
	if err != nil {
		if errors.Is(err, domain.ErrRefreshTokenInvalid) {
			return domain.User{}, AuthTokens{}, domain.ErrRefreshTokenInvalid
		}
		return domain.User{}, AuthTokens{}, fmt.Errorf("rotate refresh session: %w", err)
	}

	user, err := uc.userRepo.GetUser(ctx, session.UserID)
	if err != nil {
		return domain.User{}, AuthTokens{}, fmt.Errorf("load user for refresh session: %w", err)
	}

	accessToken, err := uc.tokenSvc.GenerateToken(user.ID)
	if err != nil {
		return domain.User{}, AuthTokens{}, fmt.Errorf("generate access token: %w", err)
	}

	return user, AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: newRawRefreshToken,
	}, nil
}

// Logout revokes a single refresh session represented by the provided refresh token.
func (uc *AuthUseCase) Logout(ctx context.Context, rawRefreshToken string) error {
	rawRefreshToken = strings.TrimSpace(rawRefreshToken)
	if rawRefreshToken == "" {
		return domain.ErrRefreshTokenInvalid
	}
	return uc.userRepo.RevokeRefreshSessionByTokenHash(ctx, hashToken(rawRefreshToken))
}

// LogoutAll revokes all active refresh sessions for the given user.
func (uc *AuthUseCase) LogoutAll(ctx context.Context, userID string) error {
	return uc.userRepo.RevokeAllRefreshSessions(ctx, userID)
}

// ListAuthSessions returns active refresh sessions for the current user.
func (uc *AuthUseCase) ListAuthSessions(ctx context.Context, userID string) ([]domain.AuthSession, error) {
	return uc.userRepo.ListActiveRefreshSessions(ctx, userID)
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

	if !isTrustedDomain(email) {
		return ResendOutput{}, errors.New("unsupported email provider")
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
		slog.Error("resend verification email send failed", "user_id", authUser.User.ID, "error", err)
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

// ForgotPassword generates a password reset token and sends it via email.
func (uc *AuthUseCase) ForgotPassword(ctx context.Context, rawEmail string) (ForgotPasswordOutput, error) {
	if !uc.emailConfigured() {
		return ForgotPasswordOutput{}, domain.ErrEmailNotConfigured
	}

	email, err := normalizeEmail(rawEmail)
	if err != nil {
		return ForgotPasswordOutput{}, domain.ErrInvalidEmail
	}

	if !isTrustedDomain(email) {
		return ForgotPasswordOutput{}, errors.New("unsupported email provider")
	}

	// Get user (don't fail if not found, for security reasons)
	user, err := uc.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		// Return success anyway to prevent email enumeration attacks
		return ForgotPasswordOutput{ResetEmailSent: false}, nil
	}

	// Create password reset token
	rawToken, expiresAt, err := uc.userRepo.CreatePasswordResetToken(ctx, user.ID, uc.verifyTTL)
	if err != nil {
		slog.Error("forgot password: create token failed", "user_id", user.ID, "error", err)
		return ForgotPasswordOutput{}, fmt.Errorf("create recovery token: %w", err)
	}

	// Build password reset link
	resetLink, err := uc.buildResetLink(rawToken)
	if err != nil {
		slog.Error("forgot password: build link failed", "user_id", user.ID, "error", err)
		return ForgotPasswordOutput{}, fmt.Errorf("build recovery link: %w", err)
	}

	// Send reset email
	if err := uc.emailSvc.SendPasswordResetEmail(ctx, user.Email, user.Name, resetLink); err != nil {
		slog.Error("forgot password: send email failed", "user_id", user.ID, "error", err)
		return ForgotPasswordOutput{}, fmt.Errorf("send recovery email: %w", err)
	}

	return ForgotPasswordOutput{ResetEmailSent: true, ExpiresAt: expiresAt.UTC()}, nil
}

// ResetPassword changes the password using a reset token.
func (uc *AuthUseCase) ResetPassword(ctx context.Context, rawToken, newPassword string) (domain.User, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return domain.User{}, domain.ErrInvalidToken
	}

	if len(newPassword) < 8 {
		return domain.User{}, domain.ErrWeakPassword
	}
	if len(newPassword) > 72 {
		return domain.User{}, domain.ErrWeakPassword
	}

	// Hash new password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}

	// Reset password using token
	user, err := uc.userRepo.ResetPasswordByToken(ctx, rawToken, string(passwordHash))
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (uc *AuthUseCase) issueAuthTokens(ctx context.Context, user domain.User, userAgent, ipAddress string) (AuthTokens, error) {
	accessToken, err := uc.tokenSvc.GenerateToken(user.ID)
	if err != nil {
		return AuthTokens{}, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := generateRawToken(32)
	if err != nil {
		return AuthTokens{}, fmt.Errorf("generate refresh token: %w", err)
	}

	if _, err := uc.userRepo.CreateRefreshSession(ctx, user.ID, hashToken(refreshToken), userAgent, ipAddress, uc.refreshTTL); err != nil {
		return AuthTokens{}, fmt.Errorf("create refresh session: %w", err)
	}

	return AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// ---------- helpers ----------

func isTrustedDomain(email string) bool {
	email = strings.ToLower(strings.TrimSpace(email))
	idx := strings.LastIndex(email, "@")
	if idx <= 0 || idx == len(email)-1 {
		return false
	}
	domain := email[idx+1:]
	return trustedDomains[domain]
}

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

func (uc *AuthUseCase) buildResetLink(token string) (string, error) {
	base := strings.TrimSpace(uc.resetURL)
	if base == "" {
		return "", errors.New("password reset url base is not configured")
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

func generateRawToken(byteLength int) (string, error) {
	if byteLength <= 0 {
		return "", errors.New("invalid token length")
	}
	buf := make([]byte, byteLength)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(raw)))
	return hex.EncodeToString(sum[:])
}
