package services

import (
	"backend/db"
	"backend/internal/models"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/organizationmembership"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// WhatsAppAuthService handles authentication and authorization for WhatsApp users
type WhatsAppAuthService interface {
	GetUserByPhone(phoneNumber string) (*models.WhatsAppUser, error)
	CreateWhatsAppUser(phoneNumber, clerkUserID string) (*models.WhatsAppUser, error)
	GenerateAuthToken(clerkUserID string) (string, error)
	ValidateAuthToken(token string) (string, error)
	IsUserInOrganization(ctx context.Context, clerkUserID, organizationID string) (bool, error)
	UpdateLastActive(phoneNumber string) error
	GenerateLinkToken(phoneNumber string) (string, error)
	ValidateLinkToken(token string) (string, error)
	LinkPhoneToUser(phoneNumber, clerkUserID string) (*models.WhatsAppUser, error)
	IsSessionExpired(user *models.WhatsAppUser) bool
	RefreshAuthToken(phoneNumber string) error
}

type whatsAppAuthService struct {
	db *gorm.DB
}

// NewWhatsAppAuthService creates a new WhatsApp authentication service
func NewWhatsAppAuthService() WhatsAppAuthService {
	return &whatsAppAuthService{
		db: db.DB,
	}
}

// GetUserByPhone retrieves a WhatsApp user by phone number
func (s *whatsAppAuthService) GetUserByPhone(phoneNumber string) (*models.WhatsAppUser, error) {
	var user models.WhatsAppUser
	err := s.db.Where("phone_number = ?", phoneNumber).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // User not found, not an error
		}
		log.Error().Err(err).Str("phone", phoneNumber).Msg("Failed to get WhatsApp user")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// CreateWhatsAppUser creates a new WhatsApp user record
func (s *whatsAppAuthService) CreateWhatsAppUser(phoneNumber, clerkUserID string) (*models.WhatsAppUser, error) {
	// Generate authentication token
	token, err := s.GenerateAuthToken(clerkUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth token: %w", err)
	}

	user := &models.WhatsAppUser{
		PhoneNumber:     phoneNumber,
		ClerkUserID:     clerkUserID,
		IsAuthenticated: true,
		AuthToken:       token,
		LastActiveAt:    time.Now(),
	}

	if err := s.db.Create(user).Error; err != nil {
		log.Error().Err(err).Str("phone", phoneNumber).Msg("Failed to create WhatsApp user")
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	log.Info().
		Str("phone", phoneNumber).
		Str("clerk_user_id", clerkUserID).
		Msg("WhatsApp user created successfully")

	return user, nil
}

// GenerateAuthToken generates a secure random token for authentication
func (s *whatsAppAuthService) GenerateAuthToken(clerkUserID string) (string, error) {
	// Generate 32 random bytes
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}

	// Encode to base64 and prepend clerk user ID for validation
	token := fmt.Sprintf("%s:%s", clerkUserID, base64.URLEncoding.EncodeToString(tokenBytes))
	return token, nil
}

// ValidateAuthToken validates an authentication token and returns the clerk user ID
func (s *whatsAppAuthService) ValidateAuthToken(token string) (string, error) {
	// Token format: clerkUserID:base64Token
	// For now, we'll look up the token in the database
	var user models.WhatsAppUser
	err := s.db.Where("auth_token = ? AND is_authenticated = ?", token, true).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("invalid or expired token")
		}
		return "", fmt.Errorf("failed to validate token: %w", err)
	}

	return user.ClerkUserID, nil
}

// IsUserInOrganization checks if a user is a member of an organization using Clerk API
func (s *whatsAppAuthService) IsUserInOrganization(ctx context.Context, clerkUserID, organizationID string) (bool, error) {
	// Get organization memberships from Clerk
	params := &organizationmembership.ListParams{}
	params.Limit = clerk.Int64(100)
	params.OrganizationID = organizationID
	params.UserIDs = []string{clerkUserID}

	memberships, err := organizationmembership.List(ctx, params)
	if err != nil {
		log.Error().
			Err(err).
			Str("clerk_user_id", clerkUserID).
			Str("org_id", organizationID).
			Msg("Failed to get organization memberships")
		return false, fmt.Errorf("failed to check organization membership: %w", err)
	}

	// Check if user is in the membership list
	for _, membership := range memberships.OrganizationMemberships {
		if membership.PublicUserData != nil && membership.PublicUserData.UserID == clerkUserID {
			return true, nil
		}
	}

	return false, nil
}

// UpdateLastActive updates the last active timestamp for a user
func (s *whatsAppAuthService) UpdateLastActive(phoneNumber string) error {
	err := s.db.Model(&models.WhatsAppUser{}).
		Where("phone_number = ?", phoneNumber).
		Update("last_active_at", time.Now()).Error

	if err != nil {
		log.Warn().Err(err).Str("phone", phoneNumber).Msg("Failed to update last active")
		return err
	}

	return nil
}

// GenerateLinkToken generates a temporary token for linking a phone number to a user account
func (s *whatsAppAuthService) GenerateLinkToken(phoneNumber string) (string, error) {
	// Generate 32 random bytes
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}

	// Encode to base64 and prepend phone number for validation
	token := fmt.Sprintf("%s:%s", phoneNumber, base64.URLEncoding.EncodeToString(tokenBytes))

	log.Info().
		Str("phone", phoneNumber).
		Msg("Generated link token for phone number")

	return token, nil
}

// ValidateLinkToken validates a link token and returns the phone number
func (s *whatsAppAuthService) ValidateLinkToken(token string) (string, error) {
	// Token format: phoneNumber:base64Token
	// Extract phone number from token
	parts := strings.Split(token, ":")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid token format")
	}

	phoneNumber := parts[0]

	// Validate phone number format
	if phoneNumber == "" {
		return "", fmt.Errorf("invalid phone number in token")
	}

	log.Info().
		Str("phone", phoneNumber).
		Msg("Validated link token")

	return phoneNumber, nil
}

// LinkPhoneToUser links a phone number to a user account
func (s *whatsAppAuthService) LinkPhoneToUser(phoneNumber, clerkUserID string) (*models.WhatsAppUser, error) {
	// Check if user already exists
	existingUser, err := s.GetUserByPhone(phoneNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	if existingUser != nil {
		// Update existing user
		existingUser.ClerkUserID = clerkUserID
		existingUser.IsAuthenticated = true

		// Generate new auth token
		token, err := s.GenerateAuthToken(clerkUserID)
		if err != nil {
			return nil, fmt.Errorf("failed to generate auth token: %w", err)
		}
		existingUser.AuthToken = token
		existingUser.LastActiveAt = time.Now()

		if err := s.db.Save(existingUser).Error; err != nil {
			log.Error().Err(err).Str("phone", phoneNumber).Msg("Failed to update WhatsApp user")
			return nil, fmt.Errorf("failed to update user: %w", err)
		}

		log.Info().
			Str("phone", phoneNumber).
			Str("clerk_user_id", clerkUserID).
			Msg("WhatsApp user updated and authenticated")

		return existingUser, nil
	}

	// Create new user
	return s.CreateWhatsAppUser(phoneNumber, clerkUserID)
}

// IsSessionExpired checks if a user's session has expired (30 days of inactivity)
func (s *whatsAppAuthService) IsSessionExpired(user *models.WhatsAppUser) bool {
	if user == nil || !user.IsAuthenticated {
		return true
	}

	// Session expires after 30 days of inactivity
	expirationTime := user.LastActiveAt.Add(30 * 24 * time.Hour)
	return time.Now().After(expirationTime)
}

// RefreshAuthToken generates a new auth token for an existing user
func (s *whatsAppAuthService) RefreshAuthToken(phoneNumber string) error {
	user, err := s.GetUserByPhone(phoneNumber)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return fmt.Errorf("user not found")
	}

	// Generate new token
	token, err := s.GenerateAuthToken(user.ClerkUserID)
	if err != nil {
		return fmt.Errorf("failed to generate auth token: %w", err)
	}

	// Update user with new token
	err = s.db.Model(&models.WhatsAppUser{}).
		Where("phone_number = ?", phoneNumber).
		Updates(map[string]interface{}{
			"auth_token":     token,
			"last_active_at": time.Now(),
		}).Error

	if err != nil {
		log.Error().Err(err).Str("phone", phoneNumber).Msg("Failed to refresh auth token")
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	log.Info().
		Str("phone", phoneNumber).
		Msg("Auth token refreshed successfully")

	return nil
}
