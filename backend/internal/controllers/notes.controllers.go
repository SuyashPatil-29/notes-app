package controllers

import (
	"backend/db"
	"backend/internal/middleware"
	"backend/internal/models"
	"encoding/json"
	"net/http"
	"strings"

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

func GenerateNoteVideo(c *gin.Context) {
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
		log.Print("Unauthorized video generation attempt on note: ", id, " by user: ", userID)
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	// Generate video data based on note content
	videoData := generateVideoData(note.Name, note.Content)

	// Convert video data to JSON string
	videoDataJSON, err := json.Marshal(videoData)
	if err != nil {
		log.Print("Error marshaling video data: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate video data"})
		return
	}

	// Update the note with video data
	update := models.Notes{
		VideoData: string(videoDataJSON),
		HasVideo:  true,
	}

	if err := db.DB.Model(&note).Updates(update).Error; err != nil {
		log.Print("Error updating note with video data: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save video data"})
		return
	}

	// Reload the note to get updated data
	if err := db.DB.Preload("Chapter.Notebook").Where("id = ?", id).First(&note).Error; err != nil {
		log.Print("Error reloading note after video generation: ", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Video generated successfully", "note": note})
}

func DeleteNoteVideo(c *gin.Context) {
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
		log.Print("Unauthorized video deletion attempt on note: ", id, " by user: ", userID)
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	// Remove video data from note - use Select to force update empty strings
	if err := db.DB.Model(&note).Select("VideoData", "HasVideo").Updates(map[string]interface{}{
		"video_data": "",
		"has_video":  false,
	}).Error; err != nil {
		log.Print("Error removing video data from note: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove video data"})
		return
	}

	// Reload the note to get updated data
	if err := db.DB.Preload("Chapter.Notebook").Where("id = ?", id).First(&note).Error; err != nil {
		log.Print("Error reloading note after video deletion: ", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Video removed successfully", "note": note})
}

// Helper function to extract text from ProseMirror JSON content
func extractTextFromJSON(content string) string {
	// Try to parse as JSON
	var jsonContent map[string]interface{}
	if err := json.Unmarshal([]byte(content), &jsonContent); err != nil {
		// If not JSON, return as is
		return content
	}

	// Extract text from ProseMirror JSON structure
	var extractText func(node interface{}) string
	extractText = func(node interface{}) string {
		nodeMap, ok := node.(map[string]interface{})
		if !ok {
			return ""
		}

		var text strings.Builder

		// If node has text, add it
		if textVal, ok := nodeMap["text"].(string); ok {
			text.WriteString(textVal)
			text.WriteString(" ")
		}

		// If node has content array, recursively extract text
		if contentArr, ok := nodeMap["content"].([]interface{}); ok {
			for _, child := range contentArr {
				text.WriteString(extractText(child))
			}
		}

		return text.String()
	}

	extractedText := extractText(jsonContent)
	if extractedText == "" {
		// Fallback to raw content if extraction fails
		return content
	}

	return strings.TrimSpace(extractedText)
}

// Helper function to generate video composition data from note content
func generateVideoData(title, content string) map[string]interface{} {
	// Extract text from JSON content if needed
	extractedContent := extractTextFromJSON(content)

	// Truncate content if too long for video
	truncatedContent := extractedContent
	if len(extractedContent) > 500 {
		// Try to truncate at a word boundary
		truncatedContent = extractedContent[:500]
		if lastSpace := strings.LastIndex(truncatedContent, " "); lastSpace > 0 {
			truncatedContent = truncatedContent[:lastSpace] + "..."
		} else {
			truncatedContent += "..."
		}
	}

	return map[string]interface{}{
		"title":            title,
		"content":          truncatedContent,
		"durationInFrames": 180, // 6 seconds at 30fps
		"fps":              30,
		"theme":            "light",
		"themeColors":      nil, // Will be set by frontend
	}
}
