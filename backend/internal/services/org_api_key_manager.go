package services

import (
	"backend/db"
	"backend/internal/models"
	"backend/pkg/utils"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// OrgAPIKeyStatus represents the status of an organization API key
type OrgAPIKeyStatus struct {
	Provider  string    `json:"provider"`
	HasKey    bool      `json:"hasKey"`
	MaskedKey *string   `json:"maskedKey,omitempty"`
	CreatedBy string    `json:"createdBy"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// OrgAPIKeyManager interface defines methods for managing organization API keys
type OrgAPIKeyManager interface {
	GetOrgAPICredentials(organizationID string) ([]OrgAPIKeyStatus, error)
	SetOrgAPICredential(organizationID string, provider string, apiKey string, createdBy string) error
	DeleteOrgAPICredential(organizationID string, provider string) error
	HasOrgAPICredential(organizationID string, provider string) (bool, error)
}

// orgAPIKeyManagerImpl implements the OrgAPIKeyManager interface
type orgAPIKeyManagerImpl struct {
	db *gorm.DB
}

// NewOrgAPIKeyManager creates a new OrgAPIKeyManager instance
func NewOrgAPIKeyManager() OrgAPIKeyManager {
	return &orgAPIKeyManagerImpl{
		db: db.DB,
	}
}

// GetOrgAPICredentials returns the status of all API credentials for an organization
func (m *orgAPIKeyManagerImpl) GetOrgAPICredentials(organizationID string) ([]OrgAPIKeyStatus, error) {
	var credentials []models.OrganizationAPICredential
	err := m.db.Where("organization_id = ?", organizationID).Find(&credentials).Error
	if err != nil {
		log.Error().
			Err(err).
			Str("organizationId", organizationID).
			Msg("Failed to fetch organization API credentials")
		return nil, fmt.Errorf("failed to fetch organization API credentials: %w", err)
	}

	// Convert to status objects
	statuses := make([]OrgAPIKeyStatus, len(credentials))
	for i, cred := range credentials {
		// Decrypt and mask the API key for display
		var maskedKey *string
		if len(cred.KeyCipher) > 0 {
			apiKey, err := utils.Decrypt(cred.KeyCipher)
			if err != nil {
				log.Warn().
					Err(err).
					Str("organizationId", organizationID).
					Str("provider", cred.Provider).
					Msg("Failed to decrypt API key for masking")
			} else {
				masked := utils.MaskAPIKey(strings.TrimSpace(apiKey))
				maskedKey = &masked
			}
		}

		statuses[i] = OrgAPIKeyStatus{
			Provider:  cred.Provider,
			HasKey:    len(cred.KeyCipher) > 0,
			MaskedKey: maskedKey,
			CreatedBy: cred.CreatedBy,
			CreatedAt: cred.CreatedAt,
			UpdatedAt: cred.UpdatedAt,
		}
	}

	log.Info().
		Str("organizationId", organizationID).
		Int("credentialCount", len(statuses)).
		Msg("Retrieved organization API credentials")

	return statuses, nil
}

// SetOrgAPICredential creates or updates an organization API credential
func (m *orgAPIKeyManagerImpl) SetOrgAPICredential(organizationID string, provider string, apiKey string, createdBy string) error {
	// Validate inputs
	if strings.TrimSpace(organizationID) == "" {
		return fmt.Errorf("organization ID cannot be empty")
	}
	if strings.TrimSpace(provider) == "" {
		return fmt.Errorf("provider cannot be empty")
	}
	if strings.TrimSpace(apiKey) == "" {
		return fmt.Errorf("API key cannot be empty")
	}
	if strings.TrimSpace(createdBy) == "" {
		return fmt.Errorf("createdBy cannot be empty")
	}

	// Trim whitespace from API key before encrypting
	apiKey = strings.TrimSpace(apiKey)

	// Encrypt the API key
	encryptedKey, err := utils.Encrypt(apiKey)
	if err != nil {
		log.Error().
			Err(err).
			Str("organizationId", organizationID).
			Str("provider", provider).
			Msg("Failed to encrypt organization API key")
		return fmt.Errorf("failed to encrypt API key: %w", err)
	}

	// Create or update the credential using GORM's Upsert functionality
	credential := models.OrganizationAPICredential{
		OrganizationID: organizationID,
		Provider:       provider,
		KeyCipher:      encryptedKey,
		CreatedBy:      createdBy,
	}

	// Use GORM's Clauses for upsert (ON CONFLICT DO UPDATE)
	err = m.db.Where(models.OrganizationAPICredential{
		OrganizationID: organizationID,
		Provider:       provider,
	}).Assign(models.OrganizationAPICredential{
		KeyCipher: encryptedKey,
		CreatedBy: createdBy,
	}).FirstOrCreate(&credential).Error

	if err != nil {
		log.Error().
			Err(err).
			Str("organizationId", organizationID).
			Str("provider", provider).
			Str("createdBy", createdBy).
			Msg("Failed to save organization API credential")
		return fmt.Errorf("failed to save organization API credential: %w", err)
	}

	// Log the operation for audit purposes
	log.Info().
		Str("organizationId", organizationID).
		Str("provider", provider).
		Str("createdBy", createdBy).
		Uint("credentialId", credential.ID).
		Msg("Organization API credential saved successfully")

	return nil
}

// DeleteOrgAPICredential removes an organization API credential
func (m *orgAPIKeyManagerImpl) DeleteOrgAPICredential(organizationID string, provider string) error {
	// Validate inputs
	if strings.TrimSpace(organizationID) == "" {
		return fmt.Errorf("organization ID cannot be empty")
	}
	if strings.TrimSpace(provider) == "" {
		return fmt.Errorf("provider cannot be empty")
	}

	// Check if credential exists before deletion for better error handling
	var credential models.OrganizationAPICredential
	err := m.db.Where("organization_id = ? AND provider = ?", organizationID, provider).
		First(&credential).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("organization API credential not found for provider %s", provider)
		}
		log.Error().
			Err(err).
			Str("organizationId", organizationID).
			Str("provider", provider).
			Msg("Failed to check organization API credential existence")
		return fmt.Errorf("failed to check organization API credential: %w", err)
	}

	// Delete the credential
	err = m.db.Where("organization_id = ? AND provider = ?", organizationID, provider).
		Delete(&models.OrganizationAPICredential{}).Error
	if err != nil {
		log.Error().
			Err(err).
			Str("organizationId", organizationID).
			Str("provider", provider).
			Msg("Failed to delete organization API credential")
		return fmt.Errorf("failed to delete organization API credential: %w", err)
	}

	// Log the operation for audit purposes
	log.Info().
		Str("organizationId", organizationID).
		Str("provider", provider).
		Str("deletedBy", credential.CreatedBy). // Log who originally created it
		Uint("credentialId", credential.ID).
		Msg("Organization API credential deleted successfully")

	return nil
}

// HasOrgAPICredential checks if an organization has an API credential for a specific provider
func (m *orgAPIKeyManagerImpl) HasOrgAPICredential(organizationID string, provider string) (bool, error) {
	// Validate inputs
	if strings.TrimSpace(organizationID) == "" {
		return false, fmt.Errorf("organization ID cannot be empty")
	}
	if strings.TrimSpace(provider) == "" {
		return false, fmt.Errorf("provider cannot be empty")
	}

	var count int64
	err := m.db.Model(&models.OrganizationAPICredential{}).
		Where("organization_id = ? AND provider = ?", organizationID, provider).
		Count(&count).Error
	if err != nil {
		log.Error().
			Err(err).
			Str("organizationId", organizationID).
			Str("provider", provider).
			Msg("Failed to check organization API credential existence")
		return false, fmt.Errorf("failed to check organization API credential: %w", err)
	}

	return count > 0, nil
}
