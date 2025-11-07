package commands

import (
	"backend/internal/models"
	"backend/internal/services"
	"backend/internal/whatsapp"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/organization"
	"github.com/clerk/clerk-sdk-go/v2/organizationmembership"
	"github.com/rs/zerolog/log"
)

// LinkOrganizationCommand handles the /link command for linking WhatsApp groups to organizations
type LinkOrganizationCommand struct {
	contextService *services.WhatsAppContextService
	authService    services.WhatsAppAuthService
}

// NewLinkOrganizationCommand creates a new link organization command
func NewLinkOrganizationCommand(contextService *services.WhatsAppContextService, authService services.WhatsAppAuthService) *LinkOrganizationCommand {
	return &LinkOrganizationCommand{
		contextService: contextService,
		authService:    authService,
	}
}

// Name returns the command name
func (c *LinkOrganizationCommand) Name() string {
	return "link"
}

// Description returns the command description
func (c *LinkOrganizationCommand) Description() string {
	return "Link this WhatsApp group to an organization (admin only)"
}

// Usage returns usage instructions
func (c *LinkOrganizationCommand) Usage() string {
	return "/link [organization_id] - Link this WhatsApp group to your organization (requires admin permissions)"
}

// RequiresAuth returns whether authentication is required
func (c *LinkOrganizationCommand) RequiresAuth() bool {
	return true
}

// Execute runs the link organization command
func (c *LinkOrganizationCommand) Execute(ctx *whatsapp.CommandContext) error {
	// Check if this is a group message
	if ctx.GroupID == nil {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ This command can only be used in WhatsApp groups.\n\nTo link a group to your organization, use this command from within the group.")
	}

	// Check if organization ID was provided
	if len(ctx.Args) == 0 {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Please provide an organization ID.\n\n*Usage:* /link [organization_id]\n\n_You can find your organization ID in the application settings._")
	}

	organizationID := ctx.Args[0]

	// Verify the organization exists and user is an admin
	isAdmin, err := c.verifyOrganizationAdmin(ctx.User.ClerkUserID, organizationID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to verify organization admin")
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Failed to verify organization permissions. Please try again.")
	}

	if !isAdmin {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ You must be an organization administrator to link a WhatsApp group.\n\nPlease contact your organization admin for assistance.")
	}

	// Check if group is already linked
	var existingLink models.WhatsAppGroupLink
	err = ctx.DB.Where("group_id = ? AND is_active = ?", *ctx.GroupID, true).First(&existingLink).Error
	if err == nil {
		// Group is already linked
		if existingLink.OrganizationID == organizationID {
			return ctx.Client.SendTextMessage(ctx.PhoneNumber,
				"ℹ️ This group is already linked to the specified organization.")
		}
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ This group is already linked to another organization.\n\nUse /unlink first to remove the existing link.")
	}

	// Create the link
	groupLink := models.WhatsAppGroupLink{
		GroupID:        *ctx.GroupID,
		OrganizationID: organizationID,
		LinkedBy:       ctx.User.ClerkUserID,
		IsActive:       true,
	}

	if err := ctx.DB.Create(&groupLink).Error; err != nil {
		log.Error().Err(err).Msg("Failed to create group link")
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Failed to link group to organization. Please try again.")
	}

	// Get organization name for confirmation message
	orgName, err := c.getOrganizationName(organizationID)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get organization name")
		orgName = organizationID
	}

	message := fmt.Sprintf("✅ *Group linked successfully!*\n\n"+
		"🏢 *Organization:* %s\n\n"+
		"This WhatsApp group is now linked to your organization. "+
		"All organization members can use commands in this group to manage shared notes, notebooks, and chapters.\n\n"+
		"_Use /unlink to remove this association._", orgName)

	return ctx.Client.SendTextMessage(ctx.PhoneNumber, message)
}

// verifyOrganizationAdmin checks if the user is an admin of the organization
func (c *LinkOrganizationCommand) verifyOrganizationAdmin(clerkUserID, organizationID string) (bool, error) {
	ctx := context.Background()

	// Check if user is a member
	isMember, err := c.authService.IsUserInOrganization(ctx, clerkUserID, organizationID)
	if err != nil {
		return false, err
	}

	if !isMember {
		return false, nil
	}

	// Get organization memberships to check role
	params := &organizationmembership.ListParams{}
	params.Limit = clerk.Int64(100)
	params.OrganizationID = organizationID
	params.UserIDs = []string{clerkUserID}

	memberships, err := organizationmembership.List(ctx, params)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get organization memberships")
		return false, fmt.Errorf("failed to get memberships: %w", err)
	}

	// Check if user has admin role
	for _, membership := range memberships.OrganizationMemberships {
		if membership.PublicUserData != nil && membership.PublicUserData.UserID == clerkUserID {
			// Check if role is admin
			if membership.Role == "org:admin" || membership.Role == "admin" {
				return true, nil
			}
		}
	}

	return false, nil
}

// getOrganizationName retrieves the organization name from Clerk
func (c *LinkOrganizationCommand) getOrganizationName(organizationID string) (string, error) {
	ctx := context.Background()
	org, err := organization.Get(ctx, organizationID)
	if err != nil {
		return "", err
	}
	return org.Name, nil
}

// UnlinkOrganizationCommand handles the /unlink command for unlinking WhatsApp groups from organizations
type UnlinkOrganizationCommand struct {
	contextService *services.WhatsAppContextService
	authService    services.WhatsAppAuthService
}

// NewUnlinkOrganizationCommand creates a new unlink organization command
func NewUnlinkOrganizationCommand(contextService *services.WhatsAppContextService, authService services.WhatsAppAuthService) *UnlinkOrganizationCommand {
	return &UnlinkOrganizationCommand{
		contextService: contextService,
		authService:    authService,
	}
}

// Name returns the command name
func (c *UnlinkOrganizationCommand) Name() string {
	return "unlink"
}

// Description returns the command description
func (c *UnlinkOrganizationCommand) Description() string {
	return "Unlink this WhatsApp group from its organization (admin only)"
}

// Usage returns usage instructions
func (c *UnlinkOrganizationCommand) Usage() string {
	return "/unlink - Remove the organization link from this WhatsApp group (requires admin permissions)"
}

// RequiresAuth returns whether authentication is required
func (c *UnlinkOrganizationCommand) RequiresAuth() bool {
	return true
}

// Execute runs the unlink organization command
func (c *UnlinkOrganizationCommand) Execute(ctx *whatsapp.CommandContext) error {
	// Check if this is a group message
	if ctx.GroupID == nil {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ This command can only be used in WhatsApp groups.")
	}

	// Check if we have an active context (awaiting confirmation)
	if ctx.ConversationCtx != nil && ctx.ConversationCtx.Command == "unlink" {
		return c.handleConfirmation(ctx)
	}

	// Check if group is linked
	var groupLink models.WhatsAppGroupLink
	err := ctx.DB.Where("group_id = ? AND is_active = ?", *ctx.GroupID, true).First(&groupLink).Error
	if err != nil {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ This group is not linked to any organization.")
	}

	// Verify user is an admin of the organization
	isAdmin, err := c.verifyOrganizationAdmin(ctx.User.ClerkUserID, groupLink.OrganizationID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to verify organization admin")
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Failed to verify organization permissions. Please try again.")
	}

	if !isAdmin {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ You must be an organization administrator to unlink a WhatsApp group.\n\nPlease contact your organization admin for assistance.")
	}

	// Get organization name
	orgName, err := c.getOrganizationName(groupLink.OrganizationID)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get organization name")
		orgName = groupLink.OrganizationID
	}

	// Request confirmation
	message := fmt.Sprintf("⚠️ *Confirm Unlink*\n\n"+
		"Are you sure you want to unlink this group from:\n\n"+
		"🏢 *%s*\n\n"+
		"After unlinking, organization members will no longer be able to use commands in this group.\n\n"+
		"Reply with 'yes' to confirm or 'no' to cancel.", orgName)

	// Store context for confirmation
	contextData := map[string]interface{}{
		"link_id":           groupLink.ID,
		"organization_id":   groupLink.OrganizationID,
		"organization_name": orgName,
	}

	if err := c.contextService.SetContext(ctx.PhoneNumber, "unlink", contextData); err != nil {
		log.Error().Err(err).Msg("Failed to set context for unlink command")
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ An error occurred. Please try again.")
	}

	return ctx.Client.SendTextMessage(ctx.PhoneNumber, message)
}

// handleConfirmation processes the unlink confirmation
func (c *UnlinkOrganizationCommand) handleConfirmation(ctx *whatsapp.CommandContext) error {
	response := strings.ToLower(strings.TrimSpace(ctx.Message))

	if response == "yes" || response == "y" {
		// Parse context data
		var contextData map[string]interface{}
		if err := json.Unmarshal([]byte(ctx.ConversationCtx.Data), &contextData); err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal context data")
			return ctx.Client.SendTextMessage(ctx.PhoneNumber,
				"❌ An error occurred. Please start over with /unlink")
		}

		linkID, _ := contextData["link_id"].(string)
		orgName, _ := contextData["organization_name"].(string)

		// Deactivate the link
		if err := ctx.DB.Model(&models.WhatsAppGroupLink{}).
			Where("id = ?", linkID).
			Update("is_active", false).Error; err != nil {
			log.Error().Err(err).Msg("Failed to unlink group")

			if err := c.contextService.ClearContext(ctx.PhoneNumber); err != nil {
				log.Error().Err(err).Msg("Failed to clear context")
			}

			return ctx.Client.SendTextMessage(ctx.PhoneNumber,
				"❌ Failed to unlink group. Please try again.")
		}

		// Clear context
		if err := c.contextService.ClearContext(ctx.PhoneNumber); err != nil {
			log.Error().Err(err).Msg("Failed to clear context")
		}

		message := fmt.Sprintf("✅ *Group unlinked successfully!*\n\n"+
			"This group has been unlinked from *%s*.\n\n"+
			"Organization commands are no longer available in this group.", orgName)
		return ctx.Client.SendTextMessage(ctx.PhoneNumber, message)
	}

	if response == "no" || response == "n" {
		if err := c.contextService.ClearContext(ctx.PhoneNumber); err != nil {
			log.Error().Err(err).Msg("Failed to clear context")
		}
		return ctx.Client.SendTextMessage(ctx.PhoneNumber, "❌ Unlink cancelled.")
	}

	return ctx.Client.SendTextMessage(ctx.PhoneNumber,
		"❌ Invalid response. Please reply with 'yes' to confirm or 'no' to cancel.")
}

// verifyOrganizationAdmin checks if the user is an admin of the organization
func (c *UnlinkOrganizationCommand) verifyOrganizationAdmin(clerkUserID, organizationID string) (bool, error) {
	ctx := context.Background()

	// Check if user is a member
	isMember, err := c.authService.IsUserInOrganization(ctx, clerkUserID, organizationID)
	if err != nil {
		return false, err
	}

	if !isMember {
		return false, nil
	}

	// Get organization memberships to check role
	params := &organizationmembership.ListParams{}
	params.Limit = clerk.Int64(100)
	params.OrganizationID = organizationID
	params.UserIDs = []string{clerkUserID}

	memberships, err := organizationmembership.List(ctx, params)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get organization memberships")
		return false, fmt.Errorf("failed to get memberships: %w", err)
	}

	// Check if user has admin role
	for _, membership := range memberships.OrganizationMemberships {
		if membership.PublicUserData != nil && membership.PublicUserData.UserID == clerkUserID {
			// Check if role is admin
			if membership.Role == "org:admin" || membership.Role == "admin" {
				return true, nil
			}
		}
	}

	return false, nil
}

// getOrganizationName retrieves the organization name from Clerk
func (c *UnlinkOrganizationCommand) getOrganizationName(organizationID string) (string, error) {
	ctx := context.Background()
	org, err := organization.Get(ctx, organizationID)
	if err != nil {
		return "", err
	}
	return org.Name, nil
}
