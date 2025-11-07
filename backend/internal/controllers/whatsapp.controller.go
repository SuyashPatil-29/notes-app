package controllers

import (
	"backend/internal/services"
	whatsappclient "backend/pkg/whatsapp"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// WhatsAppController handles WhatsApp webhook endpoints
type WhatsAppController struct {
	messageProcessor *services.WhatsAppMessageProcessor
	authService      services.WhatsAppAuthService
	metricsService   *services.WhatsAppMetricsService
	client           whatsappclient.WhatsAppClient
	verifyToken      string
}

// NewWhatsAppController creates a new WhatsApp controller
func NewWhatsAppController(
	messageProcessor *services.WhatsAppMessageProcessor,
	authService services.WhatsAppAuthService,
	metricsService *services.WhatsAppMetricsService,
	client whatsappclient.WhatsAppClient,
	verifyToken string,
) *WhatsAppController {
	return &WhatsAppController{
		messageProcessor: messageProcessor,
		authService:      authService,
		metricsService:   metricsService,
		client:           client,
		verifyToken:      verifyToken,
	}
}

// VerifyWebhook handles webhook verification from WhatsApp (GET request)
func (ctrl *WhatsAppController) VerifyWebhook(c *gin.Context) {
	// Get query parameters
	mode := c.Query("hub.mode")
	token := c.Query("hub.verify_token")
	challenge := c.Query("hub.challenge")

	log.Info().
		Str("mode", mode).
		Str("token_received", token).
		Str("challenge", challenge).
		Msg("Webhook verification request received")

	// Verify the mode and token
	if mode == "subscribe" && token == ctrl.verifyToken {
		log.Info().Msg("Webhook verified successfully")
		c.String(http.StatusOK, challenge)
		return
	}

	log.Warn().
		Str("mode", mode).
		Str("expected_token", ctrl.verifyToken).
		Str("received_token", token).
		Msg("Webhook verification failed")

	c.JSON(http.StatusForbidden, gin.H{"error": "Verification failed"})
}

// HandleWebhook handles incoming webhook events from WhatsApp (POST request)
func (ctrl *WhatsAppController) HandleWebhook(c *gin.Context) {
	// Start timing webhook processing
	timer := ctrl.metricsService.WebhookTimer()
	defer timer.ObserveDuration()

	// Read the request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read webhook body")
		ctrl.metricsService.RecordWebhookRequest("error")
		ctrl.metricsService.RecordError("webhook", "read_body_failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Log the raw webhook payload for debugging
	log.Debug().
		Str("body", string(body)).
		Msg("Received webhook payload")

	// Parse the webhook payload
	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Error().Err(err).Msg("Failed to parse webhook payload")
		ctrl.metricsService.RecordWebhookRequest("error")
		ctrl.metricsService.RecordError("webhook", "parse_failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	// Extract phone number for rate limiting
	if len(payload.Entry) > 0 && len(payload.Entry[0].Changes) > 0 {
		change := payload.Entry[0].Changes[0]
		if len(change.Value.Messages) > 0 {
			phoneNumber := change.Value.Messages[0].From
			c.Set("whatsapp_phone_number", phoneNumber)
		}
	}

	// Record successful webhook receipt
	ctrl.metricsService.RecordWebhookRequest("success")

	// Acknowledge receipt immediately (WhatsApp requires 200 OK within 20 seconds)
	c.JSON(http.StatusOK, gin.H{"status": "received"})

	// Process messages asynchronously
	go ctrl.processWebhookPayload(&payload)
}

// processWebhookPayload processes the webhook payload asynchronously
func (ctrl *WhatsAppController) processWebhookPayload(payload *WebhookPayload) {
	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			// Handle message events
			if change.Value.Messages != nil {
				for _, message := range change.Value.Messages {
					ctrl.processIncomingMessage(&message, &change.Value)
				}
			}

			// Handle status updates for outbound messages
			if change.Value.Statuses != nil {
				for _, status := range change.Value.Statuses {
					ctrl.processStatusUpdate(&status)
				}
			}
		}
	}
}

// processIncomingMessage processes a single incoming message
func (ctrl *WhatsAppController) processIncomingMessage(message *WebhookMessage, value *WebhookValue) {
	// Only process text messages for now
	if message.Type != "text" || message.Text == nil {
		log.Info().
			Str("message_id", message.ID).
			Str("type", message.Type).
			Msg("Skipping non-text message")
		return
	}

	// Extract phone number (remove WhatsApp prefix if present)
	phoneNumber := message.From

	// Extract group ID if this is a group message
	var groupID *string
	if value.Metadata.DisplayPhoneNumber != message.From {
		// This might be a group message - check if we have group context
		// For now, we'll handle this in the message processor
	}

	// Convert timestamp string to int64
	timestampInt, err := strconv.ParseInt(message.Timestamp, 10, 64)
	if err != nil {
		log.Error().
			Err(err).
			Str("timestamp", message.Timestamp).
			Msg("Failed to parse message timestamp")
		timestampInt = time.Now().Unix()
	}

	// Build incoming message
	incomingMsg := &services.IncomingMessage{
		MessageID:   message.ID,
		PhoneNumber: phoneNumber,
		Content:     message.Text.Body,
		Timestamp:   time.Unix(timestampInt, 0),
		GroupID:     groupID,
	}

	// Process the message
	if err := ctrl.messageProcessor.ProcessMessage(incomingMsg); err != nil {
		log.Error().
			Err(err).
			Str("message_id", message.ID).
			Str("phone", phoneNumber).
			Msg("Failed to process incoming message")
	}
}

// processStatusUpdate processes a status update for an outbound message
func (ctrl *WhatsAppController) processStatusUpdate(status *WebhookStatus) {
	log.Info().
		Str("message_id", status.ID).
		Str("status", status.Status).
		Str("recipient", status.RecipientID).
		Msg("Received message status update")

	// Update message status in database
	var errorCode, errorMessage *string
	if len(status.Errors) > 0 {
		code := fmt.Sprintf("%d", status.Errors[0].Code)
		errorCode = &code
		errorMessage = &status.Errors[0].Title
	}

	if err := ctrl.messageProcessor.UpdateMessageStatus(status.ID, status.Status, errorCode, errorMessage); err != nil {
		log.Error().
			Err(err).
			Str("message_id", status.ID).
			Msg("Failed to update message status")
	}
}

// WebhookPayload represents the top-level webhook payload from WhatsApp
type WebhookPayload struct {
	Object string         `json:"object"`
	Entry  []WebhookEntry `json:"entry"`
}

// WebhookEntry represents an entry in the webhook payload
type WebhookEntry struct {
	ID      string          `json:"id"`
	Changes []WebhookChange `json:"changes"`
}

// WebhookChange represents a change in the webhook entry
type WebhookChange struct {
	Value WebhookValue `json:"value"`
	Field string       `json:"field"`
}

// WebhookValue contains the actual webhook data
type WebhookValue struct {
	MessagingProduct string           `json:"messaging_product"`
	Metadata         WebhookMetadata  `json:"metadata"`
	Contacts         []WebhookContact `json:"contacts,omitempty"`
	Messages         []WebhookMessage `json:"messages,omitempty"`
	Statuses         []WebhookStatus  `json:"statuses,omitempty"`
}

// WebhookMetadata contains metadata about the webhook
type WebhookMetadata struct {
	DisplayPhoneNumber string `json:"display_phone_number"`
	PhoneNumberID      string `json:"phone_number_id"`
}

// WebhookContact represents a contact in the webhook
type WebhookContact struct {
	Profile WebhookProfile `json:"profile"`
	WaID    string         `json:"wa_id"`
}

// WebhookProfile represents a contact profile
type WebhookProfile struct {
	Name string `json:"name"`
}

// WebhookMessage represents an incoming message
type WebhookMessage struct {
	From      string              `json:"from"`
	ID        string              `json:"id"`
	Timestamp string              `json:"timestamp"`
	Type      string              `json:"type"`
	Text      *WebhookMessageText `json:"text,omitempty"`
}

// WebhookMessageText represents text message content
type WebhookMessageText struct {
	Body string `json:"body"`
}

// WebhookStatus represents a message status update
type WebhookStatus struct {
	ID          string               `json:"id"`
	Status      string               `json:"status"`
	Timestamp   string               `json:"timestamp"`
	RecipientID string               `json:"recipient_id"`
	Errors      []WebhookStatusError `json:"errors,omitempty"`
}

// WebhookStatusError represents an error in a status update
type WebhookStatusError struct {
	Code  int    `json:"code"`
	Title string `json:"title"`
}

// Global WhatsApp controller instance
var globalWhatsAppController *WhatsAppController

// SetWhatsAppController sets the global WhatsApp controller instance
func SetWhatsAppController(controller *WhatsAppController) {
	globalWhatsAppController = controller
}

// VerifyWhatsAppWebhook is a global handler for webhook verification
func VerifyWhatsAppWebhook(c *gin.Context) {
	if globalWhatsAppController == nil {
		log.Error().Msg("WhatsApp controller not initialized")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Service not initialized"})
		return
	}
	globalWhatsAppController.VerifyWebhook(c)
}

// HandleWhatsAppWebhook is a global handler for webhook events
func HandleWhatsAppWebhook(c *gin.Context) {
	if globalWhatsAppController == nil {
		log.Error().Msg("WhatsApp controller not initialized")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Service not initialized"})
		return
	}
	globalWhatsAppController.HandleWebhook(c)
}

// LinkWhatsAppAccount handles the authentication callback to link a phone number to a user account
func LinkWhatsAppAccount(c *gin.Context) {
	if globalWhatsAppController == nil {
		log.Error().Msg("WhatsApp controller not initialized")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Service not initialized"})
		return
	}
	globalWhatsAppController.LinkAccount(c)
}

// LinkAccount handles linking a WhatsApp phone number to a user account
func (ctrl *WhatsAppController) LinkAccount(c *gin.Context) {
	// Get the link token from query parameter
	linkToken := c.Query("token")
	if linkToken == "" {
		log.Warn().Msg("Missing link token in authentication request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing authentication token"})
		return
	}

	// Validate the link token and extract phone number
	phoneNumber, err := ctrl.authService.ValidateLinkToken(linkToken)
	if err != nil {
		log.Error().Err(err).Msg("Invalid link token")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired authentication token"})
		return
	}

	// Get the authenticated user's Clerk ID from the context (set by Clerk middleware)
	clerkUserID, exists := c.Get("clerk_user_id")
	if !exists {
		log.Error().Msg("User ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	clerkUserIDStr, ok := clerkUserID.(string)
	if !ok {
		log.Error().Msg("Invalid user ID type in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Link the phone number to the user account
	user, err := ctrl.authService.LinkPhoneToUser(phoneNumber, clerkUserIDStr)
	if err != nil {
		log.Error().
			Err(err).
			Str("phone", phoneNumber).
			Str("clerk_user_id", clerkUserIDStr).
			Msg("Failed to link WhatsApp account")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link WhatsApp account"})
		return
	}

	log.Info().
		Str("phone", phoneNumber).
		Str("clerk_user_id", clerkUserIDStr).
		Msg("WhatsApp account linked successfully")

	// Send welcome message to the user
	go ctrl.sendWelcomeMessage(phoneNumber)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "WhatsApp account linked successfully",
		"user": gin.H{
			"id":          user.ID,
			"phoneNumber": user.PhoneNumber,
		},
	})
}

// sendWelcomeMessage sends a welcome message to a newly authenticated user
func (ctrl *WhatsAppController) sendWelcomeMessage(phoneNumber string) {
	message := "🎉 *Authentication Successful!*\n\n" +
		"Your WhatsApp account has been linked successfully.\n\n" +
		"You can now use the following commands:\n" +
		"• /add - Create a new note\n" +
		"• /retrieve - Search and retrieve notes\n" +
		"• /list - List your notes, notebooks, or chapters\n" +
		"• /delete - Delete a note\n" +
		"• /create - Create notebooks or chapters\n" +
		"• /help - Get detailed help on commands\n\n" +
		"_Start by typing /help to learn more!_"

	if err := ctrl.client.SendTextMessage(phoneNumber, message); err != nil {
		log.Error().
			Err(err).
			Str("phone", phoneNumber).
			Msg("Failed to send welcome message")
	} else {
		log.Info().
			Str("phone", phoneNumber).
			Msg("Welcome message sent successfully")
	}
}
