package services

import (
	"backend/internal/config"
	"backend/internal/models"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err, "Failed to open test database")

	// Migrate the schema
	err = db.AutoMigrate(&models.WhatsAppConversationContext{})
	require.NoError(t, err, "Failed to migrate test database")

	return db
}

// setupTestService creates a test context service with test configuration
func setupTestService(t *testing.T) (*WhatsAppContextService, *gorm.DB) {
	db := setupTestDB(t)

	config := &config.WhatsAppConfig{
		ContextExpiration: 10 * time.Minute,
	}

	service := NewWhatsAppContextService(db, config)
	return service, db
}

func TestSetContext(t *testing.T) {
	service, db := setupTestService(t)

	phoneNumber := "+1234567890"
	command := "add_note"
	data := map[string]interface{}{
		"title": "Test Note",
		"step":  "awaiting_notebook",
	}

	// Set context
	err := service.SetContext(phoneNumber, command, data)
	require.NoError(t, err, "SetContext should not return an error")

	// Verify context was created in database
	var context models.WhatsAppConversationContext
	err = db.Where("phone_number = ?", phoneNumber).First(&context).Error
	require.NoError(t, err, "Context should exist in database")

	assert.Equal(t, phoneNumber, context.PhoneNumber)
	assert.Equal(t, command, context.Command)
	assert.Equal(t, 0, context.Step)
	assert.True(t, context.ExpiresAt.After(time.Now()))

	// Verify data was stored correctly
	var storedData map[string]interface{}
	err = json.Unmarshal([]byte(context.Data), &storedData)
	require.NoError(t, err, "Should be able to unmarshal stored data")
	assert.Equal(t, "Test Note", storedData["title"])
	assert.Equal(t, "awaiting_notebook", storedData["step"])
}

func TestGetContext(t *testing.T) {
	service, _ := setupTestService(t)

	phoneNumber := "+1234567890"
	command := "add_note"
	data := map[string]interface{}{
		"title": "Test Note",
	}

	// Set context first
	err := service.SetContext(phoneNumber, command, data)
	require.NoError(t, err)

	// Get context
	context, err := service.GetContext(phoneNumber)
	require.NoError(t, err, "GetContext should not return an error")
	require.NotNil(t, context, "Context should not be nil")

	assert.Equal(t, phoneNumber, context.PhoneNumber)
	assert.Equal(t, command, context.Command)
	assert.Equal(t, 0, context.Step)
}

func TestGetContext_NotFound(t *testing.T) {
	service, _ := setupTestService(t)

	phoneNumber := "+9999999999"

	// Get context for non-existent phone number
	context, err := service.GetContext(phoneNumber)
	require.NoError(t, err, "GetContext should not return an error for non-existent context")
	assert.Nil(t, context, "Context should be nil when not found")
}

func TestUpdateContext(t *testing.T) {
	service, _ := setupTestService(t)

	phoneNumber := "+1234567890"
	command := "add_note"
	initialData := map[string]interface{}{
		"title": "Test Note",
	}

	// Set initial context
	err := service.SetContext(phoneNumber, command, initialData)
	require.NoError(t, err)

	// Update context with new data
	updateData := map[string]interface{}{
		"notebook": "Work Notes",
	}
	err = service.UpdateContext(phoneNumber, updateData)
	require.NoError(t, err, "UpdateContext should not return an error")

	// Get updated context
	context, err := service.GetContext(phoneNumber)
	require.NoError(t, err)
	require.NotNil(t, context)

	// Verify step was incremented
	assert.Equal(t, 1, context.Step)

	// Verify data was merged
	var storedData map[string]interface{}
	err = json.Unmarshal([]byte(context.Data), &storedData)
	require.NoError(t, err)
	assert.Equal(t, "Test Note", storedData["title"])
	assert.Equal(t, "Work Notes", storedData["notebook"])
}

func TestUpdateContext_NoActiveContext(t *testing.T) {
	service, _ := setupTestService(t)

	phoneNumber := "+9999999999"
	updateData := map[string]interface{}{
		"notebook": "Work Notes",
	}

	// Try to update non-existent context
	err := service.UpdateContext(phoneNumber, updateData)
	assert.Error(t, err, "UpdateContext should return an error when no context exists")
	assert.Contains(t, err.Error(), "no active context found")
}

func TestClearContext(t *testing.T) {
	service, db := setupTestService(t)

	phoneNumber := "+1234567890"
	command := "add_note"
	data := map[string]interface{}{
		"title": "Test Note",
	}

	// Set context
	err := service.SetContext(phoneNumber, command, data)
	require.NoError(t, err)

	// Clear context
	err = service.ClearContext(phoneNumber)
	require.NoError(t, err, "ClearContext should not return an error")

	// Verify context was deleted
	var count int64
	db.Model(&models.WhatsAppConversationContext{}).
		Where("phone_number = ?", phoneNumber).
		Count(&count)
	assert.Equal(t, int64(0), count, "Context should be deleted from database")
}

func TestClearContext_NoContext(t *testing.T) {
	service, _ := setupTestService(t)

	phoneNumber := "+9999999999"

	// Clear non-existent context (should not error)
	err := service.ClearContext(phoneNumber)
	assert.NoError(t, err, "ClearContext should not error when no context exists")
}

func TestIsContextExpired(t *testing.T) {
	service, _ := setupTestService(t)

	t.Run("Not expired", func(t *testing.T) {
		context := &models.WhatsAppConversationContext{
			ExpiresAt: time.Now().Add(5 * time.Minute),
		}

		expired := service.IsContextExpired(context)
		assert.False(t, expired, "Context should not be expired")
	})

	t.Run("Expired", func(t *testing.T) {
		context := &models.WhatsAppConversationContext{
			ExpiresAt: time.Now().Add(-5 * time.Minute),
		}

		expired := service.IsContextExpired(context)
		assert.True(t, expired, "Context should be expired")
	})

	t.Run("Nil context", func(t *testing.T) {
		expired := service.IsContextExpired(nil)
		assert.True(t, expired, "Nil context should be considered expired")
	})
}

func TestGetContext_ExpiredContext(t *testing.T) {
	service, db := setupTestService(t)

	phoneNumber := "+1234567890"

	// Create an expired context directly in database
	expiredContext := models.WhatsAppConversationContext{
		PhoneNumber: phoneNumber,
		Command:     "add_note",
		Step:        0,
		Data:        `{"title":"Test"}`,
		ExpiresAt:   time.Now().Add(-5 * time.Minute), // Expired 5 minutes ago
	}
	err := db.Create(&expiredContext).Error
	require.NoError(t, err)

	// Try to get expired context
	context, err := service.GetContext(phoneNumber)
	require.NoError(t, err)
	assert.Nil(t, context, "Expired context should be cleared and return nil")

	// Verify context was deleted
	var count int64
	db.Model(&models.WhatsAppConversationContext{}).
		Where("phone_number = ?", phoneNumber).
		Count(&count)
	assert.Equal(t, int64(0), count, "Expired context should be deleted")
}

func TestCleanupExpiredContexts(t *testing.T) {
	service, db := setupTestService(t)

	// Create multiple contexts, some expired
	contexts := []models.WhatsAppConversationContext{
		{
			PhoneNumber: "+1111111111",
			Command:     "add_note",
			Data:        `{"title":"Test1"}`,
			ExpiresAt:   time.Now().Add(-10 * time.Minute), // Expired
		},
		{
			PhoneNumber: "+2222222222",
			Command:     "delete_note",
			Data:        `{"title":"Test2"}`,
			ExpiresAt:   time.Now().Add(10 * time.Minute), // Not expired
		},
		{
			PhoneNumber: "+3333333333",
			Command:     "retrieve_note",
			Data:        `{"title":"Test3"}`,
			ExpiresAt:   time.Now().Add(-5 * time.Minute), // Expired
		},
	}

	for _, ctx := range contexts {
		err := db.Create(&ctx).Error
		require.NoError(t, err)
	}

	// Cleanup expired contexts
	err := service.CleanupExpiredContexts()
	require.NoError(t, err, "CleanupExpiredContexts should not return an error")

	// Verify only non-expired context remains
	var count int64
	db.Model(&models.WhatsAppConversationContext{}).Count(&count)
	assert.Equal(t, int64(1), count, "Only one non-expired context should remain")

	// Verify the correct context remains
	var remainingContext models.WhatsAppConversationContext
	err = db.First(&remainingContext).Error
	require.NoError(t, err)
	assert.Equal(t, "+2222222222", remainingContext.PhoneNumber)
}

func TestGetContextData(t *testing.T) {
	service, _ := setupTestService(t)

	phoneNumber := "+1234567890"
	command := "add_note"
	data := map[string]interface{}{
		"title":    "Test Note",
		"notebook": "Work Notes",
		"step":     "awaiting_chapter",
	}

	// Set context
	err := service.SetContext(phoneNumber, command, data)
	require.NoError(t, err)

	// Get context data
	retrievedData, err := service.GetContextData(phoneNumber)
	require.NoError(t, err, "GetContextData should not return an error")
	require.NotNil(t, retrievedData, "Context data should not be nil")

	assert.Equal(t, "Test Note", retrievedData["title"])
	assert.Equal(t, "Work Notes", retrievedData["notebook"])
	assert.Equal(t, "awaiting_chapter", retrievedData["step"])
}

func TestGetContextData_NoContext(t *testing.T) {
	service, _ := setupTestService(t)

	phoneNumber := "+9999999999"

	// Get context data for non-existent context
	data, err := service.GetContextData(phoneNumber)
	require.NoError(t, err, "GetContextData should not return an error for non-existent context")
	assert.Nil(t, data, "Context data should be nil when no context exists")
}

func TestSetContext_ReplacesExisting(t *testing.T) {
	service, db := setupTestService(t)

	phoneNumber := "+1234567890"

	// Set first context
	err := service.SetContext(phoneNumber, "add_note", map[string]interface{}{
		"title": "First Note",
	})
	require.NoError(t, err)

	// Set second context (should replace first)
	err = service.SetContext(phoneNumber, "delete_note", map[string]interface{}{
		"search": "Second Note",
	})
	require.NoError(t, err)

	// Verify only one context exists
	var count int64
	db.Model(&models.WhatsAppConversationContext{}).
		Where("phone_number = ?", phoneNumber).
		Count(&count)
	assert.Equal(t, int64(1), count, "Only one context should exist")

	// Verify it's the second context
	context, err := service.GetContext(phoneNumber)
	require.NoError(t, err)
	require.NotNil(t, context)
	assert.Equal(t, "delete_note", context.Command)
}
