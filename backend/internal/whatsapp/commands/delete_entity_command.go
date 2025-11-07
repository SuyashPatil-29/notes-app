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

// DeleteEntityCommand handles deletion of notebooks and chapters (not notes - that's DeleteNoteCommand)
type DeleteEntityCommand struct {
	contextService *services.WhatsAppContextService
}

// NewDeleteEntityCommand creates a new delete entity command
func NewDeleteEntityCommand(contextService *services.WhatsAppContextService) *DeleteEntityCommand {
	return &DeleteEntityCommand{
		contextService: contextService,
	}
}

// Name returns the command name
func (c *DeleteEntityCommand) Name() string {
	return "remove"
}

// Description returns the command description
func (c *DeleteEntityCommand) Description() string {
	return "Delete a notebook or chapter"
}

// Usage returns usage instructions
func (c *DeleteEntityCommand) Usage() string {
	return "/remove [notebook|chapter] [name] - Delete a notebook or chapter after confirmation"
}

// RequiresAuth returns whether authentication is required
func (c *DeleteEntityCommand) RequiresAuth() bool {
	return true
}

// Execute runs the delete entity command
func (c *DeleteEntityCommand) Execute(ctx *whatsapp.CommandContext) error {
	// Verify organization access if in organization mode
	if err := whatsapp.VerifyOrganizationAccess(ctx); err != nil {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			fmt.Sprintf("❌ %s", err.Error()))
	}

	// Check if we have an active context
	if ctx.ConversationCtx != nil && ctx.ConversationCtx.Command == "remove" {
		return c.continueFlow(ctx)
	}

	// Start new delete flow
	return c.startFlow(ctx)
}

// startFlow initiates the delete entity flow
func (c *DeleteEntityCommand) startFlow(ctx *whatsapp.CommandContext) error {
	// Check if user provided entity type
	if len(ctx.Args) == 0 {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Please specify what to delete.\n\n*Usage:* /remove [notebook|chapter] [name]")
	}

	entityType := strings.ToLower(ctx.Args[0])

	switch entityType {
	case "notebook":
		return c.deleteNotebook(ctx)
	case "chapter":
		return c.deleteChapter(ctx)
	default:
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Invalid entity type. Use 'notebook' or 'chapter'.\n\n*Usage:* /remove [notebook|chapter] [name]\n\n_Note: To delete a note, use /delete [note name]_")
	}
}

// deleteNotebook handles notebook deletion
func (c *DeleteEntityCommand) deleteNotebook(ctx *whatsapp.CommandContext) error {
	// Check if name was provided
	if len(ctx.Args) < 2 {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Please provide a notebook name.\n\n*Usage:* /remove notebook [name]")
	}

	notebookName := strings.Join(ctx.Args[1:], " ")

	// Search for notebooks
	var notebooks []models.Notebook
	query := ctx.DB.Preload("Chapters").
		Where("clerk_user_id = ? AND LOWER(name) LIKE ?",
			ctx.User.ClerkUserID, "%"+strings.ToLower(notebookName)+"%")

	if ctx.OrganizationID != nil {
		query = query.Where("organization_id = ?", *ctx.OrganizationID)
	} else {
		query = query.Where("organization_id IS NULL")
	}

	if err := query.Limit(10).Find(&notebooks).Error; err != nil {
		log.Error().Err(err).Msg("Failed to search notebooks for deletion")
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ An error occurred while searching. Please try again.")
	}

	// Handle results
	if len(notebooks) == 0 {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			fmt.Sprintf("❌ No notebooks found matching: *%s*\n\nUse /list notebooks to see all your notebooks.", notebookName))
	}

	if len(notebooks) == 1 {
		// Single match, ask for confirmation
		return c.requestNotebookConfirmation(ctx, &notebooks[0])
	}

	// Multiple matches, ask user to select
	return c.showNotebookSelectionList(ctx, notebooks, notebookName)
}

// deleteChapter handles chapter deletion
func (c *DeleteEntityCommand) deleteChapter(ctx *whatsapp.CommandContext) error {
	// Check if name was provided
	if len(ctx.Args) < 2 {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Please provide a chapter name.\n\n*Usage:* /remove chapter [name]")
	}

	chapterName := strings.Join(ctx.Args[1:], " ")

	// Search for chapters
	var chapters []models.Chapter
	query := ctx.DB.Preload("Notebook").Preload("Files").
		Joins("JOIN notebooks ON notebooks.id = chapters.notebook_id").
		Where("notebooks.clerk_user_id = ? AND LOWER(chapters.name) LIKE ?",
			ctx.User.ClerkUserID, "%"+strings.ToLower(chapterName)+"%")

	if ctx.OrganizationID != nil {
		query = query.Where("chapters.organization_id = ?", *ctx.OrganizationID)
	} else {
		query = query.Where("chapters.organization_id IS NULL")
	}

	if err := query.Limit(10).Find(&chapters).Error; err != nil {
		log.Error().Err(err).Msg("Failed to search chapters for deletion")
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ An error occurred while searching. Please try again.")
	}

	// Handle results
	if len(chapters) == 0 {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			fmt.Sprintf("❌ No chapters found matching: *%s*\n\nUse /list chapters to see all your chapters.", chapterName))
	}

	if len(chapters) == 1 {
		// Single match, ask for confirmation
		return c.requestChapterConfirmation(ctx, &chapters[0])
	}

	// Multiple matches, ask user to select
	return c.showChapterSelectionList(ctx, chapters, chapterName)
}

// showNotebookSelectionList displays a numbered list of matching notebooks
func (c *DeleteEntityCommand) showNotebookSelectionList(ctx *whatsapp.CommandContext, notebooks []models.Notebook, searchTerm string) error {
	var message strings.Builder
	message.WriteString(fmt.Sprintf("🔍 Found %d notebooks matching: *%s*\n\n", len(notebooks), searchTerm))
	message.WriteString("Please select a notebook to delete by replying with its number:\n\n")

	notebookIDs := make([]string, len(notebooks))
	notebookNames := make([]string, len(notebooks))

	for i, notebook := range notebooks {
		notebookIDs[i] = notebook.ID
		notebookNames[i] = notebook.Name
		chapterCount := len(notebook.Chapters)
		message.WriteString(fmt.Sprintf("%d. *%s* (%d chapter(s))\n", i+1, notebook.Name, chapterCount))
	}

	message.WriteString(fmt.Sprintf("\n_Reply with a number (1-%d) or 'cancel' to abort_", len(notebooks)))

	contextData := map[string]interface{}{
		"entity_type":  "notebook",
		"search_term":  searchTerm,
		"entity_ids":   notebookIDs,
		"entity_names": notebookNames,
		"step":         "awaiting_selection",
	}

	if err := c.contextService.SetContext(ctx.PhoneNumber, "remove", contextData); err != nil {
		log.Error().Err(err).Msg("Failed to set context for remove command")
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ An error occurred. Please try again.")
	}

	return ctx.Client.SendTextMessage(ctx.PhoneNumber, message.String())
}

// showChapterSelectionList displays a numbered list of matching chapters
func (c *DeleteEntityCommand) showChapterSelectionList(ctx *whatsapp.CommandContext, chapters []models.Chapter, searchTerm string) error {
	var message strings.Builder
	message.WriteString(fmt.Sprintf("🔍 Found %d chapters matching: *%s*\n\n", len(chapters), searchTerm))
	message.WriteString("Please select a chapter to delete by replying with its number:\n\n")

	chapterIDs := make([]string, len(chapters))
	chapterNames := make([]string, len(chapters))

	for i, chapter := range chapters {
		chapterIDs[i] = chapter.ID
		chapterNames[i] = chapter.Name
		notebookName := "Unknown"
		if chapter.Notebook.Name != "" {
			notebookName = chapter.Notebook.Name
		}
		noteCount := len(chapter.Files)
		message.WriteString(fmt.Sprintf("%d. *%s* (📓 %s, %d note(s))\n", i+1, chapter.Name, notebookName, noteCount))
	}

	message.WriteString(fmt.Sprintf("\n_Reply with a number (1-%d) or 'cancel' to abort_", len(chapters)))

	contextData := map[string]interface{}{
		"entity_type":  "chapter",
		"search_term":  searchTerm,
		"entity_ids":   chapterIDs,
		"entity_names": chapterNames,
		"step":         "awaiting_selection",
	}

	if err := c.contextService.SetContext(ctx.PhoneNumber, "remove", contextData); err != nil {
		log.Error().Err(err).Msg("Failed to set context for remove command")
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ An error occurred. Please try again.")
	}

	return ctx.Client.SendTextMessage(ctx.PhoneNumber, message.String())
}

// continueFlow continues the multi-step delete flow
func (c *DeleteEntityCommand) continueFlow(ctx *whatsapp.CommandContext) error {
	var contextData map[string]interface{}
	if err := json.Unmarshal([]byte(ctx.ConversationCtx.Data), &contextData); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal context data")
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ An error occurred. Please start over with /remove")
	}

	step, _ := contextData["step"].(string)

	switch step {
	case "awaiting_selection":
		return c.handleSelection(ctx, contextData)
	case "awaiting_confirmation":
		return c.handleConfirmation(ctx, contextData)
	default:
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Invalid state. Please start over with /remove")
	}
}

// handleSelection processes the user's selection
func (c *DeleteEntityCommand) handleSelection(ctx *whatsapp.CommandContext, contextData map[string]interface{}) error {
	response := strings.ToLower(strings.TrimSpace(ctx.Message))

	if response == "cancel" {
		if err := c.contextService.ClearContext(ctx.PhoneNumber); err != nil {
			log.Error().Err(err).Msg("Failed to clear context")
		}
		return ctx.Client.SendTextMessage(ctx.PhoneNumber, "❌ Deletion cancelled.")
	}

	selection, err := strconv.Atoi(response)
	if err != nil {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Invalid selection. Please reply with a number or 'cancel'.")
	}

	entityIDsInterface, ok := contextData["entity_ids"].([]interface{})
	if !ok {
		log.Error().Msg("Failed to get entity IDs from context")
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ An error occurred. Please start over with /remove")
	}

	entityIDs := make([]string, len(entityIDsInterface))
	for i, id := range entityIDsInterface {
		entityIDs[i] = id.(string)
	}

	if selection < 1 || selection > len(entityIDs) {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			fmt.Sprintf("❌ Invalid selection. Please choose a number between 1 and %d.", len(entityIDs)))
	}

	selectedEntityID := entityIDs[selection-1]
	entityType, _ := contextData["entity_type"].(string)

	if entityType == "notebook" {
		var notebook models.Notebook
		if err := ctx.DB.Preload("Chapters").First(&notebook, "id = ?", selectedEntityID).Error; err != nil {
			log.Error().Err(err).Msg("Failed to get selected notebook")
			return ctx.Client.SendTextMessage(ctx.PhoneNumber,
				"❌ Failed to retrieve notebook. Please try again.")
		}
		return c.requestNotebookConfirmation(ctx, &notebook)
	} else {
		var chapter models.Chapter
		if err := ctx.DB.Preload("Notebook").Preload("Files").First(&chapter, "id = ?", selectedEntityID).Error; err != nil {
			log.Error().Err(err).Msg("Failed to get selected chapter")
			return ctx.Client.SendTextMessage(ctx.PhoneNumber,
				"❌ Failed to retrieve chapter. Please try again.")
		}
		return c.requestChapterConfirmation(ctx, &chapter)
	}
}

// requestNotebookConfirmation asks for notebook deletion confirmation
func (c *DeleteEntityCommand) requestNotebookConfirmation(ctx *whatsapp.CommandContext, notebook *models.Notebook) error {
	chapterCount := len(notebook.Chapters)

	message := fmt.Sprintf("⚠️ *Confirm Notebook Deletion*\n\n"+
		"Are you sure you want to delete this notebook?\n\n"+
		"📓 *Notebook:* %s\n"+
		"📑 *Chapters:* %d\n\n"+
		"⚠️ _This will delete the notebook and ALL its chapters and notes!_\n"+
		"⚠️ _This action cannot be undone!_\n\n"+
		"Reply with 'yes' to confirm or 'no' to cancel.",
		notebook.Name, chapterCount)

	contextData := map[string]interface{}{
		"entity_type": "notebook",
		"entity_id":   notebook.ID,
		"entity_name": notebook.Name,
		"step":        "awaiting_confirmation",
	}

	if err := c.contextService.SetContext(ctx.PhoneNumber, "remove", contextData); err != nil {
		log.Error().Err(err).Msg("Failed to set context for confirmation")
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ An error occurred. Please try again.")
	}

	return ctx.Client.SendTextMessage(ctx.PhoneNumber, message)
}

// requestChapterConfirmation asks for chapter deletion confirmation
func (c *DeleteEntityCommand) requestChapterConfirmation(ctx *whatsapp.CommandContext, chapter *models.Chapter) error {
	notebookName := "Unknown"
	if chapter.Notebook.Name != "" {
		notebookName = chapter.Notebook.Name
	}
	noteCount := len(chapter.Files)

	message := fmt.Sprintf("⚠️ *Confirm Chapter Deletion*\n\n"+
		"Are you sure you want to delete this chapter?\n\n"+
		"📑 *Chapter:* %s\n"+
		"📓 *Notebook:* %s\n"+
		"📝 *Notes:* %d\n\n"+
		"⚠️ _This will delete the chapter and ALL its notes!_\n"+
		"⚠️ _This action cannot be undone!_\n\n"+
		"Reply with 'yes' to confirm or 'no' to cancel.",
		chapter.Name, notebookName, noteCount)

	contextData := map[string]interface{}{
		"entity_type": "chapter",
		"entity_id":   chapter.ID,
		"entity_name": chapter.Name,
		"step":        "awaiting_confirmation",
	}

	if err := c.contextService.SetContext(ctx.PhoneNumber, "remove", contextData); err != nil {
		log.Error().Err(err).Msg("Failed to set context for confirmation")
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ An error occurred. Please try again.")
	}

	return ctx.Client.SendTextMessage(ctx.PhoneNumber, message)
}

// handleConfirmation processes the deletion confirmation
func (c *DeleteEntityCommand) handleConfirmation(ctx *whatsapp.CommandContext, contextData map[string]interface{}) error {
	response := strings.ToLower(strings.TrimSpace(ctx.Message))

	if response == "yes" || response == "y" {
		entityType, _ := contextData["entity_type"].(string)
		entityID, _ := contextData["entity_id"].(string)
		entityName, _ := contextData["entity_name"].(string)

		var err error
		if entityType == "notebook" {
			err = ctx.DB.Delete(&models.Notebook{}, "id = ?", entityID).Error
		} else {
			err = ctx.DB.Delete(&models.Chapter{}, "id = ?", entityID).Error
		}

		if err != nil {
			log.Error().Err(err).Str("entity_type", entityType).Str("entity_id", entityID).Msg("Failed to delete entity")
			if err := c.contextService.ClearContext(ctx.PhoneNumber); err != nil {
				log.Error().Err(err).Msg("Failed to clear context")
			}
			return ctx.Client.SendTextMessage(ctx.PhoneNumber,
				fmt.Sprintf("❌ Failed to delete %s. Please try again.", entityType))
		}

		if err := c.contextService.ClearContext(ctx.PhoneNumber); err != nil {
			log.Error().Err(err).Msg("Failed to clear context")
		}

		emoji := "📓"
		if entityType == "chapter" {
			emoji = "📑"
		}

		message := fmt.Sprintf("✅ *%s deleted successfully!*\n\n%s *%s* has been permanently deleted.",
			strings.Title(entityType), emoji, entityName)
		return ctx.Client.SendTextMessage(ctx.PhoneNumber, message)
	}

	if response == "no" || response == "n" {
		if err := c.contextService.ClearContext(ctx.PhoneNumber); err != nil {
			log.Error().Err(err).Msg("Failed to clear context")
		}
		return ctx.Client.SendTextMessage(ctx.PhoneNumber, "❌ Deletion cancelled.")
	}

	return ctx.Client.SendTextMessage(ctx.PhoneNumber,
		"❌ Invalid response. Please reply with 'yes' to confirm deletion or 'no' to cancel.")
}
