package httpapi

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// NewRouter builds the Gin engine with all routes wired.
func NewRouter(handler *Handler, corsOrigin string, enableDevLogin bool, validateToken func(string) (string, error)) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{corsOrigin},
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Public routes
	router.GET("/health", handler.Health)
	router.GET("/health/db", handler.DBHealth)
	router.HEAD("/health/db", handler.DBHealth)
	if enableDevLogin {
		router.POST("/api/v1/auth/dev-login", handler.DevLogin)
	}
	router.POST("/api/v1/auth/register", handler.Register)
	router.POST("/api/v1/auth/login", handler.Login)
	router.POST("/api/v1/auth/verify-email/resend", handler.ResendVerificationEmail)
	router.GET("/api/v1/auth/verify-email", handler.VerifyEmail)

	// Bot-to-bot routes (authenticated via X-Telegram-Bot-Auth header)
	router.POST("/api/v1/integrations/telegram/link", handler.TelegramLinkByCode)
	router.GET("/api/v1/integrations/telegram/status", handler.TelegramStatusByUserID)
	router.PATCH("/api/v1/integrations/telegram/notifications", handler.TelegramSetNotificationsByUserID)

	// Protected routes (JWT)
	api := router.Group("/api/v1")
	api.Use(AuthMiddleware(validateToken))
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
