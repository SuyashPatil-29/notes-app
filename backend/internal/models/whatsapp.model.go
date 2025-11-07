package models

import (
	"time"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

// WhatsAppUser stores the mapping between WhatsApp phone numbers and application user accounts
type WhatsAppUser struct {
	ID              string    `json:"id" gorm:"primaryKey;type:varchar(255)"`
	PhoneNumber     string    `json:"phoneNumber" gorm:"uniqueIndex;not null;type:varchar(20)"`
	ClerkUserID     string    `json:"clerkUserId" gorm:"index;not null;type:varchar(255)"`
	OrganizationID  *string   `json:"organizationId,omitempty" gorm:"type:varchar(255);index"`
	IsAuthenticated bool      `json:"isAuthenticated" gorm:"default:false"`
	AuthToken       string    `json:"authToken,omitempty" gorm:"type:varchar(500)"`
	LastActiveAt    time.Time `json:"lastActiveAt"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// BeforeCreate hook to generate CUID before creating a WhatsApp user
func (w *WhatsAppUser) BeforeCreate(tx *gorm.DB) error {
	if w.ID == "" {
		w.ID = cuid.New()
	}
	return nil
}

// WhatsAppConversationContext manages multi-step conversation state for command execution
type WhatsAppConversationContext struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(255)"`
	PhoneNumber string    `json:"phoneNumber" gorm:"index;not null;type:varchar(20)"`
	Command     string    `json:"command" gorm:"not null;type:varchar(50)"`
	Step        int       `json:"step" gorm:"default:0"`
	Data        string    `json:"data" gorm:"type:text"` // JSON string, will be marshaled/unmarshaled
	ExpiresAt   time.Time `json:"expiresAt"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// BeforeCreate hook to generate CUID before creating a conversation context
func (w *WhatsAppConversationContext) BeforeCreate(tx *gorm.DB) error {
	if w.ID == "" {
		w.ID = cuid.New()
	}
	return nil
}

// WhatsAppGroupLink links WhatsApp groups to organizations for group mode functionality
type WhatsAppGroupLink struct {
	ID             string    `json:"id" gorm:"primaryKey;type:varchar(255)"`
	GroupID        string    `json:"groupId" gorm:"uniqueIndex;not null;type:varchar(255)"`
	OrganizationID string    `json:"organizationId" gorm:"index;not null;type:varchar(255)"`
	LinkedBy       string    `json:"linkedBy" gorm:"not null;type:varchar(255)"` // ClerkUserID of admin
	IsActive       bool      `json:"isActive" gorm:"default:true"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// BeforeCreate hook to generate CUID before creating a group link
func (w *WhatsAppGroupLink) BeforeCreate(tx *gorm.DB) error {
	if w.ID == "" {
		w.ID = cuid.New()
	}
	return nil
}

// WhatsAppMessage provides audit log for all WhatsApp interactions
type WhatsAppMessage struct {
	ID           string    `json:"id" gorm:"primaryKey;type:varchar(255)"`
	MessageID    string    `json:"messageId" gorm:"uniqueIndex;type:varchar(255)"` // WhatsApp message ID
	PhoneNumber  string    `json:"phoneNumber" gorm:"index;type:varchar(20)"`
	Direction    string    `json:"direction" gorm:"type:varchar(10)"`   // "inbound" or "outbound"
	MessageType  string    `json:"messageType" gorm:"type:varchar(20)"` // "text", "image", etc.
	Content      string    `json:"content" gorm:"type:text"`
	Status       string    `json:"status" gorm:"type:varchar(20)"` // "sent", "delivered", "read", "failed"
	ErrorCode    *string   `json:"errorCode,omitempty" gorm:"type:varchar(50)"`
	ErrorMessage *string   `json:"errorMessage,omitempty" gorm:"type:text"`
	CreatedAt    time.Time `json:"createdAt"`
}

// BeforeCreate hook to generate CUID before creating a message
func (w *WhatsAppMessage) BeforeCreate(tx *gorm.DB) error {
	if w.ID == "" {
		w.ID = cuid.New()
	}
	return nil
}
