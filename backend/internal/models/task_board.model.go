package models

import (
	"time"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

type TaskBoard struct {
	ID             string    `json:"id" gorm:"primaryKey;type:varchar(255)"`
	Name           string    `json:"name"`
	Description    string    `json:"description" gorm:"type:text"`
	NoteID         *string   `json:"noteId,omitempty" gorm:"type:varchar(255);index"`
	ClerkUserID    string    `json:"clerkUserId" gorm:"type:varchar(255);index"`
	OrganizationID *string   `json:"organizationId,omitempty" gorm:"type:varchar(255);index"`
	IsStandalone   bool      `json:"isStandalone" gorm:"default:false"`
	Tasks          []Task    `json:"tasks" gorm:"foreignKey:TaskBoardID"`
	Note           *Notes    `json:"note,omitempty" gorm:"foreignKey:NoteID"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// BeforeCreate hook to generate CUID before creating a task board
func (tb *TaskBoard) BeforeCreate(tx *gorm.DB) error {
	if tb.ID == "" {
		tb.ID = cuid.New()
	}
	return nil
}
