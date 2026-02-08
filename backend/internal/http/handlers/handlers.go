package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"mathalama-focus/backend/internal/auth"
	"mathalama-focus/backend/internal/domain"
	"mathalama-focus/backend/internal/http/middleware"
	"mathalama-focus/backend/internal/repository/postgresql"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Repository interface {
	DevLogin(ctx context.Context, email, name string) (domain.User, error)
	CreateGoal(ctx context.Context, userID string, input postgresql.CreateGoalInput) (domain.Goal, error)
	ListGoals(ctx context.Context, userID string) ([]domain.Goal, error)
	StartSession(ctx context.Context, userID string, input postgresql.StartSessionInput) (domain.FocusSession, error)
	PauseSession(ctx context.Context, userID, sessionID string) (domain.FocusSession, error)
	ResumeSession(ctx context.Context, userID, sessionID string) (domain.FocusSession, error)
	AddInterruption(ctx context.Context, userID, sessionID, reason string) (domain.Interruption, error)
	CompleteSession(ctx context.Context, userID, sessionID string) (domain.FocusSession, error)
	UpsertReflection(ctx context.Context, userID, sessionID string, input postgresql.ReflectionInput) (domain.Reflection, error)
	AnalyticsOverview(ctx context.Context, userID string) (domain.AnalyticsOverview, error)
}

type Handler struct {
	repo      Repository
	jwtSecret string
}

func New(repo Repository, jwtSecret string) *Handler {
	return &Handler{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type devLoginRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (h *Handler) DevLogin(c *gin.Context) {
	var req devLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Name = strings.TrimSpace(req.Name)

	if req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}
	if req.Name == "" {
		req.Name = "Focus Learner"
	}

	user, err := h.repo.DevLogin(c.Request.Context(), req.Email, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login"})
		return
	}

	token, err := auth.GenerateToken(h.jwtSecret, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": token,
	})
}

type createGoalRequest struct {
	Topic              string `json:"topic"`
	DesiredResult      string `json:"desired_result"`
	RecommendedMinutes int    `json:"recommended_minutes"`
}

func (h *Handler) CreateGoal(c *gin.Context) {
	userID := middleware.UserID(c)

	var req createGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	req.Topic = strings.TrimSpace(req.Topic)
	req.DesiredResult = strings.TrimSpace(req.DesiredResult)

	if req.Topic == "" || req.DesiredResult == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "topic and desired_result are required"})
		return
	}

	goal, err := h.repo.CreateGoal(c.Request.Context(), userID, postgresql.CreateGoalInput{
		Topic:              req.Topic,
		DesiredResult:      req.DesiredResult,
		RecommendedMinutes: req.RecommendedMinutes,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create goal"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"goal": goal})
}

func (h *Handler) ListGoals(c *gin.Context) {
	userID := middleware.UserID(c)

	goals, err := h.repo.ListGoals(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list goals"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"goals": goals})
}

type startSessionRequest struct {
	GoalID             string `json:"goal_id"`
	RecommendedMinutes int    `json:"recommended_minutes"`
}

func (h *Handler) StartSession(c *gin.Context) {
	userID := middleware.UserID(c)

	var req startSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	req.GoalID = strings.TrimSpace(req.GoalID)
	if req.GoalID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "goal_id is required"})
		return
	}

	session, err := h.repo.StartSession(c.Request.Context(), userID, postgresql.StartSessionInput{
		GoalID:             req.GoalID,
		RecommendedMinutes: req.RecommendedMinutes,
	})
	if errors.Is(err, postgresql.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "goal not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start session"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"session": session})
}

func (h *Handler) PauseSession(c *gin.Context) {
	h.sessionAction(c, func(userID, sessionID string) (any, error) {
		return h.repo.PauseSession(c.Request.Context(), userID, sessionID)
	})
}

func (h *Handler) ResumeSession(c *gin.Context) {
	h.sessionAction(c, func(userID, sessionID string) (any, error) {
		return h.repo.ResumeSession(c.Request.Context(), userID, sessionID)
	})
}

type interruptionRequest struct {
	Reason string `json:"reason"`
}

func (h *Handler) AddInterruption(c *gin.Context) {
	userID := middleware.UserID(c)
	sessionID := strings.TrimSpace(c.Param("sessionID"))
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessionID is required"})
		return
	}

	if _, err := uuid.Parse(sessionID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sessionID format"})
		return
	}

	var req interruptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	req.Reason = strings.TrimSpace(req.Reason)
	if req.Reason == "" {
		req.Reason = "manual interruption"
	}

	interruption, err := h.repo.AddInterruption(c.Request.Context(), userID, sessionID, req.Reason)
	if errors.Is(err, postgresql.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to log interruption"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"interruption": interruption})
}

func (h *Handler) CompleteSession(c *gin.Context) {
	h.sessionAction(c, func(userID, sessionID string) (any, error) {
		return h.repo.CompleteSession(c.Request.Context(), userID, sessionID)
	})
}

type reflectionRequest struct {
	WhatLearned string `json:"what_learned"`
	WhatWasHard string `json:"what_was_hard"`
	NextAction  string `json:"next_action"`
}

func (h *Handler) UpsertReflection(c *gin.Context) {
	userID := middleware.UserID(c)
	sessionID := strings.TrimSpace(c.Param("sessionID"))
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessionID is required"})
		return
	}

	if _, err := uuid.Parse(sessionID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sessionID format"})
		return
	}

	var req reflectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	req.WhatLearned = strings.TrimSpace(req.WhatLearned)
	req.WhatWasHard = strings.TrimSpace(req.WhatWasHard)
	req.NextAction = strings.TrimSpace(req.NextAction)

	if req.WhatLearned == "" || req.WhatWasHard == "" || req.NextAction == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "all reflection fields are required"})
		return
	}

	reflection, err := h.repo.UpsertReflection(c.Request.Context(), userID, sessionID, postgresql.ReflectionInput{
		WhatLearned: req.WhatLearned,
		WhatWasHard: req.WhatWasHard,
		NextAction:  req.NextAction,
	})
	if errors.Is(err, postgresql.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save reflection"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"reflection": reflection})
}

func (h *Handler) AnalyticsOverview(c *gin.Context) {
	userID := middleware.UserID(c)

	overview, err := h.repo.AnalyticsOverview(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load analytics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"overview": overview})
}

func (h *Handler) sessionAction(c *gin.Context, action func(userID, sessionID string) (any, error)) {
	userID := middleware.UserID(c)
	sessionID := strings.TrimSpace(c.Param("sessionID"))
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessionID is required"})
		return
	}

	if _, err := uuid.Parse(sessionID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sessionID format"})
		return
	}

	payload, err := action(userID, sessionID)
	if errors.Is(err, postgresql.ErrPauseLimitReached) {
		c.JSON(http.StatusConflict, gin.H{"error": "pause limit reached"})
		return
	}
	if errors.Is(err, postgresql.ErrInvalidState) {
		c.JSON(http.StatusConflict, gin.H{"error": "session cannot perform this action"})
		return
	}
	if errors.Is(err, postgresql.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "action failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"session": payload})
}
