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

	// Check authorization efficiently
	hasAccess, err := middleware.CheckNotebookAccess(c.Request.Context(), db.DB, chapter.NotebookID, clerkUserID)
	if err != nil {
		log.Print("Notebook not found: ", chapter.NotebookID, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Notebook not found"})
		return
	}
	if !hasAccess {
		log.Warn().Str("notebook_id", chapter.NotebookID).Str("user_id", clerkUserID).Msg("User not authorized to create chapter in notebook")
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized: You cannot create chapters in this notebook"})
		return
	}

	// Get organization_id from parent notebook
	var notebook models.Notebook
	if err := db.DB.Select("organization_id").Where("id = ?", chapter.NotebookID).First(&notebook).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create chapter"})
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

	// Check authorization efficiently
	hasAccess, err := middleware.CheckNotebookAccess(c.Request.Context(), db.DB, notebookID, clerkUserID)
	if err != nil {
		log.Print("Notebook not found: ", notebookID, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Notebook not found"})
		return
	}
	if !hasAccess {
		log.Warn().Str("notebook_id", notebookID).Str("user_id", clerkUserID).Msg("User not authorized to access notebook chapters")
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	// Get chapters without preloading notes (optimized)
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

	// Get chapter without preloading notebook
	if err := db.DB.Where("id = ?", id).First(&chapter).Error; err != nil {
		log.Print("Chapter not found with id: ", id, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Chapter not found"})
		return
	}

	// Check authorization efficiently
	hasAccess, err := middleware.CheckChapterAccess(c.Request.Context(), db.DB, id, clerkUserID)
	if err != nil || !hasAccess {
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

	// Check authorization efficiently
	hasAccess, err := middleware.CheckChapterAccess(c.Request.Context(), db.DB, id, clerkUserID)
	if err != nil {
		log.Print("Chapter not found with id: ", id, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Chapter not found"})
		return
	}
	if !hasAccess {
		log.Warn().Str("chapter_id", id).Str("user_id", clerkUserID).Msg("User not authorized to delete chapter")
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	// Delete the chapter
	if err := db.DB.Delete(&models.Chapter{}, "id = ?", id).Error; err != nil {
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

	// Check authorization efficiently
	hasAccess, err := middleware.CheckChapterAccess(c.Request.Context(), db.DB, id, clerkUserID)
	if err != nil {
		log.Print("Chapter not found with id: ", id, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Chapter not found"})
		return
	}
	if !hasAccess {
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

	// Get current chapter to preserve protected fields
	var chapter models.Chapter
	if err := db.DB.Where("id = ?", id).First(&chapter).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update chapter"})
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

	// Check authorization for source chapter efficiently
	hasAccess, err := middleware.CheckChapterAccess(c.Request.Context(), db.DB, id, clerkUserID)
	if err != nil {
		log.Print("Chapter not found with id: ", id, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Chapter not found"})
		return
	}
	if !hasAccess {
		log.Warn().Str("chapter_id", id).Str("user_id", clerkUserID).Msg("User not authorized to move chapter")
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	// Bind the move data from request body
	var moveData struct {
		NotebookID     string  `json:"notebook_id" binding:"required"`
		OrganizationID *string `json:"organization_id"`
	}
	if err := c.ShouldBindJSON(&moveData); err != nil {
		log.Print("Invalid move data for chapter: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log the received organization ID
	if moveData.OrganizationID != nil {
		log.Printf("DEBUG: Move chapter request with organization ID: %s", *moveData.OrganizationID)
	}

	// Check authorization for target notebook efficiently
	targetHasAccess, err := middleware.CheckNotebookAccess(c.Request.Context(), db.DB, moveData.NotebookID, clerkUserID)
	if err != nil {
		log.Print("Target notebook not found: ", moveData.NotebookID, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Target notebook not found"})
		return
	}
	if !targetHasAccess {
		log.Warn().Str("notebook_id", moveData.NotebookID).Str("user_id", clerkUserID).Msg("User not authorized to move chapter to target notebook")
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized: Cannot move to target notebook"})
		return
	}

	// Get target notebook's organization_id
	var targetNotebook models.Notebook
	if err := db.DB.Select("organization_id").Where("id = ?", moveData.NotebookID).First(&targetNotebook).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to move chapter"})
		return
	}

	// Log before update
	log.Printf("DEBUG: Before update - Chapter ID: %s, Target Notebook ID: %s", id, moveData.NotebookID)

	// Determine which organization ID to use
	var orgIDToUse *string
	if moveData.OrganizationID != nil && *moveData.OrganizationID != "" {
		// Use the provided organization ID from request
		orgIDToUse = moveData.OrganizationID
		log.Printf("DEBUG: Using organization ID from request: %s", *orgIDToUse)
	} else {
		// Fall back to target notebook's organization ID
		orgIDToUse = targetNotebook.OrganizationID
		if orgIDToUse != nil {
			log.Printf("DEBUG: Using organization ID from target notebook: %s", *orgIDToUse)
		}
	}

	// Update the chapter's notebook_id and organization_id using Updates with Select
	updateData := map[string]interface{}{
		"notebook_id":     moveData.NotebookID,
		"organization_id": orgIDToUse,
	}

	log.Printf("DEBUG: About to update with data: %+v", updateData)

	// Use Model().Select().Updates() for explicit column updates
	result := db.DB.Model(&models.Chapter{}).Where("id = ?", id).Select("notebook_id", "organization_id").Updates(updateData)
	if result.Error != nil {
		log.Printf("ERROR: Failed to update chapter - Error: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	log.Printf("DEBUG: GORM Update Result - Rows Affected: %d, Error: %v", result.RowsAffected, result.Error)

	if result.RowsAffected == 0 {
		log.Printf("WARNING: Update returned 0 rows affected for chapter ID: %s", id)
	}

	// Reload the chapter from scratch to verify the update
	var updatedChapter models.Chapter
	if err := db.DB.Preload("Notebook").Where("id = ?", id).First(&updatedChapter).Error; err != nil {
		log.Printf("ERROR: Failed to reload chapter after move: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload chapter"})
		return
	}

	log.Printf("DEBUG: After reload - Chapter ID: %s, Notebook ID: %s (expected: %s)",
		updatedChapter.ID, updatedChapter.NotebookID, moveData.NotebookID)
	log.Printf("Chapter moved successfully. Chapter ID: %s, New Notebook ID: %s", updatedChapter.ID, moveData.NotebookID)

	c.JSON(http.StatusOK, updatedChapter)
}
