package httpapi

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// DBHealth verifies API-to-database connectivity with a lightweight query.
func (h *Handler) DBHealth(c *gin.Context) {
	if h.dbKeepAlive == nil {
		respondError(c, http.StatusServiceUnavailable, "database checker not configured")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if err := h.dbKeepAlive.KeepAlive(ctx); err != nil {
		respondError(c, http.StatusServiceUnavailable, "database unavailable")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "ok",
		"database": "ok",
	})
}
