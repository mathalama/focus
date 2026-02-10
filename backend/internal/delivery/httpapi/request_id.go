package httpapi

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const requestIDContextKey = "request_id"

// RequestIDMiddleware ensures each request has an ID for tracing and support.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := strings.TrimSpace(c.GetHeader("X-Request-ID"))
		if requestID == "" {
			requestID = uuid.NewString()
		}

		c.Set(requestIDContextKey, requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Next()
	}
}

func getRequestID(c *gin.Context) string {
	if requestID, ok := c.Get(requestIDContextKey); ok {
		if typed, castOK := requestID.(string); castOK {
			return typed
		}
	}
	return ""
}
