package models

// User model is no longer needed as Clerk is the source of truth
// Keeping this file for reference of what data Clerk stores

// AuthenticatedUser represents user data fetched from Clerk
// This is not a database model, just a DTO for API responses
type AuthenticatedUser struct {
	ClerkUserID         string  `json:"clerkUserId"`
	Name                string  `json:"name"`
	Email               string  `json:"email"`
	ImageUrl            *string `json:"imageUrl"`
	OnboardingCompleted bool    `json:"onboardingCompleted"`
	HasApiKey           bool    `json:"hasApiKey"`
}
