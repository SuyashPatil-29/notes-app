package whatsapp

import (
	"backend/internal/models"
	"backend/pkg/whatsapp"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// Command interface defines the contract for all WhatsApp command handlers
type Command interface {
	// Name returns the command name (e.g., "add", "help")
	Name() string

	// Description returns a brief description of what the command does
	Description() string

	// Usage returns usage instructions for the command
	Usage() string

	// RequiresAuth returns true if the command requires authentication
	RequiresAuth() bool

	// Execute runs the command with the given context
	Execute(ctx *CommandContext) error
}

// CommandContext contains all necessary information for command execution
type CommandContext struct {
	// PhoneNumber is the WhatsApp phone number of the user
	PhoneNumber string

	// Message is the full message text received from the user
	Message string

	// Args are the parsed command arguments (excluding the command name)
	Args []string

	// User is the authenticated WhatsApp user (nil if not authenticated)
	User *models.WhatsAppUser

	// ConversationCtx is the active conversation context (nil if none)
	ConversationCtx *models.WhatsAppConversationContext

	// Client is the WhatsApp API client for sending messages
	Client whatsapp.WhatsAppClient

	// DB is the database connection
	DB *gorm.DB

	// GroupID is the WhatsApp group ID (for organization mode)
	GroupID *string

	// OrganizationID is the organization ID (for organization mode)
	OrganizationID *string
}

// CommandRegistry manages registration and lookup of commands
type CommandRegistry struct {
	commands map[string]Command
}

// NewCommandRegistry creates a new command registry
func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{
		commands: make(map[string]Command),
	}
}

// Register adds a command to the registry
func (r *CommandRegistry) Register(cmd Command) {
	r.commands[cmd.Name()] = cmd
}

// Get retrieves a command by name
func (r *CommandRegistry) Get(name string) (Command, bool) {
	cmd, exists := r.commands[name]
	return cmd, exists
}

// GetAll returns all registered commands
func (r *CommandRegistry) GetAll() []Command {
	commands := make([]Command, 0, len(r.commands))
	for _, cmd := range r.commands {
		commands = append(commands, cmd)
	}
	return commands
}

// ParseCommand extracts the command name and arguments from a message
func ParseCommand(message string) (commandName string, args []string, isCommand bool) {
	message = strings.TrimSpace(message)

	// Check if message starts with /
	if !strings.HasPrefix(message, "/") {
		return "", nil, false
	}

	// Remove the leading /
	message = strings.TrimPrefix(message, "/")

	// Split by whitespace
	parts := strings.Fields(message)
	if len(parts) == 0 {
		return "", nil, false
	}

	// First part is the command name (lowercase)
	commandName = strings.ToLower(parts[0])

	// Rest are arguments
	if len(parts) > 1 {
		args = parts[1:]
	}

	return commandName, args, true
}

// GetCommandFromMessage extracts and retrieves a command from a message
func (r *CommandRegistry) GetCommandFromMessage(message string) (Command, []string, error) {
	commandName, args, isCommand := ParseCommand(message)

	if !isCommand {
		return nil, nil, fmt.Errorf("message is not a command")
	}

	cmd, exists := r.Get(commandName)
	if !exists {
		return nil, nil, fmt.Errorf("unknown command: %s", commandName)
	}

	return cmd, args, nil
}
