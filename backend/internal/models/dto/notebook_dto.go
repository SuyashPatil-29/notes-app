package dto

import "time"

// NotebookListItem represents a lightweight notebook for list views
type NotebookListItem struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	ClerkUserID    string    `json:"clerkUserId"`
	OrganizationID *string   `json:"organizationId,omitempty"`
	IsPublic       bool      `json:"isPublic"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	ChapterCount   int       `json:"chapterCount"`
}
