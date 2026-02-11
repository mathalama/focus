package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"mathalama-focus/backend/internal/domain"
	"mathalama-focus/backend/internal/usecase"

	"github.com/gin-gonic/gin"
)

// ---------- request types ----------

type startSessionRequest struct {
	GoalID             string `json:"goal_id"`
	RecommendedMinutes int    `json:"recommended_minutes"`
	IsStrict           bool   `json:"is_strict"`
}

type interruptionRequest struct {
	Reason string `json:"reason"`
}

type reflectionRequest struct {
	WhatLearned string `json:"what_learned"`
	WhatWasHard string `json:"what_was_hard"`
	NextAction  string `json:"next_action"`
}

// ---------- handlers ----------

func (h *Handler) StartSession(c *gin.Context) {
	userID := getUserID(c)

	var req startSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	req.GoalID = strings.TrimSpace(req.GoalID)
	if req.GoalID == "" {
		respondError(c, http.StatusBadRequest, "goal_id is required")
		return
	}

	session, err := h.session.Start(c.Request.Context(), userID, usecase.StartSessionInput{
		GoalID:             req.GoalID,
		RecommendedMinutes: req.RecommendedMinutes,
		IsStrict:           req.IsStrict,
	})
	if errors.Is(err, domain.ErrGoalCompleted) {
		respondError(c, http.StatusConflict, "goal already completed")
		return
	}
	if errors.Is(err, domain.ErrNotFound) {
		respondError(c, http.StatusNotFound, "goal not found")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to start session")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"session": session})
	h.trackEvent(c.Request.Context(), userID, "session.start", map[string]any{
		"goal_id":             req.GoalID,
		"recommended_minutes": session.RecommendedMinutes,
		"is_strict":           session.IsStrict,
	})
}

func (h *Handler) GetActiveSession(c *gin.Context) {
	userID := getUserID(c)

	session, err := h.session.GetActive(c.Request.Context(), userID)
	if errors.Is(err, domain.ErrNotFound) {
		c.JSON(http.StatusOK, gin.H{"session": nil})
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to get active session")
		return
	}

	c.JSON(http.StatusOK, gin.H{"session": session})
}

func (h *Handler) GetSession(c *gin.Context) {
	userID := getUserID(c)
	sessionID, ok := validateSessionID(c)
	if !ok {
		return
	}

	session, err := h.session.Get(c.Request.Context(), userID, sessionID)
	if errors.Is(err, domain.ErrNotFound) {
		respondError(c, http.StatusNotFound, "session not found")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to get session")
		return
	}

	c.JSON(http.StatusOK, gin.H{"session": session})
}

func (h *Handler) ListSessionHistory(c *gin.Context) {
	userID := getUserID(c)

	minMinutes, err := parseOptionalInt(c.Query("min_minutes"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "min_minutes must be a positive integer")
		return
	}
	maxMinutes, err := parseOptionalInt(c.Query("max_minutes"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "max_minutes must be a positive integer")
		return
	}
	limit, err := parseOptionalInt(c.Query("limit"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "limit must be a positive integer")
		return
	}

	tags := make([]string, 0)
	for _, tag := range strings.Split(strings.TrimSpace(c.Query("tags")), ",") {
		clean := strings.TrimSpace(tag)
		if clean != "" {
			tags = append(tags, clean)
		}
	}

	sessions, summary, err := h.session.ListHistory(c.Request.Context(), userID, usecase.SessionHistoryFilter{
		Period:     strings.TrimSpace(c.DefaultQuery("period", "all")),
		Timezone:   strings.TrimSpace(c.Query("timezone")),
		Tags:       tags,
		MinMinutes: minMinutes,
		MaxMinutes: maxMinutes,
		Status:     strings.TrimSpace(c.DefaultQuery("status", "completed")),
		Limit:      limit,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load session history")
		return
	}

	c.JSON(http.StatusOK, gin.H{"sessions": sessions, "summary": summary})
}

func (h *Handler) PauseSession(c *gin.Context) {
	h.sessionAction(c, func(userID, sessionID string) (any, error) {
		return h.session.Pause(c.Request.Context(), userID, sessionID)
	})
}

func (h *Handler) ResumeSession(c *gin.Context) {
	h.sessionAction(c, func(userID, sessionID string) (any, error) {
		return h.session.Resume(c.Request.Context(), userID, sessionID)
	})
}

func (h *Handler) AbandonSession(c *gin.Context) {
	userID := getUserID(c)
	sessionID, ok := validateSessionID(c)
	if !ok {
		return
	}
	session, err := h.session.Abandon(c.Request.Context(), userID, sessionID)
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

	c.JSON(http.StatusOK, gin.H{"session": session})
	h.trackEvent(c.Request.Context(), userID, "session.abandon", map[string]any{
		"session_id": sessionID,
	})
}

func (h *Handler) ResetSession(c *gin.Context) {
	h.sessionAction(c, func(userID, sessionID string) (any, error) {
		return h.session.Reset(c.Request.Context(), userID, sessionID)
	})
}

func (h *Handler) CompleteSession(c *gin.Context) {
	userID := getUserID(c)
	sessionID, ok := validateSessionID(c)
	if !ok {
		return
	}
	session, err := h.session.Complete(c.Request.Context(), userID, sessionID)
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

	c.JSON(http.StatusOK, gin.H{"session": session})
	h.trackEvent(c.Request.Context(), userID, "session.complete", map[string]any{
		"session_id": sessionID,
	})
}

func (h *Handler) AddInterruption(c *gin.Context) {
	userID := getUserID(c)
	sessionID, ok := validateSessionID(c)
	if !ok {
		return
	}

	var req interruptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		reason = "manual interruption"
	}

	interruption, err := h.session.AddInterruption(c.Request.Context(), userID, sessionID, reason)
	if errors.Is(err, domain.ErrNotFound) {
		respondError(c, http.StatusNotFound, "session not found")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to log interruption")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"interruption": interruption})
}

func (h *Handler) UpsertReflection(c *gin.Context) {
	userID := getUserID(c)
	sessionID, ok := validateSessionID(c)
	if !ok {
		return
	}

	var req reflectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	req.WhatLearned = strings.TrimSpace(req.WhatLearned)
	req.WhatWasHard = strings.TrimSpace(req.WhatWasHard)
	req.NextAction = strings.TrimSpace(req.NextAction)

	if req.WhatLearned == "" || req.WhatWasHard == "" || req.NextAction == "" {
		respondError(c, http.StatusBadRequest, "all reflection fields are required")
		return
	}

	reflection, err := h.session.UpsertReflection(c.Request.Context(), userID, sessionID, usecase.ReflectionInput{
		WhatLearned: req.WhatLearned,
		WhatWasHard: req.WhatWasHard,
		NextAction:  req.NextAction,
	})
	if errors.Is(err, domain.ErrNotFound) {
		respondError(c, http.StatusNotFound, "session not found")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to save reflection")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"reflection": reflection})
}

func (h *Handler) DeleteSession(c *gin.Context) {
	userID := getUserID(c)
	sessionID := strings.TrimSpace(c.Param("sessionID"))

	if sessionID == "" {
		respondError(c, http.StatusBadRequest, "session id is required")
		return
	}

	err := h.session.DeleteSession(c.Request.Context(), userID, sessionID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			respondError(c, http.StatusNotFound, "session not found")
			return
		}
		if errors.Is(err, domain.ErrInvalidState) {
			respondError(c, http.StatusBadRequest, "cannot delete active or paused sessions")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to delete session")
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true})
	h.trackEvent(c.Request.Context(), userID, "session.deleted", map[string]any{
		"session_id": sessionID,
	})
}

func (h *Handler) DeleteReflection(c *gin.Context) {
	userID := getUserID(c)
	sessionID := strings.TrimSpace(c.Param("sessionID"))

	if sessionID == "" {
		respondError(c, http.StatusBadRequest, "session id is required")
		return
	}

	err := h.session.DeleteReflection(c.Request.Context(), userID, sessionID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			respondError(c, http.StatusNotFound, "session not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to delete reflection")
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}
