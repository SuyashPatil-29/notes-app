package dto

import "time"

// ChapterListItem represents a lightweight chapter for list views
type ChapterListItem struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	NotebookID     string    `json:"notebookId"`
	OrganizationID *string   `json:"organizationId,omitempty"`
	IsPublic       bool      `json:"isPublic"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	NoteCount      int       `json:"noteCount"`
}
