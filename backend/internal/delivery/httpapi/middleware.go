package httpapi

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const userIDContextKey = "user_id"

func AuthMiddleware(validateToken func(string) (string, error)) gin.HandlerFunc {
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

		userID, err := validateToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token: " + err.Error()})
			return
		}

		c.Set(userIDContextKey, userID)
		c.Next()
	}
}

func getUserID(c *gin.Context) string {
	if userID, ok := c.Get(userIDContextKey); ok {
		if typed, castOK := userID.(string); castOK {
			return typed
		}
	}
	return ""
}
