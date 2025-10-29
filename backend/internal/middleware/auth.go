package middleware

import (
	"backend/internal/auth"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// RequireAuth is middleware that checks if user is authenticated
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := auth.Store.Get(c.Request, "auth-session")
		if err != nil {
			log.Error().Err(err).Msg("Error getting session")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			c.Abort()
			return
		}

		userID, ok := session.Values["user_id"]
		if !ok || userID == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			c.Abort()
			return
		}

		// Store user info in context for use in handlers
		c.Set("user_id", userID)
		if email, ok := session.Values["email"]; ok {
			c.Set("email", email)
		}
		if name, ok := session.Values["name"]; ok {
			c.Set("name", name)
		}

		c.Next()
	}
}

// GetUserID retrieves the user ID from the context
func GetUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	// Handle different types that might be stored
	switch v := userID.(type) {
	case uint:
		return v, true
	case uint64:
		return uint(v), true
	case int:
		return uint(v), true
	case int64:
		return uint(v), true
	default:
		return 0, false
	}
}

// GetUserEmail retrieves the user email from the context
func GetUserEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get("email")
	if !exists {
		return "", false
	}
	emailStr, ok := email.(string)
	return emailStr, ok
}

// GetUserName retrieves the user name from the context
func GetUserName(c *gin.Context) (string, bool) {
	name, exists := c.Get("name")
	if !exists {
		return "", false
	}
	nameStr, ok := name.(string)
	return nameStr, ok
}
