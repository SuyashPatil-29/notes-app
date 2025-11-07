package services

import (
	"backend/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAuditTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto-migrate the WhatsAppMessage model
	err = db.AutoMigrate(&models.WhatsAppMessage{})
	assert.NoError(t, err)

	return db
}

func TestLogInboundMessage(t *testing.T) {
	db := setupAuditTestDB(t)
	service := NewWhatsAppAuditService(db)

	messageID := "test_msg_123"
	phoneNumber := "+1234567890"
	content := "Hello, bot!"
	timestamp := time.Now()

	err := service.LogInboundMessage(messageID, phoneNumber, "text", content, timestamp)
	assert.NoError(t, err)

	// Verify the message was logged
	var msg models.WhatsAppMessage
	err = db.Where("message_id = ?", messageID).First(&msg).Error
	assert.NoError(t, err)
	assert.Equal(t, messageID, msg.MessageID)
	assert.Equal(t, phoneNumber, msg.PhoneNumber)
	assert.Equal(t, "inbound", msg.Direction)
	assert.Equal(t, "text", msg.MessageType)
	assert.Equal(t, content, msg.Content)
	assert.Equal(t, "received", msg.Status)
}

func TestLogOutboundMessage(t *testing.T) {
	db := setupAuditTestDB(t)
	service := NewWhatsAppAuditService(db)

	messageID := "test_msg_456"
	phoneNumber := "+1234567890"
	content := "Hello, user!"

	err := service.LogOutboundMessage(messageID, phoneNumber, "text", content, "sent")
	assert.NoError(t, err)

	// Verify the message was logged
	var msg models.WhatsAppMessage
	err = db.Where("message_id = ?", messageID).First(&msg).Error
	assert.NoError(t, err)
	assert.Equal(t, messageID, msg.MessageID)
	assert.Equal(t, phoneNumber, msg.PhoneNumber)
	assert.Equal(t, "outbound", msg.Direction)
	assert.Equal(t, "text", msg.MessageType)
	assert.Equal(t, content, msg.Content)
	assert.Equal(t, "sent", msg.Status)
}

func TestLogMessageError(t *testing.T) {
	db := setupAuditTestDB(t)
	service := NewWhatsAppAuditService(db)

	messageID := "test_msg_789"
	phoneNumber := "+1234567890"
	errorCode := "RATE_LIMIT"
	errorMessage := "Rate limit exceeded"

	// First log an outbound message
	err := service.LogOutboundMessage(messageID, phoneNumber, "text", "Test", "sending")
	assert.NoError(t, err)

	// Then log an error for it
	err = service.LogMessageError(messageID, phoneNumber, errorCode, errorMessage)
	assert.NoError(t, err)

	// Verify the error was logged
	var msg models.WhatsAppMessage
	err = db.Where("message_id = ?", messageID).First(&msg).Error
	assert.NoError(t, err)
	assert.Equal(t, "failed", msg.Status)
	assert.NotNil(t, msg.ErrorCode)
	assert.Equal(t, errorCode, *msg.ErrorCode)
	assert.NotNil(t, msg.ErrorMessage)
	assert.Equal(t, errorMessage, *msg.ErrorMessage)
}

func TestUpdateMessageStatus(t *testing.T) {
	db := setupAuditTestDB(t)
	service := NewWhatsAppAuditService(db)

	messageID := "test_msg_101"
	phoneNumber := "+1234567890"

	// Log an outbound message
	err := service.LogOutboundMessage(messageID, phoneNumber, "text", "Test", "sent")
	assert.NoError(t, err)

	// Update status to delivered
	err = service.UpdateMessageStatus(messageID, "delivered", nil, nil)
	assert.NoError(t, err)

	// Verify the status was updated
	var msg models.WhatsAppMessage
	err = db.Where("message_id = ?", messageID).First(&msg).Error
	assert.NoError(t, err)
	assert.Equal(t, "delivered", msg.Status)
}

func TestCleanupOldMessages(t *testing.T) {
	db := setupAuditTestDB(t)
	service := NewWhatsAppAuditService(db)

	// Create old messages (100 days old)
	oldTime := time.Now().AddDate(0, 0, -100)
	for i := 0; i < 5; i++ {
		msg := models.WhatsAppMessage{
			MessageID:   "old_msg_" + string(rune(i)),
			PhoneNumber: "+1234567890",
			Direction:   "inbound",
			MessageType: "text",
			Content:     "Old message",
			Status:      "received",
			CreatedAt:   oldTime,
		}
		err := db.Create(&msg).Error
		assert.NoError(t, err)
	}

	// Create recent messages (10 days old)
	recentTime := time.Now().AddDate(0, 0, -10)
	for i := 0; i < 3; i++ {
		msg := models.WhatsAppMessage{
			MessageID:   "recent_msg_" + string(rune(i)),
			PhoneNumber: "+1234567890",
			Direction:   "inbound",
			MessageType: "text",
			Content:     "Recent message",
			Status:      "received",
			CreatedAt:   recentTime,
		}
		err := db.Create(&msg).Error
		assert.NoError(t, err)
	}

	// Cleanup messages older than 90 days
	deletedCount, err := service.CleanupOldMessages(90)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), deletedCount)

	// Verify only recent messages remain
	var count int64
	err = db.Model(&models.WhatsAppMessage{}).Count(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestGetMessageStats(t *testing.T) {
	db := setupAuditTestDB(t)
	service := NewWhatsAppAuditService(db)

	now := time.Now()

	// Create various messages
	messages := []models.WhatsAppMessage{
		{MessageID: "msg1", PhoneNumber: "+1", Direction: "inbound", MessageType: "text", Content: "Test", Status: "received", CreatedAt: now},
		{MessageID: "msg2", PhoneNumber: "+1", Direction: "inbound", MessageType: "text", Content: "Test", Status: "received", CreatedAt: now},
		{MessageID: "msg3", PhoneNumber: "+1", Direction: "outbound", MessageType: "text", Content: "Test", Status: "sent", CreatedAt: now},
		{MessageID: "msg4", PhoneNumber: "+1", Direction: "outbound", MessageType: "text", Content: "Test", Status: "delivered", CreatedAt: now},
		{MessageID: "msg5", PhoneNumber: "+1", Direction: "outbound", MessageType: "text", Content: "Test", Status: "read", CreatedAt: now},
		{MessageID: "msg6", PhoneNumber: "+1", Direction: "outbound", MessageType: "text", Content: "Test", Status: "failed", CreatedAt: now},
	}

	for _, msg := range messages {
		err := db.Create(&msg).Error
		assert.NoError(t, err)
	}

	// Get stats
	stats, err := service.GetMessageStats(now.Add(-1 * time.Hour))
	assert.NoError(t, err)
	assert.Equal(t, int64(6), stats.TotalMessages)
	assert.Equal(t, int64(2), stats.InboundMessages)
	assert.Equal(t, int64(4), stats.OutboundMessages)
	assert.Equal(t, int64(1), stats.FailedMessages)
	assert.Equal(t, int64(1), stats.DeliveredMessages)
	assert.Equal(t, int64(1), stats.ReadMessages)
}
