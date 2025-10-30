package controllers

import (
	"backend/db"
	"backend/internal/middleware"
	"backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// PublishRequest represents the request body for publishing a notebook
type PublishRequest struct {
	NoteIds []string `json:"noteIds" binding:"required"`
}

// PublishNotebook publishes a notebook by marking specific notes as public
// This automatically makes parent chapters and the notebook public if they contain published notes
func PublishNotebook(c *gin.Context) {
	// Get authenticated user ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	notebookID := c.Param("id")
	var req PublishRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Print("Invalid publish request: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Verify notebook ownership
	var notebook models.Notebook
	if err := db.DB.Where("id = ? AND user_id = ?", notebookID, userID).First(&notebook).Error; err != nil {
		log.Print("Notebook not found with id: ", notebookID, " for user: ", userID, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Notebook not found"})
		return
	}

	// Start transaction
	tx := db.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Reset all notes in this notebook to private
	if err := tx.Model(&models.Notes{}).
		Where("chapter_id IN (SELECT id FROM chapters WHERE notebook_id = ?)", notebookID).
		Update("is_public", false).Error; err != nil {
		tx.Rollback()
		log.Print("Error resetting notes to private: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notes"})
		return
	}

	// Reset all chapters in this notebook to private
	if err := tx.Model(&models.Chapter{}).Where("notebook_id = ?", notebookID).Update("is_public", false).Error; err != nil {
		tx.Rollback()
		log.Print("Error resetting chapters to private: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update chapters"})
		return
	}

	// Reset notebook to private
	if err := tx.Model(&notebook).Update("is_public", false).Error; err != nil {
		tx.Rollback()
		log.Print("Error resetting notebook to private: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notebook"})
		return
	}

	// If no notes selected, just unpublish everything
	if len(req.NoteIds) == 0 {
		tx.Commit()
		c.JSON(http.StatusOK, gin.H{"message": "Notebook unpublished successfully"})
		return
	}

	// Mark selected notes as public
	if err := tx.Model(&models.Notes{}).Where("id IN ? AND chapter_id IN (SELECT id FROM chapters WHERE notebook_id = ?)", req.NoteIds, notebookID).Update("is_public", true).Error; err != nil {
		tx.Rollback()
		log.Print("Error marking notes as public: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish notes"})
		return
	}

	// Mark chapters that have at least one public note as public
	if err := tx.Exec(`
		UPDATE chapters
		SET is_public = true
		WHERE notebook_id = ? AND id IN (
			SELECT DISTINCT chapter_id FROM notes WHERE is_public = true
		)
	`, notebookID).Error; err != nil {
		tx.Rollback()
		log.Print("Error marking chapters as public: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish chapters"})
		return
	}

	// Mark notebook as public since it has published content
	if err := tx.Model(&notebook).Update("is_public", true).Error; err != nil {
		tx.Rollback()
		log.Print("Error marking notebook as public: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish notebook"})
		return
	}

	// Commit transaction
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "Notebook published successfully"})
}

// UpdatePublishedNotes updates which notes are published for an already published notebook
func UpdatePublishedNotes(c *gin.Context) {
	// Get authenticated user ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	notebookID := c.Param("id")
	var req PublishRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Print("Invalid update published notes request: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Verify notebook ownership and that it's published
	var notebook models.Notebook
	if err := db.DB.Where("id = ? AND user_id = ? AND is_public = ?", notebookID, userID, true).First(&notebook).Error; err != nil {
		log.Print("Published notebook not found with id: ", notebookID, " for user: ", userID, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Notebook not found or not published"})
		return
	}

	// Start transaction
	tx := db.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Reset all notes in this notebook to private
	if err := tx.Model(&models.Notes{}).
		Where("chapter_id IN (SELECT id FROM chapters WHERE notebook_id = ?)", notebookID).
		Update("is_public", false).Error; err != nil {
		tx.Rollback()
		log.Print("Error resetting notes to private: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notes"})
		return
	}

	// Reset all chapters in this notebook to private
	if err := tx.Model(&models.Chapter{}).Where("notebook_id = ?", notebookID).Update("is_public", false).Error; err != nil {
		tx.Rollback()
		log.Print("Error resetting chapters to private: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update chapters"})
		return
	}

	// If no notes selected, unpublish the notebook
	if len(req.NoteIds) == 0 {
		if err := tx.Model(&notebook).Update("is_public", false).Error; err != nil {
			tx.Rollback()
			log.Print("Error unpublishing notebook: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unpublish notebook"})
			return
		}
		tx.Commit()
		c.JSON(http.StatusOK, gin.H{"message": "Notebook unpublished successfully"})
		return
	}

	// Mark selected notes as public
	if err := tx.Model(&models.Notes{}).Where("id IN ? AND chapter_id IN (SELECT id FROM chapters WHERE notebook_id = ?)", req.NoteIds, notebookID).Update("is_public", true).Error; err != nil {
		tx.Rollback()
		log.Print("Error marking notes as public: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish notes"})
		return
	}

	// Mark chapters that have at least one public note as public
	if err := tx.Exec(`
		UPDATE chapters
		SET is_public = true
		WHERE notebook_id = ? AND id IN (
			SELECT DISTINCT chapter_id FROM notes WHERE is_public = true
		)
	`, notebookID).Error; err != nil {
		tx.Rollback()
		log.Print("Error marking chapters as public: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish chapters"})
		return
	}

	// Commit transaction
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "Published notes updated successfully"})
}

// UnpublishNotebook unpublishes a notebook by marking it and all its content as private
func UnpublishNotebook(c *gin.Context) {
	// Get authenticated user ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	notebookID := c.Param("id")

	// Verify notebook ownership
	var notebook models.Notebook
	if err := db.DB.Where("id = ? AND user_id = ?", notebookID, userID).First(&notebook).Error; err != nil {
		log.Print("Notebook not found with id: ", notebookID, " for user: ", userID, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Notebook not found"})
		return
	}

	// Start transaction
	tx := db.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Mark notebook as private
	if err := tx.Model(&notebook).Update("is_public", false).Error; err != nil {
		tx.Rollback()
		log.Print("Error unpublishing notebook: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unpublish notebook"})
		return
	}

	// Mark all chapters as private
	if err := tx.Model(&models.Chapter{}).Where("notebook_id = ?", notebookID).Update("is_public", false).Error; err != nil {
		tx.Rollback()
		log.Print("Error unpublishing chapters: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unpublish chapters"})
		return
	}

	// Mark all notes as private
	if err := tx.Model(&models.Notes{}).
		Where("chapter_id IN (SELECT id FROM chapters WHERE notebook_id = ?)", notebookID).
		Update("is_public", false).Error; err != nil {
		tx.Rollback()
		log.Print("Error unpublishing notes: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unpublish notes"})
		return
	}

	// Commit transaction
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "Notebook unpublished successfully"})
}

// PublishNote toggles the publish status of an individual note
func PublishNote(c *gin.Context) {
	// Get authenticated user ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	noteID := c.Param("id")

	// Get note and verify ownership through notebook
	var note models.Notes
	if err := db.DB.Where("id = ?", noteID).
		Preload("Chapter.Notebook", "user_id = ?", userID).
		First(&note).Error; err != nil {
		log.Print("Note not found with id: ", noteID, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	// Verify ownership
	if note.Chapter.Notebook.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Toggle the publish status
	newStatus := !note.IsPublic

	// Start transaction
	tx := db.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update note status
	if err := tx.Model(&note).Update("is_public", newStatus).Error; err != nil {
		tx.Rollback()
		log.Print("Error updating note publish status: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update note"})
		return
	}

	// Update chapter status based on whether it has any public notes
	var publicNotesCount int64
	tx.Model(&models.Notes{}).Where("chapter_id = ? AND is_public = ?", note.ChapterID, true).Count(&publicNotesCount)

	chapterStatus := publicNotesCount > 0
	if err := tx.Model(&models.Chapter{}).Where("id = ?", note.ChapterID).Update("is_public", chapterStatus).Error; err != nil {
		tx.Rollback()
		log.Print("Error updating chapter publish status: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update chapter"})
		return
	}

	// Update notebook status based on whether it has any public chapters
	var publicChaptersCount int64
	tx.Model(&models.Chapter{}).Where("notebook_id = ? AND is_public = ?", note.Chapter.Notebook.ID, true).Count(&publicChaptersCount)

	notebookStatus := publicChaptersCount > 0
	if err := tx.Model(&models.Notebook{}).Where("id = ?", note.Chapter.Notebook.ID).Update("is_public", notebookStatus).Error; err != nil {
		tx.Rollback()
		log.Print("Error updating notebook publish status: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notebook"})
		return
	}

	// Commit transaction
	tx.Commit()

	status := "unpublished"
	if newStatus {
		status = "published"
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note " + status + " successfully", "isPublic": newStatus})
}
