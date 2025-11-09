package auth

import (
	"backend/db"
	"backend/internal/models"
	"backend/pkg/recallai"
	"backend/pkg/utils"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// CalendarOAuthConfig holds OAuth configuration for calendar providers
type CalendarOAuthConfig struct {
	GoogleClientID        string
	GoogleClientSecret    string
	GoogleRedirectURL     string
	MicrosoftClientID     string
	MicrosoftClientSecret string
	MicrosoftRedirectURL  string
}

var CalendarConfig *CalendarOAuthConfig

// InitCalendarOAuth initializes calendar OAuth configuration
func InitCalendarOAuth() {
	CalendarConfig = &CalendarOAuthConfig{
		GoogleClientID:        os.Getenv("GOOGLE_CALENDAR_CLIENT_ID"),
		GoogleClientSecret:    os.Getenv("GOOGLE_CALENDAR_CLIENT_SECRET"),
		GoogleRedirectURL:     getEnvOrDefault("GOOGLE_CALENDAR_REDIRECT_URL", "http://localhost:8080/api/calendar/google/callback"),
		MicrosoftClientID:     os.Getenv("MICROSOFT_CALENDAR_CLIENT_ID"),
		MicrosoftClientSecret: os.Getenv("MICROSOFT_CALENDAR_CLIENT_SECRET"),
		MicrosoftRedirectURL:  getEnvOrDefault("MICROSOFT_CALENDAR_REDIRECT_URL", "http://localhost:8080/api/calendar/microsoft/callback"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getFrontendURL returns the frontend URL from environment variable or default
func getFrontendURL() string {
	return getEnvOrDefault("FRONTEND_URL", "http://localhost:5173")
}

// generateSecureState generates a cryptographically secure random state string
func generateSecureState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// BeginCalendarOAuth initiates the OAuth flow for Google/Microsoft Calendar
func BeginCalendarOAuth(c *gin.Context) {
	provider := c.Param("provider") // "google" or "microsoft"

	// Get Clerk session claims
	claims, ok := clerk.SessionClaimsFromContext(c.Request.Context())
	if !ok || claims == nil {
		log.Error().Msg("No Clerk session found")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	clerkUserID := claims.Subject

	// Generate state for CSRF protection
	state, err := generateSecureState()
	if err != nil {
		log.Error().Err(err).Msg("Error generating OAuth state")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate state"})
		return
	}

	// Store state in database
	oauthState := models.CalendarOAuthState{
		ClerkUserID: clerkUserID,
		State:       state,
		Platform:    provider,
		ExpiresAt:   time.Now().Add(10 * time.Minute),
	}

	if err := db.DB.Create(&oauthState).Error; err != nil {
		log.Error().Err(err).Msg("Error storing OAuth state")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store state"})
		return
	}

	// Clean up expired states
	go cleanupExpiredStates()

	var authURL string
	switch provider {
	case "google":
		authURL = buildGoogleCalendarAuthURL(state)
	case "microsoft":
		authURL = buildMicrosoftCalendarAuthURL(state)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"authUrl": authURL})
}

// buildGoogleCalendarAuthURL builds the Google OAuth URL for calendar access
func buildGoogleCalendarAuthURL(state string) string {
	baseURL := "https://accounts.google.com/o/oauth2/v2/auth"
	params := url.Values{}
	params.Add("client_id", CalendarConfig.GoogleClientID)
	params.Add("redirect_uri", CalendarConfig.GoogleRedirectURL)
	params.Add("response_type", "code")
	params.Add("scope", "https://www.googleapis.com/auth/calendar.readonly https://www.googleapis.com/auth/userinfo.email")
	params.Add("access_type", "offline")
	params.Add("prompt", "consent")
	params.Add("state", state)

	return baseURL + "?" + params.Encode()
}

// buildMicrosoftCalendarAuthURL builds the Microsoft OAuth URL for calendar access
func buildMicrosoftCalendarAuthURL(state string) string {
	baseURL := "https://login.microsoftonline.com/common/oauth2/v2.0/authorize"
	params := url.Values{}
	params.Add("client_id", CalendarConfig.MicrosoftClientID)
	params.Add("redirect_uri", CalendarConfig.MicrosoftRedirectURL)
	params.Add("response_type", "code")
	params.Add("scope", "Calendars.Read User.Read offline_access")
	params.Add("response_mode", "query")
	params.Add("state", state)

	return baseURL + "?" + params.Encode()
}

// GoogleCalendarCallback handles the OAuth callback from Google
func GoogleCalendarCallback(c *gin.Context) {
	log.Info().Msg("==================== GOOGLE CALENDAR OAUTH CALLBACK STARTED ====================")

	code := c.Query("code")
	state := c.Query("state")

	log.Info().
		Bool("has_code", code != "").
		Bool("has_state", state != "").
		Msg("OAuth callback parameters received")

	frontendURL := getFrontendURL()

	if code == "" || state == "" {
		c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/profile?calendar_error=missing_params")
		return
	}

	// Validate state and get user
	var oauthState models.CalendarOAuthState
	result := db.DB.Where("state = ? AND platform = ? AND expires_at > ?", state, "google", time.Now()).First(&oauthState)
	if result.Error != nil {
		log.Error().Err(result.Error).Msg("Invalid or expired OAuth state")
		c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/profile?calendar_error=invalid_state")
		return
	}

	// Delete the used state
	db.DB.Delete(&oauthState)

	// Exchange code for tokens
	tokenResp, err := exchangeGoogleCode(code)
	if err != nil {
		log.Error().Err(err).Msg("Error exchanging Google code")
		c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/profile?calendar_error=token_exchange_failed")
		return
	}

	// Get user email from Google
	userEmail, err := getGoogleUserEmail(tokenResp.AccessToken)
	if err != nil {
		log.Error().Err(err).Msg("Error getting Google user email")
		c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/profile?calendar_error=email_fetch_failed")
		return
	}

	// Create calendar in Recall.ai
	recallClient := recallai.NewClient()

	// Encrypt OAuth credentials
	encryptedClientSecret, err := utils.EncryptString(CalendarConfig.GoogleClientSecret)
	if err != nil {
		log.Error().Err(err).Msg("Error encrypting client secret")
		c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/profile?calendar_error=encryption_failed")
		return
	}

	encryptedRefreshToken, err := utils.EncryptString(tokenResp.RefreshToken)
	if err != nil {
		log.Error().Err(err).Msg("Error encrypting refresh token")
		c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/profile?calendar_error=encryption_failed")
		return
	}

	// Create calendar in Recall.ai - this registers the OAuth credentials
	_, err = recallClient.CreateCalendar(recallai.CreateCalendarRequest{
		OAuthClientID:     CalendarConfig.GoogleClientID,
		OAuthClientSecret: CalendarConfig.GoogleClientSecret,
		OAuthRefreshToken: tokenResp.RefreshToken,
		Platform:          "google_calendar",
		OAuthEmail:        userEmail,
	})
	if err != nil {
		log.Error().Err(err).Msg("Error creating calendar in Recall")
		c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/profile?calendar_error=recall_api_failed")
		return
	}

	// Wait a moment for Recall to process the calendar creation
	time.Sleep(2 * time.Second)

	// Fetch ALL calendars for this email from Recall (Google users often have multiple calendars)
	allCalendars, err := recallClient.ListCalendars(userEmail)
	if err != nil {
		log.Error().Err(err).Msg("Error fetching calendars from Recall")
		c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/profile?calendar_error=fetch_calendars_failed")
		return
	}

	log.Info().
		Str("clerk_user_id", oauthState.ClerkUserID).
		Str("user_email", userEmail).
		Int("calendars_count", len(allCalendars)).
		Msg("Fetched all calendars from Recall")

	// Save all calendars to database
	savedCount := 0
	for _, calendarResp := range allCalendars {
		// Check if this calendar already exists for this user
		var existingCalendar models.Calendar
		result := db.DB.Where("clerk_user_id = ? AND recall_calendar_id = ?", oauthState.ClerkUserID, calendarResp.ID).First(&existingCalendar)

		if result.Error == nil {
			// Calendar already exists, skip
			log.Debug().
				Str("recall_calendar_id", calendarResp.ID).
				Msg("Calendar already exists, skipping")
			continue
		}

		calendar := models.Calendar{
			ClerkUserID:       oauthState.ClerkUserID,
			RecallCalendarID:  calendarResp.ID,
			Platform:          "google_calendar",
			PlatformEmail:     calendarResp.PlatformEmail,
			OAuthClientID:     CalendarConfig.GoogleClientID,
			OAuthClientSecret: encryptedClientSecret,
			OAuthRefreshToken: encryptedRefreshToken,
			Status:            calendarResp.Status,
		}

		if err := db.DB.Create(&calendar).Error; err != nil {
			log.Error().
				Err(err).
				Str("clerk_user_id", oauthState.ClerkUserID).
				Str("recall_calendar_id", calendarResp.ID).
				Msg("Error saving calendar to database")
			continue
		}

		savedCount++

		log.Info().
			Str("clerk_user_id", oauthState.ClerkUserID).
			Str("calendar_id", calendar.ID).
			Str("recall_calendar_id", calendar.RecallCalendarID).
			Str("platform_email", calendar.PlatformEmail).
			Msg("Successfully saved calendar to database")
	}

	if savedCount == 0 {
		log.Error().
			Str("clerk_user_id", oauthState.ClerkUserID).
			Msg("No calendars were saved to database")
		c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/profile?calendar_error=no_calendars_saved")
		return
	}

	log.Info().
		Str("clerk_user_id", oauthState.ClerkUserID).
		Int("saved_count", savedCount).
		Str("platform", "google_calendar").
		Msg("Successfully connected Google Calendar(s)")

	// Trigger initial sync in background for all calendars
	for _, calendarResp := range allCalendars {
		var calendar models.Calendar
		if err := db.DB.Where("recall_calendar_id = ? AND clerk_user_id = ?", calendarResp.ID, oauthState.ClerkUserID).First(&calendar).Error; err == nil {
			go performInitialCalendarSync(calendar)
		}
	}

	c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/profile?calendar_success=google")
}

// MicrosoftCalendarCallback handles the OAuth callback from Microsoft
func MicrosoftCalendarCallback(c *gin.Context) {
	frontendURL := getFrontendURL()

	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/profile?calendar_error=missing_params")
		return
	}

	// Validate state and get user
	var oauthState models.CalendarOAuthState
	result := db.DB.Where("state = ? AND platform = ? AND expires_at > ?", state, "microsoft", time.Now()).First(&oauthState)
	if result.Error != nil {
		log.Error().Err(result.Error).Msg("Invalid or expired OAuth state")
		c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/profile?calendar_error=invalid_state")
		return
	}

	// Delete the used state
	db.DB.Delete(&oauthState)

	// Exchange code for tokens
	tokenResp, err := exchangeMicrosoftCode(code)
	if err != nil {
		log.Error().Err(err).Msg("Error exchanging Microsoft code")
		c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/profile?calendar_error=token_exchange_failed")
		return
	}

	// Get user email from Microsoft
	userEmail, err := getMicrosoftUserEmail(tokenResp.AccessToken)
	if err != nil {
		log.Error().Err(err).Msg("Error getting Microsoft user email")
		c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/profile?calendar_error=email_fetch_failed")
		return
	}

	// Create calendar in Recall.ai
	recallClient := recallai.NewClient()

	// Encrypt OAuth credentials
	encryptedClientSecret, err := utils.EncryptString(CalendarConfig.MicrosoftClientSecret)
	if err != nil {
		log.Error().Err(err).Msg("Error encrypting client secret")
		c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/profile?calendar_error=encryption_failed")
		return
	}

	encryptedRefreshToken, err := utils.EncryptString(tokenResp.RefreshToken)
	if err != nil {
		log.Error().Err(err).Msg("Error encrypting refresh token")
		c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/profile?calendar_error=encryption_failed")
		return
	}

	// Create calendar in Recall.ai - this registers the OAuth credentials
	_, err = recallClient.CreateCalendar(recallai.CreateCalendarRequest{
		OAuthClientID:     CalendarConfig.MicrosoftClientID,
		OAuthClientSecret: CalendarConfig.MicrosoftClientSecret,
		OAuthRefreshToken: tokenResp.RefreshToken,
		Platform:          "microsoft_outlook",
		OAuthEmail:        userEmail,
	})
	if err != nil {
		log.Error().Err(err).Msg("Error creating calendar in Recall")
		c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/profile?calendar_error=recall_api_failed")
		return
	}

	// Wait a moment for Recall to process the calendar creation
	time.Sleep(2 * time.Second)

	// Fetch ALL calendars for this email from Recall
	allCalendars, err := recallClient.ListCalendars(userEmail)
	if err != nil {
		log.Error().Err(err).Msg("Error fetching calendars from Recall")
		c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/profile?calendar_error=fetch_calendars_failed")
		return
	}

	log.Info().
		Str("clerk_user_id", oauthState.ClerkUserID).
		Str("user_email", userEmail).
		Int("calendars_count", len(allCalendars)).
		Msg("Fetched all calendars from Recall")

	// Save all calendars to database
	savedCount := 0
	for _, calendarResp := range allCalendars {
		// Check if this calendar already exists for this user
		var existingCalendar models.Calendar
		result := db.DB.Where("clerk_user_id = ? AND recall_calendar_id = ?", oauthState.ClerkUserID, calendarResp.ID).First(&existingCalendar)

		if result.Error == nil {
			// Calendar already exists, skip
			log.Debug().
				Str("recall_calendar_id", calendarResp.ID).
				Msg("Calendar already exists, skipping")
			continue
		}

		calendar := models.Calendar{
			ClerkUserID:       oauthState.ClerkUserID,
			RecallCalendarID:  calendarResp.ID,
			Platform:          "microsoft_outlook",
			PlatformEmail:     calendarResp.PlatformEmail,
			OAuthClientID:     CalendarConfig.MicrosoftClientID,
			OAuthClientSecret: encryptedClientSecret,
			OAuthRefreshToken: encryptedRefreshToken,
			Status:            calendarResp.Status,
		}

		if err := db.DB.Create(&calendar).Error; err != nil {
			log.Error().
				Err(err).
				Str("clerk_user_id", oauthState.ClerkUserID).
				Str("recall_calendar_id", calendarResp.ID).
				Msg("Error saving calendar to database")
			continue
		}

		savedCount++

		log.Info().
			Str("clerk_user_id", oauthState.ClerkUserID).
			Str("calendar_id", calendar.ID).
			Str("recall_calendar_id", calendar.RecallCalendarID).
			Str("platform_email", calendar.PlatformEmail).
			Msg("Successfully saved calendar to database")
	}

	if savedCount == 0 {
		log.Error().
			Str("clerk_user_id", oauthState.ClerkUserID).
			Msg("No calendars were saved to database")
		c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/profile?calendar_error=no_calendars_saved")
		return
	}

	log.Info().
		Str("clerk_user_id", oauthState.ClerkUserID).
		Int("saved_count", savedCount).
		Str("platform", "microsoft_outlook").
		Msg("Successfully connected Microsoft Calendar(s)")

	// Trigger initial sync in background for all calendars
	for _, calendarResp := range allCalendars {
		var calendar models.Calendar
		if err := db.DB.Where("recall_calendar_id = ? AND clerk_user_id = ?", calendarResp.ID, oauthState.ClerkUserID).First(&calendar).Error; err == nil {
			go performInitialCalendarSync(calendar)
		}
	}

	c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/profile?calendar_success=microsoft")
}

// TokenResponse represents OAuth token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// exchangeGoogleCode exchanges the authorization code for tokens
func exchangeGoogleCode(code string) (*TokenResponse, error) {
	tokenURL := "https://oauth2.googleapis.com/token"

	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", CalendarConfig.GoogleClientID)
	data.Set("client_secret", CalendarConfig.GoogleClientSecret)
	data.Set("redirect_uri", CalendarConfig.GoogleRedirectURL)
	data.Set("grant_type", "authorization_code")

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed with status: %d", resp.StatusCode)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

// exchangeMicrosoftCode exchanges the authorization code for tokens
func exchangeMicrosoftCode(code string) (*TokenResponse, error) {
	tokenURL := "https://login.microsoftonline.com/common/oauth2/v2.0/token"

	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", CalendarConfig.MicrosoftClientID)
	data.Set("client_secret", CalendarConfig.MicrosoftClientSecret)
	data.Set("redirect_uri", CalendarConfig.MicrosoftRedirectURL)
	data.Set("grant_type", "authorization_code")

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed with status: %d", resp.StatusCode)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

// getGoogleUserEmail fetches the user's email from Google
func getGoogleUserEmail(accessToken string) (string, error) {
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var userInfo struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return "", err
	}

	return userInfo.Email, nil
}

// getMicrosoftUserEmail fetches the user's email from Microsoft
func getMicrosoftUserEmail(accessToken string) (string, error) {
	req, err := http.NewRequest("GET", "https://graph.microsoft.com/v1.0/me", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var userInfo struct {
		Mail              string `json:"mail"`
		UserPrincipalName string `json:"userPrincipalName"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return "", err
	}

	// Prefer mail, fall back to userPrincipalName
	if userInfo.Mail != "" {
		return userInfo.Mail, nil
	}
	return userInfo.UserPrincipalName, nil
}

// cleanupExpiredStates removes expired OAuth states from the database
func cleanupExpiredStates() {
	if err := db.DB.Where("expires_at < ?", time.Now()).Delete(&models.CalendarOAuthState{}).Error; err != nil {
		log.Error().Err(err).Msg("Error cleaning up expired OAuth states")
	}
}

// performInitialCalendarSync syncs calendar events immediately after calendar is connected
func performInitialCalendarSync(calendar models.Calendar) {
	// Add a small delay to ensure Recall.ai has processed the calendar
	time.Sleep(2 * time.Second)

	recallClient := recallai.NewClient()
	events, err := recallClient.ListCalendarEvents(calendar.RecallCalendarID)
	if err != nil {
		log.Error().Err(err).Str("calendar_id", calendar.ID).Msg("Error performing initial calendar sync")
		return
	}

	syncedCount := 0
	for _, event := range events {
		startTime, _ := time.Parse(time.RFC3339, event.StartTime)
		endTime, _ := time.Parse(time.RFC3339, event.EndTime)

		log.Debug().
			Str("event_id", event.ID).
			Str("title", event.Title).
			Str("start_time_raw", event.StartTime).
			Time("start_time_parsed", startTime).
			Time("current_time", time.Now()).
			Bool("is_upcoming", startTime.After(time.Now())).
			Msg("Processing calendar event")

		calendarEvent := models.CalendarEvent{
			CalendarID:      calendar.ID,
			RecallEventID:   event.ID,
			ICalUID:         event.ICalUID,
			PlatformID:      event.PlatformID,
			MeetingPlatform: event.MeetingPlatform,
			MeetingURL:      event.MeetingURL,
			Title:           event.Title,
			StartTime:       startTime,
			EndTime:         endTime,
			IsDeleted:       event.IsDeleted,
			BotScheduled:    len(event.Bots) > 0,
		}

		if len(event.Bots) > 0 {
			calendarEvent.BotID = &event.Bots[0].BotID
		}

		if err := db.DB.Where(models.CalendarEvent{RecallEventID: event.ID}).
			Assign(calendarEvent).
			FirstOrCreate(&calendarEvent).Error; err != nil {
			log.Error().Err(err).Str("event_id", event.ID).Msg("Error upserting calendar event during initial sync")
			continue
		}

		syncedCount++
	}

	// Update last synced timestamp
	db.DB.Model(&calendar).Update("last_synced_at", time.Now())

	log.Info().
		Str("calendar_id", calendar.ID).
		Int("synced_count", syncedCount).
		Msg("Initial calendar sync completed")
}
