package utils

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	// MaxMessageLength is the maximum allowed message length
	MaxMessageLength = 4096
	// MaxCommandArgLength is the maximum length for command arguments
	MaxCommandArgLength = 500
	// MaxPhoneNumberLength is the maximum length for phone numbers
	MaxPhoneNumberLength = 20
)

var (
	// phoneNumberRegex validates international phone numbers
	// Accepts formats like: +1234567890, 1234567890, +1-234-567-8900
	phoneNumberRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)

	// sanitizeRegex removes potentially dangerous characters
	sanitizeRegex = regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidatePhoneNumber validates a phone number format
func ValidatePhoneNumber(phoneNumber string) error {
	if phoneNumber == "" {
		return &ValidationError{
			Field:   "phone_number",
			Message: "phone number cannot be empty",
		}
	}

	if len(phoneNumber) > MaxPhoneNumberLength {
		return &ValidationError{
			Field:   "phone_number",
			Message: fmt.Sprintf("phone number exceeds maximum length of %d", MaxPhoneNumberLength),
		}
	}

	// Remove common formatting characters for validation
	cleaned := strings.ReplaceAll(phoneNumber, "-", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")

	if !phoneNumberRegex.MatchString(cleaned) {
		return &ValidationError{
			Field:   "phone_number",
			Message: "invalid phone number format",
		}
	}

	return nil
}

// SanitizeMessage sanitizes message content by removing control characters
// and enforcing length limits
func SanitizeMessage(message string) (string, error) {
	if message == "" {
		return "", &ValidationError{
			Field:   "message",
			Message: "message cannot be empty",
		}
	}

	// Remove control characters
	sanitized := sanitizeRegex.ReplaceAllString(message, "")

	// Trim whitespace
	sanitized = strings.TrimSpace(sanitized)

	if sanitized == "" {
		return "", &ValidationError{
			Field:   "message",
			Message: "message contains only invalid characters",
		}
	}

	// Check length
	if utf8.RuneCountInString(sanitized) > MaxMessageLength {
		return "", &ValidationError{
			Field:   "message",
			Message: fmt.Sprintf("message exceeds maximum length of %d characters", MaxMessageLength),
		}
	}

	return sanitized, nil
}

// ValidateCommandArgument validates a command argument
func ValidateCommandArgument(argName, argValue string) error {
	if argValue == "" {
		return &ValidationError{
			Field:   argName,
			Message: "argument cannot be empty",
		}
	}

	// Check length
	if utf8.RuneCountInString(argValue) > MaxCommandArgLength {
		return &ValidationError{
			Field:   argName,
			Message: fmt.Sprintf("argument exceeds maximum length of %d characters", MaxCommandArgLength),
		}
	}

	// Remove control characters for validation
	sanitized := sanitizeRegex.ReplaceAllString(argValue, "")
	if strings.TrimSpace(sanitized) == "" {
		return &ValidationError{
			Field:   argName,
			Message: "argument contains only invalid characters",
		}
	}

	return nil
}

// SanitizeCommandArgument sanitizes a command argument
func SanitizeCommandArgument(argValue string) string {
	// Remove control characters
	sanitized := sanitizeRegex.ReplaceAllString(argValue, "")

	// Trim whitespace
	sanitized = strings.TrimSpace(sanitized)

	return sanitized
}

// ValidateAndSanitizeMessage validates and sanitizes a message in one step
func ValidateAndSanitizeMessage(message string) (string, error) {
	return SanitizeMessage(message)
}

// ValidateEntityType validates entity type names
func ValidateEntityType(entityType string) error {
	if entityType == "" {
		return &ValidationError{
			Field:   "entity_type",
			Message: "entity type cannot be empty",
		}
	}

	// Entity types should be alphanumeric with underscores
	validEntityType := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)
	if !validEntityType.MatchString(entityType) {
		return &ValidationError{
			Field:   "entity_type",
			Message: "entity type must start with a letter and contain only letters, numbers, and underscores",
		}
	}

	if len(entityType) > 50 {
		return &ValidationError{
			Field:   "entity_type",
			Message: "entity type exceeds maximum length of 50 characters",
		}
	}

	return nil
}

// ValidateEntityID validates entity IDs
func ValidateEntityID(entityID string) error {
	if entityID == "" {
		return &ValidationError{
			Field:   "entity_id",
			Message: "entity ID cannot be empty",
		}
	}

	if len(entityID) > 100 {
		return &ValidationError{
			Field:   "entity_id",
			Message: "entity ID exceeds maximum length of 100 characters",
		}
	}

	return nil
}

// ValidateNoteContent validates note content
func ValidateNoteContent(content string) error {
	if content == "" {
		return &ValidationError{
			Field:   "note_content",
			Message: "note content cannot be empty",
		}
	}

	if utf8.RuneCountInString(content) > MaxMessageLength {
		return &ValidationError{
			Field:   "note_content",
			Message: fmt.Sprintf("note content exceeds maximum length of %d characters", MaxMessageLength),
		}
	}

	return nil
}
