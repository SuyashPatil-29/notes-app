package middleware

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// OrgMembershipCache provides in-memory caching for organization membership checks
type OrgMembershipCache struct {
	cache map[string]*cacheEntry
	mu    sync.RWMutex
	ttl   time.Duration
}

type cacheEntry struct {
	role      string
	isMember  bool
	expiresAt time.Time
}

var (
	// Global cache instance
	orgCache *OrgMembershipCache
	once     sync.Once
)

// GetOrgCache returns the singleton cache instance
func GetOrgCache() *OrgMembershipCache {
	once.Do(func() {
		orgCache = &OrgMembershipCache{
			cache: make(map[string]*cacheEntry),
			ttl:   5 * time.Minute, // Cache for 5 minutes
		}
		// Start cleanup goroutine
		go orgCache.cleanup()
	})
	return orgCache
}

// Get retrieves cached membership data
func (c *OrgMembershipCache) Get(orgID, userID string) (role string, isMember bool, found bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := orgID + ":" + userID
	entry, exists := c.cache[key]
	if !exists {
		return "", false, false
	}

	// Check if expired
	if time.Now().After(entry.expiresAt) {
		return "", false, false
	}

	return entry.role, entry.isMember, true
}

// Set stores membership data in cache
func (c *OrgMembershipCache) Set(orgID, userID, role string, isMember bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := orgID + ":" + userID
	c.cache[key] = &cacheEntry{
		role:      role,
		isMember:  isMember,
		expiresAt: time.Now().Add(c.ttl),
	}
}

// Invalidate removes a specific cache entry
func (c *OrgMembershipCache) Invalidate(orgID, userID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := orgID + ":" + userID
	delete(c.cache, key)
}

// InvalidateOrg removes all cache entries for an organization
func (c *OrgMembershipCache) InvalidateOrg(orgID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key := range c.cache {
		if len(key) > len(orgID) && key[:len(orgID)] == orgID {
			delete(c.cache, key)
		}
	}
}

// cleanup periodically removes expired entries
func (c *OrgMembershipCache) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
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

// GetOrgMemberRoleCached is a cached version of GetOrgMemberRole
func GetOrgMemberRoleCached(ctx context.Context, orgID, userID string) (string, bool, error) {
	cache := GetOrgCache()

	// Try cache first
	if role, isMember, found := cache.Get(orgID, userID); found {
		log.Debug().Str("org_id", orgID).Str("user_id", userID).Msg("Org membership cache hit")
		return role, isMember, nil
	}

	// Cache miss - fetch from Clerk
	log.Debug().Str("org_id", orgID).Str("user_id", userID).Msg("Org membership cache miss")
	role, isMember, err := GetOrgMemberRole(ctx, orgID, userID)
	if err != nil {
		return "", false, err
	}

	// Store in cache
	cache.Set(orgID, userID, role, isMember)
	return role, isMember, nil
}
