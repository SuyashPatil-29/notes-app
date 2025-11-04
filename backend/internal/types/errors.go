package types

import (
	"net/http"
	"strings"
)

// ErrorCode represents specific error codes for API responses
type ErrorCode string

const (
	// Authentication and Authorization errors
	ErrorCodeUnauthorized ErrorCode = "UNAUTHORIZED"
	ErrorCodeForbidden    ErrorCode = "FORBIDDEN"
	ErrorCodeInvalidToken ErrorCode = "INVALID_TOKEN"

	// Validation errors
	ErrorCodeValidationFailed ErrorCode = "VALIDATION_FAILED"
	ErrorCodeInvalidProvider  ErrorCode = "INVALID_PROVIDER"
	ErrorCodeInvalidAPIKey    ErrorCode = "INVALID_API_KEY"
	ErrorCodeMissingFields    ErrorCode = "MISSING_FIELDS"

	// Resource errors
	ErrorCodeNotFound      ErrorCode = "NOT_FOUND"
	ErrorCodeAlreadyExists ErrorCode = "ALREADY_EXISTS"
	ErrorCodeConflict      ErrorCode = "CONFLICT"

	// API Key specific errors
	ErrorCodeAPIKeyNotFound      ErrorCode = "API_KEY_NOT_FOUND"
	ErrorCodeAPIKeyInvalid       ErrorCode = "API_KEY_INVALID"
	ErrorCodeAPIKeyExpired       ErrorCode = "API_KEY_EXPIRED"
	ErrorCodeAPIKeyQuotaExceeded ErrorCode = "API_KEY_QUOTA_EXCEEDED"
	ErrorCodeAPIKeyRateLimit     ErrorCode = "API_KEY_RATE_LIMIT"
	ErrorCodeEncryptionFailed    ErrorCode = "ENCRYPTION_FAILED"
	ErrorCodeDecryptionFailed    ErrorCode = "DECRYPTION_FAILED"

	// Organization errors
	ErrorCodeOrgNotFound      ErrorCode = "ORGANIZATION_NOT_FOUND"
	ErrorCodeOrgAccessDenied  ErrorCode = "ORGANIZATION_ACCESS_DENIED"
	ErrorCodeOrgAdminRequired ErrorCode = "ORGANIZATION_ADMIN_REQUIRED"

	// Internal errors
	ErrorCodeInternalError ErrorCode = "INTERNAL_ERROR"
	ErrorCodeDatabaseError ErrorCode = "DATABASE_ERROR"
	ErrorCodeServiceError  ErrorCode = "SERVICE_ERROR"
)

// APIError represents a structured API error response
type APIError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	Suggestion string    `json:"suggestion,omitempty"`
	HTTPStatus int       `json:"-"`
}

// Error implements the error interface
func (e APIError) Error() string {
	return e.Message
}

// NewAPIError creates a new APIError
func NewAPIError(code ErrorCode, message string, httpStatus int) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// WithDetails adds details to the error
func (e *APIError) WithDetails(details string) *APIError {
	e.Details = details
	return e
}

// WithSuggestion adds a suggestion to the error
func (e *APIError) WithSuggestion(suggestion string) *APIError {
	e.Suggestion = suggestion
	return e
}

// Common API errors for organization API key management
var (
	ErrUnauthorized = NewAPIError(
		ErrorCodeUnauthorized,
		"Authentication required",
		http.StatusUnauthorized,
	).WithSuggestion("Please log in and try again")

	ErrOrgAdminRequired = NewAPIError(
		ErrorCodeOrgAdminRequired,
		"Organization admin privileges required",
		http.StatusForbidden,
	).WithSuggestion("Only organization administrators can manage API keys")

	ErrInvalidProvider = NewAPIError(
		ErrorCodeInvalidProvider,
		"Invalid AI provider",
		http.StatusBadRequest,
	).WithSuggestion("Supported providers are: openai, anthropic, google")

	ErrAPIKeyNotFound = NewAPIError(
		ErrorCodeAPIKeyNotFound,
		"API key not found",
		http.StatusNotFound,
	).WithSuggestion("The API key for this provider has not been configured")

	ErrAPIKeyInvalid = NewAPIError(
		ErrorCodeAPIKeyInvalid,
		"Invalid API key format",
		http.StatusBadRequest,
	).WithSuggestion("Please check your API key format and try again")

	ErrMissingFields = NewAPIError(
		ErrorCodeMissingFields,
		"Required fields are missing",
		http.StatusBadRequest,
	).WithSuggestion("Please provide all required fields")

	ErrEncryptionFailed = NewAPIError(
		ErrorCodeEncryptionFailed,
		"Failed to encrypt API key",
		http.StatusInternalServerError,
	).WithSuggestion("Please try again or contact support if the issue persists")

	ErrDecryptionFailed = NewAPIError(
		ErrorCodeDecryptionFailed,
		"Failed to decrypt API key",
		http.StatusInternalServerError,
	).WithSuggestion("The stored API key may be corrupted. Please reconfigure it")

	ErrDatabaseError = NewAPIError(
		ErrorCodeDatabaseError,
		"Database operation failed",
		http.StatusInternalServerError,
	).WithSuggestion("Please try again or contact support if the issue persists")

	ErrInternalError = NewAPIError(
		ErrorCodeInternalError,
		"An internal error occurred",
		http.StatusInternalServerError,
	).WithSuggestion("Please try again or contact support if the issue persists")
)

// GetAPIKeyErrorFromMessage determines the appropriate API key error based on error message
func GetAPIKeyErrorFromMessage(err error) *APIError {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	// Check for specific error patterns
	switch {
	case contains(errMsg, "401", "unauthorized", "invalid_api_key", "incorrect api key"):
		return NewAPIError(
			ErrorCodeAPIKeyInvalid,
			"Invalid API key",
			http.StatusBadRequest,
		).WithDetails(errMsg).WithSuggestion("Please check your API key and ensure it's correct")

	case contains(errMsg, "insufficient_quota", "quota"):
		return NewAPIError(
			ErrorCodeAPIKeyQuotaExceeded,
			"API quota exceeded",
			http.StatusPaymentRequired,
		).WithDetails(errMsg).WithSuggestion("Please check your API usage limits or upgrade your plan")

	case contains(errMsg, "rate_limit", "rate limit"):
		return NewAPIError(
			ErrorCodeAPIKeyRateLimit,
			"Rate limit exceeded",
			http.StatusTooManyRequests,
		).WithDetails(errMsg).WithSuggestion("Please wait a moment and try again")

	case contains(errMsg, "model_not_found", "model not found"):
		return NewAPIError(
			ErrorCodeValidationFailed,
			"Model not available",
			http.StatusBadRequest,
		).WithDetails(errMsg).WithSuggestion("Please try a different model")

	case contains(errMsg, "not found"):
		return ErrAPIKeyNotFound.WithDetails(errMsg)

	case contains(errMsg, "encrypt"):
		return ErrEncryptionFailed.WithDetails(errMsg)

	case contains(errMsg, "decrypt"):
		return ErrDecryptionFailed.WithDetails(errMsg)

	default:
		return ErrInternalError.WithDetails(errMsg)
	}
}

// contains checks if any of the substrings are present in the main string (case-insensitive)
func contains(str string, substrings ...string) bool {
	lowerStr := strings.ToLower(str)
	for _, substr := range substrings {
		if strings.Contains(lowerStr, strings.ToLower(substr)) {
			return true
		}
	}
	return false
}
