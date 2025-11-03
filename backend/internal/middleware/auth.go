package middleware

import (
	"bytes"
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// responseCapture wraps http.ResponseWriter to capture status code and body
type responseCapture struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (rc *responseCapture) WriteHeader(code int) {
	rc.statusCode = code
	rc.ResponseWriter.WriteHeader(code)
}

func (rc *responseCapture) Write(b []byte) (int, error) {
	rc.body.Write(b)
	return rc.ResponseWriter.Write(b)
}

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
		// Debug: Log the Authorization header (truncated for security)
		authHeader := c.GetHeader("Authorization")
		truncatedAuth := authHeader
		if len(authHeader) > 50 {
			truncatedAuth = authHeader[:50] + "..."
		}
		log.Debug().Str("authorization_prefix", truncatedAuth).Msg("ClerkMiddleware: Processing request")

		// Create a response recorder to capture error responses
		var handlerCalled bool
		recorder := &responseCapture{
			ResponseWriter: c.Writer,
			statusCode:     http.StatusOK,
			body:           &bytes.Buffer{},
		}

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

		// Call the Clerk middleware with our response recorder
		handler.ServeHTTP(recorder, c.Request)

		// If Clerk middleware didn't call our handler, it means auth failed
		if !handlerCalled {
			errorBody := recorder.body.String()
			log.Warn().
				Int("status_code", recorder.statusCode).
				Str("error_response", errorBody).
				Msg("ClerkMiddleware: Authentication failed")

			// Check if it's likely a token expiration issue
			if recorder.statusCode == http.StatusUnauthorized {
				log.Info().Msg("ClerkMiddleware: Token may be expired - client should refresh token")
			}

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
