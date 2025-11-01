package models

import (
	"time"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

type MeetingRecording struct {
	ID                    string     `json:"id" gorm:"primaryKey;type:varchar(255)"`
	UserID                uint       `json:"userId" gorm:"not null;index"`
	User                  User       `json:"user,omitempty" gorm:"foreignKey:UserID"`
	BotID                 string     `json:"botId" gorm:"uniqueIndex;not null"`
	MeetingURL            string     `json:"meetingUrl" gorm:"not null"`
	Status                string     `json:"status" gorm:"default:'pending'"` // pending, recording, processing, completed, failed
	RecallRecordingID     string     `json:"recallRecordingId,omitempty"`
	TranscriptDownloadURL string     `json:"transcriptDownloadUrl,omitempty"`
	VideoDownloadURL      string     `json:"videoDownloadUrl,omitempty"`
	GeneratedNoteID       *string    `json:"generatedNoteId,omitempty" gorm:"type:varchar(255)"`
	GeneratedNote         *Notes     `json:"generatedNote,omitempty" gorm:"foreignKey:GeneratedNoteID;references:ID"`
	CreatedAt             time.Time  `json:"createdAt"`
	UpdatedAt             time.Time  `json:"updatedAt"`
	CompletedAt           *time.Time `json:"completedAt,omitempty"`
}

// BeforeCreate hook to generate CUID before creating a meeting recording
func (m *MeetingRecording) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = cuid.New()
	}
	return nil
}
