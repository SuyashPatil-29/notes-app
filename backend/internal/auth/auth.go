package auth

import (
	"backend/db"
	"backend/internal/models"
	"backend/pkg/utils"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/rs/zerolog/log"
)

type CurrentUserDTO struct {
	ID                  uint    `json:"id"`
	Name                string  `json:"name"`
	Email               string  `json:"email"`
	ImageUrl            *string `json:"imageUrl"`
	OnboardingCompleted bool    `json:"onboardingCompleted"`
	HasApiKey           bool    `json:"hasApiKey"`
}

const (
	MaxAge = 86400 * 30
)

var Store *sessions.CookieStore

func InitAuth() {
	var isProd bool
	err := godotenv.Load()
	if err != nil {
		log.Warn().Msg("Error loading .env file, using environment variables")
	}

	if os.Getenv("IsProd") == "true" {
		isProd = true
	} else {
		isProd = false
	}

	googleClientId := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	callbackURL := os.Getenv("GOOGLE_CALLBACK_URL")
	sessionHashKey := os.Getenv("SESSION_HASH_KEY")
	sessionBlockKey := os.Getenv("SESSION_BLOCK_KEY")

	if callbackURL == "" {
		callbackURL = "http://localhost:8080/auth/google/callback"
	}

	if sessionHashKey == "" {
		sessionHashKey = "MyRandomHashKeyPleaseChangeThis32"
	}
	if sessionBlockKey == "" {
		sessionBlockKey = "MyBlockKey16Bytes"
	}

	// Ensure keys are proper lengths
	hashKey := []byte(sessionHashKey)
	if len(hashKey) < 32 {
		hashKey = append(hashKey, make([]byte, 32-len(hashKey))...)
	}
	blockKey := []byte(sessionBlockKey)
	if len(blockKey) != 16 && len(blockKey) != 24 && len(blockKey) != 32 {
		// Pad to 32 bytes for AES-256
		if len(blockKey) < 32 {
			blockKey = append(blockKey, make([]byte, 32-len(blockKey))...)
		} else {
			blockKey = blockKey[:32]
		}
	}

	Store = sessions.NewCookieStore(hashKey, blockKey)
	Store.MaxAge(MaxAge)

	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   MaxAge,
		HttpOnly: true,
		Secure:   isProd,
		SameSite: http.SameSiteLaxMode,
	}

	gothic.Store = Store

	goth.UseProviders(google.New(googleClientId, googleClientSecret, callbackURL))
}

// BeginAuth starts the authentication process
func BeginAuth(c *gin.Context) {
	provider := c.Param("provider")

	// Set provider in query if not already set
	q := c.Request.URL.Query()
	if q.Get("provider") == "" {
		q.Set("provider", provider)
		c.Request.URL.RawQuery = q.Encode()
	}

	gothic.BeginAuthHandler(c.Writer, c.Request)
}

// AuthCallback handles the OAuth callback
func AuthCallback(c *gin.Context) {
	provider := c.Param("provider")

	q := c.Request.URL.Query()
	if q.Get("provider") == "" {
		q.Set("provider", provider)
		c.Request.URL.RawQuery = q.Encode()
	}

	gothUser, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		log.Print("Error completing auth: ", err)
		c.Redirect(http.StatusTemporaryRedirect, "http://localhost:5173?error=auth_failed")
		return
	}

	// Check if user exists in database
	var user models.User
	result := db.DB.Where("email = ?", gothUser.Email).First(&user)

	if result.Error != nil {
		// User doesn't exist, create new user
		imageUrl := gothUser.AvatarURL
		user = models.User{
			Name:     gothUser.Name,
			Email:    gothUser.Email,
			ImageUrl: &imageUrl,
		}

		if err := db.DB.Create(&user).Error; err != nil {
			log.Print("Error creating user in db: ", err)
			c.Redirect(http.StatusTemporaryRedirect, "http://localhost:5173?error=db_error")
			return
		}
	}

	// Store user in session
	session, _ := Store.Get(c.Request, "auth-session")
	session.Values["user_id"] = user.ID
	session.Values["email"] = user.Email
	session.Values["name"] = user.Name
	if err := session.Save(c.Request, c.Writer); err != nil {
		log.Print("Error saving session: ", err)
	}

	// Redirect to frontend with user info
	c.Redirect(http.StatusTemporaryRedirect, "http://localhost:5173?login=success")
}

// GetCurrentUser returns the currently logged-in user with onboarding and API key status
func GetCurrentUser(c *gin.Context) {
	session, err := Store.Get(c.Request, "auth-session")
	if err != nil {
		log.Print("Error getting session: ", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	userID, ok := session.Values["user_id"]
	if !ok || userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Fetch user from database with onboarding info
	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		log.Print("User not found in database: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Check if user has any AI credentials
	var credentialCount int64
	db.DB.Model(&models.AICredential{}).Where("user_id = ?", userID).Count(&credentialCount)
	hasApiKey := credentialCount > 0

	// Update session with current values
	session.Values["onboarding_completed"] = user.OnboardingCompleted
	session.Values["has_api_key"] = hasApiKey
	if err := session.Save(c.Request, c.Writer); err != nil {
		log.Print("Error saving session: ", err)
	}

	currentUser := CurrentUserDTO{
		ID:                  user.ID,
		Name:                user.Name,
		Email:               user.Email,
		ImageUrl:            user.ImageUrl,
		OnboardingCompleted: user.OnboardingCompleted,
		HasApiKey:           hasApiKey,
	}

	c.JSON(http.StatusOK, currentUser)
}

// GetOnboardingStatus returns the current onboarding status for the user
func GetOnboardingStatus(c *gin.Context) {
	session, err := Store.Get(c.Request, "auth-session")
	if err != nil {
		log.Print("Error getting session: ", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	userID, ok := session.Values["user_id"]
	if !ok || userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var user models.User
	if err := db.DB.Select("onboarding_completed, onboarding_type").First(&user, userID).Error; err != nil {
		log.Print("User not found in database: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"completed": user.OnboardingCompleted,
		"type":      user.OnboardingType,
	})
}

// CompleteOnboarding completes the onboarding process for the user
func CompleteOnboarding(c *gin.Context) {
	session, err := Store.Get(c.Request, "auth-session")
	if err != nil {
		log.Print("Error getting session: ", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	userID, ok := session.Values["user_id"]
	if !ok || userID == nil {
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

	// Update user onboarding status
	if err := db.DB.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"onboarding_completed": true,
		"onboarding_type":      &requestBody.Type,
	}).Error; err != nil {
		log.Print("Error updating user onboarding: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update onboarding status"})
		return
	}

	// Update session
	session.Values["onboarding_completed"] = true
	if err := session.Save(c.Request, c.Writer); err != nil {
		log.Print("Error saving session: ", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Onboarding completed successfully"})
}

// ResetOnboarding resets the onboarding status for the user (dev only)
func ResetOnboarding(c *gin.Context) {
	session, err := Store.Get(c.Request, "auth-session")
	if err != nil {
		log.Print("Error getting session: ", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	userID, ok := session.Values["user_id"]
	if !ok || userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Reset user onboarding status
	if err := db.DB.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"onboarding_completed": false,
		"onboarding_type":      nil,
	}).Error; err != nil {
		log.Print("Error resetting user onboarding: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset onboarding status"})
		return
	}

	// Update session
	session.Values["onboarding_completed"] = false
	if err := session.Save(c.Request, c.Writer); err != nil {
		log.Print("Error saving session: ", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Onboarding reset successfully"})
}

// SetAICredential stores an encrypted AI API key for the user
func SetAICredential(c *gin.Context) {
	session, err := Store.Get(c.Request, "auth-session")
	if err != nil {
		log.Print("Error getting session: ", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	userID, ok := session.Values["user_id"]
	if !ok || userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

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
		log.Print("Error encrypting API key: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt API key"})
		return
	}

	// Create or update the credential
	credential := models.AICredential{
		UserID:    userID.(uint),
		Provider:  requestBody.Provider,
		KeyCipher: encryptedKey,
	}

	if err := db.DB.Where(models.AICredential{UserID: userID.(uint), Provider: requestBody.Provider}).
		Assign(models.AICredential{KeyCipher: encryptedKey}).
		FirstOrCreate(&credential).Error; err != nil {
		log.Print("Error saving AI credential: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save API key"})
		return
	}

	// Update session
	session.Values["has_api_key"] = true
	if err := session.Save(c.Request, c.Writer); err != nil {
		log.Print("Error saving session: ", err)
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
	session, err := Store.Get(c.Request, "auth-session")
	if err != nil {
		log.Print("Error getting session: ", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	userID, ok := session.Values["user_id"]
	if !ok || userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var requestBody struct {
		Provider string `json:"provider" binding:"required"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Delete the credential
	if err := db.DB.Where("user_id = ? AND provider = ?", userID, requestBody.Provider).
		Delete(&models.AICredential{}).Error; err != nil {
		log.Print("Error deleting AI credential: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete API key"})
		return
	}

	// Check if user still has any credentials
	var credentialCount int64
	db.DB.Model(&models.AICredential{}).Where("user_id = ?", userID).Count(&credentialCount)

	// Update session
	session.Values["has_api_key"] = credentialCount > 0
	if err := session.Save(c.Request, c.Writer); err != nil {
		log.Print("Error saving session: ", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "API key deleted successfully"})
}

// GetAICredentials returns the list of providers for which the user has API keys configured
func GetAICredentials(c *gin.Context) {
	session, err := Store.Get(c.Request, "auth-session")
	if err != nil {
		log.Print("Error getting session: ", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	userID, ok := session.Values["user_id"]
	if !ok || userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Get all credentials for the user
	var credentials []models.AICredential
	if err := db.DB.Where("user_id = ?", userID).Find(&credentials).Error; err != nil {
		log.Print("Error fetching AI credentials: ", err)
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

// Logout logs out the user
func Logout(c *gin.Context) {
	provider := c.Param("provider")

	q := c.Request.URL.Query()
	if q.Get("provider") == "" {
		q.Set("provider", provider)
		c.Request.URL.RawQuery = q.Encode()
	}

	// Clear session
	session, _ := Store.Get(c.Request, "auth-session")
	session.Values["user_id"] = nil
	session.Values["email"] = nil
	session.Values["name"] = nil
	session.Options.MaxAge = -1
	session.Save(c.Request, c.Writer)

	if err := gothic.Logout(c.Writer, c.Request); err != nil {
		log.Print("Error during logout: ", err)
		c.Redirect(http.StatusTemporaryRedirect, "http://localhost:5173?error=logout_failed")
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, "http://localhost:5173?logout=success")
}
