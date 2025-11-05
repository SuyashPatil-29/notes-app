package models

import (
	"time"
)

// NoteLink represents a relationship between two notes
type NoteLink struct {
	ID             string    `json:"id" gorm:"primaryKey;type:varchar(255)"`
	SourceNoteID   string    `json:"sourceNoteId" gorm:"type:varchar(255);not null;index"`
	TargetNoteID   string    `json:"targetNoteId" gorm:"type:varchar(255);not null;index"`
	LinkType       string    `json:"linkType" gorm:"type:varchar(50);default:'references'"`
	OrganizationID *string   `json:"organizationId,omitempty" gorm:"type:varchar(255);index"`
	CreatedBy      string    `json:"createdBy" gorm:"type:varchar(255);not null;index"`
	SourceNote     *Notes    `json:"sourceNote,omitempty" gorm:"foreignKey:SourceNoteID"`
	TargetNote     *Notes    `json:"targetNote,omitempty" gorm:"foreignKey:TargetNoteID"`
	CreatedAt      time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName specifies the table name for NoteLink
func (NoteLink) TableName() string {
	return "note_links"
}

// LinkType constants
const (
	LinkTypeReferences   = "references"
	LinkTypeBuildsOn     = "builds-on"
	LinkTypeContradicts  = "contradicts"
	LinkTypeRelated      = "related"
	LinkTypePrerequisite = "prerequisite"
)

// ValidLinkTypes returns a list of valid link types
func ValidLinkTypes() []string {
	return []string{
		LinkTypeReferences,
		LinkTypeBuildsOn,
		LinkTypeContradicts,
		LinkTypeRelated,
		LinkTypePrerequisite,
	}
}
