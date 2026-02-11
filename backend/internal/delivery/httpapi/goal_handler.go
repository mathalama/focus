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
	h.trackEvent(c.Request.Context(), userID, "goal.created", map[string]any{
		"goal_id":             goal.ID,
		"recommended_minutes": goal.RecommendedMinutes,
		"tags_count":          len(goal.Tags),
	})
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

func (h *Handler) GetGoal(c *gin.Context) {
	userID := getUserID(c)
	goalID := strings.TrimSpace(c.Param("goalId"))

	if goalID == "" {
		respondError(c, http.StatusBadRequest, "goal id is required")
		return
	}

	goal, err := h.goal.Get(c.Request.Context(), userID, goalID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			respondError(c, http.StatusNotFound, "goal not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to get goal")
		return
	}

	c.JSON(http.StatusOK, gin.H{"goal": goal})
}

type updateGoalRequest struct {
	Topic              string   `json:"topic"`
	DesiredResult      string   `json:"desired_result"`
	RecommendedMinutes int      `json:"recommended_minutes"`
	Tags               []string `json:"tags"`
}

func (h *Handler) UpdateGoal(c *gin.Context) {
	userID := getUserID(c)
	goalID := strings.TrimSpace(c.Param("goalId"))

	if goalID == "" {
		respondError(c, http.StatusBadRequest, "goal id is required")
		return
	}

	var req updateGoalRequest
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

	goal, err := h.goal.Update(c.Request.Context(), userID, goalID, usecase.UpdateGoalInput{
		Topic:              req.Topic,
		DesiredResult:      req.DesiredResult,
		RecommendedMinutes: req.RecommendedMinutes,
		Tags:               req.Tags,
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			respondError(c, http.StatusNotFound, "goal not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to update goal")
		return
	}

	c.JSON(http.StatusOK, gin.H{"goal": goal})
}

func (h *Handler) DeleteGoal(c *gin.Context) {
	userID := getUserID(c)
	goalID := strings.TrimSpace(c.Param("goalId"))

	if goalID == "" {
		respondError(c, http.StatusBadRequest, "goal id is required")
		return
	}

	err := h.goal.Delete(c.Request.Context(), userID, goalID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			respondError(c, http.StatusNotFound, "goal not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to delete goal")
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true})
	h.trackEvent(c.Request.Context(), userID, "goal.deleted", map[string]any{
		"goal_id": goalID,
	})
}
