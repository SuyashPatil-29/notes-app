package dto

import "time"

// NoteListItem represents a lightweight note for list views (without content)
type NoteListItem struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	ChapterID          string    `json:"chapterId"`
	OrganizationID     *string   `json:"organizationId,omitempty"`
	IsPublic           bool      `json:"isPublic"`
	HasVideo           bool      `json:"hasVideo"`
	MeetingRecordingID *string   `json:"meetingRecordingId,omitempty"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

// NoteDetail represents a complete note with all fields for detail views
type NoteDetail struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	Content            string    `json:"content"`
	ChapterID          string    `json:"chapterId"`
	OrganizationID     *string   `json:"organizationId,omitempty"`
	IsPublic           bool      `json:"isPublic"`
	VideoData          string    `json:"videoData"`
	HasVideo           bool      `json:"hasVideo"`
	MeetingRecordingID *string   `json:"meetingRecordingId,omitempty"`
	AISummary          string    `json:"aiSummary,omitempty"`
	TranscriptRaw      string    `json:"transcriptRaw,omitempty"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}
