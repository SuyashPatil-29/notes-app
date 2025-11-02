package dto

import "time"

// MeetingListItem represents a lightweight meeting for list views (without generated note details)
type MeetingListItem struct {
	ID                    string     `json:"id"`
	ClerkUserID           string     `json:"clerkUserId"`
	BotID                 string     `json:"botId"`
	MeetingURL            string     `json:"meetingUrl"`
	Status                string     `json:"status"`
	RecallRecordingID     string     `json:"recallRecordingId,omitempty"`
	TranscriptDownloadURL string     `json:"transcriptDownloadUrl,omitempty"`
	VideoDownloadURL      string     `json:"videoDownloadUrl,omitempty"`
	GeneratedNoteID       *string    `json:"generatedNoteId,omitempty"`
	CreatedAt             time.Time  `json:"createdAt"`
	UpdatedAt             time.Time  `json:"updatedAt"`
	CompletedAt           *time.Time `json:"completedAt,omitempty"`
}
