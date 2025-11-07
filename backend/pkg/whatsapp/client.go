package whatsapp

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"backend/internal/config"

	"github.com/rs/zerolog/log"
)

// WhatsAppClient interface defines methods for interacting with WhatsApp Cloud API
type WhatsAppClient interface {
	SendTextMessage(phoneNumber, message string) error
	SendInteractiveMessage(phoneNumber string, message InteractiveMessage) error
	VerifyWebhookSignature(payload []byte, signature string) bool
	GetPhoneNumberID() string
}

// Client implements WhatsAppClient interface
type Client struct {
	config     *config.WhatsAppConfig
	httpClient *http.Client
}

// NewClient creates a new WhatsApp API client
func NewClient(cfg *config.WhatsAppConfig) WhatsAppClient {
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// InteractiveMessage represents an interactive WhatsApp message
type InteractiveMessage struct {
	Type   string         `json:"type"` // "button" or "list"
	Header *MessageHeader `json:"header,omitempty"`
	Body   MessageBody    `json:"body"`
	Footer *MessageFooter `json:"footer,omitempty"`
	Action MessageAction  `json:"action"`
}

// MessageHeader represents message header
type MessageHeader struct {
	Type string `json:"type"` // "text"
	Text string `json:"text"`
}

// MessageBody represents message body
type MessageBody struct {
	Text string `json:"text"`
}

// MessageFooter represents message footer
type MessageFooter struct {
	Text string `json:"text"`
}

// MessageAction represents interactive action
type MessageAction struct {
	Buttons  []Button  `json:"buttons,omitempty"`
	Button   string    `json:"button,omitempty"`   // For list messages
	Sections []Section `json:"sections,omitempty"` // For list messages
}

// Button represents an interactive button
type Button struct {
	Type  string      `json:"type"` // "reply"
	Reply ButtonReply `json:"reply"`
}

// ButtonReply represents button reply data
type ButtonReply struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// Section represents a list section
type Section struct {
	Title string `json:"title"`
	Rows  []Row  `json:"rows"`
}

// Row represents a list row
type Row struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

// SendTextMessage sends a text message to a WhatsApp user
func (c *Client) SendTextMessage(phoneNumber, message string) error {
	url := fmt.Sprintf("%s/%s/messages", c.config.APIURL, c.config.PhoneNumberID)

	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                phoneNumber,
		"type":              "text",
		"text": map[string]string{
			"body": message,
		},
	}

	return c.sendRequest(url, payload, 3)
}

// SendInteractiveMessage sends an interactive message to a WhatsApp user
func (c *Client) SendInteractiveMessage(phoneNumber string, message InteractiveMessage) error {
	url := fmt.Sprintf("%s/%s/messages", c.config.APIURL, c.config.PhoneNumberID)

	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                phoneNumber,
		"type":              "interactive",
		"interactive":       message,
	}

	return c.sendRequest(url, payload, 3)
}

// sendRequest sends an HTTP request to WhatsApp API with retry logic
func (c *Client) sendRequest(url string, payload interface{}, maxRetries int) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.AccessToken))

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed (attempt %d/%d): %w", attempt, maxRetries, err)
			log.Warn().Err(err).Int("attempt", attempt).Msg("WhatsApp API request failed, retrying")
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}

		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			log.Debug().
				Int("status", resp.StatusCode).
				Str("response", string(body)).
				Msg("WhatsApp API request successful")
			return nil
		}

		lastErr = fmt.Errorf("API returned status %d (attempt %d/%d): %s", resp.StatusCode, attempt, maxRetries, string(body))
		log.Warn().
			Int("status", resp.StatusCode).
			Str("response", string(body)).
			Int("attempt", attempt).
			Msg("WhatsApp API request failed")

		// Don't retry on client errors (4xx)
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return lastErr
		}

		time.Sleep(time.Duration(attempt) * time.Second)
	}

	return lastErr
}

// VerifyWebhookSignature verifies the webhook signature from WhatsApp
func (c *Client) VerifyWebhookSignature(payload []byte, signature string) bool {
	mac := hmac.New(sha256.New, []byte(c.config.AppSecret))
	mac.Write(payload)
	expectedSignature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	isValid := hmac.Equal([]byte(signature), []byte(expectedSignature))

	if !isValid {
		log.Warn().
			Str("expected_prefix", expectedSignature[:20]).
			Str("received_prefix", signature[:min(20, len(signature))]).
			Msg("Webhook signature verification failed")
	}

	return isValid
}

// GetPhoneNumberID returns the configured phone number ID
func (c *Client) GetPhoneNumberID() string {
	return c.config.PhoneNumberID
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
