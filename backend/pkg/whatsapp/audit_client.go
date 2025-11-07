package whatsapp

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

// AuditLogger interface for logging WhatsApp messages
type AuditLogger interface {
	LogOutboundMessage(messageID, phoneNumber, messageType, content, status string) error
	LogMessageError(messageID, phoneNumber, errorCode, errorMessage string) error
}

// MetricsRecorder interface for recording metrics
type MetricsRecorder interface {
	RecordOutboundMessage(messageType, status string)
	RecordMessageFailed(direction, errorType string)
}

// AuditClient wraps a WhatsAppClient and adds audit logging
type AuditClient struct {
	client  WhatsAppClient
	logger  AuditLogger
	metrics MetricsRecorder
}

// NewAuditClient creates a new audit client wrapper
func NewAuditClient(client WhatsAppClient, logger AuditLogger, metrics MetricsRecorder) WhatsAppClient {
	return &AuditClient{
		client:  client,
		logger:  logger,
		metrics: metrics,
	}
}

// SendTextMessage sends a text message and logs it
func (a *AuditClient) SendTextMessage(phoneNumber, message string) error {
	// Generate a unique message ID for tracking
	messageID := generateMessageID()

	// Log the outbound message before sending
	if err := a.logger.LogOutboundMessage(messageID, phoneNumber, "text", message, "sending"); err != nil {
		log.Warn().Err(err).Msg("Failed to log outbound message, continuing with send")
	}

	// Send the message
	err := a.client.SendTextMessage(phoneNumber, message)
	if err != nil {
		// Log the error
		if logErr := a.logger.LogMessageError(messageID, phoneNumber, "send_failed", err.Error()); logErr != nil {
			log.Error().Err(logErr).Msg("Failed to log message error")
		}

		// Record metrics
		if a.metrics != nil {
			a.metrics.RecordMessageFailed("outbound", "send_failed")
		}

		return err
	}

	// Update status to sent
	if err := a.logger.LogOutboundMessage(messageID, phoneNumber, "text", message, "sent"); err != nil {
		log.Warn().Err(err).Msg("Failed to update message status to sent")
	}

	// Record metrics
	if a.metrics != nil {
		a.metrics.RecordOutboundMessage("text", "sent")
	}

	return nil
}

// SendInteractiveMessage sends an interactive message and logs it
func (a *AuditClient) SendInteractiveMessage(phoneNumber string, message InteractiveMessage) error {
	// Generate a unique message ID for tracking
	messageID := generateMessageID()

	// Serialize the interactive message for logging
	content, err := json.Marshal(message)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to serialize interactive message for logging")
		content = []byte(fmt.Sprintf("Interactive message type: %s", message.Type))
	}

	// Log the outbound message before sending
	if err := a.logger.LogOutboundMessage(messageID, phoneNumber, "interactive", string(content), "sending"); err != nil {
		log.Warn().Err(err).Msg("Failed to log outbound interactive message, continuing with send")
	}

	// Send the message
	err = a.client.SendInteractiveMessage(phoneNumber, message)
	if err != nil {
		// Log the error
		if logErr := a.logger.LogMessageError(messageID, phoneNumber, "send_failed", err.Error()); logErr != nil {
			log.Error().Err(logErr).Msg("Failed to log message error")
		}

		// Record metrics
		if a.metrics != nil {
			a.metrics.RecordMessageFailed("outbound", "send_failed")
		}

		return err
	}

	// Update status to sent
	if err := a.logger.LogOutboundMessage(messageID, phoneNumber, "interactive", string(content), "sent"); err != nil {
		log.Warn().Err(err).Msg("Failed to update message status to sent")
	}

	// Record metrics
	if a.metrics != nil {
		a.metrics.RecordOutboundMessage("interactive", "sent")
	}

	return nil
}

// VerifyWebhookSignature delegates to the underlying client
func (a *AuditClient) VerifyWebhookSignature(payload []byte, signature string) bool {
	return a.client.VerifyWebhookSignature(payload, signature)
}

// GetPhoneNumberID delegates to the underlying client
func (a *AuditClient) GetPhoneNumberID() string {
	return a.client.GetPhoneNumberID()
}

// generateMessageID generates a unique message ID for tracking
func generateMessageID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if random generation fails
		return fmt.Sprintf("msg_%d", time.Now().UnixNano())
	}
	return "msg_" + hex.EncodeToString(bytes)
}
