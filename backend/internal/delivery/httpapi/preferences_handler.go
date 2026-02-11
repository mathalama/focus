package httpapi

import (
	"net/http"

	"mathalama-focus/backend/internal/domain"
	"mathalama-focus/backend/internal/usecase"

	"github.com/gin-gonic/gin"
)

type PreferencesHandler struct {
	preferencesUC *usecase.PreferencesUseCase
}

func NewPreferencesHandler(preferencesUC *usecase.PreferencesUseCase) *PreferencesHandler {
	return &PreferencesHandler{preferencesUC: preferencesUC}
}

// GetSessionPreferences GET /api/v1/preferences/session
func (h *PreferencesHandler) GetSessionPreferences(c *gin.Context) {
	userID := c.GetString("user_id")

	prefs, err := h.preferencesUC.GetSessionPreferences(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, prefs)
}

// UpdateSessionPreferences PUT /api/v1/preferences/session
func (h *PreferencesHandler) UpdateSessionPreferences(c *gin.Context) {
	userID := c.GetString("user_id")

	var prefs domain.UserSessionPreferences
	if err := c.BindJSON(&prefs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	updated, err := h.preferencesUC.UpdateSessionPreferences(c.Request.Context(), userID, prefs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// ListNotificationSchedules GET /api/v1/preferences/notifications
func (h *PreferencesHandler) ListNotificationSchedules(c *gin.Context) {
	userID := c.GetString("user_id")

	schedules, err := h.preferencesUC.GetNotificationSchedules(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if schedules == nil {
		schedules = []domain.NotificationSchedule{}
	}

	c.JSON(http.StatusOK, schedules)
}

// CreateNotificationSchedule POST /api/v1/preferences/notifications
func (h *PreferencesHandler) CreateNotificationSchedule(c *gin.Context) {
	userID := c.GetString("user_id")

	var schedule domain.NotificationSchedule
	if err := c.BindJSON(&schedule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	created, err := h.preferencesUC.CreateNotificationSchedule(c.Request.Context(), userID, schedule)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

// UpdateNotificationSchedule PUT /api/v1/preferences/notifications/:id
func (h *PreferencesHandler) UpdateNotificationSchedule(c *gin.Context) {
	userID := c.GetString("user_id")
	scheduleID := c.Param("id")

	var schedule domain.NotificationSchedule
	if err := c.BindJSON(&schedule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	updated, err := h.preferencesUC.UpdateNotificationSchedule(c.Request.Context(), userID, scheduleID, schedule)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// DeleteNotificationSchedule DELETE /api/v1/preferences/notifications/:id
func (h *PreferencesHandler) DeleteNotificationSchedule(c *gin.Context) {
	userID := c.GetString("user_id")
	scheduleID := c.Param("id")

	if err := h.preferencesUC.DeleteNotificationSchedule(c.Request.Context(), userID, scheduleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
