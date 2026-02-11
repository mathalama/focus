package httpapi

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// NewRouter builds the Gin engine with all routes wired.
func NewRouter(handler *Handler, corsOrigin string, enableDevLogin bool, validateToken func(string) (string, error)) *gin.Engine {
	router := gin.New()
	router.Use(RequestIDMiddleware(), StructuredLogger(), MetricsMiddleware(), gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{corsOrigin},
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "Idempotency-Key", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: true,
	}))
	authLimiter := NewIPRateLimiter(rate.Every(12*time.Second), 10, 10*time.Minute).Middleware()
	botLimiter := NewIPRateLimiter(rate.Every(2*time.Second), 20, 10*time.Minute).Middleware()
	telegramIdempotency := NewIdempotencyStore(30 * time.Minute).Middleware()
	userIdempotency := NewIdempotencyStore(30 * time.Minute).Middleware()

	// Public routes
	router.GET("/health", handler.Health)
	router.GET("/health/db", handler.DBHealth)
	router.HEAD("/health/db", handler.DBHealth)
	router.GET("/metrics", MetricsRoute())
	if enableDevLogin {
		router.POST("/api/v1/auth/dev-login", authLimiter, handler.DevLogin)
	}
	router.POST("/api/v1/auth/register", authLimiter, handler.Register)
	router.POST("/api/v1/auth/login", authLimiter, handler.Login)
	router.POST("/api/v1/auth/refresh", authLimiter, handler.Refresh)
	router.POST("/api/v1/auth/logout", authLimiter, handler.Logout)
	router.POST("/api/v1/auth/verify-email/resend", authLimiter, handler.ResendVerificationEmail)
	router.GET("/api/v1/auth/verify-email", handler.VerifyEmail)
	router.POST("/api/v1/auth/forgot-password", authLimiter, handler.ForgotPassword)
	router.POST("/api/v1/auth/reset-password", authLimiter, handler.ResetPassword)

	// Bot-to-bot routes (authenticated via X-Telegram-Bot-Auth header)
	router.POST("/api/v1/integrations/telegram/link", botLimiter, telegramIdempotency, handler.TelegramLinkByCode)
	router.GET("/api/v1/integrations/telegram/status", botLimiter, handler.TelegramStatusByUserID)
	router.PATCH("/api/v1/integrations/telegram/notifications", botLimiter, handler.TelegramSetNotificationsByUserID)

	// Protected routes (JWT)
	api := router.Group("/api/v1")
	api.Use(AuthMiddleware(validateToken))
	{
		api.GET("/me", handler.GetMe)
		api.POST("/auth/logout-all", handler.LogoutAll)
		api.GET("/auth/sessions", handler.ListAuthSessions)
		api.POST("/auth/telegram/link-code", handler.CreateTelegramLinkCode)
		api.GET("/integrations/telegram", handler.GetTelegramIdentity)
		api.DELETE("/integrations/telegram", handler.UnlinkTelegram)

		api.POST("/goals", handler.CreateGoal)
		api.GET("/goals", handler.ListGoals)
		api.GET("/goals/history", handler.ListGoalHistory)
		api.GET("/goals/:goalId", handler.GetGoal)
		api.PATCH("/goals/:goalId", handler.UpdateGoal)
		api.DELETE("/goals/:goalId", handler.DeleteGoal)

		api.POST("/sessions", handler.StartSession)
		api.GET("/sessions/active", handler.GetActiveSession)
		api.GET("/sessions/history", handler.ListSessionHistory)
		api.GET("/sessions/:sessionID", handler.GetSession)
		api.PATCH("/sessions/:sessionID/pause", handler.PauseSession)
		api.PATCH("/sessions/:sessionID/resume", handler.ResumeSession)
		api.PATCH("/sessions/:sessionID/reset", handler.ResetSession)
		api.PATCH("/sessions/:sessionID/abandon", handler.AbandonSession)
		api.POST("/sessions/:sessionID/interruption", handler.AddInterruption)
		api.PATCH("/sessions/:sessionID/complete", userIdempotency, handler.CompleteSession)
		api.POST("/sessions/:sessionID/reflection", handler.UpsertReflection)
		api.DELETE("/sessions/:sessionID", handler.DeleteSession)
		api.DELETE("/sessions/:sessionID/reflection", handler.DeleteReflection)

		api.GET("/analytics/overview", handler.AnalyticsOverview)
		api.GET("/analytics/activity", handler.GetDailyActivity)
		api.GET("/analytics/activity/day", handler.GetDailyContributions)
		api.GET("/analytics/insights", handler.GetInsights)

		api.GET("/notifications/sounds", handler.GetNotificationSounds)
		api.PATCH("/notifications/sounds", handler.UpdateNotificationSounds)

		api.GET("/leaderboard", handler.GetLeaderboard)

		api.GET("/shop/items", handler.ListItems)
		api.POST("/shop/items/:itemID/buy", userIdempotency, handler.BuyItem)

		admin := api.Group("/admin")
		{
			admin.GET("/health", handler.AdminHealth)
			admin.GET("/email-deliveries", handler.AdminListEmailDeliveries)
			admin.GET("/events", handler.AdminListEvents)
		}
	}

	return router
}
