package httpapi

import (
	"errors"
	"net/http"

	"mathalama-focus/backend/internal/domain"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetLeaderboard(c *gin.Context) {
	leaderboard, err := h.shop.Leaderboard(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load leaderboard")
		return
	}
	c.JSON(http.StatusOK, gin.H{"leaderboard": leaderboard})
}

func (h *Handler) ListItems(c *gin.Context) {
	items, err := h.shop.ListItems(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to list items")
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) BuyItem(c *gin.Context) {
	userID := getUserID(c)
	itemID := c.Param("itemID")
	if itemID == "" {
		respondError(c, http.StatusBadRequest, "itemID is required")
		return
	}

	userItem, err := h.shop.BuyItem(c.Request.Context(), userID, itemID)
	if errors.Is(err, domain.ErrInsufficientBalance) {
		respondError(c, http.StatusBadRequest, "insufficient nectar balance")
		return
	}
	if errors.Is(err, domain.ErrNotFound) {
		respondError(c, http.StatusNotFound, "item not found")
		return
	}
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_item": userItem})
}
