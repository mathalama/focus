package http

import (
	"mathalama-focus/backend/internal/http/handlers"
	"mathalama-focus/backend/internal/http/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewRouter(handler *handlers.Handler, corsOrigin string) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{corsOrigin},
		AllowMethods:     []string{"GET", "POST", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "X-User-ID"},
		AllowCredentials: true,
	}))

	router.GET("/health", handler.Health)
	router.POST("/api/v1/auth/dev-login", handler.DevLogin)

	api := router.Group("/api/v1")
	api.Use(middleware.RequireUserID())
	{
		api.POST("/goals", handler.CreateGoal)
		api.GET("/goals", handler.ListGoals)

		api.POST("/sessions", handler.StartSession)
		api.PATCH("/sessions/:sessionID/pause", handler.PauseSession)
		api.PATCH("/sessions/:sessionID/resume", handler.ResumeSession)
		api.POST("/sessions/:sessionID/interruption", handler.AddInterruption)
		api.PATCH("/sessions/:sessionID/complete", handler.CompleteSession)
		api.POST("/sessions/:sessionID/reflection", handler.UpsertReflection)

		api.GET("/analytics/overview", handler.AnalyticsOverview)
	}

	return router
}
