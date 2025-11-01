package middleware

import (
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// RequireAuth is middleware that validates Clerk JWT tokens
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract Clerk session claims from context (set by Clerk middleware)
		claims, ok := clerk.SessionClaimsFromContext(c.Request.Context())
		if !ok || claims == nil {
			log.Error().Msg("No Clerk session found in context")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			c.Abort()
			return
		}

		clerkUserID := claims.Subject

		// Store Clerk user ID in context (no database user lookup needed)
		c.Set("clerk_user_id", clerkUserID)

		c.Next()
	}
}

// ClerkMiddleware wraps Clerk's header authorization middleware for Gin
func ClerkMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Debug: Log the Authorization header
		authHeader := c.GetHeader("Authorization")
		log.Debug().Str("authorization", authHeader).Msg("ClerkMiddleware: Processing request")

		// Create a response recorder to intercept writes
		var handlerCalled bool

		// Create a handler that uses Clerk's middleware
		handler := clerkhttp.WithHeaderAuthorization()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Update the request with the modified context (contains Clerk session claims)
			c.Request = r
			handlerCalled = true

			// Debug: Check if session claims are in context
			claims, ok := clerk.SessionClaimsFromContext(r.Context())
			if ok && claims != nil {
				log.Debug().Str("user_id", claims.Subject).Msg("ClerkMiddleware: Session claims found")
			} else {
				log.Warn().Msg("ClerkMiddleware: No session claims in context after Clerk middleware")
			}
		}))

		// Call the Clerk middleware
		handler.ServeHTTP(c.Writer, c.Request)

		// If Clerk middleware didn't call our handler, it means auth failed
		if !handlerCalled {
			log.Warn().Msg("ClerkMiddleware: Handler not called - auth failed")
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetClerkUserID retrieves the Clerk user ID from the context
func GetClerkUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("clerk_user_id")
	if !exists {
		return "", false
	}

	userIDStr, ok := userID.(string)
	return userIDStr, ok
}

// GetUserID is deprecated - use GetClerkUserID instead
// Kept for backwards compatibility but will be removed
func GetUserID(c *gin.Context) (string, bool) {
	return GetClerkUserID(c)
}
