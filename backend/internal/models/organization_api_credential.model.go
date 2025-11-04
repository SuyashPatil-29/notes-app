package models

import (
	"time"
)

// OrganizationAPICredential represents API credentials configured at the organization level
// These credentials are used by all organization members for AI requests
type OrganizationAPICredential struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	OrganizationID string    `json:"organizationId" gorm:"not null;index;type:varchar(255);uniqueIndex:idx_org_provider"`
	Provider       string    `json:"provider" gorm:"not null;index;type:varchar(50);uniqueIndex:idx_org_provider"` // e.g., "openai", "anthropic", "google"
	KeyCipher      []byte    `json:"keyCipher" gorm:"not null;type:bytea"`                                         // AES-GCM encrypted API key
	CreatedBy      string    `json:"createdBy" gorm:"not null;type:varchar(255)"`                                  // Clerk user ID of the admin who created this
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`

	// Ensure unique combination of organization and provider
	// This prevents duplicate API keys for the same provider within an organization
}

// TableName specifies the table name for GORM
func (OrganizationAPICredential) TableName() string {
	return "organization_api_credentials"
}
