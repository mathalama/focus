package httpapi

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// StructuredLogger emits one JSON log line per HTTP request.
func StructuredLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()
		status := c.Writer.Status()
		duration := time.Since(startedAt)

		args := []any{
			slog.String("request_id", getRequestID(c)),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("route", c.FullPath()),
			slog.Int("status", status),
			slog.Duration("duration", duration),
			slog.String("ip", c.ClientIP()),
			slog.String("user_agent", c.Request.UserAgent()),
		}

		if len(c.Errors) > 0 {
			args = append(args, slog.String("error", c.Errors.String()))
		}

		if status >= 500 {
			slog.Error("http_request", args...)
		} else {
			slog.Info("http_request", args...)
		}
	}
}
