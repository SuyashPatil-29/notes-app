package whatsapp

import (
	"context"
	"fmt"

	"github.com/clerk/clerk-sdk-go/v2/organizationmembership"
	"github.com/rs/zerolog/log"
)

// CheckOrganizationMembership verifies if a user is a member of an organization
func CheckOrganizationMembership(ctx context.Context, clerkUserID, organizationID string) (bool, error) {
	// List organization memberships for the user
	params := &organizationmembership.ListParams{}
	params.OrganizationID = organizationID
	params.UserIDs = []string{clerkUserID}

	memberships, err := organizationmembership.List(ctx, params)
	if err != nil {
		log.Error().
			Err(err).
			Str("user_id", clerkUserID).
			Str("org_id", organizationID).
			Msg("Failed to check organization membership")
		return false, fmt.Errorf("failed to check organization membership: %w", err)
	}

	// Check if user has any membership in the organization
	return len(memberships.OrganizationMemberships) > 0, nil
}

// CheckOrganizationPermission verifies if a user has permission to perform an action in an organization
func CheckOrganizationPermission(ctx context.Context, clerkUserID, organizationID string, requireAdmin bool) (bool, string, error) {
	// List organization memberships for the user
	params := &organizationmembership.ListParams{}
	params.OrganizationID = organizationID
	params.UserIDs = []string{clerkUserID}

	memberships, err := organizationmembership.List(ctx, params)
	if err != nil {
		log.Error().
			Err(err).
			Str("user_id", clerkUserID).
			Str("org_id", organizationID).
			Msg("Failed to check organization membership")
		return false, "", fmt.Errorf("failed to check organization membership: %w", err)
	}

	// Check if user has any membership in the organization
	if len(memberships.OrganizationMemberships) == 0 {
		return false, "", nil
	}

	membership := memberships.OrganizationMemberships[0]
	role := membership.Role

	// If admin is required, check the role
	if requireAdmin && role != "org:admin" {
		return false, role, nil
	}

	return true, role, nil
}

// VerifyOrganizationAccess checks if a user can access organization resources
func VerifyOrganizationAccess(cmdCtx *CommandContext) error {
	if cmdCtx.OrganizationID == nil {
		// No organization context, personal mode
		return nil
	}

	// Check if user is a member of the organization
	isMember, err := CheckOrganizationMembership(
		context.Background(),
		cmdCtx.User.ClerkUserID,
		*cmdCtx.OrganizationID,
	)

	if err != nil {
		log.Error().
			Err(err).
			Str("user_id", cmdCtx.User.ClerkUserID).
			Str("org_id", *cmdCtx.OrganizationID).
			Msg("Failed to verify organization access")
		return fmt.Errorf("failed to verify organization access")
	}

	if !isMember {
		return fmt.Errorf("you are not a member of this organization")
	}

	return nil
}

// VerifyOrganizationAdminAccess checks if a user has admin access to organization resources
func VerifyOrganizationAdminAccess(cmdCtx *CommandContext) error {
	if cmdCtx.OrganizationID == nil {
		// No organization context, personal mode
		return nil
	}

	// Check if user is an admin of the organization
	hasPermission, role, err := CheckOrganizationPermission(
		context.Background(),
		cmdCtx.User.ClerkUserID,
		*cmdCtx.OrganizationID,
		true,
	)

	if err != nil {
		log.Error().
			Err(err).
			Str("user_id", cmdCtx.User.ClerkUserID).
			Str("org_id", *cmdCtx.OrganizationID).
			Msg("Failed to verify organization admin access")
		return fmt.Errorf("failed to verify organization access")
	}

	if !hasPermission {
		if role == "" {
			return fmt.Errorf("you are not a member of this organization")
		}
		return fmt.Errorf("you do not have admin permissions in this organization")
	}

	return nil
}
