package models

import (
	"time"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

// YjsDocument stores the current state of a collaborative Yjs document
type YjsDocument struct {
	ID        string    `json:"id" gorm:"primaryKey;type:varchar(255)"`
	NoteID    string    `json:"noteId" gorm:"uniqueIndex;type:varchar(255);not null"`
	YjsState  []byte    `json:"-" gorm:"type:bytea;not null"` // Binary Yjs state
	Version   int       `json:"version" gorm:"default:0;not null"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// BeforeCreate hook to generate CUID before creating a Yjs document
func (y *YjsDocument) BeforeCreate(tx *gorm.DB) error {
	if y.ID == "" {
		y.ID = cuid.New()
	}
	return nil
}

// YjsUpdate stores incremental Yjs updates for version history
type YjsUpdate struct {
	ID         uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	NoteID     string    `json:"noteId" gorm:"index;type:varchar(255);not null"`
	UpdateData []byte    `json:"-" gorm:"type:bytea;not null"`
	Clock      int       `json:"clock" gorm:"not null"`
	CreatedAt  time.Time `json:"createdAt"`
}

