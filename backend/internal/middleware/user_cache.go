package middleware

import (
	"context"
	"sync"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/rs/zerolog/log"
)

// UserCache provides in-memory caching for Clerk user data
type UserCache struct {
	cache map[string]*userCacheEntry
	mu    sync.RWMutex
	ttl   time.Duration
}

type userCacheEntry struct {
	user      *clerk.User
	expiresAt time.Time
}

var (
	userCache     *UserCache
	userCacheOnce sync.Once
)

// GetUserCache returns the singleton user cache instance
func GetUserCache() *UserCache {
	userCacheOnce.Do(func() {
		userCache = &UserCache{
			cache: make(map[string]*userCacheEntry),
			ttl:   2 * time.Minute, // Cache for 2 minutes (shorter than org cache)
		}
		// Start cleanup goroutine
		go userCache.cleanup()
	})
	return userCache
}

// Get retrieves cached user data
func (c *UserCache) Get(userID string) (*clerk.User, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.cache[userID]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().After(entry.expiresAt) {
		return nil, false
	}

	return entry.user, true
}

// Set stores user data in cache
func (c *UserCache) Set(userID string, user *clerk.User) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[userID] = &userCacheEntry{
		user:      user,
		expiresAt: time.Now().Add(c.ttl),
	}
}

// Invalidate removes a specific cache entry
func (c *UserCache) Invalidate(userID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.cache, userID)
}

// cleanup periodically removes expired entries
func (c *UserCache) cleanup() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.cache {
			if now.After(entry.expiresAt) {
				delete(c.cache, key)
			}
		}
		c.mu.Unlock()
	}
}

// GetUserCached is a cached version of user.Get
func GetUserCached(ctx context.Context, userID string) (*clerk.User, error) {
	cache := GetUserCache()

	// Try cache first
	if cachedUser, found := cache.Get(userID); found {
		log.Debug().Str("user_id", userID).Msg("User cache hit")
		return cachedUser, nil
	}

	// Cache miss - fetch from Clerk
	log.Debug().Str("user_id", userID).Msg("User cache miss")
	clerkUser, err := user.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Store in cache
	cache.Set(userID, clerkUser)
	return clerkUser, nil
}
