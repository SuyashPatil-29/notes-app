package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
)

// Request-level cache keys
const (
	requestCacheKey = "request_cache"
)

type requestCache struct {
	orgMemberships map[string]orgMembershipData
}

type orgMembershipData struct {
	role     string
	isMember bool
}

// GetRequestCache retrieves or creates a request-level cache
func GetRequestCache(c *gin.Context) *requestCache {
	if cache, exists := c.Get(requestCacheKey); exists {
		return cache.(*requestCache)
	}

	cache := &requestCache{
		orgMemberships: make(map[string]orgMembershipData),
	}
	c.Set(requestCacheKey, cache)
	return cache
}

// GetOrgMemberRoleWithRequestCache checks request cache first, then memory cache, then Clerk API
func GetOrgMemberRoleWithRequestCache(ctx context.Context, c *gin.Context, orgID, userID string) (string, bool, error) {
	// Check request-level cache first (fastest)
	reqCache := GetRequestCache(c)
	key := orgID + ":" + userID
	if data, exists := reqCache.orgMemberships[key]; exists {
		return data.role, data.isMember, nil
	}

	// Fall back to memory cache and Clerk API
	role, isMember, err := GetOrgMemberRoleCached(ctx, orgID, userID)
	if err != nil {
		return "", false, err
	}

	// Store in request cache for this request
	reqCache.orgMemberships[key] = orgMembershipData{
		role:     role,
		isMember: isMember,
	}

	return role, isMember, nil
}
