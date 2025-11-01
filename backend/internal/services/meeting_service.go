package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"backend/db"
	"backend/internal/models"
	"backend/pkg/recallai"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type MeetingService struct {
	db           *gorm.DB
	recallClient *recallai.Client
	aiService    *AIService
}

// NewMeetingService creates a new meeting service instance
func NewMeetingService() *MeetingService {
	return &MeetingService{
		db:           db.DB,
		recallClient: recallai.NewClient(),
		aiService:    NewAIService(),
	}
}

// StartMeetingRecording initiates a new meeting recording by creating a Recall.ai bot
func (s *MeetingService) StartMeetingRecording(ctx context.Context, userID uint, meetingURL string) (*models.MeetingRecording, error) {
	if meetingURL == "" {
		return nil, fmt.Errorf("meeting URL cannot be empty")
	}

	log.Info().
		Uint("user_id", userID).
		Str("meeting_url", meetingURL).
		Msg("Starting meeting recording")

	// Create bot in Recall.ai
	botResp, err := s.recallClient.CreateBot(meetingURL)
	if err != nil {
		log.Error().
			Err(err).
			Uint("user_id", userID).
			Str("meeting_url", meetingURL).
			Msg("Failed to create Recall.ai bot")
		return nil, fmt.Errorf("failed to create recall.ai bot: %w", err)
	}

	// Save meeting recording to database
	recording := &models.MeetingRecording{
		UserID:     userID,
		BotID:      botResp.ID,
		MeetingURL: meetingURL,
		Status:     "pending",
	}

	if err := s.db.Create(recording).Error; err != nil {
		log.Error().
			Err(err).
			Uint("user_id", userID).
			Str("bot_id", botResp.ID).
			Msg("Failed to save meeting recording to database")
		return nil, fmt.Errorf("failed to save meeting recording: %w", err)
	}

	log.Info().
		Str("recording_id", recording.ID).
		Str("bot_id", botResp.ID).
		Msg("Successfully started meeting recording")

	return recording, nil
}

// GetUserMeetings retrieves all meeting recordings for a specific user
func (s *MeetingService) GetUserMeetings(ctx context.Context, userID uint) ([]models.MeetingRecording, error) {
	var meetings []models.MeetingRecording

	err := s.db.Where("user_id = ?", userID).
		Preload("Participants").
		Preload("GeneratedNote").
		Order("created_at DESC").
		Find(&meetings).Error

	if err != nil {
		log.Error().
			Err(err).
			Uint("user_id", userID).
			Msg("Failed to retrieve user meetings")
		return nil, fmt.Errorf("failed to retrieve meetings: %w", err)
	}

	log.Debug().
		Uint("user_id", userID).
		Int("meetings_count", len(meetings)).
		Msg("Retrieved user meetings")

	return meetings, nil
}

// ProcessCompletedMeeting processes a completed meeting transcript
func (s *MeetingService) ProcessCompletedMeeting(ctx context.Context, botID string) error {
	if botID == "" {
		return fmt.Errorf("bot ID cannot be empty")
	}

	log.Info().
		Str("bot_id", botID).
		Msg("Processing completed meeting")

	// Get meeting recording from database
	var recording models.MeetingRecording
	if err := s.db.Where("bot_id = ?", botID).First(&recording).Error; err != nil {
		log.Error().
			Err(err).
			Str("bot_id", botID).
			Msg("Meeting recording not found")
		return fmt.Errorf("meeting recording not found: %w", err)
	}

	// Update status to processing
	recording.Status = "processing"
	if err := s.db.Save(&recording).Error; err != nil {
		log.Error().
			Err(err).
			Str("recording_id", recording.ID).
			Msg("Failed to update recording status to processing")
	}

	// Get bot details from Recall.ai
	botDetails, err := s.recallClient.GetBot(botID)
	if err != nil {
		s.updateRecordingStatus(&recording, "failed")
		log.Error().
			Err(err).
			Str("bot_id", botID).
			Msg("Failed to get bot details from Recall.ai")
		return fmt.Errorf("failed to get bot details: %w", err)
	}

	if len(botDetails.Recordings) == 0 {
		s.updateRecordingStatus(&recording, "failed")
		log.Error().
			Str("bot_id", botID).
			Msg("No recordings found for bot")
		return fmt.Errorf("no recordings found for bot")
	}

	// Get transcript and video download URLs
	transcriptURL := botDetails.Recordings[0].MediaShortcuts.Transcript.Data.DownloadURL
	videoURL := botDetails.Recordings[0].MediaShortcuts.VideoMixed.Data.DownloadURL
	recording.RecallRecordingID = botDetails.Recordings[0].ID
	recording.TranscriptDownloadURL = transcriptURL
	recording.VideoDownloadURL = videoURL
	if err := s.db.Save(&recording).Error; err != nil {
		log.Error().
			Err(err).
			Str("recording_id", recording.ID).
			Msg("Failed to update recording with transcript and video URLs")
	}
	
	log.Info().
		Str("recording_id", recording.ID).
		Bool("has_transcript", transcriptURL != "").
		Bool("has_video", videoURL != "").
		Msg("Retrieved recording URLs from Recall.ai")

	// Download transcript
	transcript, err := s.recallClient.DownloadTranscript(transcriptURL)
	if err != nil {
		s.updateRecordingStatus(&recording, "failed")
		log.Error().
			Err(err).
			Str("transcript_url", transcriptURL).
			Msg("Failed to download transcript")
		return fmt.Errorf("failed to download transcript: %w", err)
	}

	// Process transcript and create note
	if err := s.processTranscriptAndCreateNote(ctx, &recording, transcript); err != nil {
		s.updateRecordingStatus(&recording, "failed")
		return fmt.Errorf("failed to process transcript: %w", err)
	}

	// Update recording status to completed
	now := time.Now()
	recording.Status = "completed"
	recording.CompletedAt = &now
	if err := s.db.Save(&recording).Error; err != nil {
		log.Error().
			Err(err).
			Str("recording_id", recording.ID).
			Msg("Failed to update recording status to completed")
	}

	log.Info().
		Str("recording_id", recording.ID).
		Msg("Successfully processed completed meeting")

	return nil
}

// updateRecordingStatus is a helper method to update recording status
func (s *MeetingService) updateRecordingStatus(recording *models.MeetingRecording, status string) {
	recording.Status = status
	if err := s.db.Save(recording).Error; err != nil {
		log.Error().
			Err(err).
			Str("recording_id", recording.ID).
			Str("status", status).
			Msg("Failed to update recording status")
	}
}

// processTranscriptAndCreateNote processes the transcript with AI and creates a note
func (s *MeetingService) processTranscriptAndCreateNote(ctx context.Context, recording *models.MeetingRecording, transcript []recallai.TranscriptEntry) error {
	// Convert transcript to plain text for AI analysis
	transcriptText := s.aiService.TranscriptToPlainText(transcript)

	// Get user's existing notebooks
	var notebooks []models.Notebook
	if err := s.db.Where("user_id = ?", recording.UserID).Find(&notebooks).Error; err != nil {
		log.Error().
			Err(err).
			Uint("user_id", recording.UserID).
			Msg("Failed to retrieve existing notebooks")
		return err
	}

	existingNotebooks := make([]string, len(notebooks))
	for i, nb := range notebooks {
		existingNotebooks[i] = nb.Name
	}

	// Analyze with AI
	analysis, err := s.aiService.AnalyzeTranscript(ctx, transcriptText, existingNotebooks)
	if err != nil {
		log.Warn().
			Err(err).
			Str("recording_id", recording.ID).
			Msg("AI analysis failed, using fallback")
		// Use fallback analysis
		analysis = s.aiService.CreateFallbackAnalysis(recording.MeetingURL)
	}

	// Find or create notebook
	notebook, err := s.findOrCreateNotebook(recording.UserID, analysis.NotebookName)
	if err != nil {
		return fmt.Errorf("failed to find or create notebook: %w", err)
	}

	// Find or create chapter
	chapter, err := s.findOrCreateChapter(notebook.ID, analysis.ChapterName)
	if err != nil {
		return fmt.Errorf("failed to find or create chapter: %w", err)
	}

	// Format transcript as markdown
	markdownContent := s.aiService.FormatTranscriptAsMarkdown(transcript, analysis)

	// Store raw transcript as JSON
	transcriptJSON, err := json.Marshal(transcript)
	if err != nil {
		log.Error().
			Err(err).
			Str("recording_id", recording.ID).
			Msg("Failed to marshal transcript to JSON")
		transcriptJSON = []byte("[]") // Empty array as fallback
	}

	// Create note
	note := &models.Notes{
		Name:               analysis.NoteName,
		Content:            markdownContent,
		ChapterID:          chapter.ID,
		MeetingRecordingID: &recording.ID,
		AISummary:          analysis.Summary,
		TranscriptRaw:      string(transcriptJSON),
	}

	if err := s.db.Create(note).Error; err != nil {
		log.Error().
			Err(err).
			Str("recording_id", recording.ID).
			Msg("Failed to create note")
		return fmt.Errorf("failed to create note: %w", err)
	}

	log.Info().
		Str("recording_id", recording.ID).
		Str("note_id", note.ID).
		Str("notebook_name", analysis.NotebookName).
		Str("chapter_name", analysis.ChapterName).
		Msg("Successfully created note from meeting transcript")

	return nil
}

// findOrCreateNotebook finds an existing notebook or creates a new one
func (s *MeetingService) findOrCreateNotebook(userID uint, notebookName string) (*models.Notebook, error) {
	var notebook models.Notebook

	// Try to find existing notebook
	err := s.db.Where("name = ? AND user_id = ?", notebookName, userID).First(&notebook).Error
	if err == nil {
		return &notebook, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Create new notebook
	notebook = models.Notebook{
		Name:   notebookName,
		UserID: userID,
	}

	if err := s.db.Create(&notebook).Error; err != nil {
		log.Error().
			Err(err).
			Uint("user_id", userID).
			Str("notebook_name", notebookName).
			Msg("Failed to create notebook")
		return nil, err
	}

	log.Info().
		Str("notebook_id", notebook.ID).
		Str("notebook_name", notebookName).
		Uint("user_id", userID).
		Msg("Created new notebook for meeting")

	return &notebook, nil
}

// findOrCreateChapter finds an existing chapter or creates a new one
func (s *MeetingService) findOrCreateChapter(notebookID, chapterName string) (*models.Chapter, error) {
	var chapter models.Chapter

	// Try to find existing chapter
	err := s.db.Where("name = ? AND notebook_id = ?", chapterName, notebookID).First(&chapter).Error
	if err == nil {
		return &chapter, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Create new chapter
	chapter = models.Chapter{
		Name:       chapterName,
		NotebookID: notebookID,
	}

	if err := s.db.Create(&chapter).Error; err != nil {
		log.Error().
			Err(err).
			Str("notebook_id", notebookID).
			Str("chapter_name", chapterName).
			Msg("Failed to create chapter")
		return nil, err
	}

	log.Info().
		Str("chapter_id", chapter.ID).
		Str("chapter_name", chapterName).
		Str("notebook_id", notebookID).
		Msg("Created new chapter for meeting")

	return &chapter, nil
}
