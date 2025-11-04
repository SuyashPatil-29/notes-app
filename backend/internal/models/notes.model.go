package models

import (
	"time"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

type Notes struct {
	ID                 string     `json:"id" gorm:"primaryKey;type:varchar(255)"`
	Name               string     `json:"name"`
	Content            string     `json:"content" gorm:"type:text"`
	ChapterID          string     `json:"chapterId" gorm:"type:varchar(255);index"`
	OrganizationID     *string    `json:"organizationId,omitempty" gorm:"type:varchar(255);index"`
	Chapter            Chapter    `json:"chapter" gorm:"foreignKey:ChapterID"`
	IsPublic           bool       `json:"isPublic" gorm:"default:false"`
	VideoData          string     `json:"videoData" gorm:"type:text"`
	HasVideo           bool       `json:"hasVideo" gorm:"default:false"`
	MeetingRecordingID *string    `json:"meetingRecordingId,omitempty" gorm:"type:varchar(255)"`
	AISummary          string     `json:"aiSummary,omitempty" gorm:"type:text"`
	TranscriptRaw      string     `json:"transcriptRaw,omitempty" gorm:"type:text"`
	TaskBoard          *TaskBoard `json:"taskBoard,omitempty" gorm:"foreignKey:NoteID"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

// BeforeCreate hook to generate CUID before creating a note
func (n *Notes) BeforeCreate(tx *gorm.DB) error {
	if n.ID == "" {
		n.ID = cuid.New()
	}
	return nil
}
