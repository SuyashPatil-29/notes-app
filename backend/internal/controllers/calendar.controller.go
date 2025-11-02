package controllers

import (
	"backend/db"
	"backend/internal/middleware"
	"backend/internal/models"
	"backend/pkg/recallai"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// GetUserCalendars returns all calendars connected by the user
func GetUserCalendars(c *gin.Context) {
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var calendars []models.Calendar
	if err := db.DB.Where("clerk_user_id = ?", clerkUserID).Find(&calendars).Error; err != nil {
		log.Error().Err(err).Msg("Error fetching user calendars")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch calendars"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"calendars": calendars})
}

// SyncMissingCalendars fetches all calendars from Recall for user's existing connections and saves missing ones
func SyncMissingCalendars(c *gin.Context) {
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Get user's existing calendars
	var existingCalendars []models.Calendar
	if err := db.DB.Where("clerk_user_id = ?", clerkUserID).Find(&existingCalendars).Error; err != nil {
		log.Error().Err(err).Msg("Error fetching existing calendars")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch existing calendars"})
		return
	}

	if len(existingCalendars) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No calendars to sync", "added": 0})
		return
	}

	recallClient := recallai.NewClient()
	addedCount := 0

	// Get unique emails from existing calendars
	emailsMap := make(map[string]models.Calendar)
	for _, cal := range existingCalendars {
		if cal.PlatformEmail != "" {
			emailsMap[cal.PlatformEmail] = cal
		}
	}

	// For each unique email, fetch all calendars from Recall
	for email, referenceCal := range emailsMap {
		log.Info().
			Str("email", email).
			Msg("Fetching calendars from Recall for email")

		allCalendars, err := recallClient.ListCalendars(email)
		if err != nil {
			log.Error().Err(err).Str("email", email).Msg("Error fetching calendars from Recall")
			continue
		}

		log.Info().
			Str("email", email).
			Int("calendars_count", len(allCalendars)).
			Msg("Found calendars in Recall")

		// Check each calendar and add if missing
		for _, recallCal := range allCalendars {
			// Check if calendar already exists
			var existingCal models.Calendar
			result := db.DB.Where("recall_calendar_id = ? AND clerk_user_id = ?", recallCal.ID, clerkUserID).First(&existingCal)

			if result.Error == nil {
				// Calendar already exists
				log.Debug().Str("recall_calendar_id", recallCal.ID).Msg("Calendar already exists")
				continue
			}

			// Add missing calendar
			newCalendar := models.Calendar{
				ClerkUserID:       clerkUserID,
				RecallCalendarID:  recallCal.ID,
				Platform:          recallCal.Platform,
				PlatformEmail:     recallCal.PlatformEmail,
				OAuthClientID:     referenceCal.OAuthClientID,
				OAuthClientSecret: referenceCal.OAuthClientSecret,
				OAuthRefreshToken: referenceCal.OAuthRefreshToken,
				Status:            recallCal.Status,
			}

			if err := db.DB.Create(&newCalendar).Error; err != nil {
				log.Error().
					Err(err).
					Str("recall_calendar_id", recallCal.ID).
					Msg("Error saving missing calendar")
				continue
			}

			addedCount++
			log.Info().
				Str("calendar_id", newCalendar.ID).
				Str("recall_calendar_id", newCalendar.RecallCalendarID).
				Str("platform_email", newCalendar.PlatformEmail).
				Msg("Added missing calendar")

			// Trigger initial sync for the newly added calendar
			go syncCalendarInBackground(newCalendar)
		}
	}

	log.Info().
		Str("clerk_user_id", clerkUserID).
		Int("added_count", addedCount).
		Msg("Completed syncing missing calendars")

	c.JSON(http.StatusOK, gin.H{
		"message": "Missing calendars synced successfully",
		"added":   addedCount,
	})
}

// DisconnectCalendar removes a calendar connection
func DisconnectCalendar(c *gin.Context) {
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	calendarID := c.Param("id")

	// Find calendar
	var calendar models.Calendar
	if err := db.DB.Where("id = ? AND clerk_user_id = ?", calendarID, clerkUserID).First(&calendar).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Calendar not found"})
		return
	}

	// Delete from Recall.ai
	recallClient := recallai.NewClient()
	if err := recallClient.DeleteCalendar(calendar.RecallCalendarID); err != nil {
		log.Error().Err(err).Str("recall_calendar_id", calendar.RecallCalendarID).Msg("Error deleting calendar from Recall")
		// Continue with local deletion even if Recall API fails
	}

	// Delete from database (cascade will delete events)
	if err := db.DB.Delete(&calendar).Error; err != nil {
		log.Error().Err(err).Msg("Error deleting calendar from database")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete calendar"})
		return
	}

	log.Info().
		Str("clerk_user_id", clerkUserID).
		Str("calendar_id", calendarID).
		Msg("Calendar disconnected successfully")

	c.JSON(http.StatusOK, gin.H{"message": "Calendar disconnected successfully"})
}

// GetCalendarEvents returns all events for a specific calendar with pagination
func GetCalendarEvents(c *gin.Context) {
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	calendarID := c.Param("id")

	// Parse pagination params with defaults
	page := 1
	pageSize := 100
	if pageParam := c.Query("page"); pageParam != "" {
		fmt.Sscanf(pageParam, "%d", &page)
	}
	if pageSizeParam := c.Query("page_size"); pageSizeParam != "" {
		fmt.Sscanf(pageSizeParam, "%d", &pageSize)
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 100
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// Verify calendar belongs to user
	var calendar models.Calendar
	if err := db.DB.Where("id = ? AND clerk_user_id = ?", calendarID, clerkUserID).First(&calendar).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Calendar not found"})
		return
	}

	// Build query for events
	query := db.DB.Where("calendar_id = ? AND is_deleted = ?", calendarID, false)

	// Optional: filter upcoming events only
	currentTime := time.Now()
	if c.Query("upcoming") == "true" {
		// Show events that haven't ended yet (including ongoing meetings)
		// Also show events from the past hour in case meeting just started
		oneHourAgo := currentTime.Add(-1 * time.Hour)
		query = query.Where("end_time >= ?", oneHourAgo)
		log.Debug().
			Str("calendar_id", calendarID).
			Time("current_time", currentTime).
			Time("filter_from", oneHourAgo).
			Msg("Filtering for upcoming and ongoing events")
	}

	// Get total count
	var total int64
	query.Model(&models.CalendarEvent{}).Count(&total)

	// Get events with pagination
	var events []models.CalendarEvent
	offset := (page - 1) * pageSize
	if err := query.Order("start_time ASC").Limit(pageSize).Offset(offset).Find(&events).Error; err != nil {
		log.Error().Err(err).Msg("Error fetching calendar events")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}

	log.Debug().
		Str("calendar_id", calendarID).
		Int("events_count", len(events)).
		Bool("upcoming_filter", c.Query("upcoming") == "true").
		Msg("Fetched calendar events from database")

	// Calculate pagination metadata
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	if totalPages == 0 {
		totalPages = 1
	}

	paginatedResponse := gin.H{
		"events":     events,
		"page":       page,
		"pageSize":   pageSize,
		"total":      total,
		"totalPages": totalPages,
		"hasNext":    page < totalPages,
		"hasPrev":    page > 1,
	}

	c.JSON(http.StatusOK, paginatedResponse)
}

// SyncCalendarEvents syncs events from Recall.ai to local database
func SyncCalendarEvents(c *gin.Context) {
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	calendarID := c.Param("id")

	// Verify calendar belongs to user
	var calendar models.Calendar
	if err := db.DB.Where("id = ? AND clerk_user_id = ?", calendarID, clerkUserID).First(&calendar).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Calendar not found"})
		return
	}

	// Fetch events from Recall.ai
	recallClient := recallai.NewClient()
	events, err := recallClient.ListCalendarEvents(calendar.RecallCalendarID)
	if err != nil {
		log.Error().Err(err).Str("recall_calendar_id", calendar.RecallCalendarID).Msg("Error fetching events from Recall")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events from Recall.ai"})
		return
	}

	// Sync events to database
	syncedCount := 0
	for _, event := range events {
		startTime, _ := time.Parse(time.RFC3339, event.StartTime)
		endTime, _ := time.Parse(time.RFC3339, event.EndTime)

		calendarEvent := models.CalendarEvent{
			CalendarID:      calendarID,
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

		// Upsert event
		if err := db.DB.Where(models.CalendarEvent{RecallEventID: event.ID}).
			Assign(calendarEvent).
			FirstOrCreate(&calendarEvent).Error; err != nil {
			log.Error().Err(err).Str("event_id", event.ID).Msg("Error upserting calendar event")
			continue
		}

		syncedCount++
	}

	// Update last synced timestamp
	db.DB.Model(&calendar).Update("last_synced_at", time.Now())

	log.Info().
		Str("clerk_user_id", clerkUserID).
		Str("calendar_id", calendarID).
		Int("synced_count", syncedCount).
		Msg("Calendar events synced successfully")

	c.JSON(http.StatusOK, gin.H{
		"message":      "Events synced successfully",
		"syncedCount":  syncedCount,
		"lastSyncedAt": time.Now(),
	})
}

// ScheduleBotForEvent schedules a bot to join a specific calendar event
func ScheduleBotForEvent(c *gin.Context) {
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	eventID := c.Param("eventId")

	// Find event and verify ownership
	var event models.CalendarEvent
	if err := db.DB.Preload("Calendar").Where("id = ?", eventID).First(&event).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	if event.Calendar.ClerkUserID != clerkUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	if event.BotScheduled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bot already scheduled for this event"})
		return
	}

	// Schedule bot via Recall.ai
	recallClient := recallai.NewClient()

	// Use ICalUID as deduplication key (ensures one bot per recurring event series)
	deduplicationKey := event.ICalUID
	if deduplicationKey == "" {
		deduplicationKey = event.RecallEventID
	}

	botConfig := map[string]interface{}{
		"recording_config": map[string]interface{}{
			"transcript": map[string]interface{}{
				"provider": map[string]interface{}{
					"recallai_streaming": map[string]interface{}{
						"mode": "prioritize_accuracy",
					},
				},
			},
			"video_mixed_layout": "gallery_view_v2",
			"video_separate_mp4": map[string]interface{}{},
		},
	}

	updatedEvent, err := recallClient.ScheduleBotForEvent(event.RecallEventID, deduplicationKey, botConfig)
	if err != nil {
		log.Error().Err(err).Str("event_id", eventID).Msg("Error scheduling bot for event")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to schedule bot"})
		return
	}

	// Update event in database
	event.BotScheduled = true
	if len(updatedEvent.Bots) > 0 {
		event.BotID = &updatedEvent.Bots[0].BotID
	}

	if err := db.DB.Save(&event).Error; err != nil {
		log.Error().Err(err).Msg("Error updating event with bot info")
	}

	log.Info().
		Str("clerk_user_id", clerkUserID).
		Str("event_id", eventID).
		Msg("Bot scheduled for calendar event")

	c.JSON(http.StatusOK, gin.H{
		"message": "Bot scheduled successfully",
		"event":   event,
	})
}

// CancelBotForEvent cancels a scheduled bot for an event
func CancelBotForEvent(c *gin.Context) {
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	eventID := c.Param("eventId")

	// Find event and verify ownership
	var event models.CalendarEvent
	if err := db.DB.Preload("Calendar").Where("id = ?", eventID).First(&event).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	if event.Calendar.ClerkUserID != clerkUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	if !event.BotScheduled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No bot scheduled for this event"})
		return
	}

	// Cancel bot via Recall.ai
	// Note: Recall's managed scheduling endpoint doesn't require botID in the path
	recallClient := recallai.NewClient()
	botIDStr := ""
	if event.BotID != nil {
		botIDStr = *event.BotID
	}
	if err := recallClient.RemoveBotFromEvent(event.RecallEventID, botIDStr); err != nil {
		log.Error().Err(err).Str("event_id", eventID).Msg("Error removing bot from event")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel bot"})
		return
	}

	// Update event in database
	event.BotScheduled = false
	event.BotID = nil

	if err := db.DB.Save(&event).Error; err != nil {
		log.Error().Err(err).Msg("Error updating event")
	}

	log.Info().
		Str("clerk_user_id", clerkUserID).
		Str("event_id", eventID).
		Msg("Bot cancelled for calendar event")

	c.JSON(http.StatusOK, gin.H{
		"message": "Bot cancelled successfully",
	})
}

// verifyWebhookSignature verifies the Recall.ai webhook signature
func verifyWebhookSignature(c *gin.Context, payload []byte) bool {
	webhookSecret := os.Getenv("RECALL_CALENDAR_WEBHOOK_SECRET")
	if webhookSecret == "" {
		// If no secret is configured, skip verification (dev mode)
		log.Warn().Msg("RECALL_CALENDAR_WEBHOOK_SECRET not set, skipping signature verification")
		return true
	}

	// Get signature from header
	signature := c.GetHeader("X-Recall-Signature")
	if signature == "" {
		signature = c.GetHeader("Recall-Signature")
	}

	if signature == "" {
		log.Error().Msg("Missing webhook signature header")
		return false
	}

	// Compute expected signature
	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	// Compare signatures
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// HandleCalendarWebhook handles webhooks from Recall.ai for calendar sync events
func HandleCalendarWebhook(c *gin.Context) {
	// Read raw body for signature verification
	rawBody, err := c.GetRawData()
	if err != nil {
		log.Error().Err(err).Msg("Failed to read webhook body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Verify webhook signature
	if !verifyWebhookSignature(c, rawBody) {
		log.Error().Msg("Invalid webhook signature")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
		return
	}

	// Parse webhook payload from raw body
	var webhook recallai.CalendarSyncWebhook
	if err := json.Unmarshal(rawBody, &webhook); err != nil {
		log.Error().Err(err).Msg("Invalid webhook payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	log.Info().
		Str("event", webhook.Event).
		Str("calendar_id", webhook.Data.CalendarID).
		Str("last_updated_ts", webhook.Data.LastUpdatedTS).
		Msg("Received calendar webhook")

	// Find calendar in database
	var calendar models.Calendar
	if err := db.DB.Where("recall_calendar_id = ?", webhook.Data.CalendarID).First(&calendar).Error; err != nil {
		log.Error().Err(err).Str("recall_calendar_id", webhook.Data.CalendarID).Msg("Calendar not found for webhook")
		c.JSON(http.StatusNotFound, gin.H{"error": "Calendar not found"})
		return
	}

	// Trigger sync in background
	go syncCalendarInBackground(calendar)

	c.JSON(http.StatusOK, gin.H{"message": "Webhook received"})
}

// syncCalendarInBackground syncs calendar events in the background
func syncCalendarInBackground(calendar models.Calendar) {
	recallClient := recallai.NewClient()
	events, err := recallClient.ListCalendarEvents(calendar.RecallCalendarID)
	if err != nil {
		log.Error().Err(err).Str("calendar_id", calendar.ID).Msg("Error syncing calendar events from webhook")
		return
	}

	syncedCount := 0
	for _, event := range events {
		startTime, _ := time.Parse(time.RFC3339, event.StartTime)
		endTime, _ := time.Parse(time.RFC3339, event.EndTime)

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
			log.Error().Err(err).Str("event_id", event.ID).Msg("Error upserting calendar event from webhook")
			continue
		}

		syncedCount++
	}

	// Update last synced timestamp
	db.DB.Model(&calendar).Update("last_synced_at", time.Now())

	log.Info().
		Str("calendar_id", calendar.ID).
		Int("synced_count", syncedCount).
		Msg("Background calendar sync completed")
}
