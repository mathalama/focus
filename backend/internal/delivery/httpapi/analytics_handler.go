package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"mathalama-focus/backend/internal/domain"

	"github.com/gin-gonic/gin"
)

func (h *Handler) AnalyticsOverview(c *gin.Context) {
	userID := getUserID(c)

	overview, err := h.analytics.Overview(c.Request.Context(), userID)
	if errors.Is(err, domain.ErrNotFound) {
		respondError(c, http.StatusUnauthorized, "invalid session, please login again")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load analytics")
		return
	}

	c.JSON(http.StatusOK, gin.H{"overview": overview})
}

func (h *Handler) GetDailyActivity(c *gin.Context) {
	userID := getUserID(c)

	activity, err := h.analytics.DailyActivity(c.Request.Context(), userID, c.Query("timezone"))
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load daily activity")
		return
	}

	c.JSON(http.StatusOK, gin.H{"activity": activity})
}

func (h *Handler) GetDailyContributions(c *gin.Context) {
	userID := getUserID(c)
	date := strings.TrimSpace(c.Query("date"))
	if date == "" {
		respondError(c, http.StatusBadRequest, "date is required (YYYY-MM-DD)")
		return
	}

	contributions, err := h.analytics.DailyContributions(c.Request.Context(), userID, c.Query("timezone"), date)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load daily contributions")
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
	userID := getUserID(c)

	insight, err := h.analytics.Insights(c.Request.Context(), userID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load insights")
		return
	}

	c.JSON(http.StatusOK, gin.H{"insight": insight})
}
