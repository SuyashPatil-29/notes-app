package commands

import (
	"backend/internal/whatsapp"
	"fmt"
	"sort"
	"strings"
)

// HelpCommand provides help information about available commands
type HelpCommand struct {
	registry *whatsapp.CommandRegistry
}

// NewHelpCommand creates a new help command
func NewHelpCommand(registry *whatsapp.CommandRegistry) *HelpCommand {
	return &HelpCommand{
		registry: registry,
	}
}

// Name returns the command name
func (c *HelpCommand) Name() string {
	return "help"
}

// Description returns the command description
func (c *HelpCommand) Description() string {
	return "Display available commands and usage information"
}

// Usage returns usage instructions
func (c *HelpCommand) Usage() string {
	return "/help [command] - Show help for all commands or a specific command"
}

// RequiresAuth returns whether authentication is required
func (c *HelpCommand) RequiresAuth() bool {
	return false
}

// Execute runs the help command
func (c *HelpCommand) Execute(ctx *whatsapp.CommandContext) error {
	// If a specific command is requested, show detailed help
	if len(ctx.Args) > 0 {
		return c.showCommandHelp(ctx, ctx.Args[0])
	}

	// Otherwise, show list of all commands
	return c.showAllCommands(ctx)
}

// showAllCommands displays a list of all available commands
func (c *HelpCommand) showAllCommands(ctx *whatsapp.CommandContext) error {
	commands := c.registry.GetAll()

	// Sort commands by name for consistent display
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name() < commands[j].Name()
	})

	var message strings.Builder
	message.WriteString("📚 *Available Commands*\n\n")

	// Determine if user is in organization mode
	isOrgMode := ctx.OrganizationID != nil

	if isOrgMode {
		message.WriteString("_Organization Group Mode_\n\n")
	} else {
		message.WriteString("_Personal Account Mode_\n\n")
	}

	// List all commands with descriptions
	for _, cmd := range commands {
		// Skip commands that require auth if user is not authenticated
		if cmd.RequiresAuth() && ctx.User == nil {
			continue
		}

		// Filter organization-specific commands
		if isOrgMode {
			// In org mode, show all commands
			message.WriteString(fmt.Sprintf("/%s - %s\n", cmd.Name(), cmd.Description()))
		} else {
			// In personal mode, skip organization-specific commands
			if cmd.Name() == "link" || cmd.Name() == "unlink" {
				continue
			}
			message.WriteString(fmt.Sprintf("/%s - %s\n", cmd.Name(), cmd.Description()))
		}
	}

	message.WriteString("\n💡 _Tip: Use /help [command] for detailed usage information_")

	return ctx.Client.SendTextMessage(ctx.PhoneNumber, message.String())
}

// showCommandHelp displays detailed help for a specific command
func (c *HelpCommand) showCommandHelp(ctx *whatsapp.CommandContext, commandName string) error {
	cmd, exists := c.registry.Get(strings.ToLower(commandName))
	if !exists {
		errorMsg := fmt.Sprintf("❌ Unknown command: *%s*\n\nUse /help to see all available commands.", commandName)
		return ctx.Client.SendTextMessage(ctx.PhoneNumber, errorMsg)
	}

	var message strings.Builder
	message.WriteString(fmt.Sprintf("📖 *Help: /%s*\n\n", cmd.Name()))
	message.WriteString(fmt.Sprintf("*Description:*\n%s\n\n", cmd.Description()))
	message.WriteString(fmt.Sprintf("*Usage:*\n%s\n", cmd.Usage()))

	if cmd.RequiresAuth() {
		message.WriteString("\n🔒 _This command requires authentication_")
	}

	return ctx.Client.SendTextMessage(ctx.PhoneNumber, message.String())
}
