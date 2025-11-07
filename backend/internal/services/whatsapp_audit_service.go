package services

import (
	"backend/internal/models"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// WhatsAppAuditService handles audit logging for WhatsApp messages
type WhatsAppAuditService struct {
	db *gorm.DB
}

// NewWhatsAppAuditService creates a new audit service
func NewWhatsAppAuditService(db *gorm.DB) *WhatsAppAuditService {
	return &WhatsAppAuditService{
		db: db,
	}
}

// LogInboundMessage logs an inbound message to the audit table
func (s *WhatsAppAuditService) LogInboundMessage(messageID, phoneNumber, messageType, content string, timestamp time.Time) error {
	auditMsg := models.WhatsAppMessage{
		MessageID:   messageID,
		PhoneNumber: phoneNumber,
		Direction:   "inbound",
		MessageType: messageType,
		Content:     content,
		Status:      "received",
		CreatedAt:   timestamp,
	}

	if err := s.db.Create(&auditMsg).Error; err != nil {
		log.Error().
			Err(err).
			Str("message_id", messageID).
			Str("phone", phoneNumber).
			Msg("Failed to log inbound message")
		return err
	}

	log.Debug().
		Str("message_id", messageID).
		Str("phone", phoneNumber).
		Str("type", messageType).
		Msg("Logged inbound message")

	return nil
}

// LogOutboundMessage logs an outbound message to the audit table
func (s *WhatsAppAuditService) LogOutboundMessage(messageID, phoneNumber, messageType, content, status string) error {
	// Check if message already exists (to avoid duplicate key errors)
	var existing models.WhatsAppMessage
	err := s.db.Where("message_id = ?", messageID).First(&existing).Error

	if err == nil {
		// Message already exists, update it instead
		result := s.db.Model(&existing).Updates(map[string]interface{}{
			"status":       status,
			"content":      content,
			"message_type": messageType,
		})
		if result.Error != nil {
			log.Warn().
				Err(result.Error).
				Str("message_id", messageID).
				Str("phone", phoneNumber).
				Msg("Failed to update existing outbound message")
		}
		return nil
	}

	// Message doesn't exist, create it
	auditMsg := models.WhatsAppMessage{
		MessageID:   messageID,
		PhoneNumber: phoneNumber,
		Direction:   "outbound",
		MessageType: messageType,
		Content:     content,
		Status:      status,
		CreatedAt:   time.Now(),
	}

	if err := s.db.Create(&auditMsg).Error; err != nil {
		log.Error().
			Err(err).
			Str("message_id", messageID).
			Str("phone", phoneNumber).
			Msg("Failed to log outbound message")
		return err
	}

	log.Debug().
		Str("message_id", messageID).
		Str("phone", phoneNumber).
		Str("type", messageType).
		Str("status", status).
		Msg("Logged outbound message")

	return nil
}

// LogMessageError logs an error that occurred during message processing
func (s *WhatsAppAuditService) LogMessageError(messageID, phoneNumber, errorCode, errorMessage string) error {
	// Try to update existing message first
	result := s.db.Model(&models.WhatsAppMessage{}).
		Where("message_id = ?", messageID).
		Updates(map[string]interface{}{
			"status":        "failed",
			"error_code":    errorCode,
			"error_message": errorMessage,
		})

	if result.Error != nil {
		log.Error().
			Err(result.Error).
			Str("message_id", messageID).
			Msg("Failed to update message with error")
		return result.Error
	}

	// If no existing message was found, create a new error log entry
	if result.RowsAffected == 0 {
		errCode := errorCode
		errMsg := errorMessage
		auditMsg := models.WhatsAppMessage{
			MessageID:    messageID,
			PhoneNumber:  phoneNumber,
			Direction:    "outbound",
			MessageType:  "text",
			Content:      "",
			Status:       "failed",
			ErrorCode:    &errCode,
			ErrorMessage: &errMsg,
			CreatedAt:    time.Now(),
		}

		if err := s.db.Create(&auditMsg).Error; err != nil {
			log.Error().
				Err(err).
				Str("message_id", messageID).
				Msg("Failed to create error log entry")
			return err
		}
	}

	log.Warn().
		Str("message_id", messageID).
		Str("phone", phoneNumber).
		Str("error_code", errorCode).
		Str("error_message", errorMessage).
		Msg("Logged message error")

	return nil
}

// UpdateMessageStatus updates the status of a message
func (s *WhatsAppAuditService) UpdateMessageStatus(messageID, status string, errorCode, errorMessage *string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if errorCode != nil {
		updates["error_code"] = *errorCode
	}
	if errorMessage != nil {
		updates["error_message"] = *errorMessage
	}

	err := s.db.Model(&models.WhatsAppMessage{}).
		Where("message_id = ?", messageID).
		Updates(updates).Error

	if err != nil {
		log.Error().
			Err(err).
			Str("message_id", messageID).
			Str("status", status).
			Msg("Failed to update message status")
		return err
	}

	log.Debug().
		Str("message_id", messageID).
		Str("status", status).
		Msg("Updated message status")

	return nil
}

// CleanupOldMessages removes audit logs older than the specified retention period
func (s *WhatsAppAuditService) CleanupOldMessages(retentionDays int) (int64, error) {
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	result := s.db.Where("created_at < ?", cutoffDate).Delete(&models.WhatsAppMessage{})
	if result.Error != nil {
		log.Error().
			Err(result.Error).
			Int("retention_days", retentionDays).
			Msg("Failed to cleanup old messages")
		return 0, result.Error
	}

	log.Info().
		Int64("deleted_count", result.RowsAffected).
		Int("retention_days", retentionDays).
		Time("cutoff_date", cutoffDate).
		Msg("Cleaned up old audit messages")

	return result.RowsAffected, nil
}

// GetMessageStats returns statistics about message audit logs
func (s *WhatsAppAuditService) GetMessageStats(since time.Time) (*MessageStats, error) {
	var stats MessageStats

	// Count total messages
	if err := s.db.Model(&models.WhatsAppMessage{}).
		Where("created_at >= ?", since).
		Count(&stats.TotalMessages).Error; err != nil {
		return nil, err
	}

	// Count inbound messages
	if err := s.db.Model(&models.WhatsAppMessage{}).
		Where("created_at >= ? AND direction = ?", since, "inbound").
		Count(&stats.InboundMessages).Error; err != nil {
		return nil, err
	}

	// Count outbound messages
	if err := s.db.Model(&models.WhatsAppMessage{}).
		Where("created_at >= ? AND direction = ?", since, "outbound").
		Count(&stats.OutboundMessages).Error; err != nil {
		return nil, err
	}

	// Count failed messages
	if err := s.db.Model(&models.WhatsAppMessage{}).
		Where("created_at >= ? AND status = ?", since, "failed").
		Count(&stats.FailedMessages).Error; err != nil {
		return nil, err
	}

	// Count delivered messages
	if err := s.db.Model(&models.WhatsAppMessage{}).
		Where("created_at >= ? AND status = ?", since, "delivered").
		Count(&stats.DeliveredMessages).Error; err != nil {
		return nil, err
	}

	// Count read messages
	if err := s.db.Model(&models.WhatsAppMessage{}).
		Where("created_at >= ? AND status = ?", since, "read").
		Count(&stats.ReadMessages).Error; err != nil {
		return nil, err
	}

	return &stats, nil
}

// MessageStats represents statistics about message audit logs
type MessageStats struct {
	TotalMessages     int64 `json:"totalMessages"`
	InboundMessages   int64 `json:"inboundMessages"`
	OutboundMessages  int64 `json:"outboundMessages"`
	FailedMessages    int64 `json:"failedMessages"`
	DeliveredMessages int64 `json:"deliveredMessages"`
	ReadMessages      int64 `json:"readMessages"`
}
