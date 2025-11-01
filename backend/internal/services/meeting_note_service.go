package services

import (
	"backend/db"
	"backend/internal/models"
	"backend/pkg/recallai"
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type MeetingNoteService struct {
	aiService *AIService
}

func NewMeetingNoteService() *MeetingNoteService {
	return &MeetingNoteService{
		aiService: NewAIService(),
	}
}

// ProcessMeetingTranscript downloads transcript, analyzes it with AI, and creates a note
func (s *MeetingNoteService) ProcessMeetingTranscript(ctx context.Context, recording *models.MeetingRecording) error {
	if recording.TranscriptDownloadURL == "" {
		return fmt.Errorf("transcript download URL not available")
	}

	log.Info().
		Str("meeting_id", recording.ID).
		Str("transcript_url", recording.TranscriptDownloadURL).
		Msg("Processing meeting transcript")

	// Download transcript from Recall.ai
	recallClient := recallai.NewClient()
	transcript, err := recallClient.DownloadTranscript(recording.TranscriptDownloadURL)
	if err != nil {
		log.Error().
			Err(err).
			Str("meeting_id", recording.ID).
			Msg("Failed to download transcript")
		return fmt.Errorf("failed to download transcript: %w", err)
	}

	if len(transcript) == 0 {
		log.Warn().
			Str("meeting_id", recording.ID).
			Msg("Transcript is empty, skipping note creation")
		return fmt.Errorf("transcript is empty")
	}

	// Convert transcript to plain text for AI analysis
	plainTextTranscript := s.aiService.TranscriptToPlainText(transcript)

	// Get user's existing notebooks for context
	var notebooks []models.Notebook
	if err := db.DB.Where("clerk_user_id = ?", recording.ClerkUserID).
		Select("name").
		Find(&notebooks).Error; err != nil {
		log.Warn().
			Err(err).
			Str("clerk_user_id", recording.ClerkUserID).
			Msg("Failed to fetch user notebooks, continuing without context")
	}

	existingNotebookNames := make([]string, len(notebooks))
	for i, nb := range notebooks {
		existingNotebookNames[i] = nb.Name
	}

	// Analyze transcript with AI
	analysis, err := s.aiService.AnalyzeTranscript(ctx, plainTextTranscript, existingNotebookNames)
	if err != nil {
		log.Warn().
			Err(err).
			Str("meeting_id", recording.ID).
			Msg("AI analysis failed, using fallback")
		// Use fallback analysis
		analysis = s.aiService.CreateFallbackAnalysis(recording.MeetingURL)
	}

	log.Info().
		Str("meeting_id", recording.ID).
		Str("notebook", analysis.NotebookName).
		Str("chapter", analysis.ChapterName).
		Str("note", analysis.NoteName).
		Msg("AI analysis complete")

	// Create or find notebook and chapter, then create note
	noteID, err := s.createNoteFromAnalysis(recording.ClerkUserID, analysis, transcript)
	if err != nil {
		log.Error().
			Err(err).
			Str("meeting_id", recording.ID).
			Msg("Failed to create note from analysis")
		return fmt.Errorf("failed to create note: %w", err)
	}

	// Update meeting recording with generated note ID
	recording.GeneratedNoteID = &noteID
	if err := db.DB.Save(recording).Error; err != nil {
		log.Error().
			Err(err).
			Str("meeting_id", recording.ID).
			Str("note_id", noteID).
			Msg("Failed to update meeting recording with note ID")
		return fmt.Errorf("failed to update meeting recording: %w", err)
	}

	log.Info().
		Str("meeting_id", recording.ID).
		Str("note_id", noteID).
		Msg("Successfully created note from meeting transcript")

	return nil
}

// NoteCreationResult holds the IDs of created/found notebook, chapter, and note
type NoteCreationResult struct {
	NoteID     string
	ChapterID  string
	NotebookID string
}

// createNoteFromAnalysis creates notebook, chapter, and note from AI analysis
func (s *MeetingNoteService) createNoteFromAnalysis(
	clerkUserID string,
	analysis *TranscriptAnalysis,
	transcript []recallai.TranscriptEntry,
) (string, error) {
	var noteID string

	// Use transaction to ensure consistency
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Find or create notebook
		var notebook models.Notebook
		result := tx.Where("clerk_user_id = ? AND name = ?", clerkUserID, analysis.NotebookName).
			First(&notebook)

		if result.Error == gorm.ErrRecordNotFound {
			// Create new notebook
			notebook = models.Notebook{
				ClerkUserID: clerkUserID,
				Name:        analysis.NotebookName,
				IsPublic:    false,
			}
			if err := tx.Create(&notebook).Error; err != nil {
				return fmt.Errorf("failed to create notebook: %w", err)
			}
			log.Info().
				Str("notebook_id", notebook.ID).
				Str("notebook_name", notebook.Name).
				Msg("Created new notebook")
		} else if result.Error != nil {
			return fmt.Errorf("failed to query notebook: %w", result.Error)
		} else {
			log.Info().
				Str("notebook_id", notebook.ID).
				Str("notebook_name", notebook.Name).
				Msg("Using existing notebook")
		}

		// 2. Find or create chapter
		var chapter models.Chapter
		result = tx.Where("notebook_id = ? AND name = ?", notebook.ID, analysis.ChapterName).
			First(&chapter)

		if result.Error == gorm.ErrRecordNotFound {
			// Create new chapter
			chapter = models.Chapter{
				NotebookID: notebook.ID,
				Name:       analysis.ChapterName,
				IsPublic:   false,
			}
			if err := tx.Create(&chapter).Error; err != nil {
				return fmt.Errorf("failed to create chapter: %w", err)
			}
			log.Info().
				Str("chapter_id", chapter.ID).
				Str("chapter_name", chapter.Name).
				Msg("Created new chapter")
		} else if result.Error != nil {
			return fmt.Errorf("failed to query chapter: %w", result.Error)
		} else {
			log.Info().
				Str("chapter_id", chapter.ID).
				Str("chapter_name", chapter.Name).
				Msg("Using existing chapter")
		}

		// 3. Create note with formatted content
		markdownContent := s.aiService.FormatTranscriptAsMarkdown(transcript, analysis)

		note := models.Notes{
			ChapterID: chapter.ID,
			Name:      analysis.NoteName,
			Content:   markdownContent,
			IsPublic:  false,
		}

		if err := tx.Create(&note).Error; err != nil {
			return fmt.Errorf("failed to create note: %w", err)
		}

		log.Info().
			Str("note_id", note.ID).
			Str("note_name", note.Name).
			Int("content_length", len(markdownContent)).
			Msg("Created new note")

		noteID = note.ID
		return nil
	})

	if err != nil {
		return "", err
	}

	return noteID, nil
}
