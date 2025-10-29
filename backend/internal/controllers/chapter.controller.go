package controllers

import (
	"backend/db"
	"backend/internal/middleware"
	"backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func CreateChapter(c *gin.Context) {
	// Get authenticated user ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var chapter models.Chapter

	if err := c.ShouldBindJSON(&chapter); err != nil {
		log.Print("Missing data to create a chapter: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify that the notebook belongs to the authenticated user
	var notebook models.Notebook
	if err := db.DB.Where("id = ? AND user_id = ?", chapter.NotebookID, userID).First(&notebook).Error; err != nil {
		log.Print("Notebook not found or unauthorized for user: ", userID, " Error: ", err)
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized: Notebook does not belong to user"})
		return
	}

	if err := db.DB.Create(&chapter).Error; err != nil {
		log.Print("Error creating chapter in db: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, chapter)
}

func GetChaptersByNotebook(c *gin.Context) {
	// Get authenticated user ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	notebookID := c.Param("notebookId")

	// Verify that the notebook belongs to the authenticated user
	var notebook models.Notebook
	if err := db.DB.Where("id = ? AND user_id = ?", notebookID, userID).First(&notebook).Error; err != nil {
		log.Print("Notebook not found or unauthorized for user: ", userID, " Error: ", err)
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized: Notebook does not belong to user"})
		return
	}

	// Get all chapters for this notebook
	var chapters []models.Chapter
	if err := db.DB.Where("notebook_id = ?", notebookID).Find(&chapters).Error; err != nil {
		log.Print("Error fetching chapters for notebook: ", notebookID, " Error: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, chapters)
}

func GetChapterById(c *gin.Context) {
	// Get authenticated user ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id := c.Param("id")
	var chapter models.Chapter

	// Get chapter with notebook preloaded to verify ownership
	if err := db.DB.Preload("Notebook").Where("id = ?", id).First(&chapter).Error; err != nil {
		log.Print("Chapter not found with id: ", id, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Chapter not found"})
		return
	}

	// Verify the notebook belongs to the authenticated user
	if chapter.Notebook.UserID != userID {
		log.Print("Unauthorized access attempt to chapter: ", id, " by user: ", userID)
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	c.JSON(http.StatusOK, chapter)
}

func DeleteChapter(c *gin.Context) {
	// Get authenticated user ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id := c.Param("id")
	var chapter models.Chapter

	// Get chapter with notebook preloaded to verify ownership
	if err := db.DB.Preload("Notebook").Where("id = ?", id).First(&chapter).Error; err != nil {
		log.Print("Chapter not found with id: ", id, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Chapter not found"})
		return
	}

	// Verify the notebook belongs to the authenticated user
	if chapter.Notebook.UserID != userID {
		log.Print("Unauthorized delete attempt on chapter: ", id, " by user: ", userID)
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	// Delete the chapter
	if err := db.DB.Delete(&chapter).Error; err != nil {
		log.Print("Error deleting chapter with id: ", id, " Error: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Chapter deleted successfully"})
}

func UpdateChapter(c *gin.Context) {
	// Get authenticated user ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id := c.Param("id")
	var chapter models.Chapter

	// Get chapter with notebook preloaded to verify ownership
	if err := db.DB.Preload("Notebook").Where("id = ?", id).First(&chapter).Error; err != nil {
		log.Print("Chapter not found with id: ", id, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Chapter not found"})
		return
	}

	// Verify the notebook belongs to the authenticated user
	if chapter.Notebook.UserID != userID {
		log.Print("Unauthorized update attempt on chapter: ", id, " by user: ", userID)
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	// Bind the update data from request body
	var updateData models.Chapter
	if err := c.ShouldBindJSON(&updateData); err != nil {
		log.Print("Invalid update data for chapter: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Prevent changing notebook_id through update (security)
	updateData.NotebookID = chapter.NotebookID

	// Update the chapter
	if err := db.DB.Model(&chapter).Updates(updateData).Error; err != nil {
		log.Print("Error updating chapter with id: ", id, " Error: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, chapter)
}

func MoveChapter(c *gin.Context) {
	// Get authenticated user ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id := c.Param("id")
	var chapter models.Chapter

	// Get chapter with notebook preloaded to verify ownership
	if err := db.DB.Preload("Notebook").Where("id = ?", id).First(&chapter).Error; err != nil {
		log.Print("Chapter not found with id: ", id, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Chapter not found"})
		return
	}

	// Verify the notebook belongs to the authenticated user
	if chapter.Notebook.UserID != userID {
		log.Print("Unauthorized move attempt on chapter: ", id, " by user: ", userID)
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	// Bind the move data from request body
	var moveData struct {
		NotebookID string `json:"notebook_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&moveData); err != nil {
		log.Print("Invalid move data for chapter: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify that the target notebook exists and belongs to the user
	var targetNotebook models.Notebook
	if err := db.DB.Where("id = ? AND user_id = ?", moveData.NotebookID, userID).First(&targetNotebook).Error; err != nil {
		log.Print("Target notebook not found or unauthorized for user: ", userID, " Error: ", err)
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized: Target notebook does not belong to user"})
		return
	}

	// Update the chapter's notebook_id directly using Select to force update
	result := db.DB.Model(&chapter).Select("NotebookID").Updates(models.Chapter{NotebookID: moveData.NotebookID})
	if result.Error != nil {
		log.Print("Error moving chapter with id: ", id, " Error: ", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	log.Printf("Chapter moved successfully. Chapter ID: %s, New Notebook ID: %s", chapter.ID, moveData.NotebookID)

	// Reload the chapter to get the updated data
	if err := db.DB.Preload("Notebook").Where("id = ?", id).First(&chapter).Error; err != nil {
		log.Print("Error reloading chapter after move: ", err)
	}

	c.JSON(http.StatusOK, chapter)
}
