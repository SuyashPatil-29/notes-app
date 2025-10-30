package controllers

import (
	"backend/db"
	"backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// GetPublicNotebook returns a public notebook with only public chapters and notes
func GetPublicNotebook(c *gin.Context) {
	notebookID := c.Param("notebookId")

	var notebook models.Notebook

	// Get notebook only if it's public
	if err := db.DB.Where("id = ? AND is_public = ?", notebookID, true).First(&notebook).Error; err != nil {
		log.Print("Public notebook not found with id: ", notebookID, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Notebook not found or not public"})
		return
	}

	// Get only public chapters and their public notes
	var chapters []models.Chapter
	if err := db.DB.Where("notebook_id = ? AND is_public = ?", notebookID, true).
		Preload("Files", "is_public = ?", true).
		Find(&chapters).Error; err != nil {
		log.Print("Error fetching public chapters for notebook: ", notebookID, " Error: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch chapters"})
		return
	}

	// Filter out chapters that have no public notes
	var filteredChapters []models.Chapter
	for _, chapter := range chapters {
		if len(chapter.Files) > 0 {
			filteredChapters = append(filteredChapters, chapter)
		}
	}

	notebook.Chapters = filteredChapters

	c.JSON(http.StatusOK, notebook)
}

// GetPublicChapter returns a public chapter with only public notes
func GetPublicChapter(c *gin.Context) {
	notebookID := c.Param("notebookId")
	chapterID := c.Param("chapterId")

	var chapter models.Chapter

	// Get chapter only if it's public and belongs to a public notebook
	if err := db.DB.Where("id = ? AND notebook_id = ? AND is_public = ?", chapterID, notebookID, true).
		Preload("Notebook", "is_public = ?", true).
		First(&chapter).Error; err != nil {
		log.Print("Public chapter not found with id: ", chapterID, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Chapter not found or not public"})
		return
	}

	// Verify notebook is public
	if !chapter.Notebook.IsPublic {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notebook not found or not public"})
		return
	}

	// Get only public notes for this chapter
	var notes []models.Notes
	if err := db.DB.Where("chapter_id = ? AND is_public = ?", chapterID, true).Find(&notes).Error; err != nil {
		log.Print("Error fetching public notes for chapter: ", chapterID, " Error: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notes"})
		return
	}

	chapter.Files = notes

	c.JSON(http.StatusOK, chapter)
}

// GetPublicNote returns a single public note
func GetPublicNote(c *gin.Context) {
	notebookID := c.Param("notebookId")
	chapterID := c.Param("chapterId")
	noteID := c.Param("noteId")

	var note models.Notes

	// Get note only if it's public and belongs to a public chapter and notebook
	if err := db.DB.Where("id = ? AND chapter_id = ? AND is_public = ?", noteID, chapterID, true).
		Preload("Chapter", "is_public = ?", true).
		Preload("Chapter.Notebook", "is_public = ?", true).
		First(&note).Error; err != nil {
		log.Print("Public note not found with id: ", noteID, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found or not public"})
		return
	}

	// Verify chapter and notebook are public and match the URL parameters
	if !note.Chapter.IsPublic || !note.Chapter.Notebook.IsPublic || note.Chapter.Notebook.ID != notebookID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Content not found or not public"})
		return
	}

	c.JSON(http.StatusOK, note)
}

// GetPublicUserProfile returns a user's public profile with their public notebooks
func GetPublicUserProfile(c *gin.Context) {
	email := c.Param("email")

	var user models.User

	// Get user by email
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		log.Print("User not found with email: ", email, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get user's public notebooks with public chapters and notes
	var notebooks []models.Notebook
	if err := db.DB.Where("user_id = ? AND is_public = ?", user.ID, true).
		Preload("Chapters", "is_public = ?", true).
		Preload("Chapters.Files", "is_public = ?", true).
		Find(&notebooks).Error; err != nil {
		log.Print("Error fetching public notebooks for user: ", user.ID, " Error: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notebooks"})
		return
	}

	// Filter out chapters with no public notes
	for i := range notebooks {
		var filteredChapters []models.Chapter
		for _, chapter := range notebooks[i].Chapters {
			if len(chapter.Files) > 0 {
				filteredChapters = append(filteredChapters, chapter)
			}
		}
		notebooks[i].Chapters = filteredChapters
	}

	// Filter out notebooks with no public chapters
	var filteredNotebooks []models.Notebook
	for _, notebook := range notebooks {
		if len(notebook.Chapters) > 0 {
			filteredNotebooks = append(filteredNotebooks, notebook)
		}
	}

	user.Notebooks = filteredNotebooks

	// Return only necessary user info for public profile
	publicUser := struct {
		ID        uint              `json:"id"`
		Name      string            `json:"name"`
		Email     string            `json:"email"`
		ImageUrl  *string           `json:"imageUrl"`
		Notebooks []models.Notebook `json:"notebooks"`
	}{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		ImageUrl:  user.ImageUrl,
		Notebooks: filteredNotebooks,
	}

	c.JSON(http.StatusOK, publicUser)
}
