package middleware

import (
	"context"
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/organizationmembership"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// GetActiveOrgID extracts organization ID from query parameter or header
func GetActiveOrgID(c *gin.Context) string {
	// Try query parameter first
	orgID := c.Query("organizationId")
	if orgID != "" {
		return orgID
	}

	// Try header
	orgID = c.GetHeader("X-Organization-ID")
	return orgID
}

// GetOrgMemberRole fetches the user's role in an organization from Clerk
// Returns role ("admin" or "member") and whether the user is a member
func GetOrgMemberRole(ctx context.Context, orgID, userID string) (string, bool, error) {
	// Check if user is a member by fetching memberships for this specific organization
	params := &organizationmembership.ListParams{}
	params.Limit = clerk.Int64(100)
	params.OrganizationID = orgID
	params.UserIDs = []string{userID}

	memberships, err := organizationmembership.List(ctx, params)
	if err != nil {
		log.Error().Err(err).Str("org_id", orgID).Str("user_id", userID).Msg("Failed to fetch org memberships")
		return "", false, err
	}

	// Find the membership for this user in this org
	var membership *clerk.OrganizationMembership
	for _, m := range memberships.OrganizationMemberships {
		if m.Organization.ID == orgID && m.PublicUserData.UserID == userID {
			membership = m
			break
		}
	}

	// Check if user is a member
	if membership == nil {
		return "", false, nil
	}

	// Extract role (admin or member)
	// In Clerk, the role is stored as "org:admin" or "org:member"
	role := membership.Role
	isAdmin := role == "org:admin"

	if isAdmin {
		return "admin", true, nil
	}
	return "member", true, nil
}

// RequireOrgMembership middleware verifies that the user is a member of the organization
func RequireOrgMembership() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Clerk user ID from context (set by RequireAuth middleware)
		clerkUserID, exists := GetClerkUserID(c)
		if !exists {
			log.Error().Msg("RequireOrgMembership: No Clerk user ID in context")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			c.Abort()
			return
		}

		// Get organization ID from URL parameter
		orgID := c.Param("orgId")
		if orgID == "" {
			// Try getting from query or header
			orgID = GetActiveOrgID(c)
		}

		if orgID == "" {
			log.Error().Msg("RequireOrgMembership: No organization ID provided")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Organization ID required"})
			c.Abort()
			return
		}

		// Check membership using Clerk API (cached)
		role, isMember, err := GetOrgMemberRoleCached(c.Request.Context(), orgID, clerkUserID)
		if err != nil {
			log.Error().Err(err).Str("org_id", orgID).Str("user_id", clerkUserID).Msg("Failed to check org membership")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify organization membership"})
			c.Abort()
			return
		}

		if !isMember {
			log.Warn().Str("org_id", orgID).Str("user_id", clerkUserID).Msg("User is not a member of organization")
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not a member of this organization"})
			c.Abort()
			return
		}

		// Store org ID and role in context for later use
		c.Set("organization_id", orgID)
		c.Set("organization_role", role)

		log.Debug().Str("org_id", orgID).Str("user_id", clerkUserID).Str("role", role).Msg("Organization membership verified")
		c.Next()
	}
}

// RequireOrgAdmin middleware verifies that the user is an admin of the organization
func RequireOrgAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Clerk user ID from context
		clerkUserID, exists := GetClerkUserID(c)
		if !exists {
			log.Error().Msg("RequireOrgAdmin: No Clerk user ID in context")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			c.Abort()
			return
		}

		// Get organization ID from URL parameter
		orgID := c.Param("orgId")
		if orgID == "" {
			orgID = GetActiveOrgID(c)
		}

		if orgID == "" {
			log.Error().Msg("RequireOrgAdmin: No organization ID provided")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Organization ID required"})
			c.Abort()
			return
		}

		// Check membership and role using Clerk API (cached)
		role, isMember, err := GetOrgMemberRoleCached(c.Request.Context(), orgID, clerkUserID)
		if err != nil {
			log.Error().Err(err).Str("org_id", orgID).Str("user_id", clerkUserID).Msg("Failed to check org membership")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify organization membership"})
			c.Abort()
			return
		}

		if !isMember {
			log.Warn().Str("org_id", orgID).Str("user_id", clerkUserID).Msg("User is not a member of organization")
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not a member of this organization"})
			c.Abort()
			return
		}

		if role != "admin" {
			log.Warn().Str("org_id", orgID).Str("user_id", clerkUserID).Str("role", role).Msg("User is not an admin of organization")
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin privileges required"})
			c.Abort()
			return
		}

		// Store org ID and role in context
		c.Set("organization_id", orgID)
		c.Set("organization_role", role)

		log.Debug().Str("org_id", orgID).Str("user_id", clerkUserID).Msg("Organization admin verified")
		c.Next()
	}
}

// GetOrgIDFromContext retrieves organization ID from Gin context
func GetOrgIDFromContext(c *gin.Context) (string, bool) {
	orgID, exists := c.Get("organization_id")
	if !exists {
		return "", false
	}

	orgIDStr, ok := orgID.(string)
	return orgIDStr, ok
}

// GetOrgRoleFromContext retrieves organization role from Gin context
func GetOrgRoleFromContext(c *gin.Context) (string, bool) {
	role, exists := c.Get("organization_role")
	if !exists {
		return "", false
	}

	roleStr, ok := role.(string)
	return roleStr, ok
}
