package services

import (
	"backend/db"
	"backend/internal/models"
	"backend/pkg/recallai"
	"time"

	"github.com/rs/zerolog/log"
)

// CalendarSchedulerService handles automatic bot scheduling for calendar events
type CalendarSchedulerService struct {
	recallClient *recallai.Client
}

// NewCalendarSchedulerService creates a new calendar scheduler service
func NewCalendarSchedulerService() *CalendarSchedulerService {
	return &CalendarSchedulerService{
		recallClient: recallai.NewClient(),
	}
}

// StartAutoScheduler starts the background scheduler that auto-schedules bots for upcoming meetings
// It runs periodically to check for new events and schedule bots accordingly
func (s *CalendarSchedulerService) StartAutoScheduler() {
	ticker := time.NewTicker(5 * time.Minute) // Check every 5 minutes
	defer ticker.Stop()

	log.Info().Msg("Calendar auto-scheduler started")

	// Run immediately on start
	s.scheduleUpcomingMeetings()

	// Then run periodically
	for range ticker.C {
		s.scheduleUpcomingMeetings()
	}
}

// scheduleUpcomingMeetings finds upcoming meetings and schedules bots for them
func (s *CalendarSchedulerService) scheduleUpcomingMeetings() {
	log.Debug().Msg("Running calendar auto-scheduler")

	// Get all active calendars
	var calendars []models.Calendar
	if err := db.DB.Where("status = ?", "active").Find(&calendars).Error; err != nil {
		log.Error().Err(err).Msg("Error fetching active calendars")
		return
	}

	for _, calendar := range calendars {
		s.syncAndScheduleCalendar(calendar)
	}
}

// syncAndScheduleCalendar syncs events for a calendar and schedules bots for eligible meetings
func (s *CalendarSchedulerService) syncAndScheduleCalendar(calendar models.Calendar) {
	// Fetch latest events from Recall
	events, err := s.recallClient.ListCalendarEvents(calendar.RecallCalendarID)
	if err != nil {
		log.Error().
			Err(err).
			Str("calendar_id", calendar.ID).
			Msg("Error fetching events for calendar")
		return
	}

	scheduledCount := 0
	for _, event := range events {
		if s.shouldScheduleBot(event) {
			if err := s.scheduleBot(calendar, event); err != nil {
				log.Error().
					Err(err).
					Str("event_id", event.ID).
					Msg("Error scheduling bot for event")
				continue
			}
			scheduledCount++
		}

		// Also update local database
		s.upsertEvent(calendar.ID, event)
	}

	if scheduledCount > 0 {
		log.Info().
			Str("calendar_id", calendar.ID).
			Int("scheduled_count", scheduledCount).
			Msg("Auto-scheduled bots for calendar events")
	}

	// Update last synced timestamp
	db.DB.Model(&calendar).Update("last_synced_at", time.Now())
}

// shouldScheduleBot determines if a bot should be scheduled for an event
func (s *CalendarSchedulerService) shouldScheduleBot(event recallai.CalendarEvent) bool {
	// Skip if event is deleted
	if event.IsDeleted {
		return false
	}

	// Skip if bot already scheduled
	if len(event.Bots) > 0 {
		return false
	}

	// Skip if no meeting URL
	if event.MeetingURL == "" {
		return false
	}

	// Only schedule for supported platforms
	supportedPlatforms := map[string]bool{
		"zoom":            true,
		"google_meet":     true,
		"microsoft_teams": true,
		"webex":           true,
	}

	if !supportedPlatforms[event.MeetingPlatform] {
		return false
	}

	// Parse start time
	startTime, err := time.Parse(time.RFC3339, event.StartTime)
	if err != nil {
		log.Error().Err(err).Str("event_id", event.ID).Msg("Error parsing event start time")
		return false
	}

	// Only schedule for future events (within next 7 days)
	now := time.Now()
	sevenDaysFromNow := now.Add(7 * 24 * time.Hour)

	if startTime.Before(now) || startTime.After(sevenDaysFromNow) {
		return false
	}

	return true
}

// scheduleBot schedules a bot for a calendar event
func (s *CalendarSchedulerService) scheduleBot(calendar models.Calendar, event recallai.CalendarEvent) error {
	// Use ICalUID as deduplication key to handle recurring events
	deduplicationKey := event.ICalUID
	if deduplicationKey == "" {
		deduplicationKey = event.ID
	}

	// Bot configuration with transcription and recording
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

	// Schedule bot via Recall API
	_, err := s.recallClient.ScheduleBotForEvent(event.ID, deduplicationKey, botConfig)
	if err != nil {
		return err
	}

	log.Info().
		Str("calendar_id", calendar.ID).
		Str("event_id", event.ID).
		Str("meeting_platform", event.MeetingPlatform).
		Str("start_time", event.StartTime).
		Msg("Auto-scheduled bot for calendar event")

	return nil
}

// upsertEvent updates or creates a calendar event in the local database
func (s *CalendarSchedulerService) upsertEvent(calendarID string, event recallai.CalendarEvent) {
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

	if err := db.DB.Where(models.CalendarEvent{RecallEventID: event.ID}).
		Assign(calendarEvent).
		FirstOrCreate(&calendarEvent).Error; err != nil {
		log.Error().Err(err).Str("event_id", event.ID).Msg("Error upserting calendar event")
	}
}

// SyncCalendar manually syncs a specific calendar
func (s *CalendarSchedulerService) SyncCalendar(calendarID string) error {
	var calendar models.Calendar
	if err := db.DB.Where("id = ?", calendarID).First(&calendar).Error; err != nil {
		return err
	}

	s.syncAndScheduleCalendar(calendar)
	return nil
}
