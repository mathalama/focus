package httpapi

import (
	"net/http"

	"mathalama-focus/backend/internal/usecase"

	"github.com/gin-gonic/gin"
)

// GetNotificationSounds returns user's notification sound preferences
func (h *Handler) GetNotificationSounds(c *gin.Context) {
	userID := getUserID(c)

	sound, err := h.notification.GetSoundPreferences(c.Request.Context(), userID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to get notification sounds")
		return
	}

	c.JSON(http.StatusOK, gin.H{"sound": sound})
}

// UpdateNotificationSounds updates user's notification sound preferences
func (h *Handler) UpdateNotificationSounds(c *gin.Context) {
	userID := getUserID(c)

	var req usecase.UpdateSoundPreferencesInput
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	sound, err := h.notification.UpdateSoundPreferences(c.Request.Context(), userID, req)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to update notification sounds")
		return
	}

	c.JSON(http.StatusOK, gin.H{"sound": sound})
	h.trackEvent(c.Request.Context(), userID, "notification.settings_updated", map[string]any{
		"sounds_enabled": sound.SoundsEnabled,
		"volume":         sound.Volume,
	})
}
