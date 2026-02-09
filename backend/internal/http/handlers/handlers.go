package handlers

import (
	"context"
	"crypto/subtle"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"mathalama-focus/backend/internal/auth"
	"mathalama-focus/backend/internal/domain"
	"mathalama-focus/backend/internal/http/middleware"
	"mathalama-focus/backend/internal/repository/postgresql"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type Repository interface {
	GetUser(ctx context.Context, userID string) (domain.User, error)
	DevLogin(ctx context.Context, email, name string) (domain.User, error)
	CreateTelegramLinkCode(ctx context.Context, userID string, ttl time.Duration) (string, time.Time, error)
	LinkTelegramByCode(ctx context.Context, input postgresql.TelegramLinkInput) (domain.User, error)
	GetTelegramIdentity(ctx context.Context, userID string) (domain.TelegramIdentity, error)
	GetUserByTelegramUserID(ctx context.Context, telegramUserID int64) (domain.User, bool, error)
	SetTelegramNotificationsByTelegramUserID(ctx context.Context, telegramUserID int64, enabled bool) error
	UnlinkTelegram(ctx context.Context, userID string) error
	CreateGoal(ctx context.Context, userID string, input postgresql.CreateGoalInput) (domain.Goal, error)
	ListGoals(ctx context.Context, userID string) ([]domain.Goal, error)
	ListGoalHistory(ctx context.Context, userID string) ([]domain.Goal, error)
	StartSession(ctx context.Context, userID string, input postgresql.StartSessionInput) (domain.FocusSession, error)
	PauseSession(ctx context.Context, userID, sessionID string) (domain.FocusSession, error)
	ResumeSession(ctx context.Context, userID, sessionID string) (domain.FocusSession, error)
	AbandonSession(ctx context.Context, userID, sessionID string) (domain.FocusSession, error)
	ResetSession(ctx context.Context, userID, sessionID string) (domain.FocusSession, error)
	GetActiveSession(ctx context.Context, userID string) (domain.FocusSession, error)
	ListSessionHistory(ctx context.Context, userID string, filter postgresql.SessionHistoryFilter) ([]domain.SessionHistoryEntry, domain.SessionHistorySummary, error)
	GetSession(ctx context.Context, userID, sessionID string) (domain.FocusSession, error)
	AddInterruption(ctx context.Context, userID, sessionID, reason string) (domain.Interruption, error)
	CompleteSession(ctx context.Context, userID, sessionID string) (domain.FocusSession, error)
	UpsertReflection(ctx context.Context, userID, sessionID string, input postgresql.ReflectionInput) (domain.Reflection, error)
	AnalyticsOverview(ctx context.Context, userID string) (domain.AnalyticsOverview, error)
	GetLeaderboard(ctx context.Context) ([]domain.LeaderboardEntry, error)
	ListItems(ctx context.Context) ([]domain.Item, error)
	BuyItem(ctx context.Context, userID, itemID string) (domain.UserItem, error)
	GetDailyActivity(ctx context.Context, userID string, timezone string) ([]domain.DailyActivity, error)
	GetDailyContributions(ctx context.Context, userID string, timezone string, date string) ([]domain.DailyContribution, error)
	GetRecentReflections(ctx context.Context, userID string, limit int) ([]domain.Reflection, error)
}

type Handler struct {
	repo                 Repository
	jwtSecret            string
	telegramLinkCodeTTL  time.Duration
	telegramBotAuthToken string
}

func New(repo Repository, jwtSecret string, telegramLinkCodeTTL time.Duration, telegramBotAuthToken string) *Handler {
	return &Handler{
		repo:                 repo,
		jwtSecret:            jwtSecret,
		telegramLinkCodeTTL:  telegramLinkCodeTTL,
		telegramBotAuthToken: strings.TrimSpace(telegramBotAuthToken),
	}
}

const telegramBotAuthHeader = "X-Telegram-Bot-Auth"

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) GetMe(c *gin.Context) {
	userID := middleware.UserID(c)

	user, err := h.repo.GetUser(c.Request.Context(), userID)
	if errors.Is(err, postgresql.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

type devLoginRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (h *Handler) DevLogin(c *gin.Context) {
	var req devLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Name = strings.TrimSpace(req.Name)

	if req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}
	if req.Name == "" {
		req.Name = "Focus Learner"
	}

	user, err := h.repo.DevLogin(c.Request.Context(), req.Email, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login"})
		return
	}

	token, err := auth.GenerateToken(h.jwtSecret, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": token,
	})
}

func (h *Handler) CreateTelegramLinkCode(c *gin.Context) {
	userID := middleware.UserID(c)

	code, expiresAt, err := h.repo.CreateTelegramLinkCode(c.Request.Context(), userID, h.telegramLinkCodeTTL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create telegram link code"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":       code,
		"expires_at": expiresAt.UTC(),
	})
}

func (h *Handler) GetTelegramIdentity(c *gin.Context) {
	userID := middleware.UserID(c)

	identity, err := h.repo.GetTelegramIdentity(c.Request.Context(), userID)
	if errors.Is(err, postgresql.ErrNotFound) {
		c.JSON(http.StatusOK, gin.H{"identity": nil})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get telegram identity"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"identity": identity})
}

func (h *Handler) UnlinkTelegram(c *gin.Context) {
	userID := middleware.UserID(c)
	if err := h.repo.UnlinkTelegram(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unlink telegram account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"unlinked": true})
}

func (h *Handler) TelegramStatusByUserID(c *gin.Context) {
	if h.telegramBotAuthToken == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "telegram integration is not configured"})
		return
	}

	botAuthToken := strings.TrimSpace(c.GetHeader(telegramBotAuthHeader))
	if subtle.ConstantTimeCompare([]byte(botAuthToken), []byte(h.telegramBotAuthToken)) != 1 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid bot auth token"})
		return
	}

	rawTelegramUserID := strings.TrimSpace(c.Query("telegram_user_id"))
	telegramUserID, err := strconv.ParseInt(rawTelegramUserID, 10, 64)
	if err != nil || telegramUserID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "telegram_user_id must be a positive integer"})
		return
	}

	user, notificationsEnabled, err := h.repo.GetUserByTelegramUserID(c.Request.Context(), telegramUserID)
	if errors.Is(err, postgresql.ErrNotFound) {
		c.JSON(http.StatusOK, gin.H{
			"linked":                false,
			"notifications_enabled": false,
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve telegram link status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"linked":                true,
		"notifications_enabled": notificationsEnabled,
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
	})
}

type telegramNotificationsRequest struct {
	TelegramUserID int64 `json:"telegram_user_id"`
	Enabled        bool  `json:"enabled"`
}

func (h *Handler) TelegramSetNotificationsByUserID(c *gin.Context) {
	if h.telegramBotAuthToken == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "telegram integration is not configured"})
		return
	}

	botAuthToken := strings.TrimSpace(c.GetHeader(telegramBotAuthHeader))
	if subtle.ConstantTimeCompare([]byte(botAuthToken), []byte(h.telegramBotAuthToken)) != 1 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid bot auth token"})
		return
	}

	var req telegramNotificationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if req.TelegramUserID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "telegram_user_id must be a positive integer"})
		return
	}

	err := h.repo.SetTelegramNotificationsByTelegramUserID(c.Request.Context(), req.TelegramUserID, req.Enabled)
	if errors.Is(err, postgresql.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "telegram account is not linked"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update telegram notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"updated":               true,
		"notifications_enabled": req.Enabled,
	})
}

type telegramLinkByCodeRequest struct {
	Code             string `json:"code"`
	TelegramUserID   int64  `json:"telegram_user_id"`
	TelegramUsername string `json:"telegram_username"`
	TelegramFirst    string `json:"telegram_first_name"`
	TelegramLast     string `json:"telegram_last_name"`
}

func (h *Handler) TelegramLinkByCode(c *gin.Context) {
	if h.telegramBotAuthToken == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "telegram integration is not configured"})
		return
	}

	botAuthToken := strings.TrimSpace(c.GetHeader(telegramBotAuthHeader))
	if subtle.ConstantTimeCompare([]byte(botAuthToken), []byte(h.telegramBotAuthToken)) != 1 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid bot auth token"})
		return
	}

	var req telegramLinkByCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	req.Code = strings.ToUpper(strings.TrimSpace(req.Code))
	req.TelegramUsername = strings.TrimSpace(req.TelegramUsername)
	req.TelegramFirst = strings.TrimSpace(req.TelegramFirst)
	req.TelegramLast = strings.TrimSpace(req.TelegramLast)
	if req.Code == "" || req.TelegramUserID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code and telegram_user_id are required"})
		return
	}

	user, err := h.repo.LinkTelegramByCode(c.Request.Context(), postgresql.TelegramLinkInput{
		Code:             req.Code,
		TelegramUserID:   req.TelegramUserID,
		TelegramUsername: req.TelegramUsername,
		TelegramFirst:    req.TelegramFirst,
		TelegramLast:     req.TelegramLast,
	})
	if errors.Is(err, postgresql.ErrTelegramCodeInvalid) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired code"})
		return
	}
	if errors.Is(err, postgresql.ErrTelegramAlreadyLinked) {
		c.JSON(http.StatusConflict, gin.H{"error": "telegram account is already linked to another user"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to link telegram account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"linked_user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
	})
}

type createGoalRequest struct {
	Topic              string   `json:"topic"`
	DesiredResult      string   `json:"desired_result"`
	RecommendedMinutes int      `json:"recommended_minutes"`
	Tags               []string `json:"tags"`
}

func (h *Handler) CreateGoal(c *gin.Context) {
	userID := middleware.UserID(c)

	var req createGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	req.Topic = strings.TrimSpace(req.Topic)
	req.DesiredResult = strings.TrimSpace(req.DesiredResult)

	if req.Topic == "" || req.DesiredResult == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "topic and desired_result are required"})
		return
	}

	goal, err := h.repo.CreateGoal(c.Request.Context(), userID, postgresql.CreateGoalInput{
		Topic:              req.Topic,
		DesiredResult:      req.DesiredResult,
		RecommendedMinutes: req.RecommendedMinutes,
		Tags:               req.Tags,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" && pgErr.ConstraintName == "goals_user_id_fkey" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session, please login again"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create goal"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"goal": goal})
}

func (h *Handler) ListGoals(c *gin.Context) {
	userID := middleware.UserID(c)

	goals, err := h.repo.ListGoals(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list goals"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"goals": goals})
}

func (h *Handler) ListGoalHistory(c *gin.Context) {
	userID := middleware.UserID(c)

	goals, err := h.repo.ListGoalHistory(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list goal history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"goals": goals})
}

type startSessionRequest struct {
	GoalID             string `json:"goal_id"`
	RecommendedMinutes int    `json:"recommended_minutes"`
	IsStrict           bool   `json:"is_strict"`
}

func (h *Handler) StartSession(c *gin.Context) {
	userID := middleware.UserID(c)

	var req startSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	req.GoalID = strings.TrimSpace(req.GoalID)
	if req.GoalID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "goal_id is required"})
		return
	}

	session, err := h.repo.StartSession(c.Request.Context(), userID, postgresql.StartSessionInput{
		GoalID:             req.GoalID,
		RecommendedMinutes: req.RecommendedMinutes,
		IsStrict:           req.IsStrict,
	})
	if errors.Is(err, postgresql.ErrGoalCompleted) {
		c.JSON(http.StatusConflict, gin.H{"error": "goal already completed"})
		return
	}
	if errors.Is(err, postgresql.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "goal not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start session"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"session": session})
}

func (h *Handler) GetActiveSession(c *gin.Context) {
	userID := middleware.UserID(c)

	session, err := h.repo.GetActiveSession(c.Request.Context(), userID)
	if errors.Is(err, postgresql.ErrNotFound) {
		c.JSON(http.StatusOK, gin.H{"session": nil})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get active session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"session": session})
}

func (h *Handler) ListSessionHistory(c *gin.Context) {
	userID := middleware.UserID(c)

	minMinutes, err := parseOptionalInt(c.Query("min_minutes"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "min_minutes must be a positive integer"})
		return
	}
	maxMinutes, err := parseOptionalInt(c.Query("max_minutes"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "max_minutes must be a positive integer"})
		return
	}
	limit, err := parseOptionalInt(c.Query("limit"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be a positive integer"})
		return
	}

	tags := make([]string, 0)
	for _, tag := range strings.Split(strings.TrimSpace(c.Query("tags")), ",") {
		clean := strings.TrimSpace(tag)
		if clean != "" {
			tags = append(tags, clean)
		}
	}

	sessions, summary, err := h.repo.ListSessionHistory(c.Request.Context(), userID, postgresql.SessionHistoryFilter{
		Period:     strings.TrimSpace(c.DefaultQuery("period", "all")),
		Timezone:   strings.TrimSpace(c.Query("timezone")),
		Tags:       tags,
		MinMinutes: minMinutes,
		MaxMinutes: maxMinutes,
		Status:     strings.TrimSpace(c.DefaultQuery("status", "completed")),
		Limit:      limit,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load session history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
		"summary":  summary,
	})
}

func (h *Handler) PauseSession(c *gin.Context) {
	h.sessionAction(c, func(userID, sessionID string) (any, error) {
		return h.repo.PauseSession(c.Request.Context(), userID, sessionID)
	})
}

func (h *Handler) ResumeSession(c *gin.Context) {
	h.sessionAction(c, func(userID, sessionID string) (any, error) {
		return h.repo.ResumeSession(c.Request.Context(), userID, sessionID)
	})
}

func (h *Handler) AbandonSession(c *gin.Context) {
	h.sessionAction(c, func(userID, sessionID string) (any, error) {
		return h.repo.AbandonSession(c.Request.Context(), userID, sessionID)
	})
}

func (h *Handler) ResetSession(c *gin.Context) {
	h.sessionAction(c, func(userID, sessionID string) (any, error) {
		return h.repo.ResetSession(c.Request.Context(), userID, sessionID)
	})
}

func (h *Handler) GetSession(c *gin.Context) {
	userID := middleware.UserID(c)
	sessionID := strings.TrimSpace(c.Param("sessionID"))
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessionID is required"})
		return
	}

	if _, err := uuid.Parse(sessionID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sessionID format"})
		return
	}

	session, err := h.repo.GetSession(c.Request.Context(), userID, sessionID)
	if errors.Is(err, postgresql.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"session": session})
}

type interruptionRequest struct {
	Reason string `json:"reason"`
}

func (h *Handler) AddInterruption(c *gin.Context) {
	userID := middleware.UserID(c)
	sessionID := strings.TrimSpace(c.Param("sessionID"))
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessionID is required"})
		return
	}

	if _, err := uuid.Parse(sessionID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sessionID format"})
		return
	}

	var req interruptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	req.Reason = strings.TrimSpace(req.Reason)
	if req.Reason == "" {
		req.Reason = "manual interruption"
	}

	interruption, err := h.repo.AddInterruption(c.Request.Context(), userID, sessionID, req.Reason)
	if errors.Is(err, postgresql.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to log interruption"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"interruption": interruption})
}

func (h *Handler) CompleteSession(c *gin.Context) {
	h.sessionAction(c, func(userID, sessionID string) (any, error) {
		return h.repo.CompleteSession(c.Request.Context(), userID, sessionID)
	})
}

type reflectionRequest struct {
	WhatLearned string `json:"what_learned"`
	WhatWasHard string `json:"what_was_hard"`
	NextAction  string `json:"next_action"`
}

func (h *Handler) UpsertReflection(c *gin.Context) {
	userID := middleware.UserID(c)
	sessionID := strings.TrimSpace(c.Param("sessionID"))
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessionID is required"})
		return
	}

	if _, err := uuid.Parse(sessionID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sessionID format"})
		return
	}

	var req reflectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	req.WhatLearned = strings.TrimSpace(req.WhatLearned)
	req.WhatWasHard = strings.TrimSpace(req.WhatWasHard)
	req.NextAction = strings.TrimSpace(req.NextAction)

	if req.WhatLearned == "" || req.WhatWasHard == "" || req.NextAction == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "all reflection fields are required"})
		return
	}

	reflection, err := h.repo.UpsertReflection(c.Request.Context(), userID, sessionID, postgresql.ReflectionInput{
		WhatLearned: req.WhatLearned,
		WhatWasHard: req.WhatWasHard,
		NextAction:  req.NextAction,
	})
	if errors.Is(err, postgresql.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save reflection"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"reflection": reflection})
}

func (h *Handler) AnalyticsOverview(c *gin.Context) {
	userID := middleware.UserID(c)

	overview, err := h.repo.AnalyticsOverview(c.Request.Context(), userID)
	if errors.Is(err, postgresql.ErrNotFound) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session, please login again"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load analytics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"overview": overview})
}

func (h *Handler) GetLeaderboard(c *gin.Context) {
	leaderboard, err := h.repo.GetLeaderboard(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load leaderboard"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"leaderboard": leaderboard})
}

func (h *Handler) ListItems(c *gin.Context) {
	items, err := h.repo.ListItems(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list items"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) BuyItem(c *gin.Context) {
	userID := middleware.UserID(c)
	itemID := c.Param("itemID")
	if itemID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "itemID is required"})
		return
	}

	userItem, err := h.repo.BuyItem(c.Request.Context(), userID, itemID)
	if err != nil {
		// Basic error handling - could be more specific
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_item": userItem})
}

func (h *Handler) GetDailyActivity(c *gin.Context) {
	userID := middleware.UserID(c)
	timezone := c.Query("timezone")

	activity, err := h.repo.GetDailyActivity(c.Request.Context(), userID, timezone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load daily activity"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"activity": activity})
}

func (h *Handler) GetDailyContributions(c *gin.Context) {
	userID := middleware.UserID(c)
	timezone := c.Query("timezone")
	date := strings.TrimSpace(c.Query("date"))
	if date == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date is required (YYYY-MM-DD)"})
		return
	}

	contributions, err := h.repo.GetDailyContributions(c.Request.Context(), userID, timezone, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load daily contributions"})
		return
	}

	totalMinutes := 0
	for _, item := range contributions {
		totalMinutes += item.Minutes
	}

	c.JSON(http.StatusOK, gin.H{
		"date":          date,
		"session_count": len(contributions),
		"total_minutes": totalMinutes,
		"contributions": contributions,
	})
}

func (h *Handler) GetInsights(c *gin.Context) {
	userID := middleware.UserID(c)

	reflections, err := h.repo.GetRecentReflections(c.Request.Context(), userID, 5)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load reflections for AI"})
		return
	}

	// Mock AI Logic / Simple Rule-based Engine
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

	c.JSON(http.StatusOK, gin.H{"insight": insight})
}

func (h *Handler) sessionAction(c *gin.Context, action func(userID, sessionID string) (any, error)) {
	userID := middleware.UserID(c)
	sessionID := strings.TrimSpace(c.Param("sessionID"))
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessionID is required"})
		return
	}

	if _, err := uuid.Parse(sessionID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sessionID format"})
		return
	}

	payload, err := action(userID, sessionID)
	if errors.Is(err, postgresql.ErrPauseLimitReached) {
		c.JSON(http.StatusConflict, gin.H{"error": "pause limit reached"})
		return
	}
	if errors.Is(err, postgresql.ErrInvalidState) {
		c.JSON(http.StatusConflict, gin.H{"error": "session cannot perform this action"})
		return
	}
	if errors.Is(err, postgresql.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "action failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"session": payload})
}

func parseOptionalInt(raw string) (int, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return 0, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return 0, errors.New("invalid integer")
	}
	return parsed, nil
}
