package services

import (
	"backend/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupAuthTestDB creates an in-memory SQLite database for testing
func setupAuthTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err, "Failed to open test database")

	// Migrate the schema
	err = db.AutoMigrate(&models.WhatsAppUser{})
	require.NoError(t, err, "Failed to migrate test database")

	return db
}

// setupAuthTestService creates a test auth service
func setupAuthTestService(t *testing.T) (*whatsAppAuthService, *gorm.DB) {
	db := setupAuthTestDB(t)
	service := &whatsAppAuthService{db: db}
	return service, db
}

func TestGenerateLinkToken(t *testing.T) {
	service, _ := setupAuthTestService(t)

	phoneNumber := "+1234567890"

	// Generate link token
	token, err := service.GenerateLinkToken(phoneNumber)
	require.NoError(t, err, "Failed to generate link token")
	assert.NotEmpty(t, token, "Token should not be empty")
	assert.Contains(t, token, phoneNumber, "Token should contain phone number")
}

func TestValidateLinkToken(t *testing.T) {
	service, _ := setupAuthTestService(t)

	phoneNumber := "+1234567890"

	// Generate link token
	token, err := service.GenerateLinkToken(phoneNumber)
	require.NoError(t, err, "Failed to generate link token")

	// Validate the token
	extractedPhone, err := service.ValidateLinkToken(token)
	require.NoError(t, err, "Failed to validate link token")
	assert.Equal(t, phoneNumber, extractedPhone, "Extracted phone number should match")
}

func TestValidateLinkToken_InvalidFormat(t *testing.T) {
	service, _ := setupAuthTestService(t)

	// Test with invalid token format
	_, err := service.ValidateLinkToken("invalid-token")
	assert.Error(t, err, "Should fail with invalid token format")
}

func TestLinkPhoneToUser_NewUser(t *testing.T) {
	service, db := setupAuthTestService(t)

	phoneNumber := "+1234567890"
	clerkUserID := "user_test123"

	// Link phone to user (should create new user)
	user, err := service.LinkPhoneToUser(phoneNumber, clerkUserID)
	require.NoError(t, err, "Failed to link phone to user")
	assert.NotNil(t, user, "User should not be nil")
	assert.Equal(t, phoneNumber, user.PhoneNumber, "Phone number should match")
	assert.Equal(t, clerkUserID, user.ClerkUserID, "Clerk user ID should match")
	assert.True(t, user.IsAuthenticated, "User should be authenticated")
	assert.NotEmpty(t, user.AuthToken, "Auth token should be generated")

	// Verify user was created in database
	var dbUser models.WhatsAppUser
	err = db.Where("phone_number = ?", phoneNumber).First(&dbUser).Error
	require.NoError(t, err, "User should exist in database")
	assert.Equal(t, clerkUserID, dbUser.ClerkUserID, "Clerk user ID should match in database")
}

func TestLinkPhoneToUser_ExistingUser(t *testing.T) {
	service, db := setupAuthTestService(t)

	phoneNumber := "+1234567890"
	oldClerkUserID := "user_old123"
	newClerkUserID := "user_new456"

	// Create existing user
	existingUser := &models.WhatsAppUser{
		PhoneNumber:     phoneNumber,
		ClerkUserID:     oldClerkUserID,
		IsAuthenticated: false,
		LastActiveAt:    time.Now().Add(-48 * time.Hour),
	}
	err := db.Create(existingUser).Error
	require.NoError(t, err, "Failed to create existing user")

	// Link phone to new user (should update existing user)
	user, err := service.LinkPhoneToUser(phoneNumber, newClerkUserID)
	require.NoError(t, err, "Failed to link phone to user")
	assert.NotNil(t, user, "User should not be nil")
	assert.Equal(t, phoneNumber, user.PhoneNumber, "Phone number should match")
	assert.Equal(t, newClerkUserID, user.ClerkUserID, "Clerk user ID should be updated")
	assert.True(t, user.IsAuthenticated, "User should be authenticated")
	assert.NotEmpty(t, user.AuthToken, "Auth token should be generated")

	// Verify only one user exists in database
	var count int64
	db.Model(&models.WhatsAppUser{}).Where("phone_number = ?", phoneNumber).Count(&count)
	assert.Equal(t, int64(1), count, "Should have exactly one user")
}

func TestIsSessionExpired_NotExpired(t *testing.T) {
	service, _ := setupAuthTestService(t)

	user := &models.WhatsAppUser{
		PhoneNumber:     "+1234567890",
		ClerkUserID:     "user_test123",
		IsAuthenticated: true,
		LastActiveAt:    time.Now().Add(-1 * time.Hour), // 1 hour ago
	}

	expired := service.IsSessionExpired(user)
	assert.False(t, expired, "Session should not be expired")
}

func TestIsSessionExpired_Expired(t *testing.T) {
	service, _ := setupAuthTestService(t)

	user := &models.WhatsAppUser{
		PhoneNumber:     "+1234567890",
		ClerkUserID:     "user_test123",
		IsAuthenticated: true,
		LastActiveAt:    time.Now().Add(-31 * 24 * time.Hour), // 31 days ago
	}

	expired := service.IsSessionExpired(user)
	assert.True(t, expired, "Session should be expired")
}

func TestIsSessionExpired_NotAuthenticated(t *testing.T) {
	service, _ := setupAuthTestService(t)

	user := &models.WhatsAppUser{
		PhoneNumber:     "+1234567890",
		ClerkUserID:     "user_test123",
		IsAuthenticated: false,
		LastActiveAt:    time.Now(),
	}

	expired := service.IsSessionExpired(user)
	assert.True(t, expired, "Unauthenticated user should be considered expired")
}

func TestIsSessionExpired_NilUser(t *testing.T) {
	service, _ := setupAuthTestService(t)

	expired := service.IsSessionExpired(nil)
	assert.True(t, expired, "Nil user should be considered expired")
}

func TestRefreshAuthToken(t *testing.T) {
	service, db := setupAuthTestService(t)

	phoneNumber := "+1234567890"
	clerkUserID := "user_test123"

	// Create user with old token
	user := &models.WhatsAppUser{
		PhoneNumber:     phoneNumber,
		ClerkUserID:     clerkUserID,
		IsAuthenticated: true,
		AuthToken:       "old_token",
		LastActiveAt:    time.Now().Add(-24 * time.Hour),
	}
	err := db.Create(user).Error
	require.NoError(t, err, "Failed to create user")

	oldToken := user.AuthToken
	oldLastActive := user.LastActiveAt

	// Refresh token
	err = service.RefreshAuthToken(phoneNumber)
	require.NoError(t, err, "Failed to refresh auth token")

	// Verify token was updated
	var updatedUser models.WhatsAppUser
	err = db.Where("phone_number = ?", phoneNumber).First(&updatedUser).Error
	require.NoError(t, err, "Failed to get updated user")

	assert.NotEqual(t, oldToken, updatedUser.AuthToken, "Token should be updated")
	assert.True(t, updatedUser.LastActiveAt.After(oldLastActive), "Last active should be updated")
}

func TestRefreshAuthToken_UserNotFound(t *testing.T) {
	service, _ := setupAuthTestService(t)

	phoneNumber := "+1234567890"

	// Try to refresh token for non-existent user
	err := service.RefreshAuthToken(phoneNumber)
	assert.Error(t, err, "Should fail for non-existent user")
}
