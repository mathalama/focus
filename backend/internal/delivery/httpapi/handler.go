package httpapi

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"mathalama-focus/backend/internal/domain"
	"mathalama-focus/backend/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type EmailDeliveryReader interface {
	ListRecentDeliveries(ctx context.Context, limit int) ([]domain.EmailDelivery, error)
}

type EventTracker interface {
	TrackEvent(ctx context.Context, userID, eventName, source string, properties map[string]any) error
	ListRecentEvents(ctx context.Context, limit int, eventName string) ([]domain.ProductEvent, error)
}

type DBKeepAlive interface {
	KeepAlive(ctx context.Context) error
}

// Handler groups all HTTP handlers and their use-case dependencies.
type Handler struct {
	auth            *usecase.AuthUseCase
	session         *usecase.SessionUseCase
	goal            *usecase.GoalUseCase
	analytics       *usecase.AnalyticsUseCase
	shop            *usecase.ShopUseCase
	notification    *usecase.NotificationUseCase
	preferences     *usecase.PreferencesUseCase
	dbKeepAlive     DBKeepAlive
	emailDeliveries EmailDeliveryReader
	eventTracker    EventTracker
}

// NewHandler creates a new Handler with all use-case dependencies.
func NewHandler(
	auth *usecase.AuthUseCase,
	session *usecase.SessionUseCase,
	goal *usecase.GoalUseCase,
	analytics *usecase.AnalyticsUseCase,
	shop *usecase.ShopUseCase,
	notification *usecase.NotificationUseCase,
	preferences *usecase.PreferencesUseCase,
	dbKeepAlive DBKeepAlive,
	emailDeliveries EmailDeliveryReader,
	eventTracker EventTracker,
) *Handler {
	return &Handler{
		auth:            auth,
		session:         session,
		goal:            goal,
		analytics:       analytics,
		shop:            shop,
		notification:    notification,
		preferences:     preferences,
		dbKeepAlive:     dbKeepAlive,
		emailDeliveries: emailDeliveries,
		eventTracker:    eventTracker,
	}
}

// Health is a simple liveness check.
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// HandleFocusEvent handles focus session events from browser extension
func (h *Handler) HandleFocusEvent(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Action   string `json:"action" binding:"required,oneof=start stop"`
		Duration int    `json:"duration"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindingError(c, err)
		return
	}

	// Broadcast event to WebSocket clients
	hub := GetFocusHub()
	if req.Action == "start" {
		hub.BroadcastFocusStarted(userID, req.Duration)
	} else if req.Action == "stop" {
		hub.BroadcastFocusStopped(userID)
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ---------- shared helpers ----------

func respondError(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{"error": msg})
}

func (h *Handler) handleError(c *gin.Context, err error, fallbackMsg string) {
	if err == nil {
		return
	}

	status := http.StatusInternalServerError
	msg := fallbackMsg

	switch {
	case errors.Is(err, domain.ErrNotFound):
		status = http.StatusNotFound
		msg = err.Error()
	case errors.Is(err, domain.ErrAlreadyExists):
		status = http.StatusConflict
		msg = err.Error()
	case errors.Is(err, domain.ErrInvalidEmail), errors.Is(err, domain.ErrWeakPassword), errors.Is(err, domain.ErrEmailTokenInvalid), errors.Is(err, domain.ErrInvalidToken):
		status = http.StatusBadRequest
		msg = err.Error()
	case errors.Is(err, domain.ErrInvalidCredentials):
		status = http.StatusUnauthorized
		msg = "invalid email or password"
	case errors.Is(err, domain.ErrEmailNotVerified):
		status = http.StatusForbidden
		msg = "email is not verified"
	case errors.Is(err, domain.ErrEmailNotConfigured), errors.Is(err, domain.ErrAuthUnavailable):
		status = http.StatusServiceUnavailable
		msg = err.Error()
	case errors.Is(err, domain.ErrInsufficientBalance), errors.Is(err, domain.ErrPauseLimitReached), errors.Is(err, domain.ErrInvalidState), errors.Is(err, domain.ErrGoalCompleted):
		status = http.StatusConflict
		msg = err.Error()
	case errors.Is(err, domain.ErrForbidden):
		status = http.StatusForbidden
		msg = "forbidden"
	case errors.Is(err, domain.ErrRefreshTokenInvalid):
		status = http.StatusUnauthorized
		msg = "invalid session"
	}

	if status == http.StatusInternalServerError {
		slog.Error("internal error", "method", c.Request.Method, "path", c.Request.URL.Path, "error", err)
	} else {
		slog.Warn("request error", "method", c.Request.Method, "path", c.Request.URL.Path, "status", status, "error", err)
	}

	respondError(c, status, msg)
}

func handleBindingError(c *gin.Context, err error) {
	if ve, ok := err.(validator.ValidationErrors); ok {
		var errs []string
		for _, e := range ve {
			errs = append(errs, fmt.Sprintf("field %s: %s", e.Field(), e.Tag()))
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "details": errs})
		return
	}
	respondError(c, http.StatusBadRequest, "invalid request body")
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

func validateSessionID(c *gin.Context) (string, bool) {
	sessionID := strings.TrimSpace(c.Param("sessionID"))
	if sessionID == "" {
		respondError(c, http.StatusBadRequest, "sessionID is required")
		return "", false
	}
	if _, err := uuid.Parse(sessionID); err != nil {
		respondError(c, http.StatusBadRequest, "invalid sessionID format")
		return "", false
	}
	return sessionID, true
}

func (h *Handler) sessionAction(c *gin.Context, action func(userID, sessionID string) (any, error)) {
	userID := getUserID(c)
	sessionID, ok := validateSessionID(c)
	if !ok {
		return
	}

	payload, err := action(userID, sessionID)
	if err != nil {
		h.handleError(c, err, "action failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"session": payload})
}

// ---------- preferences handlers ----------

func (h *Handler) GetSessionPreferences(c *gin.Context) {
	userID := c.GetString("user_id")

	prefs, err := h.preferences.GetSessionPreferences(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err, "failed to get preferences")
		return
	}

	c.JSON(http.StatusOK, prefs)
}

func (h *Handler) UpdateSessionPreferences(c *gin.Context) {
	userID := c.GetString("user_id")

	var prefs domain.UserSessionPreferences
	if err := c.BindJSON(&prefs); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request")
		return
	}

	updated, err := h.preferences.UpdateSessionPreferences(c.Request.Context(), userID, prefs)
	if err != nil {
		h.handleError(c, err, "failed to update preferences")
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *Handler) ListNotificationSchedules(c *gin.Context) {
	userID := c.GetString("user_id")

	schedules, err := h.preferences.GetNotificationSchedules(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err, "failed to list schedules")
		return
	}

	if schedules == nil {
		schedules = []domain.NotificationSchedule{}
	}

	c.JSON(http.StatusOK, schedules)
}

func (h *Handler) CreateNotificationSchedule(c *gin.Context) {
	userID := c.GetString("user_id")

	var schedule domain.NotificationSchedule
	if err := c.BindJSON(&schedule); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request")
		return
	}

	created, err := h.preferences.CreateNotificationSchedule(c.Request.Context(), userID, schedule)
	if err != nil {
		h.handleError(c, err, "failed to create schedule")
		return
	}

	c.JSON(http.StatusCreated, created)
}

func (h *Handler) UpdateNotificationSchedule(c *gin.Context) {
	userID := c.GetString("user_id")
	scheduleID := c.Param("id")

	var schedule domain.NotificationSchedule
	if err := c.BindJSON(&schedule); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request")
		return
	}

	updated, err := h.preferences.UpdateNotificationSchedule(c.Request.Context(), userID, scheduleID, schedule)
	if err != nil {
		h.handleError(c, err, "failed to update schedule")
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *Handler) DeleteNotificationSchedule(c *gin.Context) {
	userID := c.GetString("user_id")
	scheduleID := c.Param("id")

	if err := h.preferences.DeleteNotificationSchedule(c.Request.Context(), userID, scheduleID); err != nil {
		h.handleError(c, err, "failed to delete schedule")
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *Handler) trackEvent(ctx context.Context, userID, eventName string, properties map[string]any) {
	if h.eventTracker == nil {
		return
	}
	if err := h.eventTracker.TrackEvent(ctx, userID, eventName, "api", properties); err != nil {
		// best effort analytics
	}
}
