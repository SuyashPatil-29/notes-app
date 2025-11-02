package controllers

import (
	"backend/db"
	"backend/internal/middleware"
	"backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// Helper function to check if user has access to a notebook (personal or org)
func userCanAccessNotebook(c *gin.Context, notebook *models.Notebook, clerkUserID string) bool {
	if notebook.OrganizationID != nil && *notebook.OrganizationID != "" {
		// Organization notebook - verify membership
		_, isMember, err := middleware.GetOrgMemberRole(c.Request.Context(), *notebook.OrganizationID, clerkUserID)
		return err == nil && isMember
	}
	// Personal notebook - verify ownership
	return notebook.ClerkUserID == clerkUserID
}

func CreateChapter(c *gin.Context) {
	// Get authenticated user ID
	clerkUserID, exists := middleware.GetClerkUserID(c)
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

	// Verify that the notebook exists and user has access
	var notebook models.Notebook
	if err := db.DB.Where("id = ?", chapter.NotebookID).First(&notebook).Error; err != nil {
		log.Print("Notebook not found: ", chapter.NotebookID, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Notebook not found"})
		return
	}

	if !userCanAccessNotebook(c, &notebook, clerkUserID) {
		log.Warn().Str("notebook_id", chapter.NotebookID).Str("user_id", clerkUserID).Msg("User not authorized to create chapter in notebook")
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized: You cannot create chapters in this notebook"})
		return
	}

	// Inherit organization_id from parent notebook
	chapter.OrganizationID = notebook.OrganizationID

	if err := db.DB.Create(&chapter).Error; err != nil {
		log.Print("Error creating chapter in db: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, chapter)
}

func GetChaptersByNotebook(c *gin.Context) {
	// Get authenticated user ID
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	notebookID := c.Param("notebookId")

	// Verify that the notebook exists and user has access
	var notebook models.Notebook
	if err := db.DB.Where("id = ?", notebookID).First(&notebook).Error; err != nil {
		log.Print("Notebook not found: ", notebookID, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Notebook not found"})
		return
	}

	if !userCanAccessNotebook(c, &notebook, clerkUserID) {
		log.Warn().Str("notebook_id", notebookID).Str("user_id", clerkUserID).Msg("User not authorized to access notebook chapters")
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
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
	clerkUserID, exists := middleware.GetClerkUserID(c)
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

	// Verify user has access to the parent notebook
	if !userCanAccessNotebook(c, &chapter.Notebook, clerkUserID) {
		log.Warn().Str("chapter_id", id).Str("user_id", clerkUserID).Msg("User not authorized to access chapter")
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	c.JSON(http.StatusOK, chapter)
}

func DeleteChapter(c *gin.Context) {
	// Get authenticated user ID
	clerkUserID, exists := middleware.GetClerkUserID(c)
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

	// Verify user has access to the parent notebook
	if !userCanAccessNotebook(c, &chapter.Notebook, clerkUserID) {
		log.Warn().Str("chapter_id", id).Str("user_id", clerkUserID).Msg("User not authorized to delete chapter")
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
	clerkUserID, exists := middleware.GetClerkUserID(c)
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

	// Verify user has access to the parent notebook
	if !userCanAccessNotebook(c, &chapter.Notebook, clerkUserID) {
		log.Warn().Str("chapter_id", id).Str("user_id", clerkUserID).Msg("User not authorized to update chapter")
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

	// Prevent changing notebook_id and organization_id through update (security)
	updateData.NotebookID = chapter.NotebookID
	updateData.OrganizationID = chapter.OrganizationID

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
	clerkUserID, exists := middleware.GetClerkUserID(c)
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

	// Verify user has access to the source notebook
	if !userCanAccessNotebook(c, &chapter.Notebook, clerkUserID) {
		log.Warn().Str("chapter_id", id).Str("user_id", clerkUserID).Msg("User not authorized to move chapter")
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

	// Verify that the target notebook exists and user has access
	var targetNotebook models.Notebook
	if err := db.DB.Where("id = ?", moveData.NotebookID).First(&targetNotebook).Error; err != nil {
		log.Print("Target notebook not found: ", moveData.NotebookID, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Target notebook not found"})
		return
	}

	if !userCanAccessNotebook(c, &targetNotebook, clerkUserID) {
		log.Warn().Str("notebook_id", moveData.NotebookID).Str("user_id", clerkUserID).Msg("User not authorized to move chapter to target notebook")
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized: Cannot move to target notebook"})
		return
	}

	// Update the chapter's notebook_id and inherit target notebook's organization_id
	result := db.DB.Model(&chapter).Updates(map[string]interface{}{
		"notebook_id":     moveData.NotebookID,
		"organization_id": targetNotebook.OrganizationID,
	})
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
