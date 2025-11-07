package commands

import (
	"backend/internal/models"
	"backend/internal/services"
	"backend/internal/utils"
	"backend/internal/whatsapp"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

// RetrieveNoteCommand handles the /retrieve command for searching and retrieving notes
type RetrieveNoteCommand struct {
	contextService *services.WhatsAppContextService
}

// NewRetrieveNoteCommand creates a new retrieve note command
func NewRetrieveNoteCommand(contextService *services.WhatsAppContextService) *RetrieveNoteCommand {
	return &RetrieveNoteCommand{
		contextService: contextService,
	}
}

// Name returns the command name
func (c *RetrieveNoteCommand) Name() string {
	return "retrieve"
}

// Description returns the command description
func (c *RetrieveNoteCommand) Description() string {
	return "Search and retrieve notes by title"
}

// Usage returns usage instructions
func (c *RetrieveNoteCommand) Usage() string {
	return "/retrieve [search term] - Search for notes by title and retrieve their content"
}

// RequiresAuth returns whether authentication is required
func (c *RetrieveNoteCommand) RequiresAuth() bool {
	return true
}

// Execute runs the retrieve note command
func (c *RetrieveNoteCommand) Execute(ctx *whatsapp.CommandContext) error {
	// Check if we have an active context (user is selecting from multiple matches)
	if ctx.ConversationCtx != nil && ctx.ConversationCtx.Command == "retrieve" {
		return c.handleSelection(ctx)
	}

	// Start new search
	return c.searchNotes(ctx)
}

// searchNotes searches for notes matching the search term
func (c *RetrieveNoteCommand) searchNotes(ctx *whatsapp.CommandContext) error {
	// Extract search term from arguments
	if len(ctx.Args) == 0 {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Please provide a search term.\n\n*Usage:* /retrieve [search term]")
	}

	searchTerm := strings.Join(ctx.Args, " ")

	// Search for notes
	var notes []models.Notes
	query := ctx.DB.Preload("Chapter").Preload("Chapter.Notebook").
		Joins("JOIN chapters ON chapters.id = notes.chapter_id").
		Joins("JOIN notebooks ON notebooks.id = chapters.notebook_id").
		Where("notebooks.clerk_user_id = ? AND LOWER(notes.name) LIKE ?",
			ctx.User.ClerkUserID, "%"+strings.ToLower(searchTerm)+"%")

	if ctx.OrganizationID != nil {
		query = query.Where("notes.organization_id = ?", *ctx.OrganizationID)
	} else {
		query = query.Where("notes.organization_id IS NULL")
	}

	if err := query.Limit(10).Find(&notes).Error; err != nil {
		log.Error().Err(err).Msg("Failed to search notes")
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ An error occurred while searching. Please try again.")
	}

	// Handle results
	if len(notes) == 0 {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			fmt.Sprintf("❌ No notes found matching: *%s*\n\nTry a different search term or use /list notes to see all your notes.", searchTerm))
	}

	if len(notes) == 1 {
		// Single match, send the note content directly
		return c.sendNoteContent(ctx, &notes[0])
	}

	// Multiple matches, ask user to select
	return c.showSelectionList(ctx, notes, searchTerm)
}

// showSelectionList displays a numbered list of matching notes
func (c *RetrieveNoteCommand) showSelectionList(ctx *whatsapp.CommandContext, notes []models.Notes, searchTerm string) error {
	var message strings.Builder
	message.WriteString(fmt.Sprintf("🔍 Found %d notes matching: *%s*\n\n", len(notes), searchTerm))
	message.WriteString("Please select a note by replying with its number:\n\n")

	// Store note IDs for selection
	noteIDs := make([]string, len(notes))

	for i, note := range notes {
		noteIDs[i] = note.ID
		notebookName := "Unknown"
		chapterName := "Unknown"

		if note.Chapter.Notebook.Name != "" {
			notebookName = note.Chapter.Notebook.Name
		}
		if note.Chapter.Name != "" {
			chapterName = note.Chapter.Name
		}

		message.WriteString(fmt.Sprintf("%d. *%s*\n   📓 %s › 📑 %s\n\n", i+1, note.Name, notebookName, chapterName))
	}

	message.WriteString(fmt.Sprintf("_Reply with a number (1-%d) or 'cancel' to abort_", len(notes)))

	// Store context for selection
	contextData := map[string]interface{}{
		"search_term": searchTerm,
		"note_ids":    noteIDs,
		"step":        "awaiting_selection",
	}

	if err := c.contextService.SetContext(ctx.PhoneNumber, "retrieve", contextData); err != nil {
		log.Error().Err(err).Msg("Failed to set context for retrieve command")
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ An error occurred. Please try again.")
	}

	return ctx.Client.SendTextMessage(ctx.PhoneNumber, message.String())
}

// handleSelection processes the user's selection from multiple matches
func (c *RetrieveNoteCommand) handleSelection(ctx *whatsapp.CommandContext) error {
	response := strings.ToLower(strings.TrimSpace(ctx.Message))

	// Check for cancellation
	if response == "cancel" {
		if err := c.contextService.ClearContext(ctx.PhoneNumber); err != nil {
			log.Error().Err(err).Msg("Failed to clear context")
		}
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Note retrieval cancelled.")
	}

	// Parse selection number
	selection, err := strconv.Atoi(response)
	if err != nil {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Invalid selection. Please reply with a number or 'cancel'.")
	}

	// Parse context data
	var contextData map[string]interface{}
	if err := json.Unmarshal([]byte(ctx.ConversationCtx.Data), &contextData); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal context data")
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ An error occurred. Please start over with /retrieve")
	}

	// Get note IDs from context
	noteIDsInterface, ok := contextData["note_ids"].([]interface{})
	if !ok {
		log.Error().Msg("Failed to get note IDs from context")
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ An error occurred. Please start over with /retrieve")
	}

	noteIDs := make([]string, len(noteIDsInterface))
	for i, id := range noteIDsInterface {
		noteIDs[i] = id.(string)
	}

	// Validate selection
	if selection < 1 || selection > len(noteIDs) {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			fmt.Sprintf("❌ Invalid selection. Please choose a number between 1 and %d.", len(noteIDs)))
	}

	// Get the selected note
	selectedNoteID := noteIDs[selection-1]
	var note models.Notes
	if err := ctx.DB.Preload("Chapter").Preload("Chapter.Notebook").
		First(&note, "id = ?", selectedNoteID).Error; err != nil {
		log.Error().Err(err).Msg("Failed to get selected note")
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Failed to retrieve note. Please try again.")
	}

	// Clear context
	if err := c.contextService.ClearContext(ctx.PhoneNumber); err != nil {
		log.Error().Err(err).Msg("Failed to clear context")
	}

	// Send note content
	return c.sendNoteContent(ctx, &note)
}

// sendNoteContent sends the note content to the user
func (c *RetrieveNoteCommand) sendNoteContent(ctx *whatsapp.CommandContext, note *models.Notes) error {
	var message strings.Builder

	// Header
	message.WriteString(fmt.Sprintf("📝 *%s*\n\n", note.Name))

	// Metadata
	notebookName := "Unknown"
	chapterName := "Unknown"
	if note.Chapter.Notebook.Name != "" {
		notebookName = note.Chapter.Notebook.Name
	}
	if note.Chapter.Name != "" {
		chapterName = note.Chapter.Name
	}

	message.WriteString(fmt.Sprintf("📓 *Notebook:* %s\n", notebookName))
	message.WriteString(fmt.Sprintf("📑 *Chapter:* %s\n", chapterName))
	message.WriteString(fmt.Sprintf("🕒 *Updated:* %s\n\n", note.UpdatedAt.Format("Jan 2, 2006 3:04 PM")))

	// Content
	message.WriteString("─────────────────────\n\n")

	if note.Content == "" {
		message.WriteString("_This note is empty_")
	} else {
		// Convert TipTap JSON to markdown
		markdownContent, err := utils.TipTapToMarkdown(note.Content)
		if err != nil {
			log.Error().Err(err).Msg("Failed to convert TipTap to markdown")
			markdownContent = note.Content // Fallback to raw content
		}

		// Truncate content if too long (WhatsApp has message length limits)
		content := markdownContent
		maxLength := 3000
		if len(content) > maxLength {
			content = content[:maxLength] + "\n\n... _(content truncated)_"
		}
		message.WriteString(content)
	}

	return ctx.Client.SendTextMessage(ctx.PhoneNumber, message.String())
}
