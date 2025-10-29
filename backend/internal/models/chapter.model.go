package models

import (
	"time"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

type Chapter struct {
	ID         string    `json:"id" gorm:"primaryKey;type:varchar(255)"`
	Name       string    `json:"name"`
	NotebookID string    `json:"notebookId" gorm:"type:varchar(255);index"`
	Notebook   Notebook  `json:"notebook" gorm:"foreignKey:NotebookID"`
	Files      []Notes   `json:"notes" gorm:"foreignKey:ChapterID"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// BeforeCreate hook to generate CUID before creating a chapter
func (c *Chapter) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = cuid.New()
	}
	return nil
}
