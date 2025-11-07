package commands

import (
	"backend/internal/services"
	"backend/internal/whatsapp"

	"github.com/rs/zerolog/log"
)

// CancelCommand handles the /cancel command for cancelling ongoing operations
type CancelCommand struct {
	contextService *services.WhatsAppContextService
}

// NewCancelCommand creates a new cancel command
func NewCancelCommand(contextService *services.WhatsAppContextService) *CancelCommand {
	return &CancelCommand{
		contextService: contextService,
	}
}

// Name returns the command name
func (c *CancelCommand) Name() string {
	return "cancel"
}

// Description returns the command description
func (c *CancelCommand) Description() string {
	return "Cancel the current operation"
}

// Usage returns usage instructions
func (c *CancelCommand) Usage() string {
	return "/cancel - Cancel any ongoing command or operation"
}

// RequiresAuth returns whether authentication is required
func (c *CancelCommand) RequiresAuth() bool {
	return true
}

// Execute runs the cancel command
func (c *CancelCommand) Execute(ctx *whatsapp.CommandContext) error {
	// Check if there's an active context
	if ctx.ConversationCtx == nil {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ No active operation to cancel.\n\nYou can start a new command anytime!")
	}

	// Get the command being cancelled
	commandName := ctx.ConversationCtx.Command

	// Clear the context
	if err := c.contextService.ClearContext(ctx.PhoneNumber); err != nil {
		log.Error().Err(err).Msg("Failed to clear context")
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ An error occurred while cancelling. Please try again.")
	}

	return ctx.Client.SendTextMessage(ctx.PhoneNumber,
		"✅ Operation cancelled successfully.\n\n_Previous command:_ `"+commandName+"`\n\nYou can start a new command anytime!")
}
