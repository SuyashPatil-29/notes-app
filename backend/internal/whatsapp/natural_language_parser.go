package whatsapp

import (
	"strings"
)

// NaturalLanguageCommand represents a parsed natural language command
type NaturalLanguageCommand struct {
	Command   string
	NoteTitle string
	UseAI     bool // Always true for natural language commands
}

// ParseNaturalLanguageAddCommand parses natural language "add" commands
// Takes the entire text after "add " as the note title
// AI will decide notebook, chapter, and generate content
func ParseNaturalLanguageAddCommand(message string) (*NaturalLanguageCommand, bool) {
	message = strings.TrimSpace(message)
	messageLower := strings.ToLower(message)

	// Check if it starts with "add "
	if !strings.HasPrefix(messageLower, "add ") {
		return nil, false
	}

	// Remove "add " prefix and use the rest as the note title
	noteTitle := strings.TrimSpace(message[4:])

	if noteTitle == "" {
		return nil, false
	}

	return &NaturalLanguageCommand{
		Command:   "add",
		NoteTitle: noteTitle,
		UseAI:     true,
	}, true
}

// ConvertToCommandFormat converts natural language command to standard command format
// Returns command name and args that can be used with the command system
func (nlc *NaturalLanguageCommand) ConvertToCommandFormat() (string, []string) {
	switch nlc.Command {
	case "add":
		// Return command in format: /add [note title]
		return "add", []string{nlc.NoteTitle}
	default:
		return "", nil
	}
}
