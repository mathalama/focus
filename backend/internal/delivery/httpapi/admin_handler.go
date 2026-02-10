package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"mathalama-focus/backend/internal/domain"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ensureAdmin(c *gin.Context) (string, bool) {
	userID := getUserID(c)
	if userID == "" {
		respondError(c, http.StatusUnauthorized, "unauthorized")
		return "", false
	}

	user, err := h.auth.GetUser(c.Request.Context(), userID)
	if errors.Is(err, domain.ErrNotFound) {
		respondError(c, http.StatusUnauthorized, "unauthorized")
		return "", false
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to authorize user")
		return "", false
	}

	if user.Role != domain.RoleAdmin {
		respondError(c, http.StatusForbidden, "admin access required")
		return "", false
	}

	return userID, true
}

func (h *Handler) AdminHealth(c *gin.Context) {
	userID, ok := h.ensureAdmin(c)
	if !ok {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "ok",
		"scope":    "admin",
		"actor_id": userID,
	})
}

func (h *Handler) AdminListEmailDeliveries(c *gin.Context) {
	if _, ok := h.ensureAdmin(c); !ok {
		return
	}

	if h.emailDeliveries == nil {
		respondError(c, http.StatusNotImplemented, "email delivery outbox is not configured")
		return
	}

	limit := 100
	rawLimit := strings.TrimSpace(c.Query("limit"))
	if rawLimit != "" {
		parsed, err := parseOptionalInt(rawLimit)
		if err != nil || parsed <= 0 {
			respondError(c, http.StatusBadRequest, "limit must be a positive integer")
			return
		}
		limit = parsed
	}

	deliveries, err := h.emailDeliveries.ListRecentDeliveries(c.Request.Context(), limit)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to list email deliveries")
		return
	}

	c.JSON(http.StatusOK, gin.H{"deliveries": deliveries})
}

func (h *Handler) AdminListEvents(c *gin.Context) {
	if _, ok := h.ensureAdmin(c); !ok {
		return
	}
	if h.eventTracker == nil {
		respondError(c, http.StatusNotImplemented, "event tracker is not configured")
		return
	}

	limit := 100
	rawLimit := strings.TrimSpace(c.Query("limit"))
	if rawLimit != "" {
		parsed, err := parseOptionalInt(rawLimit)
		if err != nil || parsed <= 0 {
			respondError(c, http.StatusBadRequest, "limit must be a positive integer")
			return
		}
		limit = parsed
	}

	eventName := strings.TrimSpace(c.Query("event_name"))
	events, err := h.eventTracker.ListRecentEvents(c.Request.Context(), limit, eventName)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to list product events")
		return
	}

	c.JSON(http.StatusOK, gin.H{"events": events})
}
