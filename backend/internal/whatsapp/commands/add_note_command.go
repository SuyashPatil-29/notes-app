package commands

import (
	"backend/internal/models"
	"backend/internal/services"
	"backend/internal/utils"
	"backend/internal/whatsapp"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
)

// AddNoteCommand handles the /add command for creating notes
type AddNoteCommand struct {
	contextService *services.WhatsAppContextService
	aiService      *services.AIService
}

// NewAddNoteCommand creates a new add note command
func NewAddNoteCommand(contextService *services.WhatsAppContextService) *AddNoteCommand {
	return &AddNoteCommand{
		contextService: contextService,
		aiService:      services.NewAIService(),
	}
}

// Name returns the command name
func (c *AddNoteCommand) Name() string {
	return "add"
}

// Description returns the command description
func (c *AddNoteCommand) Description() string {
	return "Create a new note in a notebook and chapter"
}

// Usage returns usage instructions
func (c *AddNoteCommand) Usage() string {
	return "/add [note title] - Start creating a new note with the specified title"
}

// RequiresAuth returns whether authentication is required
func (c *AddNoteCommand) RequiresAuth() bool {
	return true
}

// Execute runs the add note command
func (c *AddNoteCommand) Execute(ctx *whatsapp.CommandContext) error {
	// Verify organization access if in organization mode
	if err := whatsapp.VerifyOrganizationAccess(ctx); err != nil {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			fmt.Sprintf("❌ %s", err.Error()))
	}

	// Check if we have an active context
	if ctx.ConversationCtx != nil && ctx.ConversationCtx.Command == "add" {
		// Check if this is a prefilled natural language context
		var contextData map[string]interface{}
		if err := json.Unmarshal([]byte(ctx.ConversationCtx.Data), &contextData); err == nil {
			if prefilled, ok := contextData["prefilled"].(bool); ok && prefilled {
				// This is a natural language command with prefilled data
				return c.handleNaturalLanguageFlow(ctx, contextData)
			}
		}
		return c.continueFlow(ctx)
	}

	// Start new flow
	return c.startFlow(ctx)
}

// startFlow initiates the add note flow
func (c *AddNoteCommand) startFlow(ctx *whatsapp.CommandContext) error {
	// Extract note title from arguments
	if len(ctx.Args) == 0 {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Please provide a note title.\n\n*Usage:* /add [note title]")
	}

	noteTitle := strings.Join(ctx.Args, " ")

	// Validate and sanitize note title
	if err := utils.ValidateCommandArgument("note_title", noteTitle); err != nil {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			fmt.Sprintf("❌ Invalid note title: %s", err.Error()))
	}
	noteTitle = utils.SanitizeCommandArgument(noteTitle)

	// Create context data
	contextData := map[string]interface{}{
		"title": noteTitle,
		"step":  "awaiting_notebook",
	}

	// Store context
	if err := c.contextService.SetContext(ctx.PhoneNumber, "add", contextData); err != nil {
		log.Error().Err(err).Msg("Failed to set context for add command")
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ An error occurred. Please try again.")
	}

	// Ask for notebook name
	message := fmt.Sprintf("📝 Creating note: *%s*\n\nWhich notebook would you like to add this note to?\n\n_Type the notebook name or 'new' to create a new notebook_", noteTitle)
	return ctx.Client.SendTextMessage(ctx.PhoneNumber, message)
}

// continueFlow continues the multi-step add note flow
func (c *AddNoteCommand) continueFlow(ctx *whatsapp.CommandContext) error {
	// Parse context data
	var contextData map[string]interface{}
	if err := json.Unmarshal([]byte(ctx.ConversationCtx.Data), &contextData); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal context data")
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ An error occurred. Please start over with /add")
	}

	step, _ := contextData["step"].(string)

	switch step {
	case "awaiting_notebook":
		return c.handleNotebookResponse(ctx, contextData)
	case "awaiting_new_notebook_name":
		return c.handleNewNotebookName(ctx, contextData)
	case "awaiting_notebook_confirmation":
		// Check if this is from natural language with existing notebook/chapter
		if createNotebook, ok := contextData["create_notebook"].(bool); !ok || !createNotebook {
			return c.handleNaturalLanguageNotebookConfirmation(ctx, contextData)
		}
		return c.handleNotebookConfirmation(ctx, contextData)
	case "awaiting_chapter":
		return c.handleChapterResponse(ctx, contextData)
	case "awaiting_new_chapter_name":
		return c.handleNewChapterName(ctx, contextData)
	case "awaiting_chapter_confirmation":
		// Check if this is from natural language with existing chapter
		if createChapter, ok := contextData["create_chapter"].(bool); !ok || !createChapter {
			return c.handleNaturalLanguageChapterConfirmation(ctx, contextData)
		}
		return c.handleChapterConfirmation(ctx, contextData)
	case "awaiting_ai_content":
		return c.handleAIContentResponse(ctx, contextData)
	default:
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Invalid state. Please start over with /add")
	}
}

// handleNotebookResponse processes the notebook name response
func (c *AddNoteCommand) handleNotebookResponse(ctx *whatsapp.CommandContext, contextData map[string]interface{}) error {
	notebookName := strings.TrimSpace(ctx.Message)

	if notebookName == "" {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Please provide a notebook name.")
	}

	// Check if user wants to create a new notebook
	if strings.ToLower(notebookName) == "new" {
		contextData["step"] = "awaiting_new_notebook_name"
		if err := c.contextService.UpdateContext(ctx.PhoneNumber, contextData); err != nil {
			log.Error().Err(err).Msg("Failed to update context")
		}
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"📓 What would you like to name the new notebook?")
	}

	// Search for existing notebook
	var notebook models.Notebook
	query := ctx.DB.Where("clerk_user_id = ? AND LOWER(name) LIKE ?", ctx.User.ClerkUserID, "%"+strings.ToLower(notebookName)+"%")

	if ctx.OrganizationID != nil {
		query = query.Where("organization_id = ?", *ctx.OrganizationID)
	} else {
		query = query.Where("organization_id IS NULL")
	}

	err := query.First(&notebook).Error

	if err != nil {
		// Notebook not found, ask if user wants to create it
		contextData["step"] = "awaiting_notebook_confirmation"
		contextData["notebook_name"] = notebookName
		contextData["create_notebook"] = true

		if err := c.contextService.UpdateContext(ctx.PhoneNumber, contextData); err != nil {
			log.Error().Err(err).Msg("Failed to update context")
		}

		message := fmt.Sprintf("📓 Notebook *%s* not found.\n\nWould you like to create it? (yes/no)", notebookName)
		return ctx.Client.SendTextMessage(ctx.PhoneNumber, message)
	}

	// Notebook found, store it and ask for chapter
	contextData["step"] = "awaiting_chapter"
	contextData["notebook_id"] = notebook.ID
	contextData["notebook_name"] = notebook.Name

	if err := c.contextService.UpdateContext(ctx.PhoneNumber, contextData); err != nil {
		log.Error().Err(err).Msg("Failed to update context")
	}

	message := fmt.Sprintf("✅ Using notebook: *%s*\n\nWhich chapter would you like to add this note to?\n\n_Type the chapter name or 'new' to create a new chapter_", notebook.Name)
	return ctx.Client.SendTextMessage(ctx.PhoneNumber, message)
}

// handleNotebookConfirmation processes the notebook creation confirmation
func (c *AddNoteCommand) handleNotebookConfirmation(ctx *whatsapp.CommandContext, contextData map[string]interface{}) error {
	response := strings.ToLower(strings.TrimSpace(ctx.Message))

	if response == "yes" || response == "y" {
		// Create the notebook
		notebookName, _ := contextData["notebook_name"].(string)

		notebook := models.Notebook{
			Name:        notebookName,
			ClerkUserID: ctx.User.ClerkUserID,
		}

		if ctx.OrganizationID != nil {
			notebook.OrganizationID = ctx.OrganizationID
		}

		if err := ctx.DB.Create(&notebook).Error; err != nil {
			log.Error().Err(err).Msg("Failed to create notebook")
			return ctx.Client.SendTextMessage(ctx.PhoneNumber,
				"❌ Failed to create notebook. Please try again.")
		}

		// Update context with new notebook
		contextData["step"] = "awaiting_chapter"
		contextData["notebook_id"] = notebook.ID
		contextData["notebook_name"] = notebook.Name

		if err := c.contextService.UpdateContext(ctx.PhoneNumber, contextData); err != nil {
			log.Error().Err(err).Msg("Failed to update context")
		}

		message := fmt.Sprintf("✅ Created notebook: *%s*\n\nWhich chapter would you like to add this note to?\n\n_Type the chapter name or 'new' to create a new chapter_", notebook.Name)
		return ctx.Client.SendTextMessage(ctx.PhoneNumber, message)
	}

	// User declined, clear context
	if err := c.contextService.ClearContext(ctx.PhoneNumber); err != nil {
		log.Error().Err(err).Msg("Failed to clear context")
	}

	return ctx.Client.SendTextMessage(ctx.PhoneNumber,
		"❌ Note creation cancelled. Use /add to start over.")
}

// handleChapterResponse processes the chapter name response
func (c *AddNoteCommand) handleChapterResponse(ctx *whatsapp.CommandContext, contextData map[string]interface{}) error {
	chapterName := strings.TrimSpace(ctx.Message)

	if chapterName == "" {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Please provide a chapter name.")
	}

	notebookID, _ := contextData["notebook_id"].(string)

	// Check if user wants to create a new chapter
	if strings.ToLower(chapterName) == "new" {
		contextData["step"] = "awaiting_new_chapter_name"
		if err := c.contextService.UpdateContext(ctx.PhoneNumber, contextData); err != nil {
			log.Error().Err(err).Msg("Failed to update context")
		}
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"📑 What would you like to name the new chapter?")
	}

	// Search for existing chapter
	var chapter models.Chapter
	query := ctx.DB.Where("notebook_id = ? AND LOWER(name) LIKE ?", notebookID, "%"+strings.ToLower(chapterName)+"%")

	if ctx.OrganizationID != nil {
		query = query.Where("organization_id = ?", *ctx.OrganizationID)
	} else {
		query = query.Where("organization_id IS NULL")
	}

	err := query.First(&chapter).Error

	if err != nil {
		// Chapter not found, ask if user wants to create it
		contextData["step"] = "awaiting_chapter_confirmation"
		contextData["chapter_name"] = chapterName
		contextData["create_chapter"] = true

		if err := c.contextService.UpdateContext(ctx.PhoneNumber, contextData); err != nil {
			log.Error().Err(err).Msg("Failed to update context")
		}

		message := fmt.Sprintf("📑 Chapter *%s* not found.\n\nWould you like to create it? (yes/no)", chapterName)
		return ctx.Client.SendTextMessage(ctx.PhoneNumber, message)
	}

	// Chapter found, create the note
	return c.createNote(ctx, contextData, chapter.ID)
}

// handleChapterConfirmation processes the chapter creation confirmation
func (c *AddNoteCommand) handleChapterConfirmation(ctx *whatsapp.CommandContext, contextData map[string]interface{}) error {
	response := strings.ToLower(strings.TrimSpace(ctx.Message))

	if response == "yes" || response == "y" {
		// Create the chapter
		chapterName, _ := contextData["chapter_name"].(string)
		notebookID, _ := contextData["notebook_id"].(string)

		chapter := models.Chapter{
			Name:       chapterName,
			NotebookID: notebookID,
		}

		if ctx.OrganizationID != nil {
			chapter.OrganizationID = ctx.OrganizationID
		}

		if err := ctx.DB.Create(&chapter).Error; err != nil {
			log.Error().Err(err).Msg("Failed to create chapter")
			return ctx.Client.SendTextMessage(ctx.PhoneNumber,
				"❌ Failed to create chapter. Please try again.")
		}

		// Create the note
		return c.createNote(ctx, contextData, chapter.ID)
	}

	// User declined, clear context
	if err := c.contextService.ClearContext(ctx.PhoneNumber); err != nil {
		log.Error().Err(err).Msg("Failed to clear context")
	}

	return ctx.Client.SendTextMessage(ctx.PhoneNumber,
		"❌ Note creation cancelled. Use /add to start over.")
}

// handleNewNotebookName handles the response when user provides a new notebook name
func (c *AddNoteCommand) handleNewNotebookName(ctx *whatsapp.CommandContext, contextData map[string]interface{}) error {
	notebookName := strings.TrimSpace(ctx.Message)

	if notebookName == "" {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Please provide a notebook name.")
	}

	// Validate and sanitize notebook name
	if err := utils.ValidateCommandArgument("name", notebookName); err != nil {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			fmt.Sprintf("❌ Invalid notebook name: %s", err.Error()))
	}
	notebookName = utils.SanitizeCommandArgument(notebookName)

	// Create the notebook
	notebook := models.Notebook{
		Name:        notebookName,
		ClerkUserID: ctx.User.ClerkUserID,
	}

	if ctx.OrganizationID != nil {
		notebook.OrganizationID = ctx.OrganizationID
	}

	if err := ctx.DB.Create(&notebook).Error; err != nil {
		log.Error().Err(err).Msg("Failed to create notebook")
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Failed to create notebook. Please try again.")
	}

	// Update context with new notebook
	contextData["step"] = "awaiting_chapter"
	contextData["notebook_id"] = notebook.ID
	contextData["notebook_name"] = notebook.Name

	if err := c.contextService.UpdateContext(ctx.PhoneNumber, contextData); err != nil {
		log.Error().Err(err).Msg("Failed to update context")
	}

	message := fmt.Sprintf("✅ Created notebook: *%s*\n\nWhich chapter would you like to add this note to?\n\n_Type the chapter name or 'new' to create a new chapter_", notebook.Name)
	return ctx.Client.SendTextMessage(ctx.PhoneNumber, message)
}

// handleNewChapterName handles the response when user provides a new chapter name
func (c *AddNoteCommand) handleNewChapterName(ctx *whatsapp.CommandContext, contextData map[string]interface{}) error {
	chapterName := strings.TrimSpace(ctx.Message)

	if chapterName == "" {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Please provide a chapter name.")
	}

	notebookID, _ := contextData["notebook_id"].(string)

	// Validate and sanitize chapter name
	if err := utils.ValidateCommandArgument("name", chapterName); err != nil {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			fmt.Sprintf("❌ Invalid chapter name: %s", err.Error()))
	}
	chapterName = utils.SanitizeCommandArgument(chapterName)

	// Create the chapter
	chapter := models.Chapter{
		Name:       chapterName,
		NotebookID: notebookID,
	}

	if ctx.OrganizationID != nil {
		chapter.OrganizationID = ctx.OrganizationID
	}

	if err := ctx.DB.Create(&chapter).Error; err != nil {
		log.Error().Err(err).Msg("Failed to create chapter")
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Failed to create chapter. Please try again.")
	}

	// Create the note
	return c.createNote(ctx, contextData, chapter.ID)
}

// createNote creates the note with the collected information
func (c *AddNoteCommand) createNote(ctx *whatsapp.CommandContext, contextData map[string]interface{}, chapterID string) error {
	noteTitle, _ := contextData["title"].(string)
	notebookName, _ := contextData["notebook_name"].(string)

	// Get chapter name
	var chapter models.Chapter
	if err := ctx.DB.First(&chapter, "id = ?", chapterID).Error; err != nil {
		log.Error().Err(err).Msg("Failed to get chapter")
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Failed to create note. Please try again.")
	}

	// Create the note
	note := models.Notes{
		Name:      noteTitle,
		Content:   "", // Empty content initially
		ChapterID: chapterID,
	}

	if ctx.OrganizationID != nil {
		note.OrganizationID = ctx.OrganizationID
	}

	if err := ctx.DB.Create(&note).Error; err != nil {
		log.Error().Err(err).Msg("Failed to create note")
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Failed to create note. Please try again.")
	}

	// Store note ID for AI content generation
	contextData["note_id"] = note.ID
	contextData["chapter_name"] = chapter.Name
	contextData["step"] = "awaiting_ai_content"

	if err := c.contextService.UpdateContext(ctx.PhoneNumber, contextData); err != nil {
		log.Error().Err(err).Msg("Failed to update context")
		// Continue anyway, just skip AI generation
		return c.sendSuccessMessage(ctx, noteTitle, notebookName, chapter.Name, false)
	}

	// Ask if user wants AI-generated content
	message := fmt.Sprintf("✅ *Note created successfully!*\n\n📝 *Title:* %s\n📓 *Notebook:* %s\n📑 *Chapter:* %s\n\n🤖 Would you like me to generate content for this note using AI?\n\n_Reply 'yes' to generate content or 'no' to finish_", noteTitle, notebookName, chapter.Name)
	return ctx.Client.SendTextMessage(ctx.PhoneNumber, message)
}

// handleNaturalLanguageFlow processes natural language commands with prefilled data
func (c *AddNoteCommand) handleNaturalLanguageFlow(ctx *whatsapp.CommandContext, contextData map[string]interface{}) error {
	noteTitle, _ := contextData["title"].(string)
	notebookName, _ := contextData["notebook_name"].(string)
	chapterName, _ := contextData["chapter_name"].(string)

	// Confirm with user
	var message string
	if chapterName != "" {
		message = fmt.Sprintf("📝 I'll create a note:\n\n*Title:* %s\n*Notebook:* %s\n*Chapter:* %s\n\nIs this correct? (yes/no)", noteTitle, notebookName, chapterName)
	} else {
		message = fmt.Sprintf("📝 I'll create a note:\n\n*Title:* %s\n*Notebook:* %s\n\nIs this correct? (yes/no)\n\n_Note: I'll ask for the chapter next_", noteTitle, notebookName)
	}

	// Remove prefilled flag so continueFlow works normally
	delete(contextData, "prefilled")
	if err := c.contextService.UpdateContext(ctx.PhoneNumber, contextData); err != nil {
		log.Error().Err(err).Msg("Failed to update context")
	}

	return ctx.Client.SendTextMessage(ctx.PhoneNumber, message)
}

// handleNaturalLanguageNotebookConfirmation handles confirmation for prefilled notebook from natural language
func (c *AddNoteCommand) handleNaturalLanguageNotebookConfirmation(ctx *whatsapp.CommandContext, contextData map[string]interface{}) error {
	response := strings.ToLower(strings.TrimSpace(ctx.Message))

	if response != "yes" && response != "y" {
		// User declined, clear context and start over
		if err := c.contextService.ClearContext(ctx.PhoneNumber); err != nil {
			log.Error().Err(err).Msg("Failed to clear context")
		}
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Cancelled. Use /add to start over with a different configuration.")
	}

	// User confirmed, proceed with the flow
	notebookName, _ := contextData["notebook_name"].(string)
	chapterName, _ := contextData["chapter_name"].(string)

	if chapterName != "" {
		// Both notebook and chapter provided, validate and create note
		return c.validateAndCreateNote(ctx, contextData, notebookName, chapterName)
	} else {
		// Only notebook provided, ask for chapter
		contextData["step"] = "awaiting_chapter"
		if err := c.contextService.UpdateContext(ctx.PhoneNumber, contextData); err != nil {
			log.Error().Err(err).Msg("Failed to update context")
		}

		message := fmt.Sprintf("✅ Using notebook: *%s*\n\nWhich chapter would you like to add this note to?\n\n_Type the chapter name or 'new' to create a new chapter_", notebookName)
		return ctx.Client.SendTextMessage(ctx.PhoneNumber, message)
	}
}

// handleNaturalLanguageChapterConfirmation handles confirmation for prefilled chapter from natural language
func (c *AddNoteCommand) handleNaturalLanguageChapterConfirmation(ctx *whatsapp.CommandContext, contextData map[string]interface{}) error {
	response := strings.ToLower(strings.TrimSpace(ctx.Message))

	if response != "yes" && response != "y" {
		// User declined, ask for different chapter
		contextData["step"] = "awaiting_chapter"
		delete(contextData, "chapter_name")
		if err := c.contextService.UpdateContext(ctx.PhoneNumber, contextData); err != nil {
			log.Error().Err(err).Msg("Failed to update context")
		}

		notebookName, _ := contextData["notebook_name"].(string)
		message := fmt.Sprintf("Which chapter would you like to add this note to in *%s*?\n\n_Type the chapter name or 'new' to create a new chapter_", notebookName)
		return ctx.Client.SendTextMessage(ctx.PhoneNumber, message)
	}

	// User confirmed, validate and create note
	notebookName, _ := contextData["notebook_name"].(string)
	chapterName, _ := contextData["chapter_name"].(string)
	return c.validateAndCreateNote(ctx, contextData, notebookName, chapterName)
}

// validateAndCreateNote validates notebook and chapter exist, then creates the note
func (c *AddNoteCommand) validateAndCreateNote(ctx *whatsapp.CommandContext, contextData map[string]interface{}, notebookName, chapterName string) error {
	// Find the notebook
	var notebook models.Notebook
	query := ctx.DB.Where("LOWER(name) = ? AND clerk_user_id = ?", strings.ToLower(notebookName), ctx.User.ClerkUserID)
	if ctx.OrganizationID != nil {
		query = query.Where("organization_id = ?", *ctx.OrganizationID)
	} else {
		query = query.Where("organization_id IS NULL")
	}

	if err := query.First(&notebook).Error; err != nil {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			fmt.Sprintf("❌ Notebook '%s' not found. Use 'new' to create it or type an existing notebook name.", notebookName))
	}

	// Find the chapter
	var chapter models.Chapter
	chapterQuery := ctx.DB.Where("LOWER(name) = ? AND notebook_id = ?", strings.ToLower(chapterName), notebook.ID)
	if err := chapterQuery.First(&chapter).Error; err != nil {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			fmt.Sprintf("❌ Chapter '%s' not found in notebook '%s'. Use 'new' to create it or type an existing chapter name.", chapterName, notebookName))
	}

	// Create the note
	return c.createNote(ctx, contextData, chapter.ID)
}

// sendSuccessMessage sends the final success message
func (c *AddNoteCommand) sendSuccessMessage(ctx *whatsapp.CommandContext, noteTitle, notebookName, chapterName string, aiGenerated bool) error {
	var message string
	if aiGenerated {
		message = fmt.Sprintf("✅ *Note created with AI content!*\n\n📝 *Title:* %s\n📓 *Notebook:* %s\n📑 *Chapter:* %s\n\n🤖 AI-generated content has been added to your note.\nYou can view and edit it in the application.", noteTitle, notebookName, chapterName)
	} else {
		message = fmt.Sprintf("✅ *Note created!*\n\n📝 *Title:* %s\n📓 *Notebook:* %s\n📑 *Chapter:* %s\n\nYou can now edit this note in the application.", noteTitle, notebookName, chapterName)
	}
	return ctx.Client.SendTextMessage(ctx.PhoneNumber, message)
}

// handleAIContentResponse handles the response for AI content generation
func (c *AddNoteCommand) handleAIContentResponse(ctx *whatsapp.CommandContext, contextData map[string]interface{}) error {
	response := strings.ToLower(strings.TrimSpace(ctx.Message))
	noteID, _ := contextData["note_id"].(string)
	noteTitle, _ := contextData["title"].(string)
	notebookName, _ := contextData["notebook_name"].(string)
	chapterName, _ := contextData["chapter_name"].(string)

	// Clear context before proceeding
	if err := c.contextService.ClearContext(ctx.PhoneNumber); err != nil {
		log.Error().Err(err).Msg("Failed to clear context")
	}

	if response == "yes" || response == "y" {
		// Generate AI content (returns markdown)
		markdownContent, err := c.aiService.GenerateNoteContent(context.Background(), services.NoteContentGenerationRequest{
			NoteTitle: noteTitle,
			UserID:    ctx.User.ClerkUserID,
			OrgID:     ctx.OrganizationID,
		})

		if err != nil {
			log.Error().Err(err).Msg("Failed to generate AI content")
			ctx.Client.SendTextMessage(ctx.PhoneNumber,
				"⚠️ Failed to generate AI content. Your note has been created but remains empty.")
			return c.sendSuccessMessage(ctx, noteTitle, notebookName, chapterName, false)
		}

		// Convert markdown to TipTap JSON format
		tiptapContent, err := utils.MarkdownToTipTap(markdownContent)
		if err != nil {
			log.Error().Err(err).Msg("Failed to convert markdown to TipTap format")
			ctx.Client.SendTextMessage(ctx.PhoneNumber,
				"⚠️ Failed to convert content. Your note has been created but remains empty.")
			return c.sendSuccessMessage(ctx, noteTitle, notebookName, chapterName, false)
		}

		// Update note with generated content in TipTap format
		if err := ctx.DB.Model(&models.Notes{}).Where("id = ?", noteID).Update("content", tiptapContent).Error; err != nil {
			log.Error().Err(err).Msg("Failed to update note with AI content")
			ctx.Client.SendTextMessage(ctx.PhoneNumber,
				"⚠️ Content was generated but failed to save. Your note remains empty.")
			return c.sendSuccessMessage(ctx, noteTitle, notebookName, chapterName, false)
		}

		return c.sendSuccessMessage(ctx, noteTitle, notebookName, chapterName, true)
	}

	// User declined AI generation
	return c.sendSuccessMessage(ctx, noteTitle, notebookName, chapterName, false)
}
