package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

// WhatsAppConfig holds all WhatsApp Cloud API configuration
type WhatsAppConfig struct {
	PhoneNumberID        string
	BusinessAccountID    string
	AccessToken          string
	AppSecret            string
	VerifyToken          string
	APIURL               string
	FrontendURL          string
	ContextExpiration    time.Duration
	RateLimitPerMin      int
	AuditRetentionDays   int
	AuditCleanupInterval int // hours
}

// LoadWhatsAppConfig loads WhatsApp configuration from environment variables
func LoadWhatsAppConfig() (*WhatsAppConfig, error) {
	config := &WhatsAppConfig{
		PhoneNumberID:        os.Getenv("WHATSAPP_PHONE_NUMBER_ID"),
		BusinessAccountID:    os.Getenv("WHATSAPP_BUSINESS_ACCOUNT_ID"),
		AccessToken:          os.Getenv("WHATSAPP_ACCESS_TOKEN"),
		AppSecret:            os.Getenv("WHATSAPP_APP_SECRET"),
		VerifyToken:          os.Getenv("WHATSAPP_WEBHOOK_VERIFY_TOKEN"),
		APIURL:               getEnvOrDefault("WHATSAPP_API_URL", "https://graph.facebook.com/v18.0"),
		FrontendURL:          getEnvOrDefault("FRONTEND_URL", "http://localhost:5173"),
		ContextExpiration:    time.Duration(getEnvIntOrDefault("WHATSAPP_CONTEXT_EXPIRATION", 10)) * time.Minute,
		RateLimitPerMin:      getEnvIntOrDefault("WHATSAPP_RATE_LIMIT_PER_MINUTE", 10),
		AuditRetentionDays:   getEnvIntOrDefault("WHATSAPP_AUDIT_RETENTION_DAYS", 90),
		AuditCleanupInterval: getEnvIntOrDefault("WHATSAPP_AUDIT_CLEANUP_INTERVAL_HOURS", 24),
	}

	// Validate required fields
	if err := config.Validate(); err != nil {
		return nil, err
	}

	log.Info().
		Str("api_url", config.APIURL).
		Dur("context_expiration", config.ContextExpiration).
		Int("rate_limit", config.RateLimitPerMin).
		Int("audit_retention_days", config.AuditRetentionDays).
		Int("audit_cleanup_interval_hours", config.AuditCleanupInterval).
		Msg("WhatsApp configuration loaded successfully")

	return config, nil
}

// Validate checks if all required configuration values are present
func (c *WhatsAppConfig) Validate() error {
	if c.PhoneNumberID == "" {
		return fmt.Errorf("WHATSAPP_PHONE_NUMBER_ID is required")
	}
	if c.BusinessAccountID == "" {
		return fmt.Errorf("WHATSAPP_BUSINESS_ACCOUNT_ID is required")
	}
	if c.AccessToken == "" {
		return fmt.Errorf("WHATSAPP_ACCESS_TOKEN is required")
	}
	if c.AppSecret == "" {
		return fmt.Errorf("WHATSAPP_APP_SECRET is required")
	}
	if c.VerifyToken == "" {
		return fmt.Errorf("WHATSAPP_WEBHOOK_VERIFY_TOKEN is required")
	}
	return nil
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvIntOrDefault returns environment variable as int or default if not set or invalid
func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Warn().
			Str("key", key).
			Str("value", value).
			Int("default", defaultValue).
			Msg("Invalid integer value for environment variable, using default")
	}
	return defaultValue
}
