package models

import "time"

// OrganizationMember represents a cached membership record
// This is optional - can be used to cache Clerk org membership for performance
// Otherwise, we fetch membership info directly from Clerk API on each request
type OrganizationMember struct {
	OrganizationID string    `json:"organizationId" gorm:"primaryKey;type:varchar(255)"`
	ClerkUserID    string    `json:"clerkUserId" gorm:"primaryKey;type:varchar(255)"`
	Role           string    `json:"role" gorm:"type:varchar(50)"` // "admin" or "member"
	JoinedAt       time.Time `json:"joinedAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
