package controllers

import (
	"backend/internal/middleware"
	"backend/internal/services"
	"backend/internal/types"
	internalutils "backend/internal/utils"
	"backend/pkg/utils"
	"net/http"
	"strings"

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

// GetOrgAPICredentials returns the status of organization API credentials
func GetOrgAPICredentials(c *gin.Context) {
	orgID := c.Param("orgId")

	// Get organization API key manager
	orgAPIKeyManager := services.NewOrgAPIKeyManager()

	// Get organization API credentials
	credentials, err := orgAPIKeyManager.GetOrgAPICredentials(orgID)
	if err != nil {
		log.Error().Err(err).Str("org_id", orgID).Msg("Failed to get organization API credentials")
		apiErr := types.ErrDatabaseError.WithDetails("Failed to retrieve organization API credentials")
		internalutils.SendErrorResponse(c, apiErr)
		return
	}

	// Return the credentials status
	internalutils.SendSuccessResponse(c, gin.H{
		"credentials": credentials,
		"total":       len(credentials),
	})
}

// SetOrgAPICredential sets an API credential for the organization (admin only)
func SetOrgAPICredential(c *gin.Context) {
	orgID := c.Param("orgId")

	// Get authenticated user ID
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		internalutils.SendErrorResponse(c, types.ErrUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		Provider string `json:"provider" binding:"required"`
		ApiKey   string `json:"apiKey" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr := types.ErrMissingFields.WithDetails("Provider and API key are required")
		internalutils.SendErrorResponse(c, apiErr)
		return
	}

	// Validate provider
	validProviders := map[string]bool{
		"openai":    true,
		"anthropic": true,
		"google":    true,
	}
	if !validProviders[req.Provider] {
		internalutils.SendErrorResponse(c, types.ErrInvalidProvider)
		return
	}

	// Validate API key format
	if strings.TrimSpace(req.ApiKey) == "" {
		apiErr := types.ErrAPIKeyInvalid.WithDetails("API key cannot be empty")
		internalutils.SendErrorResponse(c, apiErr)
		return
	}

	// Get organization API key manager
	orgAPIKeyManager := services.NewOrgAPIKeyManager()

	// Set the organization API credential
	err := orgAPIKeyManager.SetOrgAPICredential(orgID, req.Provider, req.ApiKey, clerkUserID)
	if err != nil {
		log.Error().Err(err).Str("org_id", orgID).Str("provider", req.Provider).Msg("Failed to set organization API credential")

		// Determine appropriate error response
		var apiErr *types.APIError
		if strings.Contains(err.Error(), "encrypt") {
			apiErr = types.ErrEncryptionFailed.WithDetails(err.Error())
		} else {
			apiErr = types.ErrDatabaseError.WithDetails("Failed to save API credential")
		}

		internalutils.SendErrorResponse(c, apiErr)
		return
	}

	// Return masked key for UI display
	maskedKey := utils.MaskAPIKey(strings.TrimSpace(req.ApiKey))

	log.Info().Str("org_id", orgID).Str("provider", req.Provider).Str("created_by", clerkUserID).Msg("Organization API credential set")

	internalutils.SendSuccessMessageResponse(c, "Organization API key saved successfully", gin.H{
		"provider":  req.Provider,
		"maskedKey": maskedKey,
	})
}

// DeleteOrgAPICredential removes an API credential from the organization (admin only)
func DeleteOrgAPICredential(c *gin.Context) {
	orgID := c.Param("orgId")

	// Parse request body
	var req struct {
		Provider string `json:"provider" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr := types.ErrMissingFields.WithDetails("Provider is required")
		internalutils.SendErrorResponse(c, apiErr)
		return
	}

	// Validate provider
	validProviders := map[string]bool{
		"openai":    true,
		"anthropic": true,
		"google":    true,
	}
	if !validProviders[req.Provider] {
		internalutils.SendErrorResponse(c, types.ErrInvalidProvider)
		return
	}

	// Get organization API key manager
	orgAPIKeyManager := services.NewOrgAPIKeyManager()

	// Delete the organization API credential
	err := orgAPIKeyManager.DeleteOrgAPICredential(orgID, req.Provider)
	if err != nil {
		log.Error().Err(err).Str("org_id", orgID).Str("provider", req.Provider).Msg("Failed to delete organization API credential")

		// Determine appropriate error response
		var apiErr *types.APIError
		if strings.Contains(err.Error(), "not found") {
			apiErr = types.ErrAPIKeyNotFound.WithDetails("API credential not found for this provider")
		} else {
			apiErr = types.ErrDatabaseError.WithDetails("Failed to delete API credential")
		}

		internalutils.SendErrorResponse(c, apiErr)
		return
	}

	log.Info().Str("org_id", orgID).Str("provider", req.Provider).Msg("Organization API credential deleted")

	internalutils.SendSuccessMessageResponse(c, "Organization API key deleted successfully", gin.H{
		"provider": req.Provider,
	})
}

// GetOrganizationMembers fetches all members of an organization for task assignment
func GetOrganizationMembers(c *gin.Context) {
	orgID := c.Param("orgId")

	// Get authenticated user ID
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Verify user is a member of the organization
	_, isMember, err := middleware.GetOrgMemberRole(c.Request.Context(), orgID, clerkUserID)
	if err != nil || !isMember {
		log.Warn().Str("org_id", orgID).Str("user_id", clerkUserID).Msg("User not authorized to view organization members")
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not a member of this organization"})
		return
	}

	// Fetch organization memberships from Clerk
	params := &organizationmembership.ListParams{}
	params.Limit = clerk.Int64(100) // Fetch up to 100 members
	params.OrganizationID = orgID

	memberships, err := organizationmembership.List(c.Request.Context(), params)
	if err != nil {
		log.Error().Err(err).Str("org_id", orgID).Msg("Failed to fetch organization members")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch organization members"})
		return
	}

	// Build response with member details
	members := make([]gin.H, 0)
	for _, membership := range memberships.OrganizationMemberships {
		// Map Clerk role to our simplified role format
		role := "member"
		if membership.Role == "org:admin" {
			role = "admin"
		}

		members = append(members, gin.H{
			"id":       membership.PublicUserData.UserID,
			"name":     getDisplayName(membership.PublicUserData),
			"email":    membership.PublicUserData.Identifier,
			"imageUrl": membership.PublicUserData.ImageURL,
			"role":     role,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"members": members,
		"total":   len(members),
	})
}

// Helper function to get display name from Clerk user data
func getDisplayName(userData *clerk.OrganizationMembershipPublicUserData) string {
	if userData.FirstName != nil && userData.LastName != nil {
		return *userData.FirstName + " " + *userData.LastName
	}
	if userData.FirstName != nil {
		return *userData.FirstName
	}
	if userData.LastName != nil {
		return *userData.LastName
	}
	return userData.Identifier // Fallback to email/username
}
