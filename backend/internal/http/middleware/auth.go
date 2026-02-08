package middleware

import (
	"mathalama-focus/backend/internal/auth"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const UserIDContextKey = "user_id"

func RequireUserID(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing Authorization header"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid Authorization header format"})
			return
		}

		tokenString := parts[1]
		userID, err := auth.ValidateToken(jwtSecret, tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token: " + err.Error()})
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
