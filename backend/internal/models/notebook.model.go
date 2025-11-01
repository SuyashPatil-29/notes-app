package models

import (
	"time"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

type Notebook struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(255)"`
	Name        string    `json:"name"`
	ClerkUserID string    `json:"clerkUserId" gorm:"not null;index"`
	Chapters    []Chapter `json:"chapters" gorm:"foreignKey:NotebookID"`
	IsPublic    bool      `json:"isPublic" gorm:"default:false"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// BeforeCreate hook to generate CUID before creating a notebook
func (n *Notebook) BeforeCreate(tx *gorm.DB) error {
	if n.ID == "" {
		n.ID = cuid.New()
	}
	return nil
}
