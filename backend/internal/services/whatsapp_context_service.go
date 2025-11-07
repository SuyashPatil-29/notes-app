package services

import (
	"backend/internal/config"
	"backend/internal/models"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// WhatsAppContextService manages conversation context and state for WhatsApp interactions
type WhatsAppContextService struct {
	db     *gorm.DB
	config *config.WhatsAppConfig
}

// NewWhatsAppContextService creates a new WhatsApp context service instance
func NewWhatsAppContextService(db *gorm.DB, config *config.WhatsAppConfig) *WhatsAppContextService {
	return &WhatsAppContextService{
		db:     db,
		config: config,
	}
}

// GetContext retrieves the active conversation context for a phone number
func (s *WhatsAppContextService) GetContext(phoneNumber string) (*models.WhatsAppConversationContext, error) {
	var context models.WhatsAppConversationContext

	err := s.db.Where("phone_number = ?", phoneNumber).
		Order("created_at DESC").
		First(&context).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No context found, not an error
		}
		log.Error().
			Err(err).
			Str("phone_number", phoneNumber).
			Msg("Failed to retrieve conversation context")
		return nil, fmt.Errorf("failed to retrieve context: %w", err)
	}

	// Check if context is expired
	if s.IsContextExpired(&context) {
		log.Info().
			Str("phone_number", phoneNumber).
			Str("command", context.Command).
			Msg("Context expired, clearing")

		// Clear expired context
		if err := s.ClearContext(phoneNumber); err != nil {
			log.Warn().Err(err).Msg("Failed to clear expired context")
		}
		return nil, nil
	}

	return &context, nil
}

// SetContext creates a new conversation context for a phone number
func (s *WhatsAppContextService) SetContext(phoneNumber, command string, data map[string]interface{}) error {
	// Clear any existing context first
	if err := s.ClearContext(phoneNumber); err != nil {
		log.Warn().
			Err(err).
			Str("phone_number", phoneNumber).
			Msg("Failed to clear existing context before setting new one")
	}

	// Marshal data to JSON string
	dataJSON, err := json.Marshal(data)
	if err != nil {
		log.Error().
			Err(err).
			Str("phone_number", phoneNumber).
			Msg("Failed to marshal context data")
		return fmt.Errorf("failed to marshal context data: %w", err)
	}

	// Create new context
	context := models.WhatsAppConversationContext{
		PhoneNumber: phoneNumber,
		Command:     command,
		Step:        0,
		Data:        string(dataJSON),
		ExpiresAt:   time.Now().Add(s.config.ContextExpiration),
	}

	if err := s.db.Create(&context).Error; err != nil {
		log.Error().
			Err(err).
			Str("phone_number", phoneNumber).
			Str("command", command).
			Msg("Failed to create conversation context")
		return fmt.Errorf("failed to create context: %w", err)
	}

	log.Info().
		Str("phone_number", phoneNumber).
		Str("command", command).
		Time("expires_at", context.ExpiresAt).
		Msg("Created new conversation context")

	return nil
}

// UpdateContext modifies an existing conversation context
func (s *WhatsAppContextService) UpdateContext(phoneNumber string, data map[string]interface{}) error {
	// Get existing context
	context, err := s.GetContext(phoneNumber)
	if err != nil {
		return err
	}

	if context == nil {
		return fmt.Errorf("no active context found for phone number: %s", phoneNumber)
	}

	// Parse existing data
	var existingData map[string]interface{}
	if context.Data != "" {
		if err := json.Unmarshal([]byte(context.Data), &existingData); err != nil {
			log.Error().
				Err(err).
				Str("phone_number", phoneNumber).
				Msg("Failed to unmarshal existing context data")
			return fmt.Errorf("failed to unmarshal existing context data: %w", err)
		}
	} else {
		existingData = make(map[string]interface{})
	}

	// Merge new data with existing data
	for key, value := range data {
		existingData[key] = value
	}

	// Marshal updated data
	dataJSON, err := json.Marshal(existingData)
	if err != nil {
		log.Error().
			Err(err).
			Str("phone_number", phoneNumber).
			Msg("Failed to marshal updated context data")
		return fmt.Errorf("failed to marshal updated context data: %w", err)
	}

	// Update context
	updates := map[string]interface{}{
		"data":       string(dataJSON),
		"step":       context.Step + 1,
		"updated_at": time.Now(),
	}

	if err := s.db.Model(&models.WhatsAppConversationContext{}).
		Where("phone_number = ?", phoneNumber).
		Updates(updates).Error; err != nil {
		log.Error().
			Err(err).
			Str("phone_number", phoneNumber).
			Msg("Failed to update conversation context")
		return fmt.Errorf("failed to update context: %w", err)
	}

	log.Info().
		Str("phone_number", phoneNumber).
		Str("command", context.Command).
		Int("step", context.Step+1).
		Msg("Updated conversation context")

	return nil
}

// ClearContext removes the conversation context for a phone number
func (s *WhatsAppContextService) ClearContext(phoneNumber string) error {
	result := s.db.Where("phone_number = ?", phoneNumber).
		Delete(&models.WhatsAppConversationContext{})

	if result.Error != nil {
		log.Error().
			Err(result.Error).
			Str("phone_number", phoneNumber).
			Msg("Failed to clear conversation context")
		return fmt.Errorf("failed to clear context: %w", result.Error)
	}

	if result.RowsAffected > 0 {
		log.Info().
			Str("phone_number", phoneNumber).
			Int64("rows_affected", result.RowsAffected).
			Msg("Cleared conversation context")
	}

	return nil
}

// IsContextExpired checks if a conversation context has expired
func (s *WhatsAppContextService) IsContextExpired(context *models.WhatsAppConversationContext) bool {
	if context == nil {
		return true
	}
	return time.Now().After(context.ExpiresAt)
}

// CleanupExpiredContexts removes all expired conversation contexts
// This should be called periodically (e.g., via a cron job)
func (s *WhatsAppContextService) CleanupExpiredContexts() error {
	result := s.db.Where("expires_at < ?", time.Now()).
		Delete(&models.WhatsAppConversationContext{})

	if result.Error != nil {
		log.Error().
			Err(result.Error).
			Msg("Failed to cleanup expired contexts")
		return fmt.Errorf("failed to cleanup expired contexts: %w", result.Error)
	}

	if result.RowsAffected > 0 {
		log.Info().
			Int64("rows_affected", result.RowsAffected).
			Msg("Cleaned up expired conversation contexts")
	}

	return nil
}

// GetContextData retrieves and unmarshals the context data into a map
func (s *WhatsAppContextService) GetContextData(phoneNumber string) (map[string]interface{}, error) {
	context, err := s.GetContext(phoneNumber)
	if err != nil {
		return nil, err
	}

	if context == nil {
		return nil, nil
	}

	var data map[string]interface{}
	if context.Data != "" {
		if err := json.Unmarshal([]byte(context.Data), &data); err != nil {
			log.Error().
				Err(err).
				Str("phone_number", phoneNumber).
				Msg("Failed to unmarshal context data")
			return nil, fmt.Errorf("failed to unmarshal context data: %w", err)
		}
	}

	return data, nil
}
