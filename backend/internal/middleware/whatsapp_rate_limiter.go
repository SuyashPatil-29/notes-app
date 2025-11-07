package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// WhatsAppRateLimiter manages rate limiting for WhatsApp users
type WhatsAppRateLimiter struct {
	limiters        map[string]*rateLimiterEntry
	mu              sync.RWMutex
	rate            rate.Limit
	burst           int
	cleanupInterval time.Duration
	metricsService  MetricsService
}

// MetricsService interface for recording metrics
type MetricsService interface {
	RecordRateLimitHit()
	RecordRateLimitBlocked()
}

type rateLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewWhatsAppRateLimiter creates a new rate limiter
// messagesPerMinute: number of messages allowed per minute per user
func NewWhatsAppRateLimiter(messagesPerMinute int, metricsService MetricsService) *WhatsAppRateLimiter {
	rl := &WhatsAppRateLimiter{
		limiters:        make(map[string]*rateLimiterEntry),
		rate:            rate.Limit(float64(messagesPerMinute) / 60.0), // Convert to per-second rate
		burst:           messagesPerMinute,
		cleanupInterval: 5 * time.Minute,
		metricsService:  metricsService,
	}

	// Start cleanup goroutine
	go rl.cleanupRoutine()

	return rl
}

// Allow checks if a message from the given phone number is allowed
func (rl *WhatsAppRateLimiter) Allow(phoneNumber string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	entry, exists := rl.limiters[phoneNumber]
	if !exists {
		entry = &rateLimiterEntry{
			limiter:  rate.NewLimiter(rl.rate, rl.burst),
			lastSeen: time.Now(),
		}
		rl.limiters[phoneNumber] = entry
	} else {
		entry.lastSeen = time.Now()
	}

	allowed := entry.limiter.Allow()

	// Record metrics
	if rl.metricsService != nil {
		rl.metricsService.RecordRateLimitHit()
		if !allowed {
			rl.metricsService.RecordRateLimitBlocked()
		}
	}

	return allowed
}

// cleanupRoutine periodically removes inactive limiters
func (rl *WhatsAppRateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanupInactiveLimiters()
	}
}

// cleanupInactiveLimiters removes limiters that haven't been used recently
func (rl *WhatsAppRateLimiter) cleanupInactiveLimiters() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for phoneNumber, entry := range rl.limiters {
		if now.Sub(entry.lastSeen) > rl.cleanupInterval {
			delete(rl.limiters, phoneNumber)
		}
	}
}

// Middleware returns a Gin middleware function for rate limiting
func (rl *WhatsAppRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract phone number from request context or body
		// This will be set by the webhook handler after parsing the message
		phoneNumber, exists := c.Get("whatsapp_phone_number")
		if !exists {
			// If phone number not set, allow the request to proceed
			// The webhook handler will extract it
			c.Next()
			return
		}

		phone := phoneNumber.(string)
		if !rl.Allow(phone) {
			c.JSON(429, gin.H{
				"error":   "rate_limit_exceeded",
				"message": "Too many messages. Please wait a moment before sending another message.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetLimiterCount returns the number of active limiters (for testing/monitoring)
func (rl *WhatsAppRateLimiter) GetLimiterCount() int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return len(rl.limiters)
}

// Global rate limiter instance
var globalRateLimiter *WhatsAppRateLimiter

// SetWhatsAppRateLimiter sets the global rate limiter instance
func SetWhatsAppRateLimiter(limiter *WhatsAppRateLimiter) {
	globalRateLimiter = limiter
}

// WhatsAppRateLimiterMiddleware returns the global rate limiter middleware
func WhatsAppRateLimiterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if globalRateLimiter == nil {
			c.Next()
			return
		}
		globalRateLimiter.Middleware()(c)
	}
}
