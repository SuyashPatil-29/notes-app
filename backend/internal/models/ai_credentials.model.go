package models

import (
	"time"
)

type AICredential struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"userId" gorm:"index"`
	Provider  string    `json:"provider" gorm:"index"`       // e.g., "openai", "anthropic"
	KeyCipher []byte    `json:"keyCipher" gorm:"type:bytea"` // AES-GCM encrypted
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
