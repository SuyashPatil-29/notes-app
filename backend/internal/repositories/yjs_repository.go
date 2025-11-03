package repositories

import (
	"backend/internal/models"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type YjsRepository struct {
	db *gorm.DB
}

func NewYjsRepository(db *gorm.DB) *YjsRepository {
	return &YjsRepository{db: db}
}

// GetYjsState retrieves the current Yjs state for a note
func (r *YjsRepository) GetYjsState(noteID string) (*models.YjsDocument, error) {
	var yjsDoc models.YjsDocument

	err := r.db.Where("note_id = ?", noteID).First(&yjsDoc).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Not found is not an error, return nil
		}
		return nil, err
	}

	return &yjsDoc, nil
}

// CreateYjsDocument creates a new Yjs document with initial state
// Uses FirstOrCreate to handle race conditions where multiple requests try to create simultaneously
func (r *YjsRepository) CreateYjsDocument(noteID string, initialState []byte) (*models.YjsDocument, error) {
	yjsDoc := &models.YjsDocument{
		NoteID: noteID,
	}

	// FirstOrCreate atomically checks if record exists and creates if not
	// This prevents race conditions from concurrent initialization requests
	// Attrs sets values only when creating (not when finding existing)
	err := r.db.Where(models.YjsDocument{NoteID: noteID}).
		Attrs(models.YjsDocument{
			YjsState: initialState,
			Version:  0,
		}).
		FirstOrCreate(yjsDoc).Error

	if err != nil {
		return nil, err
	}

	return yjsDoc, nil
}

// ApplyYjsUpdate applies a Yjs update to the document with optimistic locking
// This method uses transactions and row-level locking to prevent race conditions
func (r *YjsRepository) ApplyYjsUpdate(noteID string, update []byte, newState []byte) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var yjsDoc models.YjsDocument

		// Lock the row for update (SELECT ... FOR UPDATE)
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("note_id = ?", noteID).
			First(&yjsDoc).Error

		if err != nil {
			return err
		}

		// Get current version
		currentVersion := yjsDoc.Version

		// Update the document with version check (optimistic locking)
		result := tx.Model(&yjsDoc).
			Where("version = ?", currentVersion).
			Updates(map[string]interface{}{
				"yjs_state":  newState,
				"version":    currentVersion + 1,
				"updated_at": time.Now(),
			})

		if result.Error != nil {
			return result.Error
		}

		// Check if the update was successful (version conflict check)
		if result.RowsAffected == 0 {
			return errors.New("version conflict: document was modified by another process")
		}

		// Store the update in history
		yjsUpdate := models.YjsUpdate{
			NoteID:     noteID,
			UpdateData: update,
			Clock:      currentVersion + 1,
		}

		err = tx.Create(&yjsUpdate).Error
		if err != nil {
			return err
		}

		return nil
	})
}

// UpdateYjsState updates the entire Yjs state (used for batch updates)
func (r *YjsRepository) UpdateYjsState(noteID string, newState []byte) error {
	return r.db.Model(&models.YjsDocument{}).
		Where("note_id = ?", noteID).
		Updates(map[string]interface{}{
			"yjs_state":  newState,
			"updated_at": time.Now(),
		}).Error
}

// GetYjsUpdates retrieves Yjs updates for a note since a specific time
// Useful for syncing clients or debugging
func (r *YjsRepository) GetYjsUpdates(noteID string, since time.Time) ([]models.YjsUpdate, error) {
	var updates []models.YjsUpdate

	err := r.db.Where("note_id = ? AND created_at > ?", noteID, since).
		Order("clock ASC").
		Find(&updates).Error

	if err != nil {
		return nil, err
	}

	return updates, nil
}

// DeleteYjsDocument deletes the Yjs document and all its updates
func (r *YjsRepository) DeleteYjsDocument(noteID string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete all updates first
		if err := tx.Where("note_id = ?", noteID).Delete(&models.YjsUpdate{}).Error; err != nil {
			return err
		}

		// Delete the document
		if err := tx.Where("note_id = ?", noteID).Delete(&models.YjsDocument{}).Error; err != nil {
			return err
		}

		return nil
	})
}

// GetDocumentVersion gets the current version number of a Yjs document
func (r *YjsRepository) GetDocumentVersion(noteID string) (int, error) {
	var yjsDoc models.YjsDocument

	err := r.db.Select("version").
		Where("note_id = ?", noteID).
		First(&yjsDoc).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}

	return yjsDoc.Version, nil
}
