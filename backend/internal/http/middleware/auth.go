package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const UserIDContextKey = "user_id"

func RequireUserID() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetHeader("X-User-ID")
		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing X-User-ID header"})
			return
		}

		if _, err := uuid.Parse(userID); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid X-User-ID header"})
			return
		}

		c.Set(UserIDContextKey, userID)
		c.Next()
	}
}

func UserID(c *gin.Context) string {
	if userID, ok := c.Get(UserIDContextKey); ok {
		if typed, castOK := userID.(string); castOK {
			return typed
		}
	}
	return ""
}
