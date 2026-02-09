package http

import (
	"mathalama-focus/backend/internal/http/handlers"
	"mathalama-focus/backend/internal/http/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewRouter(handler *handlers.Handler, corsOrigin string, jwtSecret string) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{corsOrigin},
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	router.GET("/health", handler.Health)
	router.POST("/api/v1/auth/dev-login", handler.DevLogin)
	router.POST("/api/v1/integrations/telegram/link", handler.TelegramLinkByCode)
	router.GET("/api/v1/integrations/telegram/status", handler.TelegramStatusByUserID)

	api := router.Group("/api/v1")
	api.Use(middleware.RequireUserID(jwtSecret))
	{
		api.GET("/me", handler.GetMe)
		api.POST("/auth/telegram/link-code", handler.CreateTelegramLinkCode)
		api.GET("/integrations/telegram", handler.GetTelegramIdentity)
		api.DELETE("/integrations/telegram", handler.UnlinkTelegram)

		api.POST("/goals", handler.CreateGoal)
		api.GET("/goals", handler.ListGoals)
		api.GET("/goals/history", handler.ListGoalHistory)

		api.POST("/sessions", handler.StartSession)
		api.GET("/sessions/active", handler.GetActiveSession)
		api.GET("/sessions/history", handler.ListSessionHistory)
		api.GET("/sessions/:sessionID", handler.GetSession)
		api.PATCH("/sessions/:sessionID/pause", handler.PauseSession)
		api.PATCH("/sessions/:sessionID/resume", handler.ResumeSession)
		api.PATCH("/sessions/:sessionID/reset", handler.ResetSession)
		api.PATCH("/sessions/:sessionID/abandon", handler.AbandonSession)
		api.POST("/sessions/:sessionID/interruption", handler.AddInterruption)
		api.PATCH("/sessions/:sessionID/complete", handler.CompleteSession)
		api.POST("/sessions/:sessionID/reflection", handler.UpsertReflection)

		api.GET("/analytics/overview", handler.AnalyticsOverview)
		api.GET("/analytics/activity", handler.GetDailyActivity)
		api.GET("/analytics/activity/day", handler.GetDailyContributions)
		api.GET("/analytics/insights", handler.GetInsights)

		api.GET("/leaderboard", handler.GetLeaderboard)

		api.GET("/shop/items", handler.ListItems)
		api.POST("/shop/items/:itemID/buy", handler.BuyItem)
	}

	return router
}
