package whatsapp

import (
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
)

// ErrorCategory represents the category of WhatsApp error
type ErrorCategory string

const (
	ErrorCategoryAuth        ErrorCategory = "authentication"
	ErrorCategoryValidation  ErrorCategory = "validation"
	ErrorCategoryNotFound    ErrorCategory = "not_found"
	ErrorCategoryPermission  ErrorCategory = "permission"
	ErrorCategoryRateLimit   ErrorCategory = "rate_limit"
	ErrorCategoryInternal    ErrorCategory = "internal"
	ErrorCategoryWhatsAppAPI ErrorCategory = "whatsapp_api"
	ErrorCategoryContext     ErrorCategory = "context"
	ErrorCategoryCommand     ErrorCategory = "command"
)

// WhatsAppError represents a structured error for WhatsApp operations
type WhatsAppError struct {
	Category      ErrorCategory
	Message       string
	UserMessage   string
	Retryable     bool
	OriginalError error
}

// Error implements the error interface
func (e *WhatsAppError) Error() string {
	if e.OriginalError != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Category, e.Message, e.OriginalError)
	}
	return fmt.Sprintf("[%s] %s", e.Category, e.Message)
}

// Unwrap returns the original error
func (e *WhatsAppError) Unwrap() error {
	return e.OriginalError
}

// Common error instances
var (
	ErrUnauthorized = &WhatsAppError{
		Category:    ErrorCategoryAuth,
		Message:     "User is not authenticated",
		UserMessage: "You need to authenticate first. Please use /help to learn how to get started.",
		Retryable:   false,
	}

	ErrInvalidToken = &WhatsAppError{
		Category:    ErrorCategoryAuth,
		Message:     "Invalid or expired authentication token",
		UserMessage: "Your session has expired. Please authenticate again.",
		Retryable:   false,
	}

	ErrPermissionDenied = &WhatsAppError{
		Category:    ErrorCategoryPermission,
		Message:     "User does not have permission to perform this action",
		UserMessage: "You don't have permission to perform this action.",
		Retryable:   false,
	}

	ErrNotFound = &WhatsAppError{
		Category:    ErrorCategoryNotFound,
		Message:     "Resource not found",
		UserMessage: "The requested item was not found.",
		Retryable:   false,
	}

	ErrRateLimitExceeded = &WhatsAppError{
		Category:    ErrorCategoryRateLimit,
		Message:     "Rate limit exceeded",
		UserMessage: "You're sending messages too quickly. Please wait a moment and try again.",
		Retryable:   true,
	}

	ErrInvalidInput = &WhatsAppError{
		Category:    ErrorCategoryValidation,
		Message:     "Invalid input provided",
		UserMessage: "The input you provided is invalid. Please check and try again.",
		Retryable:   false,
	}

	ErrContextExpired = &WhatsAppError{
		Category:    ErrorCategoryContext,
		Message:     "Conversation context has expired",
		UserMessage: "Your previous conversation has expired. Please start a new command.",
		Retryable:   false,
	}

	ErrCommandNotFound = &WhatsAppError{
		Category:    ErrorCategoryCommand,
		Message:     "Command not recognized",
		UserMessage: "I don't recognize that command. Use /help to see available commands.",
		Retryable:   false,
	}

	ErrInternalError = &WhatsAppError{
		Category:    ErrorCategoryInternal,
		Message:     "Internal server error",
		UserMessage: "Something went wrong on our end. Please try again later.",
		Retryable:   true,
	}

	ErrWhatsAppAPI = &WhatsAppError{
		Category:    ErrorCategoryWhatsAppAPI,
		Message:     "WhatsApp API error",
		UserMessage: "Unable to send message. Please try again later.",
		Retryable:   true,
	}
)

// NewWhatsAppError creates a new WhatsAppError with the given parameters
func NewWhatsAppError(category ErrorCategory, message, userMessage string, retryable bool, originalErr error) *WhatsAppError {
	return &WhatsAppError{
		Category:      category,
		Message:       message,
		UserMessage:   userMessage,
		Retryable:     retryable,
		OriginalError: originalErr,
	}
}

// WrapError wraps an existing error with WhatsApp error context
func WrapError(category ErrorCategory, message, userMessage string, retryable bool, err error) *WhatsAppError {
	return &WhatsAppError{
		Category:      category,
		Message:       message,
		UserMessage:   userMessage,
		Retryable:     retryable,
		OriginalError: err,
	}
}

// ConvertToWhatsAppError converts a standard error to a WhatsAppError
// If the error is already a WhatsAppError, it returns it as-is
func ConvertToWhatsAppError(err error) *WhatsAppError {
	if err == nil {
		return nil
	}

	// Check if it's already a WhatsAppError
	var whatsappErr *WhatsAppError
	if errors.As(err, &whatsappErr) {
		return whatsappErr
	}

	// Check for common error patterns and convert them
	errMsg := err.Error()

	// Authentication errors
	if errors.Is(err, ErrUnauthorized) || containsAny(errMsg, "unauthorized", "not authenticated") {
		return WrapError(ErrorCategoryAuth, "Authentication failed", ErrUnauthorized.UserMessage, false, err)
	}

	// Permission errors
	if errors.Is(err, ErrPermissionDenied) || containsAny(errMsg, "permission denied", "forbidden", "not allowed") {
		return WrapError(ErrorCategoryPermission, "Permission denied", ErrPermissionDenied.UserMessage, false, err)
	}

	// Not found errors
	if errors.Is(err, ErrNotFound) || containsAny(errMsg, "not found", "does not exist") {
		return WrapError(ErrorCategoryNotFound, "Resource not found", ErrNotFound.UserMessage, false, err)
	}

	// Validation errors
	if containsAny(errMsg, "invalid", "validation", "required") {
		return WrapError(ErrorCategoryValidation, "Validation error", ErrInvalidInput.UserMessage, false, err)
	}

	// Rate limit errors
	if containsAny(errMsg, "rate limit", "too many requests") {
		return WrapError(ErrorCategoryRateLimit, "Rate limit exceeded", ErrRateLimitExceeded.UserMessage, true, err)
	}

	// Default to internal error
	return WrapError(ErrorCategoryInternal, "Internal error", ErrInternalError.UserMessage, true, err)
}

// LogError logs a WhatsApp error with structured fields
func LogError(err error, phoneNumber string, additionalFields map[string]interface{}) {
	whatsappErr := ConvertToWhatsAppError(err)
	if whatsappErr == nil {
		return
	}

	// Create base log event
	logEvent := log.Error().
		Str("category", string(whatsappErr.Category)).
		Str("message", whatsappErr.Message).
		Str("user_message", whatsappErr.UserMessage).
		Bool("retryable", whatsappErr.Retryable).
		Str("phone_number", phoneNumber)

	// Add original error if present
	if whatsappErr.OriginalError != nil {
		logEvent = logEvent.Err(whatsappErr.OriginalError)
	}

	// Add additional fields
	for key, value := range additionalFields {
		logEvent = logEvent.Interface(key, value)
	}

	logEvent.Msg("WhatsApp error occurred")
}

// LogErrorWithContext logs a WhatsApp error with command context
func LogErrorWithContext(err error, ctx *CommandContext, additionalFields map[string]interface{}) {
	fields := make(map[string]interface{})

	// Add context information
	if ctx != nil {
		fields["message"] = ctx.Message
		if len(ctx.Args) > 0 {
			fields["args"] = ctx.Args
		}
		if ctx.User != nil {
			fields["user_id"] = ctx.User.ClerkUserID
			fields["organization_id"] = ctx.User.OrganizationID
		}
		if ctx.GroupID != nil && *ctx.GroupID != "" {
			fields["group_id"] = *ctx.GroupID
		}
	}

	// Merge additional fields
	for key, value := range additionalFields {
		fields[key] = value
	}

	phoneNumber := ""
	if ctx != nil {
		phoneNumber = ctx.PhoneNumber
	}

	LogError(err, phoneNumber, fields)
}

// GetUserMessage extracts the user-friendly message from an error
func GetUserMessage(err error) string {
	if err == nil {
		return ""
	}

	whatsappErr := ConvertToWhatsAppError(err)
	return whatsappErr.UserMessage
}

// IsRetryable checks if an error is retryable
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	whatsappErr := ConvertToWhatsAppError(err)
	return whatsappErr.Retryable
}

// containsAny checks if a string contains any of the given substrings
func containsAny(s string, substrings ...string) bool {
	for _, substr := range substrings {
		if contains(s, substr) {
			return true
		}
	}
	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			c1 := s[i+j]
			c2 := substr[j]
			// Simple case-insensitive comparison
			if c1 != c2 && c1 != c2+32 && c1 != c2-32 {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
