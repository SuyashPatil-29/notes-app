package services

import (
	"backend/internal/models"
	"backend/internal/repositories"
	"encoding/json"
	"errors"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type YjsService struct {
	yjsRepo *repositories.YjsRepository
	db      *gorm.DB
}

func NewYjsService(db *gorm.DB) *YjsService {
	return &YjsService{
		yjsRepo: repositories.NewYjsRepository(db),
		db:      db,
	}
}

// GetYjsStateResponse represents the response for Yjs state requests
type GetYjsStateResponse struct {
	Exists       bool   `json:"exists"`
	YjsState     []byte `json:"yjsState,omitempty"`
	Version      int    `json:"version"`
	NoteContent  string `json:"noteContent,omitempty"` // JSON content for conversion
	RequiresInit bool   `json:"requiresInit"`
}

// GetOrCreateYjsState gets the Yjs state or indicates that initialization is needed
func (s *YjsService) GetOrCreateYjsState(noteID string) (*GetYjsStateResponse, error) {
	// Try to get existing Yjs document
	yjsDoc, err := s.yjsRepo.GetYjsState(noteID)
	if err != nil {
		log.Error().Err(err).Str("noteID", noteID).Msg("Failed to get Yjs state")
		return nil, err
	}

	// If Yjs document exists, return it
	if yjsDoc != nil {
		return &GetYjsStateResponse{
			Exists:       true,
			YjsState:     yjsDoc.YjsState,
			Version:      yjsDoc.Version,
			RequiresInit: false,
		}, nil
	}

	// Yjs document doesn't exist, get the note content for frontend to convert
	var note models.Notes
	err = s.db.Where("id = ?", noteID).First(&note).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("note not found")
		}
		log.Error().Err(err).Str("noteID", noteID).Msg("Failed to get note")
		return nil, err
	}

	return &GetYjsStateResponse{
		Exists:       false,
		NoteContent:  note.Content,
		RequiresInit: true,
		Version:      0,
	}, nil
}

// InitializeYjsDocument creates a new Yjs document from initial state (sent by frontend)
func (s *YjsService) InitializeYjsDocument(noteID string, initialState []byte) error {
	// Check if already exists
	existing, err := s.yjsRepo.GetYjsState(noteID)
	if err != nil {
		return err
	}

	if existing != nil {
		log.Warn().Str("noteID", noteID).Msg("Yjs document already exists, skipping initialization")
		return nil // Already exists, not an error
	}

	// Create new Yjs document
	_, err = s.yjsRepo.CreateYjsDocument(noteID, initialState)
	if err != nil {
		log.Error().Err(err).Str("noteID", noteID).Msg("Failed to create Yjs document")
		return err
	}

	log.Info().Str("noteID", noteID).Msg("Yjs document initialized successfully")
	return nil
}

// ApplyUpdate applies a Yjs update with the new complete state
func (s *YjsService) ApplyUpdate(noteID string, update []byte, newState []byte) error {
	// Check if Yjs document exists
	yjsDoc, err := s.yjsRepo.GetYjsState(noteID)
	if err != nil {
		return err
	}

	if yjsDoc == nil {
		return errors.New("yjs document not found, must initialize first")
	}

	// Apply the update with transaction and locking
	err = s.yjsRepo.ApplyYjsUpdate(noteID, update, newState)
	if err != nil {
		log.Error().Err(err).Str("noteID", noteID).Msg("Failed to apply Yjs update")
		return err
	}

	log.Debug().Str("noteID", noteID).Msg("Yjs update applied successfully")
	return nil
}

// SyncYjsToNoteContent converts Yjs state to JSON and updates the note's content field
// This ensures the note's content field stays in sync for non-collaborative views
func (s *YjsService) SyncYjsToNoteContent(noteID string, jsonContent string) error {
	// Validate JSON
	var contentMap map[string]interface{}
	err := json.Unmarshal([]byte(jsonContent), &contentMap)
	if err != nil {
		return errors.New("invalid JSON content")
	}

	// Update the note's content field
	var note models.Notes
	err = s.db.Where("id = ?", noteID).First(&note).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("note not found")
		}
		return err
	}

	err = s.db.Model(&note).Update("content", jsonContent).Error
	if err != nil {
		log.Error().Err(err).Str("noteID", noteID).Msg("Failed to sync content to note")
		return err
	}

	log.Debug().Str("noteID", noteID).Msg("Note content synced from Yjs")
	return nil
}

// GetYjsUpdates retrieves update history for debugging or syncing
func (s *YjsService) GetYjsUpdates(noteID string, since time.Time) ([]models.YjsUpdate, error) {
	return s.yjsRepo.GetYjsUpdates(noteID, since)
}

// DeleteYjsDocument deletes the Yjs document and all its history
func (s *YjsService) DeleteYjsDocument(noteID string) error {
	return s.yjsRepo.DeleteYjsDocument(noteID)
}

// GetDocumentVersion gets the current version number
func (s *YjsService) GetDocumentVersion(noteID string) (int, error) {
	return s.yjsRepo.GetDocumentVersion(noteID)
}

// BatchUpdateYjsState performs a batch update (for periodic syncing from memory)
func (s *YjsService) BatchUpdateYjsState(noteID string, newState []byte) error {
	yjsDoc, err := s.yjsRepo.GetYjsState(noteID)
	if err != nil {
		return err
	}

	if yjsDoc == nil {
		return errors.New("yjs document not found")
	}

	return s.yjsRepo.UpdateYjsState(noteID, newState)
}

