package services

import (
	"backend/internal/config"
	"backend/internal/models"
	"backend/internal/utils"
	"backend/internal/whatsapp"
	whatsappclient "backend/pkg/whatsapp"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// WhatsAppMessageProcessor handles incoming WhatsApp messages and routes them to appropriate handlers
type WhatsAppMessageProcessor struct {
	db             *gorm.DB
	config         *config.WhatsAppConfig
	client         whatsappclient.WhatsAppClient
	authService    WhatsAppAuthService
	contextService *WhatsAppContextService
	registry       *whatsapp.CommandRegistry
	auditService   *WhatsAppAuditService
	metricsService *WhatsAppMetricsService
}

// NewWhatsAppMessageProcessor creates a new message processor
func NewWhatsAppMessageProcessor(
	db *gorm.DB,
	config *config.WhatsAppConfig,
	client whatsappclient.WhatsAppClient,
	authService WhatsAppAuthService,
	contextService *WhatsAppContextService,
	registry *whatsapp.CommandRegistry,
	auditService *WhatsAppAuditService,
	metricsService *WhatsAppMetricsService,
) *WhatsAppMessageProcessor {
	return &WhatsAppMessageProcessor{
		db:             db,
		config:         config,
		client:         client,
		authService:    authService,
		contextService: contextService,
		registry:       registry,
		auditService:   auditService,
		metricsService: metricsService,
	}
}

// IncomingMessage represents a parsed incoming WhatsApp message
type IncomingMessage struct {
	MessageID   string
	PhoneNumber string
	Content     string
	Timestamp   time.Time
	GroupID     *string
}

// ProcessMessage handles an incoming WhatsApp message
func (p *WhatsAppMessageProcessor) ProcessMessage(msg *IncomingMessage) error {
	// Start timing message processing
	timer := p.metricsService.MessageProcessingTimer()
	defer timer.ObserveDuration()

	// Record inbound message metric
	p.metricsService.RecordInboundMessage("text", "received")

	// Validate phone number
	if err := utils.ValidatePhoneNumber(msg.PhoneNumber); err != nil {
		log.Error().Err(err).Str("phone", msg.PhoneNumber).Msg("Invalid phone number")
		p.metricsService.RecordError("message_processor", "invalid_phone_number")
		return fmt.Errorf("invalid phone number: %w", err)
	}

	// Validate and sanitize message content
	sanitizedContent, err := utils.ValidateAndSanitizeMessage(msg.Content)
	if err != nil {
		log.Error().Err(err).Str("phone", msg.PhoneNumber).Msg("Invalid message content")
		p.metricsService.RecordError("message_processor", "invalid_content")
		return p.sendErrorMessage(msg.PhoneNumber, "Your message contains invalid content. Please try again.")
	}
	msg.Content = sanitizedContent

	// Log the incoming message using audit service
	if err := p.auditService.LogInboundMessage(msg.MessageID, msg.PhoneNumber, "text", msg.Content, msg.Timestamp); err != nil {
		log.Warn().Err(err).Msg("Failed to log incoming message")
	}

	// Authenticate user
	user, err := p.authService.GetUserByPhone(msg.PhoneNumber)
	if err != nil {
		log.Error().Err(err).Str("phone", msg.PhoneNumber).Msg("Failed to get user")
		return p.sendErrorMessage(msg.PhoneNumber, "An error occurred. Please try again.")
	}

	// Check if user is authenticated
	if user == nil || !user.IsAuthenticated {
		return p.handleUnauthenticatedUser(msg)
	}

	// Check if session has expired
	if p.authService.IsSessionExpired(user) {
		return p.handleExpiredSession(msg)
	}

	// Update last active timestamp
	if err := p.authService.UpdateLastActive(msg.PhoneNumber); err != nil {
		log.Warn().Err(err).Msg("Failed to update last active")
	}

	// Check for organization context (group message)
	var organizationID *string
	if msg.GroupID != nil {
		orgID, err := p.getOrganizationForGroup(*msg.GroupID)
		if err != nil {
			log.Error().Err(err).Str("group_id", *msg.GroupID).Msg("Failed to get organization for group")
		} else if orgID != nil {
			organizationID = orgID
		}
	}

	// Check for active conversation context
	conversationCtx, err := p.contextService.GetContext(msg.PhoneNumber)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get conversation context")
		return p.sendErrorMessage(msg.PhoneNumber, "An error occurred. Please try again.")
	}

	// Build command context
	cmdCtx := &whatsapp.CommandContext{
		PhoneNumber:     msg.PhoneNumber,
		Message:         msg.Content,
		User:            user,
		ConversationCtx: conversationCtx,
		Client:          p.client,
		DB:              p.db,
		GroupID:         msg.GroupID,
		OrganizationID:  organizationID,
	}

	// If we have an active context, continue the flow
	if conversationCtx != nil {
		return p.continueConversationFlow(cmdCtx)
	}

	// Otherwise, parse and execute command
	return p.executeCommand(cmdCtx)
}

// handleUnauthenticatedUser sends authentication instructions to unauthenticated users
func (p *WhatsAppMessageProcessor) handleUnauthenticatedUser(msg *IncomingMessage) error {
	// Generate a link token for authentication
	linkToken, err := p.authService.GenerateLinkToken(msg.PhoneNumber)
	if err != nil {
		log.Error().Err(err).Str("phone", msg.PhoneNumber).Msg("Failed to generate link token")
		return p.sendErrorMessage(msg.PhoneNumber, "An error occurred. Please try again.")
	}

	// Build authentication URL from config
	baseURL := p.config.FrontendURL
	if baseURL == "" {
		baseURL = "http://localhost:5173"
	}
	authURL := fmt.Sprintf("%s/whatsapp-auth?token=%s", baseURL, linkToken)

	message := "👋 *Welcome to NotesApp!*\n\n" +
		"To use this service, you need to link your WhatsApp account.\n\n" +
		"🔗 *Click here to authenticate:*\n" +
		authURL + "\n\n" +
		"Once linked, you'll be able to:\n" +
		"• Create and manage notes\n" +
		"• Search and retrieve notes\n" +
		"• Organize notebooks and chapters\n" +
		"• And much more!\n\n" +
		"_Type /help after authentication to see all available commands._"

	log.Info().
		Str("phone", msg.PhoneNumber).
		Str("auth_url", authURL).
		Msg("Sent authentication instructions to unauthenticated user")

	return p.client.SendTextMessage(msg.PhoneNumber, message)
}

// handleExpiredSession sends re-authentication instructions to users with expired sessions
func (p *WhatsAppMessageProcessor) handleExpiredSession(msg *IncomingMessage) error {
	// Clear the expired authentication
	err := p.db.Model(&models.WhatsAppUser{}).
		Where("phone_number = ?", msg.PhoneNumber).
		Updates(map[string]interface{}{
			"is_authenticated": false,
			"auth_token":       "",
		}).Error

	if err != nil {
		log.Error().Err(err).Str("phone", msg.PhoneNumber).Msg("Failed to clear expired session")
	}

	// Generate a new link token for re-authentication
	linkToken, err := p.authService.GenerateLinkToken(msg.PhoneNumber)
	if err != nil {
		log.Error().Err(err).Str("phone", msg.PhoneNumber).Msg("Failed to generate link token")
		return p.sendErrorMessage(msg.PhoneNumber, "An error occurred. Please try again.")
	}

	// Build authentication URL from config
	baseURL := p.config.FrontendURL
	if baseURL == "" {
		baseURL = "http://localhost:5173"
	}
	authURL := fmt.Sprintf("%s/whatsapp-auth?token=%s", baseURL, linkToken)

	message := "🔒 *Your session has expired*\n\n" +
		"For security reasons, your WhatsApp session has expired after 30 days of inactivity.\n\n" +
		"🔗 *Click here to re-authenticate:*\n" +
		authURL + "\n\n" +
		"Once re-authenticated, you'll be able to continue using all features."

	log.Info().
		Str("phone", msg.PhoneNumber).
		Msg("Sent re-authentication instructions for expired session")

	return p.client.SendTextMessage(msg.PhoneNumber, message)
}

// continueConversationFlow continues an active multi-step conversation
func (p *WhatsAppMessageProcessor) continueConversationFlow(ctx *whatsapp.CommandContext) error {
	// Get the command for the active context
	cmd, exists := p.registry.Get(ctx.ConversationCtx.Command)
	if !exists {
		log.Error().
			Str("command", ctx.ConversationCtx.Command).
			Msg("Command not found for active context")

		// Clear invalid context
		if err := p.contextService.ClearContext(ctx.PhoneNumber); err != nil {
			log.Error().Err(err).Msg("Failed to clear invalid context")
		}

		return p.sendErrorMessage(ctx.PhoneNumber,
			"An error occurred with your previous command. Please start over.")
	}

	// Execute the command (it will handle the continuation) with timing
	cmdTimer := p.metricsService.CommandTimer(cmd.Name())
	err := cmd.Execute(ctx)
	cmdTimer.ObserveDuration()

	if err != nil {
		log.Error().
			Err(err).
			Str("command", cmd.Name()).
			Str("phone", ctx.PhoneNumber).
			Msg("Command execution failed")

		p.metricsService.RecordCommandExecution(cmd.Name(), "failed")
		p.metricsService.RecordCommandError(cmd.Name(), "execution_error")
		p.metricsService.RecordErrorByCategory("command_execution")

		return p.sendErrorMessage(ctx.PhoneNumber,
			"Failed to process your request. Please try again or use /help for assistance.")
	}

	p.metricsService.RecordCommandExecution(cmd.Name(), "success")
	return nil
}

// executeCommand parses and executes a command
func (p *WhatsAppMessageProcessor) executeCommand(ctx *whatsapp.CommandContext) error {
	// Parse command from message
	cmd, args, err := p.registry.GetCommandFromMessage(ctx.Message)
	if err != nil {
		// Not a command or unknown command
		if strings.HasPrefix(ctx.Message, "/") {
			// User tried to use a command but it doesn't exist
			commandName := strings.TrimPrefix(strings.Fields(ctx.Message)[0], "/")
			return p.sendErrorMessage(ctx.PhoneNumber,
				fmt.Sprintf("❌ Unknown command: *%s*\n\nUse /help to see all available commands.", commandName))
		}

		// Try parsing as natural language command
		if nlCmd, isNL := whatsapp.ParseNaturalLanguageAddCommand(ctx.Message); isNL {
			return p.executeNaturalLanguageCommand(ctx, nlCmd)
		}

		// Regular message, not a command
		return p.sendErrorMessage(ctx.PhoneNumber,
			"Please use commands to interact with the bot. Type /help to see available commands.")
	}

	// Check if command requires authentication
	if cmd.RequiresAuth() && (ctx.User == nil || !ctx.User.IsAuthenticated) {
		return p.sendErrorMessage(ctx.PhoneNumber,
			"🔒 This command requires authentication.\n\nPlease link your WhatsApp account in the application first.")
	}

	// Set args in context
	ctx.Args = args

	// Execute the command with timing
	cmdTimer := p.metricsService.CommandTimer(cmd.Name())
	err = cmd.Execute(ctx)
	cmdTimer.ObserveDuration()

	if err != nil {
		log.Error().
			Err(err).
			Str("command", cmd.Name()).
			Str("phone", ctx.PhoneNumber).
			Msg("Command execution failed")

		p.metricsService.RecordCommandExecution(cmd.Name(), "failed")
		p.metricsService.RecordCommandError(cmd.Name(), "execution_error")
		p.metricsService.RecordErrorByCategory("command_execution")

		return p.sendErrorMessage(ctx.PhoneNumber,
			"❌ Failed to process your command. Please try again or use /help for assistance.")
	}

	p.metricsService.RecordCommandExecution(cmd.Name(), "success")
	return nil
}

// executeNaturalLanguageCommand handles natural language commands
func (p *WhatsAppMessageProcessor) executeNaturalLanguageCommand(ctx *whatsapp.CommandContext, nlCmd *whatsapp.NaturalLanguageCommand) error {
	// Check authentication for natural language commands
	if ctx.User == nil || !ctx.User.IsAuthenticated {
		return p.sendErrorMessage(ctx.PhoneNumber,
			"🔒 This command requires authentication.\n\nPlease link your WhatsApp account in the application first.")
	}

	// Use AI to process the entire note creation
	return p.createNoteWithAI(ctx, nlCmd.NoteTitle)
}

// createNoteWithAI creates a note using AI to organize and generate content
func (p *WhatsAppMessageProcessor) createNoteWithAI(ctx *whatsapp.CommandContext, noteTitle string) error {
	cmdTimer := p.metricsService.CommandTimer("ai_note_creation")
	defer cmdTimer.ObserveDuration()

	// Clear any existing context first (AI commands don't need conversation state)
	if err := p.contextService.ClearContext(ctx.PhoneNumber); err != nil {
		log.Debug().Err(err).Msg("No context to clear (expected for AI commands)")
	}

	// Send initial processing message
	p.client.SendTextMessage(ctx.PhoneNumber, "🤖 Creating your note with AI...\n\n_This may take a moment_")

	// Get existing notebooks
	var notebooks []models.Notebook
	query := p.db.Where("clerk_user_id = ?", ctx.User.ClerkUserID)
	if ctx.OrganizationID != nil {
		query = query.Where("organization_id = ?", *ctx.OrganizationID)
	} else {
		query = query.Where("organization_id IS NULL")
	}
	query.Find(&notebooks)

	existingNotebooks := make([]string, len(notebooks))
	for i, nb := range notebooks {
		existingNotebooks[i] = nb.Name
	}

	// Use AI to organize the note
	aiService := NewAIService()
	organization, err := aiService.OrganizeNoteWithAI(context.Background(), NoteOrganizationRequest{
		NoteTitle:         noteTitle,
		ExistingNotebooks: existingNotebooks,
		UserID:            ctx.User.ClerkUserID,
		OrgID:             ctx.OrganizationID,
	})

	if err != nil {
		log.Error().
			Err(err).
			Str("note_title", noteTitle).
			Str("user_id", ctx.User.ClerkUserID).
			Msg("Failed to organize note with AI")
		p.metricsService.RecordCommandExecution("ai_note_creation", "failed")

		// Send detailed error to user
		errorMsg := fmt.Sprintf("❌ Failed to organize your note with AI.\n\n*Error:* %s\n\nPlease check your OpenAI API key and try again.", err.Error())
		return p.sendErrorMessage(ctx.PhoneNumber, errorMsg)
	}

	// Find or create notebook
	var notebook models.Notebook
	notebookQuery := p.db.Where("LOWER(name) = ? AND clerk_user_id = ?",
		strings.ToLower(organization.NotebookName), ctx.User.ClerkUserID)
	if ctx.OrganizationID != nil {
		notebookQuery = notebookQuery.Where("organization_id = ?", *ctx.OrganizationID)
	} else {
		notebookQuery = notebookQuery.Where("organization_id IS NULL")
	}

	err = notebookQuery.First(&notebook).Error
	if err != nil {
		// Notebook doesn't exist, create it
		notebook = models.Notebook{
			Name:        organization.NotebookName,
			ClerkUserID: ctx.User.ClerkUserID,
		}
		if ctx.OrganizationID != nil {
			notebook.OrganizationID = ctx.OrganizationID
		}
		if err := p.db.Create(&notebook).Error; err != nil {
			log.Error().Err(err).Msg("Failed to create notebook")
			p.metricsService.RecordCommandExecution("ai_note_creation", "failed")
			return p.sendErrorMessage(ctx.PhoneNumber,
				"❌ Failed to create notebook. Please try again.")
		}
	}

	// Find or create chapter
	var chapter models.Chapter
	chapterQuery := p.db.Where("LOWER(name) = ? AND notebook_id = ?",
		strings.ToLower(organization.ChapterName), notebook.ID)

	err = chapterQuery.First(&chapter).Error
	if err != nil {
		// Chapter doesn't exist, create it
		chapter = models.Chapter{
			Name:       organization.ChapterName,
			NotebookID: notebook.ID,
		}
		if ctx.OrganizationID != nil {
			chapter.OrganizationID = ctx.OrganizationID
		}
		if err := p.db.Create(&chapter).Error; err != nil {
			log.Error().Err(err).Msg("Failed to create chapter")
			p.metricsService.RecordCommandExecution("ai_note_creation", "failed")
			return p.sendErrorMessage(ctx.PhoneNumber,
				"❌ Failed to create chapter. Please try again.")
		}
	}

	// Generate AI content
	markdownContent, err := aiService.GenerateNoteContent(context.Background(), NoteContentGenerationRequest{
		NoteTitle: noteTitle,
		UserID:    ctx.User.ClerkUserID,
		OrgID:     ctx.OrganizationID,
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to generate AI content")
		// Continue with empty content
		markdownContent = ""
	}

	// Convert markdown to TipTap format
	content := ""
	if markdownContent != "" {
		tiptapContent, err := utils.MarkdownToTipTap(markdownContent)
		if err != nil {
			log.Error().Err(err).Msg("Failed to convert markdown to TipTap")
			content = "" // Use empty content on conversion failure
		} else {
			content = tiptapContent
		}
	}

	// Create the note
	note := models.Notes{
		Name:      noteTitle,
		Content:   content,
		ChapterID: chapter.ID,
	}
	if ctx.OrganizationID != nil {
		note.OrganizationID = ctx.OrganizationID
	}

	if err := p.db.Create(&note).Error; err != nil {
		log.Error().Err(err).Msg("Failed to create note")
		p.metricsService.RecordCommandExecution("ai_note_creation", "failed")
		return p.sendErrorMessage(ctx.PhoneNumber,
			"❌ Failed to create note. Please try again.")
	}

	// Send success message
	aiContentStatus := ""
	if content != "" {
		aiContentStatus = "\n🤖 AI-generated content added!"
	}

	successMessage := fmt.Sprintf("✅ *Note created successfully!*\n\n📝 *Title:* %s\n📓 *Notebook:* %s\n📑 *Chapter:* %s%s\n\nYou can view and edit it in the application.",
		noteTitle, notebook.Name, chapter.Name, aiContentStatus)

	p.metricsService.RecordCommandExecution("ai_note_creation", "success")
	return p.client.SendTextMessage(ctx.PhoneNumber, successMessage)
}

// getOrganizationForGroup retrieves the organization ID for a WhatsApp group
func (p *WhatsAppMessageProcessor) getOrganizationForGroup(groupID string) (*string, error) {
	var groupLink models.WhatsAppGroupLink
	err := p.db.Where("group_id = ? AND is_active = ?", groupID, true).First(&groupLink).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No organization linked
		}
		return nil, err
	}

	return &groupLink.OrganizationID, nil
}

// sendErrorMessage sends an error message to the user
func (p *WhatsAppMessageProcessor) sendErrorMessage(phoneNumber, message string) error {
	return p.client.SendTextMessage(phoneNumber, message)
}

// UpdateMessageStatus updates the status of an outbound message (delegates to audit service)
func (p *WhatsAppMessageProcessor) UpdateMessageStatus(messageID, status string, errorCode, errorMessage *string) error {
	return p.auditService.UpdateMessageStatus(messageID, status, errorCode, errorMessage)
}
