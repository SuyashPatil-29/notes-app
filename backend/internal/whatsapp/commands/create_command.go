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

// CreateCommand handles the /create command for creating notebooks and chapters
type CreateCommand struct {
	contextService *services.WhatsAppContextService
}

// NewCreateCommand creates a new create command
func NewCreateCommand(contextService *services.WhatsAppContextService) *CreateCommand {
	return &CreateCommand{
		contextService: contextService,
	}
}

// Name returns the command name
func (c *CreateCommand) Name() string {
	return "create"
}

// Description returns the command description
func (c *CreateCommand) Description() string {
	return "Create a new notebook or chapter"
}

// Usage returns usage instructions
func (c *CreateCommand) Usage() string {
	return "/create [notebook|chapter] [name] - Create a new notebook or chapter"
}

// RequiresAuth returns whether authentication is required
func (c *CreateCommand) RequiresAuth() bool {
	return true
}

// Execute runs the create command
func (c *CreateCommand) Execute(ctx *whatsapp.CommandContext) error {
	// Verify organization access if in organization mode
	if err := whatsapp.VerifyOrganizationAccess(ctx); err != nil {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			fmt.Sprintf("❌ %s", err.Error()))
	}

	// Check if we have an active context
	if ctx.ConversationCtx != nil && ctx.ConversationCtx.Command == "create" {
		return c.continueFlow(ctx)
	}

	// Start new create flow
	return c.startFlow(ctx)
}

// startFlow initiates the create flow
func (c *CreateCommand) startFlow(ctx *whatsapp.CommandContext) error {
	// Check if user provided entity type
	if len(ctx.Args) == 0 {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Please specify what to create.\n\n*Usage:* /create [notebook|chapter] [name]")
	}

	entityType := strings.ToLower(ctx.Args[0])

	switch entityType {
	case "notebook":
		return c.createNotebook(ctx)
	case "chapter":
		return c.createChapter(ctx)
	default:
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Invalid entity type. Use 'notebook' or 'chapter'.\n\n*Usage:* /create [notebook|chapter] [name]")
	}
}

// createNotebook creates a new notebook
func (c *CreateCommand) createNotebook(ctx *whatsapp.CommandContext) error {
	// Check if name was provided
	if len(ctx.Args) < 2 {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Please provide a notebook name.\n\n*Usage:* /create notebook [name]")
	}

	notebookName := strings.Join(ctx.Args[1:], " ")

	// Validate and sanitize notebook name
	if err := utils.ValidateCommandArgument("notebook_name", notebookName); err != nil {
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

	message := fmt.Sprintf("✅ *Notebook created successfully!*\n\n📓 *%s*\n\nYou can now add chapters and notes to this notebook.", notebookName)
	return ctx.Client.SendTextMessage(ctx.PhoneNumber, message)
}

// createChapter creates a new chapter
func (c *CreateCommand) createChapter(ctx *whatsapp.CommandContext) error {
	// Check if name was provided
	if len(ctx.Args) < 2 {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Please provide a chapter name.\n\n*Usage:* /create chapter [name]")
	}

	chapterName := strings.Join(ctx.Args[1:], " ")

	// Validate and sanitize chapter name
	if err := utils.ValidateCommandArgument("chapter_name", chapterName); err != nil {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			fmt.Sprintf("❌ Invalid chapter name: %s", err.Error()))
	}
	chapterName = utils.SanitizeCommandArgument(chapterName)

	// Get user's notebooks
	var notebooks []models.Notebook
	query := ctx.DB.Where("clerk_user_id = ?", ctx.User.ClerkUserID)

	if ctx.OrganizationID != nil {
		query = query.Where("organization_id = ?", *ctx.OrganizationID)
	} else {
		query = query.Where("organization_id IS NULL")
	}

	if err := query.Order("created_at DESC").Find(&notebooks).Error; err != nil {
		log.Error().Err(err).Msg("Failed to get notebooks")
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ An error occurred. Please try again.")
	}

	if len(notebooks) == 0 {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ You don't have any notebooks yet.\n\nCreate a notebook first using: /create notebook [name]")
	}

	// If only one notebook, use it directly
	if len(notebooks) == 1 {
		return c.createChapterInNotebook(ctx, chapterName, notebooks[0].ID, notebooks[0].Name)
	}

	// Multiple notebooks, ask user to select
	return c.showNotebookSelection(ctx, notebooks, chapterName)
}

// showNotebookSelection displays a list of notebooks for selection
func (c *CreateCommand) showNotebookSelection(ctx *whatsapp.CommandContext, notebooks []models.Notebook, chapterName string) error {
	var message strings.Builder
	message.WriteString(fmt.Sprintf("📓 *Select a notebook for chapter '%s'*\n\n", chapterName))

	// Store notebook IDs and names for selection
	notebookIDs := make([]string, len(notebooks))
	notebookNames := make([]string, len(notebooks))

	for i, notebook := range notebooks {
		notebookIDs[i] = notebook.ID
		notebookNames[i] = notebook.Name
		message.WriteString(fmt.Sprintf("%d. *%s*\n", i+1, notebook.Name))
	}

	message.WriteString(fmt.Sprintf("\n_Reply with a number (1-%d) or 'cancel' to abort_", len(notebooks)))

	// Store context for selection
	contextData := map[string]interface{}{
		"chapter_name":   chapterName,
		"notebook_ids":   notebookIDs,
		"notebook_names": notebookNames,
		"step":           "awaiting_notebook_selection",
	}

	if err := c.contextService.SetContext(ctx.PhoneNumber, "create", contextData); err != nil {
		log.Error().Err(err).Msg("Failed to set context for create command")
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ An error occurred. Please try again.")
	}

	return ctx.Client.SendTextMessage(ctx.PhoneNumber, message.String())
}

// continueFlow continues the multi-step create flow
func (c *CreateCommand) continueFlow(ctx *whatsapp.CommandContext) error {
	// Parse context data
	var contextData map[string]interface{}
	if err := json.Unmarshal([]byte(ctx.ConversationCtx.Data), &contextData); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal context data")
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ An error occurred. Please start over with /create")
	}

	step, _ := contextData["step"].(string)

	switch step {
	case "awaiting_notebook_selection":
		return c.handleNotebookSelection(ctx, contextData)
	default:
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Invalid state. Please start over with /create")
	}
}

// handleNotebookSelection processes the notebook selection
func (c *CreateCommand) handleNotebookSelection(ctx *whatsapp.CommandContext, contextData map[string]interface{}) error {
	response := strings.ToLower(strings.TrimSpace(ctx.Message))

	// Check for cancellation
	if response == "cancel" {
		if err := c.contextService.ClearContext(ctx.PhoneNumber); err != nil {
			log.Error().Err(err).Msg("Failed to clear context")
		}
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Chapter creation cancelled.")
	}

	// Parse selection number
	selection, err := strconv.Atoi(response)
	if err != nil {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Invalid selection. Please reply with a number or 'cancel'.")
	}

	// Get notebook IDs from context
	notebookIDsInterface, ok := contextData["notebook_ids"].([]interface{})
	if !ok {
		log.Error().Msg("Failed to get notebook IDs from context")
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ An error occurred. Please start over with /create")
	}

	notebookIDs := make([]string, len(notebookIDsInterface))
	for i, id := range notebookIDsInterface {
		notebookIDs[i] = id.(string)
	}

	// Get notebook names from context
	notebookNamesInterface, ok := contextData["notebook_names"].([]interface{})
	if !ok {
		log.Error().Msg("Failed to get notebook names from context")
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ An error occurred. Please start over with /create")
	}

	notebookNames := make([]string, len(notebookNamesInterface))
	for i, name := range notebookNamesInterface {
		notebookNames[i] = name.(string)
	}

	// Validate selection
	if selection < 1 || selection > len(notebookIDs) {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			fmt.Sprintf("❌ Invalid selection. Please choose a number between 1 and %d.", len(notebookIDs)))
	}

	// Get selected notebook
	selectedNotebookID := notebookIDs[selection-1]
	selectedNotebookName := notebookNames[selection-1]
	chapterName, _ := contextData["chapter_name"].(string)

	// Clear context
	if err := c.contextService.ClearContext(ctx.PhoneNumber); err != nil {
		log.Error().Err(err).Msg("Failed to clear context")
	}

	// Create the chapter
	return c.createChapterInNotebook(ctx, chapterName, selectedNotebookID, selectedNotebookName)
}

// createChapterInNotebook creates a chapter in the specified notebook
func (c *CreateCommand) createChapterInNotebook(ctx *whatsapp.CommandContext, chapterName, notebookID, notebookName string) error {
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

	message := fmt.Sprintf("✅ *Chapter created successfully!*\n\n📑 *%s*\n📓 In notebook: *%s*\n\nYou can now add notes to this chapter.", chapterName, notebookName)
	return ctx.Client.SendTextMessage(ctx.PhoneNumber, message)
}
