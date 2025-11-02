package controllers

import (
	"backend/db"
	"backend/internal/middleware"
	"backend/internal/models"
	"backend/internal/services"
	"backend/pkg/recallai"
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// StartRecordingRequest represents the request payload for starting a meeting recording
type StartRecordingRequest struct {
	MeetingURL string `json:"meeting_url" binding:"required"`
}

// isValidMeetingURL validates if the provided URL is a valid meeting URL
func isValidMeetingURL(meetingURL string) bool {
	// Parse the URL
	parsedURL, err := url.Parse(meetingURL)
	if err != nil {
		return false
	}

	// Check if it's a valid HTTP/HTTPS URL
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return false
	}

	// Check if it's from a supported meeting platform
	supportedDomains := []string{
		"meet.google.com",
		"zoom.us",
		"teams.microsoft.com",
		"webex.com",
		"gotomeeting.com",
	}

	for _, domain := range supportedDomains {
		if parsedURL.Host == domain ||
			(len(parsedURL.Host) > len(domain) &&
				parsedURL.Host[len(parsedURL.Host)-len(domain)-1:] == "."+domain) {
			return true
		}
	}

	// Allow any domain for now (Recall.ai supports many platforms)
	return parsedURL.Host != ""
}

// GetUserMeetings retrieves all meeting recordings for the authenticated user
func GetUserMeetings(ctx *gin.Context) {
	// Get authenticated user ID
	clerkUserID, exists := middleware.GetClerkUserID(ctx)
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	log.Debug().
		Str("clerk_user_id", clerkUserID).
		Msg("Retrieving user meetings")

	var meetings []models.MeetingRecording
	// Remove Preload to avoid loading full note with chapter and notebook
	err := db.DB.Where("clerk_user_id = ?", clerkUserID).
		Order("created_at DESC").
		Find(&meetings).Error

	if err != nil {
		log.Error().
			Err(err).
			Str("clerk_user_id", clerkUserID).
			Msg("Failed to retrieve user meetings")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve meetings: " + err.Error()})
		return
	}

	// Build response with meeting list items
	response := make([]gin.H, len(meetings))
	for i, meeting := range meetings {
		response[i] = gin.H{
			"id":                    meeting.ID,
			"clerkUserId":           meeting.ClerkUserID,
			"botId":                 meeting.BotID,
			"meetingUrl":            meeting.MeetingURL,
			"status":                meeting.Status,
			"recallRecordingId":     meeting.RecallRecordingID,
			"transcriptDownloadUrl": meeting.TranscriptDownloadURL,
			"videoDownloadUrl":      meeting.VideoDownloadURL,
			"generatedNoteId":       meeting.GeneratedNoteID,
			"createdAt":             meeting.CreatedAt,
			"updatedAt":             meeting.UpdatedAt,
			"completedAt":           meeting.CompletedAt,
		}
	}

	log.Debug().
		Str("clerk_user_id", clerkUserID).
		Int("meetings_count", len(meetings)).
		Msg("Successfully retrieved user meetings")

	ctx.JSON(http.StatusOK, gin.H{"data": response})
}

// StartMeetingRecording is the package-level function for starting meeting recordings
func StartMeetingRecording(ctx *gin.Context) {
	// Get authenticated user ID
	clerkUserID, exists := middleware.GetClerkUserID(ctx)
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req StartRecordingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Print("Invalid request data for meeting recording: ", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}

	// Validate meeting URL format
	if !isValidMeetingURL(req.MeetingURL) {
		log.Print("Invalid meeting URL format: ", req.MeetingURL)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid meeting URL format"})
		return
	}

	log.Info().
		Str("clerk_user_id", clerkUserID).
		Str("meeting_url", req.MeetingURL).
		Msg("Starting meeting recording")

	// Create Recall.ai client and bot
	recallClient := recallai.NewClient()

	// Create bot in Recall.ai
	botResp, err := recallClient.CreateBot(req.MeetingURL)
	if err != nil {
		log.Error().
			Err(err).
			Str("clerk_user_id", clerkUserID).
			Str("meeting_url", req.MeetingURL).
			Msg("Failed to create Recall.ai bot")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create meeting bot: " + err.Error()})
		return
	}

	// Save meeting recording to database
	recording := &models.MeetingRecording{
		ClerkUserID: clerkUserID,
		BotID:       botResp.ID,
		MeetingURL:  req.MeetingURL,
		Status:      "pending",
	}

	if err := db.DB.Create(recording).Error; err != nil {
		log.Error().
			Err(err).
			Str("clerk_user_id", clerkUserID).
			Str("bot_id", botResp.ID).
			Msg("Failed to save meeting recording to database")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save meeting recording: " + err.Error()})
		return
	}

	log.Info().
		Str("recording_id", recording.ID).
		Str("bot_id", botResp.ID).
		Msg("Successfully started meeting recording")

	ctx.JSON(http.StatusCreated, gin.H{
		"data":    recording,
		"message": "Bot created successfully and joining meeting",
	})
}

// GetMeetingTranscript retrieves and formats the transcript for a specific meeting
func GetMeetingTranscript(ctx *gin.Context) {
	// Get authenticated user ID
	clerkUserID, exists := middleware.GetClerkUserID(ctx)
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	meetingID := ctx.Param("id")
	if meetingID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Meeting ID is required"})
		return
	}

	// Find the meeting recording
	var recording models.MeetingRecording
	if err := db.DB.Where("id = ? AND clerk_user_id = ?", meetingID, clerkUserID).First(&recording).Error; err != nil {
		log.Error().
			Err(err).
			Str("meeting_id", meetingID).
			Str("clerk_user_id", clerkUserID).
			Msg("Meeting recording not found")
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Meeting not found"})
		return
	}

	// Fetch fresh download URL from Recall.ai (pre-signed URLs expire)
	recallClient := recallai.NewClient()
	botDetails, err := recallClient.GetBot(recording.BotID)
	if err != nil {
		log.Error().
			Err(err).
			Str("meeting_id", meetingID).
			Str("bot_id", recording.BotID).
			Msg("Failed to fetch bot details from Recall.ai")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch meeting details"})
		return
	}

	// Check if recordings exist and get fresh transcript URL
	if len(botDetails.Recordings) == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"transcript": nil,
			"status":     recording.Status,
			"message":    "Transcript not yet available",
		})
		return
	}

	transcriptURL := botDetails.Recordings[0].MediaShortcuts.Transcript.Data.DownloadURL
	if transcriptURL == "" {
		ctx.JSON(http.StatusOK, gin.H{
			"transcript": nil,
			"status":     recording.Status,
			"message":    "Transcript not yet available",
		})
		return
	}

	// Download and parse transcript using fresh URL
	transcript, err := recallClient.DownloadTranscript(transcriptURL)
	if err != nil {
		log.Error().
			Err(err).
			Str("meeting_id", meetingID).
			Msg("Failed to download transcript")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to download transcript"})
		return
	}

	// Format transcript for display
	type TranscriptEntry struct {
		Speaker   string `json:"speaker"`
		Text      string `json:"text"`
		Timestamp string `json:"timestamp"`
	}

	var formattedTranscript []TranscriptEntry
	for _, entry := range transcript {
		speaker := entry.Participant.Name
		if speaker == "" {
			speaker = "Unknown Speaker"
		}

		// Combine all words into sentences
		var words []string
		for _, word := range entry.Words {
			words = append(words, word.Text)
		}

		if len(words) > 0 {
			text := words[0]
			for i := 1; i < len(words); i++ {
				text += " " + words[i]
			}

			timestamp := ""
			if len(entry.Words) > 0 {
				timestamp = entry.Words[0].StartTimestamp.Absolute
			}

			formattedTranscript = append(formattedTranscript, TranscriptEntry{
				Speaker:   speaker,
				Text:      text,
				Timestamp: timestamp,
			})
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"transcript":  formattedTranscript,
		"meetingUrl":  recording.MeetingURL,
		"status":      recording.Status,
		"createdAt":   recording.CreatedAt,
		"completedAt": recording.UpdatedAt,
	})
}

// HandleRecallWebhook is the package-level function for handling Recall.ai webhooks
func HandleRecallWebhook(ctx *gin.Context) {
	log.Debug().Msg("Received Recall.ai webhook")

	var payload map[string]interface{}
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		log.Error().
			Err(err).
			Msg("Invalid webhook payload")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook payload: " + err.Error()})
		return
	}

	// Check event type
	eventType, ok := payload["event"].(string)
	if !ok {
		log.Error().
			Interface("payload", payload).
			Msg("Missing or invalid event type in webhook")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid event type"})
		return
	}

	log.Info().
		Str("event_type", eventType).
		Msg("Processing Recall.ai webhook event")

	// Handle transcript.done event
	if eventType == "transcript.done" {
		// Extract bot_id from nested structure: data.bot.id
		var botID string
		if data, ok := payload["data"].(map[string]interface{}); ok {
			if bot, ok := data["bot"].(map[string]interface{}); ok {
				if id, ok := bot["id"].(string); ok {
					botID = id
				}
			}
		}

		// Fallback: try direct bot_id field for backward compatibility
		if botID == "" {
			if id, ok := payload["bot_id"].(string); ok {
				botID = id
			}
		}

		if botID == "" {
			log.Error().
				Interface("payload", payload).
				Msg("Missing bot_id in transcript.done webhook")
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing bot_id in webhook payload"})
			return
		}

		log.Info().
			Str("bot_id", botID).
			Msg("Processing transcript.done webhook")

		// Update the recording status
		var recording models.MeetingRecording
		if err := db.DB.Where("bot_id = ?", botID).First(&recording).Error; err != nil {
			log.Error().
				Err(err).
				Str("bot_id", botID).
				Msg("Meeting recording not found for webhook")
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Meeting recording not found"})
			return
		}

		// Fetch transcript download URL from Recall.ai
		recallClient := recallai.NewClient()
		botDetails, err := recallClient.GetBot(botID)
		if err != nil {
			log.Error().
				Err(err).
				Str("bot_id", botID).
				Msg("Failed to fetch bot details from Recall.ai")
		} else if len(botDetails.Recordings) > 0 {
			// Get recording ID and transcript URL
			firstRecording := botDetails.Recordings[0]
			recording.RecallRecordingID = firstRecording.ID

			if firstRecording.MediaShortcuts.Transcript.Data.DownloadURL != "" {
				recording.TranscriptDownloadURL = firstRecording.MediaShortcuts.Transcript.Data.DownloadURL
				log.Info().
					Str("recording_id", recording.ID).
					Str("transcript_url", recording.TranscriptDownloadURL).
					Msg("Retrieved transcript download URL")

				// Download and parse the transcript
				transcript, err := recallClient.DownloadTranscript(recording.TranscriptDownloadURL)
				if err != nil {
					log.Error().
						Err(err).
						Str("transcript_url", recording.TranscriptDownloadURL).
						Msg("Failed to download transcript")
				} else {
					// Convert transcript to plain text and store it
					var transcriptText string
					for _, entry := range transcript {
						speaker := entry.Participant.Name
						if speaker == "" {
							speaker = "Unknown Speaker"
						}

						// Combine all words into a single sentence
						var words []string
						for _, word := range entry.Words {
							words = append(words, word.Text)
						}
						sentence := ""
						if len(words) > 0 {
							sentence = words[0]
							for i := 1; i < len(words); i++ {
								sentence += " " + words[i]
							}
						}

						if sentence != "" {
							transcriptText += speaker + ": " + sentence + "\n\n"
						}
					}

					// Store the raw transcript as JSON in TranscriptRaw field
					_, err := json.Marshal(transcript)
					if err == nil {
						// Create a note with the transcript
						// For now, just log it - we'll create the note in the next step
						log.Info().
							Str("recording_id", recording.ID).
							Int("transcript_entries", len(transcript)).
							Int("transcript_length", len(transcriptText)).
							Str("transcript_preview", transcriptText[:min(200, len(transcriptText))]).
							Msg("Successfully downloaded and parsed transcript")

						// TODO: Create a note with the transcript content
						// This will be implemented next to organize it into notebooks/chapters
					}
				}
			}
		}

		// Update status to completed
		recording.Status = "completed"
		if err := db.DB.Save(&recording).Error; err != nil {
			log.Error().
				Err(err).
				Str("recording_id", recording.ID).
				Msg("Failed to update recording status")
		} else {
			log.Info().
				Str("recording_id", recording.ID).
				Str("bot_id", botID).
				Str("recall_recording_id", recording.RecallRecordingID).
				Msg("Successfully updated recording status to completed")
		}

		// Trigger note generation asynchronously
		go func() {
			noteService := services.NewMeetingNoteService()
			ctx := context.Background()

			log.Info().
				Str("meeting_id", recording.ID).
				Msg("Starting automatic note generation from transcript")

			if err := noteService.ProcessMeetingTranscript(ctx, &recording); err != nil {
				log.Error().
					Err(err).
					Str("meeting_id", recording.ID).
					Msg("Failed to generate note from transcript")
			} else {
				log.Info().
					Str("meeting_id", recording.ID).
					Str("note_id", *recording.GeneratedNoteID).
					Msg("Successfully generated note from transcript")
			}
		}()

		ctx.JSON(http.StatusOK, gin.H{"status": "received", "message": "Transcript processing completed"})
		return
	}

	// Handle recording.done event to also update status
	if eventType == "recording.done" || eventType == "bot.done" {
		var botID string
		if data, ok := payload["data"].(map[string]interface{}); ok {
			if bot, ok := data["bot"].(map[string]interface{}); ok {
				if id, ok := bot["id"].(string); ok {
					botID = id
				}
			}
		}

		if botID != "" {
			var recording models.MeetingRecording
			if err := db.DB.Where("bot_id = ?", botID).First(&recording).Error; err == nil {
				recording.Status = "completed"
				db.DB.Save(&recording)
				log.Info().
					Str("recording_id", recording.ID).
					Str("event_type", eventType).
					Msg("Updated recording status to completed")
			}
		}
	}

	// Handle bot status events to update recording status
	if eventType == "bot.in_call_recording" {
		var botID string
		if data, ok := payload["data"].(map[string]interface{}); ok {
			if bot, ok := data["bot"].(map[string]interface{}); ok {
				if id, ok := bot["id"].(string); ok {
					botID = id
				}
			}
		}

		if botID != "" {
			var recording models.MeetingRecording
			if err := db.DB.Where("bot_id = ?", botID).First(&recording).Error; err == nil {
				recording.Status = "recording"
				db.DB.Save(&recording)
				log.Info().
					Str("recording_id", recording.ID).
					Msg("Updated recording status to recording")
			}
		}
	}

	// Handle calendar events
	if eventType == "calendar.update" || eventType == "calendar.sync_events" {
		log.Info().
			Str("event_type", eventType).
			Msg("Detected calendar event, processing inline")

		// Process calendar webhook inline (body already consumed)
		if data, ok := payload["data"].(map[string]interface{}); ok {
			calendarID, _ := data["calendar_id"].(string)
			lastUpdatedTS, _ := data["last_updated_ts"].(string)

			log.Info().
				Str("calendar_id", calendarID).
				Str("last_updated_ts", lastUpdatedTS).
				Msg("Processing calendar sync event")

			// Find calendar in database
			var calendar models.Calendar
			if err := db.DB.Where("recall_calendar_id = ?", calendarID).First(&calendar).Error; err != nil {
				log.Error().
					Err(err).
					Str("recall_calendar_id", calendarID).
					Msg("Calendar not found for webhook")
				ctx.JSON(http.StatusOK, gin.H{"status": "calendar_not_found", "message": "Calendar not in database yet"})
				return
			}

			// Trigger sync in background
			go syncCalendarInBackground(calendar)

			ctx.JSON(http.StatusOK, gin.H{"status": "received", "message": "Calendar sync triggered"})
		} else {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid calendar webhook data"})
		}
		return
	}

	// Handle other event types (for future expansion)
	log.Info().
		Str("event_type", eventType).
		Msg("Received unhandled webhook event type")

	ctx.JSON(http.StatusOK, gin.H{"status": "received", "message": "Event acknowledged"})
}

// BackfillVideoURLs fetches and updates video URLs for existing meetings that have recording IDs
func BackfillVideoURLs(ctx *gin.Context) {
	// Get authenticated user ID
	clerkUserID, exists := middleware.GetClerkUserID(ctx)
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	log.Info().
		Str("clerk_user_id", clerkUserID).
		Msg("Starting video URL backfill for user meetings")

	// Find all completed meetings for this user that have recall_recording_id but missing video_download_url
	var meetings []models.MeetingRecording
	err := db.DB.Where("clerk_user_id = ? AND status = ? AND recall_recording_id != ? AND (video_download_url IS NULL OR video_download_url = ?)",
		clerkUserID, "completed", "", "").
		Find(&meetings).Error

	if err != nil {
		log.Error().
			Err(err).
			Str("clerk_user_id", clerkUserID).
			Msg("Failed to retrieve meetings for backfill")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve meetings"})
		return
	}

	if len(meetings) == 0 {
		log.Info().
			Str("clerk_user_id", clerkUserID).
			Msg("No meetings found needing video URL backfill")
		ctx.JSON(http.StatusOK, gin.H{
			"message":        "No meetings need video URL updates",
			"meetings_count": 0,
			"updated_count":  0,
		})
		return
	}

	log.Info().
		Str("clerk_user_id", clerkUserID).
		Int("meetings_count", len(meetings)).
		Msg("Found meetings needing video URL backfill")

	// Initialize Recall.ai client
	recallClient := recallai.NewClient()
	updatedCount := 0
	failedCount := 0
	errors := []string{}

	// Process each meeting
	for _, meeting := range meetings {
		log.Debug().
			Str("meeting_id", meeting.ID).
			Str("bot_id", meeting.BotID).
			Msg("Fetching video URL for meeting")

		// Get bot details from Recall.ai
		botDetails, err := recallClient.GetBot(meeting.BotID)
		if err != nil {
			log.Error().
				Err(err).
				Str("meeting_id", meeting.ID).
				Str("bot_id", meeting.BotID).
				Msg("Failed to get bot details from Recall.ai")
			failedCount++
			errors = append(errors, "Meeting "+meeting.ID+": "+err.Error())
			continue
		}

		// Check if recordings exist
		if len(botDetails.Recordings) == 0 {
			log.Warn().
				Str("meeting_id", meeting.ID).
				Str("bot_id", meeting.BotID).
				Msg("No recordings found for bot")
			failedCount++
			errors = append(errors, "Meeting "+meeting.ID+": No recordings found")
			continue
		}

		// Extract video URL
		videoURL := botDetails.Recordings[0].MediaShortcuts.VideoMixed.Data.DownloadURL
		if videoURL == "" {
			log.Warn().
				Str("meeting_id", meeting.ID).
				Str("bot_id", meeting.BotID).
				Msg("Video URL not available for this recording")
			failedCount++
			errors = append(errors, "Meeting "+meeting.ID+": Video URL not available")
			continue
		}

		// Update the meeting record
		meeting.VideoDownloadURL = videoURL
		if err := db.DB.Save(&meeting).Error; err != nil {
			log.Error().
				Err(err).
				Str("meeting_id", meeting.ID).
				Msg("Failed to update meeting with video URL")
			failedCount++
			errors = append(errors, "Meeting "+meeting.ID+": Failed to save")
			continue
		}

		updatedCount++
		log.Info().
			Str("meeting_id", meeting.ID).
			Bool("has_video", videoURL != "").
			Msg("Successfully updated meeting with video URL")
	}

	log.Info().
		Str("clerk_user_id", clerkUserID).
		Int("total_meetings", len(meetings)).
		Int("updated_count", updatedCount).
		Int("failed_count", failedCount).
		Msg("Completed video URL backfill")

	response := gin.H{
		"message":        "Video URL backfill completed",
		"meetings_count": len(meetings),
		"updated_count":  updatedCount,
		"failed_count":   failedCount,
	}

	if len(errors) > 0 {
		response["errors"] = errors
	}

	ctx.JSON(http.StatusOK, response)
}
