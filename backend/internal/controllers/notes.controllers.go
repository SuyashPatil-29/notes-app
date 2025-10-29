package controllers

import (
	"backend/db"
	"backend/internal/middleware"
	"backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func GetNoteById(c *gin.Context) {
	// Get authenticated user ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var note models.Notes
	id := c.Param("id")

	// Get note with chapter and notebook preloaded to verify ownership
	if err := db.DB.Preload("Chapter.Notebook").Where("id = ?", id).First(&note).Error; err != nil {
		log.Print("Note not found with id: ", id, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	// Verify the notebook belongs to the authenticated user
	if note.Chapter.Notebook.UserID != userID {
		log.Print("Unauthorized access attempt to note: ", id, " by user: ", userID)
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	c.JSON(http.StatusOK, note)
}

func GetNotesByChapter(c *gin.Context) {
	// Get authenticated user ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	chapterID := c.Param("chapterId")

	// Verify that the chapter belongs to the authenticated user
	var chapter models.Chapter
	if err := db.DB.Preload("Notebook").Where("id = ?", chapterID).First(&chapter).Error; err != nil {
		log.Print("Chapter not found with id: ", chapterID, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Chapter not found"})
		return
	}

	if chapter.Notebook.UserID != userID {
		log.Print("Unauthorized: Chapter does not belong to user: ", userID)
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized: Chapter does not belong to user"})
		return
	}

	// Get all notes for this chapter
	var notes []models.Notes
	if err := db.DB.Where("chapter_id = ?", chapterID).Find(&notes).Error; err != nil {
		log.Print("Error fetching notes for chapter: ", chapterID, " Error: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, notes)
}

func CreateNote(c *gin.Context) {
	// Get authenticated user ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var note models.Notes

	if err := c.ShouldBindJSON(&note); err != nil {
		log.Print("Missing data to create a note", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify that the chapter belongs to the authenticated user
	var chapter models.Chapter
	if err := db.DB.Preload("Notebook").Where("id = ?", note.ChapterID).First(&chapter).Error; err != nil {
		log.Print("Chapter not found with id: ", note.ChapterID, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Chapter not found"})
		return
	}

	if chapter.Notebook.UserID != userID {
		log.Print("Unauthorized: Chapter does not belong to user: ", userID)
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized: Chapter does not belong to user"})
		return
	}

	if err := db.DB.Create(&note).Error; err != nil {
		log.Print("Error creating note in db", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, note)
}

func DeleteNote(c *gin.Context) {
	// Get authenticated user ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var note models.Notes
	id := c.Param("id")

	// Get note with chapter and notebook preloaded to verify ownership
	if err := db.DB.Preload("Chapter.Notebook").Where("id = ?", id).First(&note).Error; err != nil {
		log.Print("Note not found with id: ", id, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	// Verify the notebook belongs to the authenticated user
	if note.Chapter.Notebook.UserID != userID {
		log.Print("Unauthorized delete attempt on note: ", id, " by user: ", userID)
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	if err := db.DB.Delete(&note, id).Error; err != nil {
		log.Print("Error deleting Note with id: ", id, " Error: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note deleted successfully"})
}

func UpdateNote(c *gin.Context) {
	// Get authenticated user ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id := c.Param("id")
	var note models.Notes

	// Get note with chapter and notebook preloaded to verify ownership
	if err := db.DB.Preload("Chapter.Notebook").Where("id = ?", id).First(&note).Error; err != nil {
		log.Print("Note not found with id: ", id, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	// Verify the notebook belongs to the authenticated user
	if note.Chapter.Notebook.UserID != userID {
		log.Print("Unauthorized update attempt on note: ", id, " by user: ", userID)
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	// Bind the update data from request body
	var updateData models.Notes
	if err := c.ShouldBindJSON(&updateData); err != nil {
		log.Print("Invalid update data for note: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Prevent changing chapter_id through update (security)
	updateData.ChapterID = note.ChapterID

	// Update the note
	if err := db.DB.Model(&note).Updates(updateData).Error; err != nil {
		log.Print("Error updating note with id: ", id, " Error: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, note)
}

func MoveNote(c *gin.Context) {
	// Get authenticated user ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id := c.Param("id")
	var note models.Notes

	// Get note with chapter and notebook preloaded to verify ownership
	if err := db.DB.Preload("Chapter.Notebook").Where("id = ?", id).First(&note).Error; err != nil {
		log.Print("Note not found with id: ", id, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	// Verify the notebook belongs to the authenticated user
	if note.Chapter.Notebook.UserID != userID {
		log.Print("Unauthorized move attempt on note: ", id, " by user: ", userID)
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	// Bind the move data from request body
	var moveData struct {
		ChapterID string `json:"chapter_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&moveData); err != nil {
		log.Print("Invalid move data for note: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify that the target chapter exists and belongs to the user
	var targetChapter models.Chapter
	if err := db.DB.Preload("Notebook").Where("id = ?", moveData.ChapterID).First(&targetChapter).Error; err != nil {
		log.Print("Target chapter not found with id: ", moveData.ChapterID, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Target chapter not found"})
		return
	}

	if targetChapter.Notebook.UserID != userID {
		log.Print("Unauthorized: Target chapter does not belong to user: ", userID)
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized: Target chapter does not belong to user"})
		return
	}

	// Update the note's chapter_id directly using Select to force update
	result := db.DB.Model(&note).Select("ChapterID").Updates(models.Notes{ChapterID: moveData.ChapterID})
	if result.Error != nil {
		log.Print("Error moving note with id: ", id, " Error: ", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	log.Printf("Note moved successfully. Note ID: %s, New Chapter ID: %s", note.ID, moveData.ChapterID)

	// Reload the note to get the updated data
	if err := db.DB.Preload("Chapter.Notebook").Where("id = ?", id).First(&note).Error; err != nil {
		log.Print("Error reloading note after move: ", err)
	}

	c.JSON(http.StatusOK, note)
}
