package auth

import (
	"backend/db"
	"backend/internal/middleware"
	"backend/internal/models"
	"backend/pkg/utils"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type CurrentUserDTO struct {
	ClerkUserID         string  `json:"clerkUserId"`
	Name                string  `json:"name"`
	Email               string  `json:"email"`
	ImageUrl            *string `json:"imageUrl"`
	OnboardingCompleted bool    `json:"onboardingCompleted"`
	HasApiKey           bool    `json:"hasApiKey"`
}

// GetCurrentUser returns the currently logged-in user with onboarding and API key status
func GetCurrentUser(c *gin.Context) {
	// Get Clerk session claims from context (includes publicMetadata)
	claims, ok := clerk.SessionClaimsFromContext(c.Request.Context())
	if !ok || claims == nil {
		log.Error().Msg("No Clerk session found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	clerkUserID := claims.Subject

	// Fetch user from Clerk to get full details (cached)
	clerkUser, err := middleware.GetUserCached(c.Request.Context(), clerkUserID)
	if err != nil {
		log.Error().Err(err).Msg("Error fetching user from Clerk")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}

	// Extract user data from Clerk
	imageUrl := ""
	if clerkUser.ImageURL != nil {
		imageUrl = *clerkUser.ImageURL
	}

	email := ""
	if len(clerkUser.EmailAddresses) > 0 {
		for _, emailAddr := range clerkUser.EmailAddresses {
			if emailAddr.ID == *clerkUser.PrimaryEmailAddressID {
				email = emailAddr.EmailAddress
				break
			}
		}
	}

	name := ""
	if clerkUser.FirstName != nil && clerkUser.LastName != nil {
		name = *clerkUser.FirstName + " " + *clerkUser.LastName
	} else if clerkUser.FirstName != nil {
		name = *clerkUser.FirstName
	} else if clerkUser.Username != nil {
		name = *clerkUser.Username
	}

	// Get onboarding status from Clerk's publicMetadata
	onboardingCompleted := false
	if len(clerkUser.PublicMetadata) > 0 {
		var metadata map[string]interface{}
		if err := json.Unmarshal(clerkUser.PublicMetadata, &metadata); err == nil {
			if completed, ok := metadata["onboardingCompleted"].(bool); ok {
				onboardingCompleted = completed
			}
		}
	}

	// Check if user has any AI credentials
	var credentialCount int64
	db.DB.Model(&models.AICredential{}).Where("clerk_user_id = ?", clerkUserID).Count(&credentialCount)
	hasApiKey := credentialCount > 0

	currentUser := CurrentUserDTO{
		ClerkUserID:         clerkUserID,
		Name:                name,
		Email:               email,
		ImageUrl:            &imageUrl,
		OnboardingCompleted: onboardingCompleted,
		HasApiKey:           hasApiKey,
	}

	c.JSON(http.StatusOK, currentUser)
}

// GetOnboardingStatus returns the current onboarding status for the user
func GetOnboardingStatus(c *gin.Context) {
	claims, ok := clerk.SessionClaimsFromContext(c.Request.Context())
	if !ok || claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Fetch user from Clerk to get publicMetadata (cached)
	clerkUser, err := middleware.GetUserCached(c.Request.Context(), claims.Subject)
	if err != nil {
		log.Error().Err(err).Msg("Error fetching user from Clerk")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}

	onboardingCompleted := false
	var onboardingType *string

	// Extract from publicMetadata
	if len(clerkUser.PublicMetadata) > 0 {
		var metadata map[string]interface{}
		if err := json.Unmarshal(clerkUser.PublicMetadata, &metadata); err == nil {
			if completed, ok := metadata["onboardingCompleted"].(bool); ok {
				onboardingCompleted = completed
			}
			if typeVal, ok := metadata["onboardingType"].(string); ok {
				onboardingType = &typeVal
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"completed": onboardingCompleted,
		"type":      onboardingType,
	})
}

// CompleteOnboarding completes the onboarding process for the user
func CompleteOnboarding(c *gin.Context) {
	claims, ok := clerk.SessionClaimsFromContext(c.Request.Context())
	if !ok || claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var requestBody struct {
		Type string `json:"type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Update user's publicMetadata in Clerk
	metadata := map[string]interface{}{
		"onboardingCompleted": true,
		"onboardingType":      requestBody.Type,
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		log.Error().Err(err).Msg("Error marshaling metadata")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update onboarding status"})
		return
	}
	rawMessage := json.RawMessage(metadataJSON)

	_, err = user.Update(c.Request.Context(), claims.Subject, &user.UpdateParams{
		PublicMetadata: &rawMessage,
	})

	if err != nil {
		log.Error().Err(err).Msg("Error updating user onboarding in Clerk")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update onboarding status"})
		return
	}

	log.Info().Str("clerk_user_id", claims.Subject).Str("type", requestBody.Type).Msg("Onboarding completed")
	c.JSON(http.StatusOK, gin.H{"message": "Onboarding completed successfully"})
}

// ResetOnboarding resets the onboarding status for the user (dev only)
func ResetOnboarding(c *gin.Context) {
	claims, ok := clerk.SessionClaimsFromContext(c.Request.Context())
	if !ok || claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Reset user's publicMetadata in Clerk
	metadata := map[string]interface{}{
		"onboardingCompleted": false,
		"onboardingType":      nil,
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		log.Error().Err(err).Msg("Error marshaling metadata")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset onboarding status"})
		return
	}
	rawMessage := json.RawMessage(metadataJSON)

	_, err = user.Update(c.Request.Context(), claims.Subject, &user.UpdateParams{
		PublicMetadata: &rawMessage,
	})

	if err != nil {
		log.Error().Err(err).Msg("Error resetting user onboarding in Clerk")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset onboarding status"})
		return
	}

	log.Info().Str("clerk_user_id", claims.Subject).Msg("Onboarding reset")
	c.JSON(http.StatusOK, gin.H{"message": "Onboarding reset successfully"})
}

// SetAICredential stores an encrypted AI API key for the user
func SetAICredential(c *gin.Context) {
	claims, ok := clerk.SessionClaimsFromContext(c.Request.Context())
	if !ok || claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	clerkUserID := claims.Subject

	var requestBody struct {
		Provider string `json:"provider" binding:"required"`
		ApiKey   string `json:"apiKey" binding:"required"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Trim whitespace from API key before encrypting
	apiKey := strings.TrimSpace(requestBody.ApiKey)
	if apiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "API key cannot be empty"})
		return
	}

	// Encrypt the API key
	encryptedKey, err := utils.Encrypt(apiKey)
	if err != nil {
		log.Error().Err(err).Msg("Error encrypting API key")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt API key"})
		return
	}

	// Create or update the credential
	credential := models.AICredential{
		ClerkUserID: clerkUserID,
		Provider:    requestBody.Provider,
		KeyCipher:   encryptedKey,
	}

	if err := db.DB.Where(models.AICredential{ClerkUserID: clerkUserID, Provider: requestBody.Provider}).
		Assign(models.AICredential{KeyCipher: encryptedKey}).
		FirstOrCreate(&credential).Error; err != nil {
		log.Error().Err(err).Msg("Error saving AI credential")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save API key"})
		return
	}

	// Return masked key for UI display
	maskedKey := utils.MaskAPIKey(apiKey)
	c.JSON(http.StatusOK, gin.H{
		"message":   "API key saved successfully",
		"provider":  requestBody.Provider,
		"maskedKey": maskedKey,
	})
}

// DeleteAICredential removes an AI API key for the user
func DeleteAICredential(c *gin.Context) {
	claims, ok := clerk.SessionClaimsFromContext(c.Request.Context())
	if !ok || claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	clerkUserID := claims.Subject

	var requestBody struct {
		Provider string `json:"provider" binding:"required"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Delete the credential
	if err := db.DB.Where("clerk_user_id = ? AND provider = ?", clerkUserID, requestBody.Provider).
		Delete(&models.AICredential{}).Error; err != nil {
		log.Error().Err(err).Msg("Error deleting AI credential")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete API key"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API key deleted successfully"})
}

// GetAICredentials returns the list of providers for which the user has API keys configured
func GetAICredentials(c *gin.Context) {
	claims, ok := clerk.SessionClaimsFromContext(c.Request.Context())
	if !ok || claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	clerkUserID := claims.Subject

	// Get all credentials for the user
	var credentials []models.AICredential
	if err := db.DB.Where("clerk_user_id = ?", clerkUserID).Find(&credentials).Error; err != nil {
		log.Error().Err(err).Msg("Error fetching AI credentials")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch API keys"})
		return
	}

	// Build a map of provider -> has key
	providers := make(map[string]bool)
	for _, cred := range credentials {
		providers[cred.Provider] = true
	}

	c.JSON(http.StatusOK, gin.H{
		"providers": providers,
	})
}

// UserCreatedWebhook handles Clerk user.created webhook events
// Note: With the users table removed, this webhook is no longer needed
// but kept for future extensibility if we need to trigger actions on user creation
func UserCreatedWebhook(c *gin.Context) {
	// No database sync needed anymore since Clerk is the source of truth
	// This can be used for other side effects like sending welcome emails, etc.
	c.JSON(http.StatusOK, gin.H{"message": "User webhook received"})
}
