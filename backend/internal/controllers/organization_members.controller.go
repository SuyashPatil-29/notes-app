package controllers

import (
	"backend/internal/middleware"
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/organization"
	"github.com/clerk/clerk-sdk-go/v2/organizationinvitation"
	"github.com/clerk/clerk-sdk-go/v2/organizationmembership"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// InviteMember sends an invitation to join the organization (admin only)
func InviteMember(c *gin.Context) {
	orgID := c.Param("orgId")

	// Get inviter user ID
	inviterUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Parse request body
	var req struct {
		EmailAddress string `json:"emailAddress" binding:"required,email"`
		Role         string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email address and role are required"})
		return
	}

	// Validate role
	if req.Role != "org:admin" && req.Role != "org:member" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role must be 'org:admin' or 'org:member'"})
		return
	}

	// Create invitation via Clerk API
	params := &organizationinvitation.CreateParams{
		OrganizationID: orgID,
		EmailAddress:   clerk.String(req.EmailAddress),
		InviterUserID:  clerk.String(inviterUserID),
		Role:           clerk.String(req.Role),
	}

	invitation, err := organizationinvitation.Create(c.Request.Context(), params)
	if err != nil {
		// Enhanced error logging with API error details
		if apiErr, ok := err.(*clerk.APIErrorResponse); ok {
			log.Error().
				Err(err).
				Str("org_id", orgID).
				Str("email", req.EmailAddress).
				Str("trace_id", apiErr.TraceID).
				Str("api_error", apiErr.Error()).
				Msg("Failed to create invitation - API error")
		} else {
			log.Error().Err(err).Str("org_id", orgID).Str("email", req.EmailAddress).Msg("Failed to create invitation")
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send invitation"})
		return
	}

	log.Info().Str("org_id", orgID).Str("email", req.EmailAddress).Str("invitation_id", invitation.ID).Msg("Invitation sent")
	c.JSON(http.StatusCreated, gin.H{
		"id":             invitation.ID,
		"emailAddress":   invitation.EmailAddress,
		"organizationId": invitation.OrganizationID,
		"role":           invitation.Role,
		"status":         invitation.Status,
		"createdAt":      invitation.CreatedAt,
	})
}

// ListMembers gets all members of an organization
func ListMembers(c *gin.Context) {
	orgID := c.Param("orgId")

	// Get current user ID to mark their own membership
	currentUserID, _ := middleware.GetClerkUserID(c)

	// Fetch organization memberships for this specific organization
	params := &organizationmembership.ListParams{}
	params.Limit = clerk.Int64(100)
	params.OrganizationID = orgID

	memberships, err := organizationmembership.List(c.Request.Context(), params)
	if err != nil {
		log.Error().Err(err).Str("org_id", orgID).Msg("Failed to fetch organization members")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch members"})
		return
	}

	// Build response with member details
	members := make([]gin.H, 0)
	for _, membership := range memberships.OrganizationMemberships {
		memberData := gin.H{
			"id":             membership.ID,
			"userId":         membership.PublicUserData.UserID,
			"organizationId": membership.Organization.ID,
			"role":           membership.Role,
			"createdAt":      membership.CreatedAt,
			"publicUserData": gin.H{
				"identifier": membership.PublicUserData.Identifier,
				"firstName":  membership.PublicUserData.FirstName,
				"lastName":   membership.PublicUserData.LastName,
				"imageUrl":   membership.PublicUserData.ImageURL,
			},
		}

		// Mark if this is the current user
		if membership.PublicUserData.UserID == currentUserID {
			memberData["isCurrentUser"] = true
		}

		members = append(members, memberData)
	}

	c.JSON(http.StatusOK, gin.H{
		"members": members,
		"total":   len(members),
	})
}

// UpdateMemberRole changes a member's role in the organization (admin only)
func UpdateMemberRole(c *gin.Context) {
	orgID := c.Param("orgId")
	userID := c.Param("userId")

	// Parse request body
	var req struct {
		Role string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role is required"})
		return
	}

	// Validate role
	if req.Role != "org:admin" && req.Role != "org:member" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role must be 'org:admin' or 'org:member'"})
		return
	}

	// First, find the membership ID for this user in this org
	listParams := &organizationmembership.ListParams{}
	listParams.Limit = clerk.Int64(100)
	listParams.OrganizationID = orgID
	listParams.UserIDs = []string{userID}

	memberships, err := organizationmembership.List(c.Request.Context(), listParams)
	if err != nil {
		log.Error().Err(err).Str("org_id", orgID).Str("user_id", userID).Msg("Failed to list memberships")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find member"})
		return
	}

	// Find the membership for this user in this org
	var membershipID string
	for _, m := range memberships.OrganizationMemberships {
		if m.Organization.ID == orgID && m.PublicUserData.UserID == userID {
			membershipID = m.ID
			break
		}
	}

	if membershipID == "" {
		log.Error().Str("org_id", orgID).Str("user_id", userID).Msg("Membership not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "Member not found"})
		return
	}

	// Update the membership role
	updateParams := &organizationmembership.UpdateParams{
		OrganizationID: orgID,
		UserID:         userID,
		Role:           clerk.String(req.Role),
	}

	membership, err := organizationmembership.Update(c.Request.Context(), updateParams)
	if err != nil {
		log.Error().Err(err).Str("org_id", orgID).Str("user_id", userID).Str("membership_id", membershipID).Msg("Failed to update member role")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role"})
		return
	}

	log.Info().Str("org_id", orgID).Str("user_id", userID).Str("new_role", req.Role).Msg("Member role updated")
	c.JSON(http.StatusOK, gin.H{
		"id":             membership.ID,
		"userId":         membership.PublicUserData.UserID,
		"organizationId": membership.Organization.ID,
		"role":           membership.Role,
	})
}

// RemoveMember removes a member from the organization (admin only)
func RemoveMember(c *gin.Context) {
	orgID := c.Param("orgId")
	userID := c.Param("userId")

	// First, find the membership ID
	listParams := &organizationmembership.ListParams{}
	listParams.Limit = clerk.Int64(100)
	listParams.OrganizationID = orgID
	listParams.UserIDs = []string{userID}

	memberships, err := organizationmembership.List(c.Request.Context(), listParams)
	if err != nil {
		log.Error().Err(err).Str("org_id", orgID).Str("user_id", userID).Msg("Failed to list memberships")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find member"})
		return
	}

	// Find the membership for this user in this org
	var membershipID string
	for _, m := range memberships.OrganizationMemberships {
		if m.Organization.ID == orgID && m.PublicUserData.UserID == userID {
			membershipID = m.ID
			break
		}
	}

	if membershipID == "" {
		log.Error().Str("org_id", orgID).Str("user_id", userID).Msg("Membership not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "Member not found"})
		return
	}

	// Delete the membership
	deleteParams := &organizationmembership.DeleteParams{
		OrganizationID: orgID,
		UserID:         userID,
	}
	_, err = organizationmembership.Delete(c.Request.Context(), deleteParams)
	if err != nil {
		log.Error().Err(err).Str("org_id", orgID).Str("user_id", userID).Str("membership_id", membershipID).Msg("Failed to remove member")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove member"})
		return
	}

	log.Info().Str("org_id", orgID).Str("user_id", userID).Msg("Member removed from organization")
	c.JSON(http.StatusOK, gin.H{"message": "Member removed successfully"})
}

// ListInvitations gets all pending invitations for an organization
func ListInvitations(c *gin.Context) {
	orgID := c.Param("orgId")

	// Fetch organization invitations for this specific organization
	params := &organizationinvitation.ListParams{}
	params.Limit = clerk.Int64(100)
	params.OrganizationID = orgID

	invitations, err := organizationinvitation.List(c.Request.Context(), params)
	if err != nil {
		log.Error().Err(err).Str("org_id", orgID).Msg("Failed to fetch invitations")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch invitations"})
		return
	}

	// Build response - filter for pending status
	invitationsList := make([]gin.H, 0)
	for _, invitation := range invitations.OrganizationInvitations {
		if invitation.Status == "pending" {
			invitationsList = append(invitationsList, gin.H{
				"id":             invitation.ID,
				"emailAddress":   invitation.EmailAddress,
				"organizationId": invitation.OrganizationID,
				"role":           invitation.Role,
				"status":         invitation.Status,
				"createdAt":      invitation.CreatedAt,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"invitations": invitationsList,
		"total":       len(invitationsList),
	})
}

// RevokeInvitation cancels a pending invitation (admin only)
func RevokeInvitation(c *gin.Context) {
	invitationID := c.Param("invitationId")

	// Revoke the invitation
	// Note: We need to get the organization ID from the invitation first, or from the route
	orgID := c.Param("orgId")
	revokeParams := &organizationinvitation.RevokeParams{
		OrganizationID: orgID,
		ID:             invitationID,
	}
	_, err := organizationinvitation.Revoke(c.Request.Context(), revokeParams)
	if err != nil {
		log.Error().Err(err).Str("invitation_id", invitationID).Msg("Failed to revoke invitation")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke invitation"})
		return
	}

	log.Info().Str("invitation_id", invitationID).Msg("Invitation revoked")
	c.JSON(http.StatusOK, gin.H{"message": "Invitation revoked successfully"})
}

// ListUserInvitations gets all pending invitations for the current user
func ListUserInvitations(c *gin.Context) {
	// Get authenticated user
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Fetch user details to get email
	usr, err := user.Get(c.Request.Context(), clerkUserID)
	if err != nil {
		log.Error().Err(err).Str("user_id", clerkUserID).Msg("Failed to fetch user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user details"})
		return
	}

	// Get primary email
	var primaryEmail string
	for _, email := range usr.EmailAddresses {
		if usr.PrimaryEmailAddressID != nil && email.ID == *usr.PrimaryEmailAddressID {
			primaryEmail = email.EmailAddress
			break
		}
	}

	if primaryEmail == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No email address found"})
		return
	}

	// Fetch pending invitations - we'll filter by email after
	params := &organizationinvitation.ListParams{}
	params.Limit = clerk.Int64(100)

	// Note: We need to list all invitations and filter by email since Clerk API doesn't support email filtering
	// In production, you might want to cache this or use Clerk's user invitation endpoints
	invitations, err := organizationinvitation.List(c.Request.Context(), params)
	if err != nil {
		log.Error().Err(err).Str("user_id", clerkUserID).Msg("Failed to fetch user invitations")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch invitations"})
		return
	}

	// Filter invitations for this user's email and pending status
	userInvitations := make([]gin.H, 0)
	for _, invitation := range invitations.OrganizationInvitations {
		if invitation.EmailAddress == primaryEmail && invitation.Status == "pending" {
			// Fetch organization details
			org, err := organization.Get(c.Request.Context(), invitation.OrganizationID)
			if err != nil {
				log.Warn().Err(err).Str("org_id", invitation.OrganizationID).Msg("Failed to fetch org for invitation")
				continue
			}

			userInvitations = append(userInvitations, gin.H{
				"id":             invitation.ID,
				"emailAddress":   invitation.EmailAddress,
				"organizationId": invitation.OrganizationID,
				"role":           invitation.Role,
				"status":         invitation.Status,
				"createdAt":      invitation.CreatedAt,
				"organization": gin.H{
					"id":       org.ID,
					"name":     org.Name,
					"slug":     org.Slug,
					"imageUrl": org.ImageURL,
				},
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"invitations": userInvitations,
		"total":       len(userInvitations),
	})
}

// AcceptInvitation accepts an invitation to join an organization
// Note: In Clerk v2, accepting invitations is typically done client-side via useOrganizationList().acceptInvitation()
// This endpoint will get the invitation details and the frontend should handle the actual acceptance
func AcceptInvitation(c *gin.Context) {
	invitationID := c.Param("invitationId")

	// First, we need to list invitations to find which org this invitation belongs to
	// This is a limitation of the SDK - Get requires both org ID and invitation ID
	// For now, return an error directing users to the client-side method
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":        "This endpoint is not available server-side",
		"message":      "Please use client-side Clerk SDK's useOrganizationList().acceptInvitation() method",
		"invitationId": invitationID,
	})
	return

	// TODO: If needed, implement by listing all invitations and finding the matching one
	/*
		getParams := &organizationinvitation.GetParams{
			OrganizationID: orgID, // We don't have this without listing first
			ID:             invitationID,
		}
		invitation, err := organizationinvitation.Get(c.Request.Context(), getParams)
		if err != nil {
			log.Error().Err(err).Str("invitation_id", invitationID).Msg("Failed to get invitation")
			c.JSON(http.StatusNotFound, gin.H{"error": "Invitation not found"})
			return
		}

	*/
}

// DeclineInvitation declines an invitation to join an organization
func DeclineInvitation(c *gin.Context) {
	invitationID := c.Param("invitationId")

	// Similar to AcceptInvitation, this requires organization ID which we don't have from the route
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":        "This endpoint is not available server-side",
		"message":      "Please use client-side Clerk SDK or delete the invitation from the organization's invitations list",
		"invitationId": invitationID,
	})
	return

	// TODO: If needed, implement by listing all invitations and finding the matching one
	/*
		revokeParams := &organizationinvitation.RevokeParams{
			OrganizationID: orgID, // We don't have this without listing first
			ID:             invitationID,
		}
		_, err := organizationinvitation.Revoke(c.Request.Context(), revokeParams)
	*/
}
