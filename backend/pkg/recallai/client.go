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

// getCalendarBaseURL returns the base URL for calendar API endpoints
func (c *Client) getCalendarBaseURL() string {
	// Calendar API uses v2 endpoint
	return fmt.Sprintf("https://%s.recall.ai/api/v2", c.Region)
}

// CreateBotRequest represents the request payload for creating a bot
type CreateBotRequest struct {
	MeetingURL      string          `json:"meeting_url"`
	RecordingConfig RecordingConfig `json:"recording_config"`
}

// RecordingConfig defines the recording configuration for the bot
type RecordingConfig struct {
	Transcript       TranscriptConfig `json:"transcript"`
	VideoMixedLayout string           `json:"video_mixed_layout,omitempty"` // e.g., "gallery_view_v2"
	VideoSeparateMP4 *struct{}        `json:"video_separate_mp4,omitempty"` // Empty object to enable
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
	VideoMixed VideoMixedShortcut `json:"video_mixed"`
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

// VideoMixedShortcut contains video recording shortcuts
type VideoMixedShortcut struct {
	ID   string    `json:"id"`
	Data VideoData `json:"data"`
}

// VideoData contains the actual video data and download URL
type VideoData struct {
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

// ===== Calendar V2 Integration Types =====

// CreateCalendarRequest represents the request to create a calendar in Recall
type CreateCalendarRequest struct {
	OAuthClientID     string `json:"oauth_client_id"`
	OAuthClientSecret string `json:"oauth_client_secret"`
	OAuthRefreshToken string `json:"oauth_refresh_token"`
	Platform          string `json:"platform"` // "google_calendar" or "microsoft_outlook"
	OAuthEmail        string `json:"oauth_email"`
}

// CreateCalendarResponse represents the response from creating a calendar
type CreateCalendarResponse struct {
	ID                string        `json:"id"`
	OAuthClientID     string        `json:"oauth_client_id"`
	OAuthClientSecret string        `json:"oauth_client_secret"`
	OAuthRefreshToken string        `json:"oauth_refresh_token"`
	Platform          string        `json:"platform"`
	PlatformEmail     string        `json:"platform_email"`
	Status            string        `json:"status"`
	StatusChanges     []interface{} `json:"status_changes"` // Can be array or object
	CreatedAt         string        `json:"created_at"`
	UpdatedAt         string        `json:"updated_at"`
}

// UpdateCalendarRequest represents the request to update a calendar
type UpdateCalendarRequest struct {
	OAuthRefreshToken string `json:"oauth_refresh_token"`
}

// CalendarEvent represents an event from a calendar
type CalendarEvent struct {
	ID              string          `json:"id"`
	ICalUID         string          `json:"ical_uid"`
	PlatformID      string          `json:"platform_id"`
	CalendarID      string          `json:"calendar_id"`
	MeetingPlatform string          `json:"meeting_platform"`
	MeetingURL      string          `json:"meeting_url"`
	StartTime       string          `json:"start_time"`
	EndTime         string          `json:"end_time"`
	Title           string          `json:"title,omitempty"`
	IsDeleted       bool            `json:"is_deleted"`
	CreatedAt       string          `json:"created_at"`
	UpdatedAt       string          `json:"updated_at"`
	Bots            []ScheduledBot  `json:"bots,omitempty"`
	Raw             json.RawMessage `json:"raw,omitempty"`
}

// ScheduledBot represents a bot scheduled for a calendar event
type ScheduledBot struct {
	BotID            string `json:"bot_id"`
	StartTime        string `json:"start_time"`
	DeduplicationKey string `json:"deduplication_key"`
	MeetingURL       string `json:"meeting_url"`
}

// ListCalendarEventsResponse represents the response from listing calendar events
type ListCalendarEventsResponse struct {
	Results  []CalendarEvent `json:"results"`
	Next     *string         `json:"next"`
	Previous *string         `json:"previous"`
}

// ListCalendarsResponse represents the response from listing calendars
type ListCalendarsResponse struct {
	Results  []CreateCalendarResponse `json:"results"`
	Next     *string                  `json:"next"`
	Previous *string                  `json:"previous"`
}

// ScheduleBotForEventRequest represents the request to schedule a bot for an event
type ScheduleBotForEventRequest struct {
	DeduplicationKey string                 `json:"deduplication_key"`
	BotConfig        map[string]interface{} `json:"bot_config"`
}

// ScheduleBotForEventResponse represents the response from scheduling a bot
type ScheduleBotForEventResponse struct {
	CalendarEvent CalendarEvent `json:"calendar_event"`
}

// CalendarSyncWebhook represents the webhook payload for calendar sync events
type CalendarSyncWebhook struct {
	Event string `json:"event"`
	Data  struct {
		CalendarID    string `json:"calendar_id"`
		LastUpdatedTS string `json:"last_updated_ts"`
	} `json:"data"`
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

// makeCalendarRequest is a helper method to make HTTP requests to the Recall.ai Calendar V2 API
func (c *Client) makeCalendarRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader

	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.getCalendarBaseURL()+endpoint, reqBody)
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
		Msg("Making Recall.ai Calendar API request")

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
			VideoMixedLayout: "gallery_view_v2", // Enable video recording with gallery view
			VideoSeparateMP4: &struct{}{},       // Enable separate MP4 streams
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

// ===== Calendar V2 Integration Methods =====

// CreateCalendar creates a new calendar in Recall for a user
func (c *Client) CreateCalendar(req CreateCalendarRequest) (*CreateCalendarResponse, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("RECALL_AI_API_KEY environment variable is not set")
	}

	if req.Platform != "google_calendar" && req.Platform != "microsoft_outlook" {
		return nil, fmt.Errorf("invalid platform: must be 'google_calendar' or 'microsoft_outlook'")
	}

	log.Info().
		Str("platform", req.Platform).
		Str("email", req.OAuthEmail).
		Msg("Creating calendar in Recall.ai")

	resp, err := c.makeCalendarRequest("POST", "/calendars", req)
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, c.handleAPIError(resp)
	}

	var calendarResp CreateCalendarResponse
	if err := json.NewDecoder(resp.Body).Decode(&calendarResp); err != nil {
		return nil, fmt.Errorf("failed to decode calendar creation response: %w", err)
	}

	log.Info().
		Str("calendar_id", calendarResp.ID).
		Str("status", calendarResp.Status).
		Msg("Successfully created calendar in Recall.ai")

	return &calendarResp, nil
}

// UpdateCalendar updates an existing calendar's refresh token
func (c *Client) UpdateCalendar(calendarID string, refreshToken string) error {
	if c.APIKey == "" {
		return fmt.Errorf("RECALL_AI_API_KEY environment variable is not set")
	}

	if calendarID == "" {
		return fmt.Errorf("calendar ID cannot be empty")
	}

	req := UpdateCalendarRequest{
		OAuthRefreshToken: refreshToken,
	}

	log.Info().
		Str("calendar_id", calendarID).
		Msg("Updating calendar in Recall.ai")

	resp, err := c.makeCalendarRequest("PUT", "/calendars/"+calendarID, req)
	if err != nil {
		return fmt.Errorf("failed to update calendar: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.handleAPIError(resp)
	}

	log.Info().
		Str("calendar_id", calendarID).
		Msg("Successfully updated calendar in Recall.ai")

	return nil
}

// DeleteCalendar deletes a calendar from Recall
func (c *Client) DeleteCalendar(calendarID string) error {
	if c.APIKey == "" {
		return fmt.Errorf("RECALL_AI_API_KEY environment variable is not set")
	}

	if calendarID == "" {
		return fmt.Errorf("calendar ID cannot be empty")
	}

	log.Info().
		Str("calendar_id", calendarID).
		Msg("Deleting calendar from Recall.ai")

	resp, err := c.makeCalendarRequest("DELETE", "/calendars/"+calendarID, nil)
	if err != nil {
		return fmt.Errorf("failed to delete calendar: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return c.handleAPIError(resp)
	}

	log.Info().
		Str("calendar_id", calendarID).
		Msg("Successfully deleted calendar from Recall.ai")

	return nil
}

// ListCalendars retrieves all calendars from Recall, optionally filtered by email
func (c *Client) ListCalendars(email string) ([]CreateCalendarResponse, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("RECALL_AI_API_KEY environment variable is not set")
	}

	endpoint := "/calendars"
	if email != "" {
		endpoint = endpoint + "?email=" + fmt.Sprintf("%s", email)
	}

	log.Debug().
		Str("email", email).
		Msg("Fetching calendars from Recall.ai")

	resp, err := c.makeCalendarRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list calendars: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleAPIError(resp)
	}

	var calendarsResp ListCalendarsResponse
	if err := json.NewDecoder(resp.Body).Decode(&calendarsResp); err != nil {
		return nil, fmt.Errorf("failed to decode calendars response: %w", err)
	}

	log.Debug().
		Str("email", email).
		Int("calendars_count", len(calendarsResp.Results)).
		Msg("Successfully fetched calendars")

	return calendarsResp.Results, nil
}

// ListCalendarEvents retrieves all events for a specific calendar
func (c *Client) ListCalendarEvents(calendarID string) ([]CalendarEvent, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("RECALL_AI_API_KEY environment variable is not set")
	}

	if calendarID == "" {
		return nil, fmt.Errorf("calendar ID cannot be empty")
	}

	log.Debug().
		Str("calendar_id", calendarID).
		Msg("Fetching calendar events from Recall.ai")

	resp, err := c.makeCalendarRequest("GET", "/calendar-events/?calendar_id="+calendarID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list calendar events: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleAPIError(resp)
	}

	var eventsResp ListCalendarEventsResponse
	if err := json.NewDecoder(resp.Body).Decode(&eventsResp); err != nil {
		return nil, fmt.Errorf("failed to decode calendar events response: %w", err)
	}

	log.Debug().
		Str("calendar_id", calendarID).
		Int("events_count", len(eventsResp.Results)).
		Msg("Successfully fetched calendar events")

	return eventsResp.Results, nil
}

// ScheduleBotForEvent schedules a bot to join a specific calendar event
func (c *Client) ScheduleBotForEvent(eventID string, deduplicationKey string, botConfig map[string]interface{}) (*CalendarEvent, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("RECALL_AI_API_KEY environment variable is not set")
	}

	if eventID == "" {
		return nil, fmt.Errorf("event ID cannot be empty")
	}

	req := ScheduleBotForEventRequest{
		DeduplicationKey: deduplicationKey,
		BotConfig:        botConfig,
	}

	log.Info().
		Str("event_id", eventID).
		Str("deduplication_key", deduplicationKey).
		Msg("Scheduling bot for calendar event")

	resp, err := c.makeCalendarRequest("POST", "/calendar-events/"+eventID+"/bot/", req)
	if err != nil {
		return nil, fmt.Errorf("failed to schedule bot for event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, c.handleAPIError(resp)
	}

	var scheduleResp ScheduleBotForEventResponse
	if err := json.NewDecoder(resp.Body).Decode(&scheduleResp); err != nil {
		return nil, fmt.Errorf("failed to decode schedule bot response: %w", err)
	}

	log.Info().
		Str("event_id", eventID).
		Msg("Successfully scheduled bot for calendar event")

	return &scheduleResp.CalendarEvent, nil
}

// RemoveBotFromEvent removes a scheduled bot from a calendar event
func (c *Client) RemoveBotFromEvent(eventID string, botID string) error {
	if c.APIKey == "" {
		return fmt.Errorf("RECALL_AI_API_KEY environment variable is not set")
	}

	if eventID == "" || botID == "" {
		return fmt.Errorf("event ID and bot ID cannot be empty")
	}

	log.Info().
		Str("event_id", eventID).
		Str("bot_id", botID).
		Msg("Removing bot from calendar event")

	resp, err := c.makeCalendarRequest("DELETE", "/calendar-events/"+eventID+"/bot/"+botID+"/", nil)
	if err != nil {
		return fmt.Errorf("failed to remove bot from event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return c.handleAPIError(resp)
	}

	log.Info().
		Str("event_id", eventID).
		Str("bot_id", botID).
		Msg("Successfully removed bot from calendar event")

	return nil
}
