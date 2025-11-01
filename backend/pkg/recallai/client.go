package recallai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

type Client struct {
	APIKey     string
	Region     string
	HTTPClient *http.Client
}

// NewClient creates a new Recall.ai API client
func NewClient() *Client {
	region := os.Getenv("RECALL_AI_REGION")
	if region == "" {
		region = "us-east-1" // Default to us-east-1
	}

	return &Client{
		APIKey: os.Getenv("RECALL_AI_API_KEY"),
		Region: region,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// getBaseURL returns the base URL for the configured region
func (c *Client) getBaseURL() string {
	return fmt.Sprintf("https://%s.recall.ai/api/v1", c.Region)
}

// CreateBotRequest represents the request payload for creating a bot
type CreateBotRequest struct {
	MeetingURL      string          `json:"meeting_url"`
	RecordingConfig RecordingConfig `json:"recording_config"`
}

// RecordingConfig defines the recording configuration for the bot
type RecordingConfig struct {
	Transcript TranscriptConfig `json:"transcript"`
}

// TranscriptConfig defines the transcript provider configuration
type TranscriptConfig struct {
	Provider TranscriptProvider `json:"provider"`
}

// TranscriptProvider defines the specific transcript provider settings
type TranscriptProvider struct {
	RecallAIStreaming RecallAIStreamingConfig `json:"recallai_streaming"`
}

// RecallAIStreamingConfig defines the Recall.ai streaming configuration
type RecallAIStreamingConfig struct {
	Mode string `json:"mode,omitempty"` // "prioritize_accuracy" or "prioritize_low_latency"
}

// CreateBotResponse represents the response from creating a bot
type CreateBotResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// BotDetails represents the detailed information about a bot
type BotDetails struct {
	ID         string      `json:"id"`
	Status     string      `json:"status"`
	Recordings []Recording `json:"recordings"`
}

// Recording represents a recording within bot details
type Recording struct {
	ID             string         `json:"id"`
	MediaShortcuts MediaShortcuts `json:"media_shortcuts"`
}

// MediaShortcuts contains shortcuts to media files
type MediaShortcuts struct {
	Transcript TranscriptShortcut `json:"transcript"`
}

// TranscriptShortcut contains transcript-specific shortcuts
type TranscriptShortcut struct {
	ID   string         `json:"id"`
	Data TranscriptData `json:"data"`
}

// TranscriptData contains the actual transcript data and download URL
type TranscriptData struct {
	DownloadURL string `json:"download_url"`
}

// TranscriptEntry represents a single entry in the transcript
type TranscriptEntry struct {
	Participant ParticipantInfo `json:"participant"`
	Words       []WordInfo      `json:"words"`
}

// ParticipantInfo contains information about a meeting participant
type ParticipantInfo struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	IsHost   bool   `json:"is_host"`
	Platform string `json:"platform"`
}

// WordInfo represents a single word in the transcript with timing
type WordInfo struct {
	Text           string    `json:"text"`
	StartTimestamp Timestamp `json:"start_timestamp"`
	EndTimestamp   Timestamp `json:"end_timestamp"`
}

// Timestamp represents timing information for transcript words
type Timestamp struct {
	Relative float64 `json:"relative"`
	Absolute string  `json:"absolute"`
}

// makeRequest is a helper method to make HTTP requests to the Recall.ai API
func (c *Client) makeRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader

	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.getBaseURL()+endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers
	req.Header.Set("Authorization", "Token "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	log.Debug().
		Str("method", method).
		Str("url", req.URL.String()).
		Msg("Making Recall.ai API request")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	return resp, nil
}

// handleAPIError processes API error responses
func (c *Client) handleAPIError(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("recall.ai API error: %d (failed to read response body)", resp.StatusCode)
	}

	log.Error().
		Int("status_code", resp.StatusCode).
		Str("response_body", string(body)).
		Msg("Recall.ai API error")

	return fmt.Errorf("recall.ai API error: %d - %s", resp.StatusCode, string(body))
}

// CreateBot creates a new bot to join and record a meeting
func (c *Client) CreateBot(meetingURL string) (*CreateBotResponse, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("RECALL_AI_API_KEY environment variable is not set")
	}

	reqBody := CreateBotRequest{
		MeetingURL: meetingURL,
		RecordingConfig: RecordingConfig{
			Transcript: TranscriptConfig{
				Provider: TranscriptProvider{
					RecallAIStreaming: RecallAIStreamingConfig{
						Mode: "prioritize_accuracy",
					},
				},
			},
		},
	}

	log.Info().
		Str("meeting_url", meetingURL).
		Msg("Creating Recall.ai bot")

	resp, err := c.makeRequest("POST", "/bot/", reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, c.handleAPIError(resp)
	}

	var botResp CreateBotResponse
	if err := json.NewDecoder(resp.Body).Decode(&botResp); err != nil {
		return nil, fmt.Errorf("failed to decode bot creation response: %w", err)
	}

	log.Info().
		Str("bot_id", botResp.ID).
		Str("status", botResp.Status).
		Msg("Successfully created Recall.ai bot")

	return &botResp, nil
}

// GetBot retrieves detailed information about a bot by its ID
func (c *Client) GetBot(botID string) (*BotDetails, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("RECALL_AI_API_KEY environment variable is not set")
	}

	if botID == "" {
		return nil, fmt.Errorf("bot ID cannot be empty")
	}

	log.Debug().
		Str("bot_id", botID).
		Msg("Retrieving Recall.ai bot details")

	resp, err := c.makeRequest("GET", "/bot/"+botID+"/", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get bot details: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleAPIError(resp)
	}

	var botDetails BotDetails
	if err := json.NewDecoder(resp.Body).Decode(&botDetails); err != nil {
		return nil, fmt.Errorf("failed to decode bot details response: %w", err)
	}

	log.Debug().
		Str("bot_id", botDetails.ID).
		Str("status", botDetails.Status).
		Int("recordings_count", len(botDetails.Recordings)).
		Msg("Successfully retrieved bot details")

	return &botDetails, nil
}

// DownloadTranscript downloads and parses the transcript from the provided URL
func (c *Client) DownloadTranscript(downloadURL string) ([]TranscriptEntry, error) {
	if downloadURL == "" {
		return nil, fmt.Errorf("download URL cannot be empty")
	}

	log.Info().
		Str("download_url", downloadURL).
		Msg("Downloading transcript from Recall.ai")

	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create transcript download request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download transcript: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Error().
			Int("status_code", resp.StatusCode).
			Str("response_body", string(body)).
			Msg("Failed to download transcript")
		return nil, fmt.Errorf("failed to download transcript: status %d", resp.StatusCode)
	}

	var transcript []TranscriptEntry
	if err := json.NewDecoder(resp.Body).Decode(&transcript); err != nil {
		return nil, fmt.Errorf("failed to decode transcript: %w", err)
	}

	log.Info().
		Int("entries_count", len(transcript)).
		Msg("Successfully downloaded and parsed transcript")

	return transcript, nil
}
