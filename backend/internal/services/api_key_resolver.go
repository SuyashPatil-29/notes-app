package services

import (
	"backend/db"
	"backend/internal/models"
	"backend/pkg/utils"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// APIKeySource represents the source of an API key
type APIKeySource string

const (
	APIKeySourceOrganization APIKeySource = "organization"
	APIKeySourceIndividual   APIKeySource = "individual"
)

// APIKeyResult contains the resolved API key and its source
type APIKeyResult struct {
	APIKey string
	Source APIKeySource
}

// APIKeyResolver interface defines methods for resolving API keys
type APIKeyResolver interface {
	GetAPIKey(clerkUserID string, organizationID *string, provider string) (*APIKeyResult, error)
	GetAPIKeySource(clerkUserID string, organizationID *string, provider string) (APIKeySource, error)
}

// apiKeyResolverImpl implements the APIKeyResolver interface
type apiKeyResolverImpl struct {
	db *gorm.DB
}

// NewAPIKeyResolver creates a new APIKeyResolver instance
func NewAPIKeyResolver() APIKeyResolver {
	return &apiKeyResolverImpl{
		db: db.DB,
	}
}

// GetAPIKey resolves an API key for a user and provider, prioritizing organization keys over individual keys
func (r *apiKeyResolverImpl) GetAPIKey(clerkUserID string, organizationID *string, provider string) (*APIKeyResult, error) {
	// First, try to get organization API key if organization context is provided
	if organizationID != nil && *organizationID != "" {
		orgKey, err := r.getOrganizationAPIKey(*organizationID, provider)
		if err == nil {
			log.Debug().
				Str("organizationId", *organizationID).
				Str("provider", provider).
				Msg("Using organization API key")
			return &APIKeyResult{
				APIKey: orgKey,
				Source: APIKeySourceOrganization,
			}, nil
		}

		// Log the error but continue to try individual key
		if err != gorm.ErrRecordNotFound {
			log.Warn().
				Err(err).
				Str("organizationId", *organizationID).
				Str("provider", provider).
				Msg("Failed to retrieve organization API key, falling back to individual key")
		}
	}

	// Fall back to individual API key
	individualKey, err := r.getIndividualAPIKey(clerkUserID, provider)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no API key configured for provider %s", provider)
		}
		return nil, fmt.Errorf("failed to retrieve API key: %w", err)
	}

	log.Debug().
		Str("clerkUserId", clerkUserID).
		Str("provider", provider).
		Msg("Using individual API key")

	return &APIKeyResult{
		APIKey: individualKey,
		Source: APIKeySourceIndividual,
	}, nil
}

// GetAPIKeySource determines the source of the API key without retrieving the actual key
func (r *apiKeyResolverImpl) GetAPIKeySource(clerkUserID string, organizationID *string, provider string) (APIKeySource, error) {
	// Check organization key first if organization context is provided
	if organizationID != nil && *organizationID != "" {
		var orgCredential models.OrganizationAPICredential
		err := r.db.Where("organization_id = ? AND provider = ?", *organizationID, provider).
			First(&orgCredential).Error
		if err == nil {
			return APIKeySourceOrganization, nil
		}

		if err != gorm.ErrRecordNotFound {
			log.Warn().
				Err(err).
				Str("organizationId", *organizationID).
				Str("provider", provider).
				Msg("Error checking organization API key")
		}
	}

	// Check individual key
	var individualCredential models.AICredential
	err := r.db.Where("clerk_user_id = ? AND provider = ?", clerkUserID, provider).
		First(&individualCredential).Error
	if err == nil {
		return APIKeySourceIndividual, nil
	}

	if err == gorm.ErrRecordNotFound {
		return "", fmt.Errorf("no API key configured for provider %s", provider)
	}

	return "", fmt.Errorf("failed to check API key source: %w", err)
}

// getOrganizationAPIKey retrieves and decrypts an organization API key
func (r *apiKeyResolverImpl) getOrganizationAPIKey(organizationID string, provider string) (string, error) {
	var credential models.OrganizationAPICredential
	err := r.db.Where("organization_id = ? AND provider = ?", organizationID, provider).
		First(&credential).Error
	if err != nil {
		return "", err
	}

	// Decrypt the API key
	apiKey, err := utils.Decrypt(credential.KeyCipher)
	if err != nil {
		log.Error().
			Err(err).
			Str("organizationId", organizationID).
			Str("provider", provider).
			Msg("Failed to decrypt organization API key")
		return "", fmt.Errorf("failed to decrypt organization API key: %w", err)
	}

	// Trim whitespace
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return "", fmt.Errorf("decrypted organization API key is empty")
	}

	return apiKey, nil
}

// getIndividualAPIKey retrieves and decrypts an individual user API key
func (r *apiKeyResolverImpl) getIndividualAPIKey(clerkUserID string, provider string) (string, error) {
	var credential models.AICredential
	err := r.db.Where("clerk_user_id = ? AND provider = ?", clerkUserID, provider).
		First(&credential).Error
	if err != nil {
		return "", err
	}

	// Decrypt the API key
	apiKey, err := utils.Decrypt(credential.KeyCipher)
	if err != nil {
		log.Error().
			Err(err).
			Str("clerkUserId", clerkUserID).
			Str("provider", provider).
			Msg("Failed to decrypt individual API key")
		return "", fmt.Errorf("failed to decrypt individual API key: %w", err)
	}

	// Trim whitespace
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return "", fmt.Errorf("decrypted individual API key is empty")
	}

	return apiKey, nil
}
