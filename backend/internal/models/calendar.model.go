package models

import (
	"time"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

// Calendar represents a user's connected calendar (Google/Microsoft)
type Calendar struct {
	ID                string     `json:"id" gorm:"primaryKey;type:varchar(255)"`
	UserID            uint       `json:"userId" gorm:"not null;index"`
	User              User       `json:"user,omitempty" gorm:"foreignKey:UserID"`
	RecallCalendarID  string     `json:"recallCalendarId" gorm:"uniqueIndex;not null"` // ID from Recall.ai
	Platform          string     `json:"platform" gorm:"not null"`                     // "google_calendar" or "microsoft_outlook"
	PlatformEmail     string     `json:"platformEmail" gorm:"not null"`
	OAuthClientID     string     `json:"oauthClientId" gorm:"column:oauth_client_id;not null"`
	OAuthClientSecret string     `json:"oauthClientSecret" gorm:"column:oauth_client_secret;not null"` // Encrypted
	OAuthRefreshToken string     `json:"oauthRefreshToken" gorm:"column:oauth_refresh_token;not null"` // Encrypted
	Status            string     `json:"status" gorm:"default:'active'"`                               // active, inactive, error
	LastSyncedAt      *time.Time `json:"lastSyncedAt"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

// BeforeCreate hook to generate CUID before creating a calendar
func (c *Calendar) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = cuid.New()
	}
	return nil
}

// CalendarEvent represents an event/meeting from a user's calendar
type CalendarEvent struct {
	ID                 string            `json:"id" gorm:"primaryKey;type:varchar(255)"`
	CalendarID         string            `json:"calendarId" gorm:"not null;index"`
	Calendar           Calendar          `json:"calendar,omitempty" gorm:"foreignKey:CalendarID"`
	RecallEventID      string            `json:"recallEventId" gorm:"uniqueIndex;not null"` // ID from Recall.ai
	ICalUID            string            `json:"iCalUid"`
	PlatformID         string            `json:"platformId"`
	MeetingPlatform    string            `json:"meetingPlatform"` // zoom, google_meet, microsoft_teams, etc.
	MeetingURL         string            `json:"meetingUrl"`
	Title              string            `json:"title"`
	StartTime          time.Time         `json:"startTime" gorm:"index"`
	EndTime            time.Time         `json:"endTime"`
	IsDeleted          bool              `json:"isDeleted" gorm:"default:false"`
	BotScheduled       bool              `json:"botScheduled" gorm:"default:false"`
	BotID              *string           `json:"botId,omitempty"`
	MeetingRecordingID *string           `json:"meetingRecordingId,omitempty"`
	MeetingRecording   *MeetingRecording `json:"meetingRecording,omitempty" gorm:"foreignKey:MeetingRecordingID"`
	CreatedAt          time.Time         `json:"createdAt"`
	UpdatedAt          time.Time         `json:"updatedAt"`
}

// BeforeCreate hook to generate CUID before creating a calendar event
func (ce *CalendarEvent) BeforeCreate(tx *gorm.DB) error {
	if ce.ID == "" {
		ce.ID = cuid.New()
	}
	return nil
}

// CalendarOAuthState represents temporary OAuth state for security
type CalendarOAuthState struct {
	ID        string    `json:"id" gorm:"primaryKey;type:varchar(255)"`
	UserID    uint      `json:"userId" gorm:"not null;index"`
	State     string    `json:"state" gorm:"uniqueIndex;not null"`
	Platform  string    `json:"platform" gorm:"not null"` // "google" or "microsoft"
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt" gorm:"index"`
}

// BeforeCreate hook to generate CUID before creating OAuth state
func (cos *CalendarOAuthState) BeforeCreate(tx *gorm.DB) error {
	if cos.ID == "" {
		cos.ID = cuid.New()
	}
	return nil
}
