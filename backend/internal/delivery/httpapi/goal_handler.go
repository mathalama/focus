package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"mathalama-focus/backend/internal/domain"
	"mathalama-focus/backend/internal/usecase"

	"github.com/gin-gonic/gin"
)

type createGoalRequest struct {
	Topic              string   `json:"topic"`
	DesiredResult      string   `json:"desired_result"`
	RecommendedMinutes int      `json:"recommended_minutes"`
	Tags               []string `json:"tags"`
}

func (h *Handler) CreateGoal(c *gin.Context) {
	userID := getUserID(c)

	var req createGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Topic = strings.TrimSpace(req.Topic)
	req.DesiredResult = strings.TrimSpace(req.DesiredResult)

	if req.Topic == "" || req.DesiredResult == "" {
		respondError(c, http.StatusBadRequest, "topic and desired_result are required")
		return
	}

	goal, err := h.goal.Create(c.Request.Context(), userID, usecase.CreateGoalInput{
		Topic:              req.Topic,
		DesiredResult:      req.DesiredResult,
		RecommendedMinutes: req.RecommendedMinutes,
		Tags:               req.Tags,
	})
	if errors.Is(err, domain.ErrNotFound) {
		respondError(c, http.StatusUnauthorized, "invalid session, please login again")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create goal")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"goal": goal})
}

func (h *Handler) ListGoals(c *gin.Context) {
	userID := getUserID(c)

	goals, err := h.goal.List(c.Request.Context(), userID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to list goals")
		return
	}

	c.JSON(http.StatusOK, gin.H{"goals": goals})
}

func (h *Handler) ListGoalHistory(c *gin.Context) {
	userID := getUserID(c)

	goals, err := h.goal.ListHistory(c.Request.Context(), userID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to list goal history")
		return
	}

	c.JSON(http.StatusOK, gin.H{"goals": goals})
}
