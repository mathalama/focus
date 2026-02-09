package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"mathalama-focus/backend/internal/domain"
	"mathalama-focus/backend/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler groups all HTTP handlers and their use-case dependencies.
type Handler struct {
	auth                 *usecase.AuthUseCase
	session              *usecase.SessionUseCase
	goal                 *usecase.GoalUseCase
	analytics            *usecase.AnalyticsUseCase
	shop                 *usecase.ShopUseCase
	telegram             *usecase.TelegramUseCase
	telegramBotAuthToken string
}

// NewHandler creates a new Handler with all use-case dependencies.
func NewHandler(
	auth *usecase.AuthUseCase,
	session *usecase.SessionUseCase,
	goal *usecase.GoalUseCase,
	analytics *usecase.AnalyticsUseCase,
	shop *usecase.ShopUseCase,
	telegram *usecase.TelegramUseCase,
	telegramBotAuthToken string,
) *Handler {
	return &Handler{
		auth:                 auth,
		session:              session,
		goal:                 goal,
		analytics:            analytics,
		shop:                 shop,
		telegram:             telegram,
		telegramBotAuthToken: strings.TrimSpace(telegramBotAuthToken),
	}
}

// Health is a simple liveness check.
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ---------- shared helpers ----------

func respondError(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{"error": msg})
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
	if errors.Is(err, domain.ErrPauseLimitReached) {
		respondError(c, http.StatusConflict, "pause limit reached")
		return
	}
	if errors.Is(err, domain.ErrInvalidState) {
		respondError(c, http.StatusConflict, "session cannot perform this action")
		return
	}
	if errors.Is(err, domain.ErrNotFound) {
		respondError(c, http.StatusNotFound, "session not found")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, "action failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"session": payload})
}
