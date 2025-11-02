package controllers

import (
	"backend/internal/middleware"
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/organization"
	"github.com/clerk/clerk-sdk-go/v2/organizationmembership"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// CreateOrganization creates a new organization via Clerk
func CreateOrganization(c *gin.Context) {
	// Get authenticated user ID
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Parse request body
	var req struct {
		Name string `json:"name" binding:"required"`
		Slug string `json:"slug"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Organization name is required"})
		return
	}

	// Create organization via Clerk API
	params := &organization.CreateParams{
		Name:      clerk.String(req.Name),
		CreatedBy: clerk.String(clerkUserID),
	}

	if req.Slug != "" {
		params.Slug = clerk.String(req.Slug)
	}

	org, err := organization.Create(c.Request.Context(), params)
	if err != nil {
		log.Error().Err(err).Str("user_id", clerkUserID).Str("name", req.Name).Msg("Failed to create organization")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create organization"})
		return
	}

	log.Info().Str("org_id", org.ID).Str("name", org.Name).Str("user_id", clerkUserID).Msg("Organization created")
	c.JSON(http.StatusCreated, gin.H{
		"id":           org.ID,
		"name":         org.Name,
		"slug":         org.Slug,
		"imageUrl":     org.ImageURL,
		"createdAt":    org.CreatedAt,
		"membersCount": org.MembersCount,
	})
}

// ListUserOrganizations fetches all organizations the user is a member of
func ListUserOrganizations(c *gin.Context) {
	// Get authenticated user ID
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Fetch user's organization memberships
	params := &organizationmembership.ListParams{}
	params.Limit = clerk.Int64(100) // Fetch up to 100 orgs
	params.UserIDs = []string{clerkUserID}

	memberships, err := organizationmembership.List(c.Request.Context(), params)
	if err != nil {
		log.Error().Err(err).Str("user_id", clerkUserID).Msg("Failed to fetch user organizations")
		// Return empty list instead of error if it's just a 404 (no orgs found)
		c.JSON(http.StatusOK, gin.H{
			"organizations": []gin.H{},
			"total":         0,
		})
		return
	}

	// Build response with organization details for this user
	orgs := make([]gin.H, 0)
	for _, membership := range memberships.OrganizationMemberships {
		// Map Clerk role to our simplified role format
		role := "member"
		if membership.Role == "org:admin" {
			role = "admin"
		}

		orgs = append(orgs, gin.H{
			"id":           membership.Organization.ID,
			"name":         membership.Organization.Name,
			"slug":         membership.Organization.Slug,
			"imageUrl":     membership.Organization.ImageURL,
			"role":         role,
			"createdAt":    membership.Organization.CreatedAt,
			"membersCount": membership.Organization.MembersCount,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"organizations": orgs,
		"total":         len(orgs),
	})
}

// GetOrganization fetches details for a specific organization
func GetOrganization(c *gin.Context) {
	orgID := c.Param("orgId")

	// Fetch organization from Clerk
	org, err := organization.Get(c.Request.Context(), orgID)
	if err != nil {
		log.Error().Err(err).Str("org_id", orgID).Msg("Failed to fetch organization")
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		return
	}

	// Get user's role in this org (set by RequireOrgMembership middleware)
	role, _ := middleware.GetOrgRoleFromContext(c)

	c.JSON(http.StatusOK, gin.H{
		"id":           org.ID,
		"name":         org.Name,
		"slug":         org.Slug,
		"imageUrl":     org.ImageURL,
		"createdAt":    org.CreatedAt,
		"membersCount": org.MembersCount,
		"role":         role,
	})
}

// UpdateOrganization updates organization settings (admin only)
func UpdateOrganization(c *gin.Context) {
	orgID := c.Param("orgId")

	// Parse request body
	var req struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Build update params
	params := &organization.UpdateParams{}
	if req.Name != "" {
		params.Name = clerk.String(req.Name)
	}
	if req.Slug != "" {
		params.Slug = clerk.String(req.Slug)
	}

	// Update organization via Clerk API
	org, err := organization.Update(c.Request.Context(), orgID, params)
	if err != nil {
		log.Error().Err(err).Str("org_id", orgID).Msg("Failed to update organization")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update organization"})
		return
	}

	log.Info().Str("org_id", orgID).Msg("Organization updated")
	c.JSON(http.StatusOK, gin.H{
		"id":           org.ID,
		"name":         org.Name,
		"slug":         org.Slug,
		"imageUrl":     org.ImageURL,
		"createdAt":    org.CreatedAt,
		"membersCount": org.MembersCount,
	})
}

// DeleteOrganization deletes an organization (admin only)
func DeleteOrganization(c *gin.Context) {
	orgID := c.Param("orgId")

	// Delete organization via Clerk API
	_, err := organization.Delete(c.Request.Context(), orgID)
	if err != nil {
		log.Error().Err(err).Str("org_id", orgID).Msg("Failed to delete organization")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete organization"})
		return
	}

	log.Info().Str("org_id", orgID).Msg("Organization deleted")
	c.JSON(http.StatusOK, gin.H{"message": "Organization deleted successfully"})
}
