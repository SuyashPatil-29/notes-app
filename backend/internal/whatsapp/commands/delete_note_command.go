package commands

import (
	"backend/internal/models"
	"backend/internal/services"
	"backend/internal/whatsapp"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

// DeleteNoteCommand handles the /delete command for deleting notes
type DeleteNoteCommand struct {
	contextService *services.WhatsAppContextService
}

// NewDeleteNoteCommand creates a new delete note command
func NewDeleteNoteCommand(contextService *services.WhatsAppContextService) *DeleteNoteCommand {
	return &DeleteNoteCommand{
		contextService: contextService,
	}
}

// Name returns the command name
func (c *DeleteNoteCommand) Name() string {
	return "delete"
}

// Description returns the command description
func (c *DeleteNoteCommand) Description() string {
	return "Delete a note by searching for its title"
}

// Usage returns usage instructions
func (c *DeleteNoteCommand) Usage() string {
	return "/delete [note title] - Search for a note and delete it after confirmation"
}

// RequiresAuth returns whether authentication is required
func (c *DeleteNoteCommand) RequiresAuth() bool {
	return true
}

// Execute runs the delete note command
func (c *DeleteNoteCommand) Execute(ctx *whatsapp.CommandContext) error {
	// Verify organization access if in organization mode
	if err := whatsapp.VerifyOrganizationAccess(ctx); err != nil {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			fmt.Sprintf("❌ %s", err.Error()))
	}

	// Check if we have an active context
	if ctx.ConversationCtx != nil && ctx.ConversationCtx.Command == "delete" {
		return c.continueFlow(ctx)
	}

	// Start new delete flow
	return c.startFlow(ctx)
}

// startFlow initiates the delete note flow
func (c *DeleteNoteCommand) startFlow(ctx *whatsapp.CommandContext) error {
	// Extract search term from arguments
	if len(ctx.Args) == 0 {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Please provide a note title to search for.\n\n*Usage:* /delete [note title]")
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
		log.Error().Err(err).Msg("Failed to search notes for deletion")
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ An error occurred while searching. Please try again.")
	}

	// Handle results
	if len(notes) == 0 {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			fmt.Sprintf("❌ No notes found matching: *%s*\n\nTry a different search term or use /list notes to see all your notes.", searchTerm))
	}

	if len(notes) == 1 {
		// Single match, ask for confirmation
		return c.requestConfirmation(ctx, &notes[0])
	}

	// Multiple matches, ask user to select
	return c.showSelectionList(ctx, notes, searchTerm)
}

// showSelectionList displays a numbered list of matching notes for selection
func (c *DeleteNoteCommand) showSelectionList(ctx *whatsapp.CommandContext, notes []models.Notes, searchTerm string) error {
	var message strings.Builder
	message.WriteString(fmt.Sprintf("🔍 Found %d notes matching: *%s*\n\n", len(notes), searchTerm))
	message.WriteString("Please select a note to delete by replying with its number:\n\n")

	// Store note IDs and names for selection
	noteIDs := make([]string, len(notes))
	noteNames := make([]string, len(notes))

	for i, note := range notes {
		noteIDs[i] = note.ID
		noteNames[i] = note.Name

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
		"note_names":  noteNames,
		"step":        "awaiting_selection",
	}

	if err := c.contextService.SetContext(ctx.PhoneNumber, "delete", contextData); err != nil {
		log.Error().Err(err).Msg("Failed to set context for delete command")
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ An error occurred. Please try again.")
	}

	return ctx.Client.SendTextMessage(ctx.PhoneNumber, message.String())
}

// continueFlow continues the multi-step delete flow
func (c *DeleteNoteCommand) continueFlow(ctx *whatsapp.CommandContext) error {
	// Parse context data
	var contextData map[string]interface{}
	if err := json.Unmarshal([]byte(ctx.ConversationCtx.Data), &contextData); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal context data")
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ An error occurred. Please start over with /delete")
	}

	step, _ := contextData["step"].(string)

	switch step {
	case "awaiting_selection":
		return c.handleSelection(ctx, contextData)
	case "awaiting_confirmation":
		return c.handleConfirmation(ctx, contextData)
	default:
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Invalid state. Please start over with /delete")
	}
}

// handleSelection processes the user's selection from multiple matches
func (c *DeleteNoteCommand) handleSelection(ctx *whatsapp.CommandContext, contextData map[string]interface{}) error {
	response := strings.ToLower(strings.TrimSpace(ctx.Message))

	// Check for cancellation
	if response == "cancel" {
		if err := c.contextService.ClearContext(ctx.PhoneNumber); err != nil {
			log.Error().Err(err).Msg("Failed to clear context")
		}
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Note deletion cancelled.")
	}

	// Parse selection number
	selection, err := strconv.Atoi(response)
	if err != nil {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Invalid selection. Please reply with a number or 'cancel'.")
	}

	// Get note IDs from context
	noteIDsInterface, ok := contextData["note_ids"].([]interface{})
	if !ok {
		log.Error().Msg("Failed to get note IDs from context")
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ An error occurred. Please start over with /delete")
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

	// Request confirmation
	return c.requestConfirmation(ctx, &note)
}

// requestConfirmation asks the user to confirm deletion
func (c *DeleteNoteCommand) requestConfirmation(ctx *whatsapp.CommandContext, note *models.Notes) error {
	notebookName := "Unknown"
	chapterName := "Unknown"

	if note.Chapter.Notebook.Name != "" {
		notebookName = note.Chapter.Notebook.Name
	}
	if note.Chapter.Name != "" {
		chapterName = note.Chapter.Name
	}

	message := fmt.Sprintf("⚠️ *Confirm Deletion*\n\n"+
		"Are you sure you want to delete this note?\n\n"+
		"📝 *Note:* %s\n"+
		"📓 *Notebook:* %s\n"+
		"📑 *Chapter:* %s\n\n"+
		"⚠️ _This action cannot be undone!_\n\n"+
		"Reply with 'yes' to confirm or 'no' to cancel.",
		note.Name, notebookName, chapterName)

	// Store context for confirmation
	contextData := map[string]interface{}{
		"note_id":   note.ID,
		"note_name": note.Name,
		"step":      "awaiting_confirmation",
	}

	if err := c.contextService.SetContext(ctx.PhoneNumber, "delete", contextData); err != nil {
		log.Error().Err(err).Msg("Failed to set context for delete confirmation")
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ An error occurred. Please try again.")
	}

	return ctx.Client.SendTextMessage(ctx.PhoneNumber, message)
}

// handleConfirmation processes the deletion confirmation
func (c *DeleteNoteCommand) handleConfirmation(ctx *whatsapp.CommandContext, contextData map[string]interface{}) error {
	response := strings.ToLower(strings.TrimSpace(ctx.Message))

	if response == "yes" || response == "y" {
		// Get note ID from context
		noteID, _ := contextData["note_id"].(string)
		noteName, _ := contextData["note_name"].(string)

		// Delete the note
		if err := ctx.DB.Delete(&models.Notes{}, "id = ?", noteID).Error; err != nil {
			log.Error().Err(err).Str("note_id", noteID).Msg("Failed to delete note")

			// Clear context
			if err := c.contextService.ClearContext(ctx.PhoneNumber); err != nil {
				log.Error().Err(err).Msg("Failed to clear context")
			}

			return ctx.Client.SendTextMessage(ctx.PhoneNumber,
				"❌ Failed to delete note. Please try again.")
		}

		// Clear context
		if err := c.contextService.ClearContext(ctx.PhoneNumber); err != nil {
			log.Error().Err(err).Msg("Failed to clear context")
		}

		// Send success message
		message := fmt.Sprintf("✅ *Note deleted successfully!*\n\n📝 *%s* has been permanently deleted.", noteName)
		return ctx.Client.SendTextMessage(ctx.PhoneNumber, message)
	}

	// User declined or invalid response
	if response == "no" || response == "n" {
		// Clear context
		if err := c.contextService.ClearContext(ctx.PhoneNumber); err != nil {
			log.Error().Err(err).Msg("Failed to clear context")
		}

		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Note deletion cancelled.")
	}

	// Invalid response
	return ctx.Client.SendTextMessage(ctx.PhoneNumber,
		"❌ Invalid response. Please reply with 'yes' to confirm deletion or 'no' to cancel.")
}
