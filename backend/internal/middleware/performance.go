package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// ResponseTimeMiddleware tracks and logs API request performance
func ResponseTimeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		durationMs := duration.Milliseconds()

		// Add response time header
		c.Header("X-Response-Time", fmt.Sprintf("%dms", durationMs))

		// Log slow queries
		if durationMs > 200 {
			log.Warn().
				Str("method", c.Request.Method).
				Str("path", c.Request.URL.Path).
				Int64("duration_ms", durationMs).
				Int("status", c.Writer.Status()).
				Msg("Slow API request detected")
		}

		// Log all requests in debug mode
		log.Debug().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int64("duration_ms", durationMs).
			Int("status", c.Writer.Status()).
			Msg("API request completed")
	}
}
