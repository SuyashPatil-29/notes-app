package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"backend/db"
	"backend/internal/middleware"
	"backend/internal/models"
	"backend/pkg/utils"

	"github.com/anthropics/anthropic-sdk-go"
	anthropicoption "github.com/anthropics/anthropic-sdk-go/option"
	"github.com/coder/aisdk-go"
	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go"
	openaioption "github.com/openai/openai-go/option"
	"github.com/rs/zerolog/log"
	"google.golang.org/genai"
)

// ChatRequest wraps the aisdk.Chat with optional extra fields
type ChatRequest struct {
	aisdk.Chat
	Provider string `json:"provider"`
	Model    string `json:"model"`
	Thinking bool   `json:"thinking"`
}

var (
	lastMessages []aisdk.Message
)

func getOpenAIClient(apiKey string) *openai.Client {
	client := openai.NewClient(openaioption.WithAPIKey(apiKey))
	return &client
}

func getAnthropicClient(apiKey string) *anthropic.Client {
	client := anthropic.NewClient(anthropicoption.WithAPIKey(apiKey))
	return &client
}

func getGoogleClient(ctx context.Context, apiKey string) (*genai.Client, error) {
	return genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
}

// getUserAPIKey retrieves the user's API key for a specific provider
func getUserAPIKey(userID uint, provider string) (string, error) {
	var credential models.AICredential
	if err := db.DB.Where("user_id = ? AND provider = ?", userID, provider).First(&credential).Error; err != nil {
		return "", err
	}

	// Decrypt the API key
	apiKey, err := utils.Decrypt(credential.KeyCipher)
	if err != nil {
		log.Error().Err(err).Msg("Failed to decrypt API key")
		return "", err
	}

	// Trim whitespace (leading/trailing spaces/newlines that might have been accidentally saved)
	apiKey = strings.TrimSpace(apiKey)

	if apiKey == "" {
		return "", fmt.Errorf("decrypted API key is empty")
	}

	return apiKey, nil
}

// handleNotesToolCall handles tool calls for notes-related operations
func handleNotesToolCall(toolCall aisdk.ToolCall, userID uint) any {
	log.Info().Str("tool", toolCall.Name).Interface("args", toolCall.Args).Msg("Tool called")

	switch toolCall.Name {
	case "searchNotes":
		query, ok := toolCall.Args["query"].(string)
		if !ok {
			return map[string]string{"error": "Invalid query parameter"}
		}
		return searchNotes(userID, query)

	case "listNotebooks":
		return listNotebooks(userID)

	case "listChapters":
		notebookID, ok := toolCall.Args["notebookId"].(string)
		if !ok {
			return map[string]string{"error": "Invalid notebookId parameter"}
		}
		return listChapters(userID, notebookID)

	case "getNoteContent":
		noteID, ok := toolCall.Args["noteId"].(string)
		if !ok {
			return map[string]string{"error": "Invalid noteId parameter"}
		}
		return getNoteContent(userID, noteID)

	case "listNotesInChapter":
		chapterID, ok := toolCall.Args["chapterId"].(string)
		if !ok {
			return map[string]string{"error": "Invalid chapterId parameter"}
		}
		return listNotesInChapter(userID, chapterID)

	case "createNote":
		chapterID, ok := toolCall.Args["chapterId"].(string)
		if !ok {
			return map[string]string{"error": "Invalid chapterId parameter"}
		}
		title, ok := toolCall.Args["title"].(string)
		if !ok {
			return map[string]string{"error": "Invalid title parameter"}
		}
		// Content is now optional - defaults to empty string
		content := ""
		if contentArg, ok := toolCall.Args["content"].(string); ok {
			content = contentArg
		}
		return createNote(userID, chapterID, title, content)

	case "moveNote":
		noteID, ok := toolCall.Args["noteId"].(string)
		if !ok {
			return map[string]string{"error": "Invalid noteId parameter"}
		}
		targetChapterID, ok := toolCall.Args["targetChapterId"].(string)
		if !ok {
			return map[string]string{"error": "Invalid targetChapterId parameter"}
		}
		return moveNote(userID, noteID, targetChapterID)

	case "moveChapter":
		chapterID, ok := toolCall.Args["chapterId"].(string)
		if !ok {
			return map[string]string{"error": "Invalid chapterId parameter"}
		}
		targetNotebookID, ok := toolCall.Args["targetNotebookId"].(string)
		if !ok {
			return map[string]string{"error": "Invalid targetNotebookId parameter"}
		}
		return moveChapter(userID, chapterID, targetNotebookID)

	case "generateNoteVideo":
		noteID, ok := toolCall.Args["noteId"].(string)
		if !ok {
			return map[string]string{"error": "Invalid noteId parameter"}
		}
		return generateNoteVideo(userID, noteID)

	case "deleteNoteVideo":
		noteID, ok := toolCall.Args["noteId"].(string)
		if !ok {
			return map[string]string{"error": "Invalid noteId parameter"}
		}
		return deleteNoteVideo(userID, noteID)

	case "renameNote":
		noteID, ok := toolCall.Args["noteId"].(string)
		if !ok {
			return map[string]string{"error": "Invalid noteId parameter"}
		}
		newName, ok := toolCall.Args["newName"].(string)
		if !ok {
			return map[string]string{"error": "Invalid newName parameter"}
		}
		return renameNote(userID, noteID, newName)

	case "deleteNote":
		noteID, ok := toolCall.Args["noteId"].(string)
		if !ok {
			return map[string]string{"error": "Invalid noteId parameter"}
		}
		return deleteNote(userID, noteID)

	case "updateNoteContent":
		log.Info().Interface("args", toolCall.Args).Msg("updateNoteContent tool called")
		noteID, ok := toolCall.Args["noteId"].(string)
		if !ok {
			log.Error().Interface("noteId", toolCall.Args["noteId"]).Msg("Invalid noteId parameter")
			return map[string]string{"error": "Invalid noteId parameter"}
		}
		content, ok := toolCall.Args["content"].(string)
		if !ok {
			log.Error().Interface("content", toolCall.Args["content"]).Msg("Invalid content parameter")
			return map[string]string{"error": "Invalid content parameter"}
		}
		log.Info().Str("noteID", noteID).Int("contentLen", len(content)).Msg("Calling updateNoteContent")
		return updateNoteContent(userID, noteID, content)

	default:
		return map[string]string{"error": "Unknown tool: " + toolCall.Name}
	}
}

// searchNotes searches for notes by query in title and content
func searchNotes(userID uint, query string) any {
	var notes []models.Notes

	// Search in both title and content
	searchQuery := "%" + strings.ToLower(query) + "%"
	err := db.DB.Preload("Chapter.Notebook").
		Joins("JOIN chapters ON notes.chapter_id = chapters.id").
		Joins("JOIN notebooks ON chapters.notebook_id = notebooks.id").
		Where("notebooks.user_id = ? AND (LOWER(notes.name) LIKE ? OR LOWER(notes.content) LIKE ?)",
			userID, searchQuery, searchQuery).
		Limit(10).
		Find(&notes).Error

	if err != nil {
		log.Error().Err(err).Msg("Failed to search notes")
		return map[string]string{"error": "Failed to search notes"}
	}

	if len(notes) == 0 {
		return map[string]any{
			"message": "No notes found matching the query",
			"query":   query,
			"count":   0,
		}
	}

	// Format results
	results := make([]map[string]any, len(notes))
	for i, note := range notes {
		// Truncate content for preview
		preview := note.Content
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}

		results[i] = map[string]any{
			"id":           note.ID,
			"name":         note.Name,
			"preview":      preview,
			"chapterName":  note.Chapter.Name,
			"notebookName": note.Chapter.Notebook.Name,
			"updatedAt":    note.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return map[string]any{
		"query":   query,
		"count":   len(notes),
		"results": results,
	}
}

// listNotebooks lists all notebooks for a user
func listNotebooks(userID uint) any {
	var notebooks []models.Notebook

	err := db.DB.Preload("Chapters").
		Where("user_id = ?", userID).
		Order("updated_at DESC").
		Find(&notebooks).Error

	if err != nil {
		log.Error().Err(err).Msg("Failed to list notebooks")
		return map[string]string{"error": "Failed to list notebooks"}
	}

	if len(notebooks) == 0 {
		return map[string]any{
			"message":   "No notebooks found. Create one to get started!",
			"count":     0,
			"notebooks": []any{},
		}
	}

	// Format results
	results := make([]map[string]any, len(notebooks))
	for i, notebook := range notebooks {
		results[i] = map[string]any{
			"id":           notebook.ID,
			"name":         notebook.Name,
			"chapterCount": len(notebook.Chapters),
			"createdAt":    notebook.CreatedAt.Format("2006-01-02"),
			"updatedAt":    notebook.UpdatedAt.Format("2006-01-02"),
		}
	}

	return map[string]any{
		"count":     len(notebooks),
		"notebooks": results,
	}
}

// listChapters lists all chapters in a notebook
func listChapters(userID uint, notebookID string) any {
	// Verify notebook belongs to user
	var notebook models.Notebook
	err := db.DB.Where("id = ? AND user_id = ?", notebookID, userID).First(&notebook).Error
	if err != nil {
		log.Error().Err(err).Msg("Notebook not found")
		return map[string]string{"error": "Notebook not found or access denied"}
	}

	var chapters []models.Chapter
	err = db.DB.Preload("Files").
		Where("notebook_id = ?", notebookID).
		Order("created_at ASC").
		Find(&chapters).Error

	if err != nil {
		log.Error().Err(err).Msg("Failed to list chapters")
		return map[string]string{"error": "Failed to list chapters"}
	}

	if len(chapters) == 0 {
		return map[string]any{
			"notebookId":   notebookID,
			"notebookName": notebook.Name,
			"message":      "No chapters found in this notebook",
			"count":        0,
			"chapters":     []any{},
		}
	}

	// Format results
	results := make([]map[string]any, len(chapters))
	for i, chapter := range chapters {
		results[i] = map[string]any{
			"id":        chapter.ID,
			"name":      chapter.Name,
			"noteCount": len(chapter.Files),
			"createdAt": chapter.CreatedAt.Format("2006-01-02"),
		}
	}

	return map[string]any{
		"notebookId":   notebookID,
		"notebookName": notebook.Name,
		"count":        len(chapters),
		"chapters":     results,
	}
}

// getNoteContent gets the full content of a note
func getNoteContent(userID uint, noteID string) any {
	var note models.Notes

	err := db.DB.Preload("Chapter.Notebook").
		Joins("JOIN chapters ON notes.chapter_id = chapters.id").
		Joins("JOIN notebooks ON chapters.notebook_id = notebooks.id").
		Where("notes.id = ? AND notebooks.user_id = ?", noteID, userID).
		First(&note).Error

	if err != nil {
		log.Error().Err(err).Msg("Note not found")
		return map[string]string{"error": "Note not found or access denied"}
	}

	return map[string]any{
		"id":           note.ID,
		"name":         note.Name,
		"content":      note.Content,
		"chapterName":  note.Chapter.Name,
		"notebookName": note.Chapter.Notebook.Name,
		"createdAt":    note.CreatedAt.Format("2006-01-02 15:04:05"),
		"updatedAt":    note.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// listNotesInChapter lists all notes in a chapter
func listNotesInChapter(userID uint, chapterID string) any {
	// Verify chapter belongs to user's notebook
	var chapter models.Chapter
	err := db.DB.Preload("Notebook").
		Joins("JOIN notebooks ON chapters.notebook_id = notebooks.id").
		Where("chapters.id = ? AND notebooks.user_id = ?", chapterID, userID).
		First(&chapter).Error

	if err != nil {
		log.Error().Err(err).Msg("Chapter not found")
		return map[string]string{"error": "Chapter not found or access denied"}
	}

	var notes []models.Notes
	err = db.DB.Where("chapter_id = ?", chapterID).
		Order("updated_at DESC").
		Find(&notes).Error

	if err != nil {
		log.Error().Err(err).Msg("Failed to list notes")
		return map[string]string{"error": "Failed to list notes"}
	}

	if len(notes) == 0 {
		return map[string]any{
			"chapterId":    chapterID,
			"chapterName":  chapter.Name,
			"notebookName": chapter.Notebook.Name,
			"message":      "No notes found in this chapter",
			"count":        0,
			"notes":        []any{},
		}
	}

	// Format results
	results := make([]map[string]any, len(notes))
	for i, note := range notes {
		// Truncate content for preview
		preview := note.Content
		if len(preview) > 150 {
			preview = preview[:150] + "..."
		}

		results[i] = map[string]any{
			"id":        note.ID,
			"name":      note.Name,
			"preview":   preview,
			"updatedAt": note.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return map[string]any{
		"chapterId":    chapterID,
		"chapterName":  chapter.Name,
		"notebookName": chapter.Notebook.Name,
		"count":        len(notes),
		"notes":        results,
	}
}

// createNote creates a new note in a chapter
func createNote(userID uint, chapterID string, title string, content string) any {
	// Verify chapter belongs to user's notebook
	var chapter models.Chapter
	err := db.DB.Preload("Notebook").
		Joins("JOIN notebooks ON chapters.notebook_id = notebooks.id").
		Where("chapters.id = ? AND notebooks.user_id = ?", chapterID, userID).
		First(&chapter).Error

	if err != nil {
		log.Error().Err(err).Msg("Chapter not found")
		return map[string]string{"error": "Chapter not found or access denied"}
	}

	// Create the note
	note := models.Notes{
		Name:      title,
		Content:   content,
		ChapterID: chapterID,
	}

	err = db.DB.Create(&note).Error
	if err != nil {
		log.Error().Err(err).Msg("Failed to create note")
		return map[string]string{"error": "Failed to create note"}
	}

	// Reload note with relationships
	err = db.DB.Preload("Chapter.Notebook").First(&note, "id = ?", note.ID).Error
	if err != nil {
		log.Error().Err(err).Msg("Failed to reload note")
		// Still return success even if reload fails
	}

	return map[string]any{
		"success":      true,
		"message":      "Note created successfully!",
		"noteId":       note.ID,
		"noteName":     note.Name,
		"chapterId":    chapter.ID,
		"chapterName":  chapter.Name,
		"notebookId":   chapter.Notebook.ID,
		"notebookName": chapter.Notebook.Name,
		"contentSize":  len(content),
		"createdAt":    note.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// moveChapter moves a chapter to a different notebook
func moveChapter(userID uint, chapterID string, targetNotebookID string) any {
	// Verify chapter belongs to user's notebook
	var chapter models.Chapter
	err := db.DB.Preload("Notebook").
		Joins("JOIN notebooks ON chapters.notebook_id = notebooks.id").
		Where("chapters.id = ? AND notebooks.user_id = ?", chapterID, userID).
		First(&chapter).Error

	if err != nil {
		log.Error().Err(err).Msg("Chapter not found")
		return map[string]string{"error": "Chapter not found or access denied"}
	}

	// Verify target notebook belongs to user
	var targetNotebook models.Notebook
	err = db.DB.Where("id = ? AND user_id = ?", targetNotebookID, userID).First(&targetNotebook).Error
	if err != nil {
		log.Error().Err(err).Msg("Target notebook not found")
		return map[string]string{"error": "Target notebook not found or access denied"}
	}

	oldNotebookName := chapter.Notebook.Name

	// Count notes in chapter
	var noteCount int64
	db.DB.Model(&models.Notes{}).Where("chapter_id = ?", chapterID).Count(&noteCount)

	// Update chapter's notebook using Select to force update (same as REST API)
	result := db.DB.Model(&chapter).Select("NotebookID").Updates(models.Chapter{NotebookID: targetNotebookID})
	if result.Error != nil {
		log.Error().Err(result.Error).Msg("Failed to move chapter")
		return map[string]string{"error": "Failed to move chapter"}
	}

	log.Info().Str("chapterId", chapter.ID).Str("newNotebookId", targetNotebookID).Int64("noteCount", noteCount).Msg("Chapter moved successfully via AI tool")

	return map[string]any{
		"success":      true,
		"message":      "Chapter moved successfully!",
		"chapterId":    chapter.ID,
		"chapterName":  chapter.Name,
		"fromNotebook": oldNotebookName,
		"toNotebook":   targetNotebook.Name,
		"notesCount":   noteCount,
	}
}

// moveNote moves a note to a different chapter
func moveNote(userID uint, noteID string, targetChapterID string) any {
	// Verify note belongs to user
	var note models.Notes
	err := db.DB.Preload("Chapter.Notebook").
		Joins("JOIN chapters ON notes.chapter_id = chapters.id").
		Joins("JOIN notebooks ON chapters.notebook_id = notebooks.id").
		Where("notes.id = ? AND notebooks.user_id = ?", noteID, userID).
		First(&note).Error

	if err != nil {
		log.Error().Err(err).Msg("Note not found")
		return map[string]string{"error": "Note not found or access denied"}
	}

	// Verify target chapter belongs to user
	var targetChapter models.Chapter
	err = db.DB.Preload("Notebook").
		Joins("JOIN notebooks ON chapters.notebook_id = notebooks.id").
		Where("chapters.id = ? AND notebooks.user_id = ?", targetChapterID, userID).
		First(&targetChapter).Error

	if err != nil {
		log.Error().Err(err).Msg("Target chapter not found")
		return map[string]string{"error": "Target chapter not found or access denied"}
	}

	oldChapterName := note.Chapter.Name
	oldNotebookName := note.Chapter.Notebook.Name

	// Update note's chapter using Select to force update (same as REST API)
	result := db.DB.Model(&note).Select("ChapterID").Updates(models.Notes{ChapterID: targetChapterID})
	if result.Error != nil {
		log.Error().Err(result.Error).Msg("Failed to move note")
		return map[string]string{"error": "Failed to move note"}
	}

	log.Info().Str("noteId", note.ID).Str("newChapterId", targetChapterID).Msg("Note moved successfully via AI tool")

	return map[string]any{
		"success":      true,
		"message":      "Note moved successfully!",
		"noteId":       note.ID,
		"noteName":     note.Name,
		"fromChapter":  oldChapterName,
		"fromNotebook": oldNotebookName,
		"toChapter":    targetChapter.Name,
		"toNotebook":   targetChapter.Notebook.Name,
	}
}

// renameNote renames a note
func renameNote(userID uint, noteID string, newName string) any {
	// Verify note belongs to user
	var note models.Notes
	err := db.DB.Preload("Chapter.Notebook").
		Joins("JOIN chapters ON notes.chapter_id = chapters.id").
		Joins("JOIN notebooks ON chapters.notebook_id = notebooks.id").
		Where("notes.id = ? AND notebooks.user_id = ?", noteID, userID).
		First(&note).Error

	if err != nil {
		log.Error().Err(err).Msg("Note not found")
		return map[string]string{"error": "Note not found or access denied"}
	}

	oldName := note.Name

	// Update note's name using Select to force update
	result := db.DB.Model(&note).Select("Name").Updates(models.Notes{Name: newName})
	if result.Error != nil {
		log.Error().Err(result.Error).Msg("Failed to rename note")
		return map[string]string{"error": "Failed to rename note"}
	}

	log.Info().Str("noteId", note.ID).Str("oldName", oldName).Str("newName", newName).Msg("Note renamed successfully via AI tool")

	return map[string]any{
		"success":      true,
		"message":      "Note renamed successfully!",
		"noteId":       note.ID,
		"oldName":      oldName,
		"newName":      newName,
		"chapterName":  note.Chapter.Name,
		"notebookName": note.Chapter.Notebook.Name,
	}
}

// deleteNote deletes a note
func deleteNote(userID uint, noteID string) any {
	// Verify note belongs to user
	var note models.Notes
	err := db.DB.Preload("Chapter.Notebook").
		Joins("JOIN chapters ON notes.chapter_id = chapters.id").
		Joins("JOIN notebooks ON chapters.notebook_id = notebooks.id").
		Where("notes.id = ? AND notebooks.user_id = ?", noteID, userID).
		First(&note).Error

	if err != nil {
		log.Error().Err(err).Msg("Note not found")
		return map[string]string{"error": "Note not found or access denied"}
	}

	noteName := note.Name
	chapterName := note.Chapter.Name
	notebookName := note.Chapter.Notebook.Name
	noteIDToDelete := note.ID

	// Delete the note
	err = db.DB.Delete(&note).Error
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete note")
		return map[string]string{"error": "Failed to delete note"}
	}

	log.Info().Str("noteId", noteIDToDelete).Str("noteName", noteName).Msg("Note deleted successfully via AI tool")

	return map[string]any{
		"success":      true,
		"message":      "Note deleted successfully!",
		"noteName":     noteName,
		"chapterName":  chapterName,
		"notebookName": notebookName,
	}
}

// updateNoteContent updates the content of a note
func updateNoteContent(userID uint, noteID string, content string) any {
	log.Info().
		Uint("userID", userID).
		Str("noteID", noteID).
		Int("contentLength", len(content)).
		Str("contentPreview", func() string {
			if len(content) > 100 {
				return content[:100] + "..."
			}
			return content
		}()).
		Msg("updateNoteContent called")

	// Verify note belongs to user
	var note models.Notes
	err := db.DB.Preload("Chapter.Notebook").
		Joins("JOIN chapters ON notes.chapter_id = chapters.id").
		Joins("JOIN notebooks ON chapters.notebook_id = notebooks.id").
		Where("notes.id = ? AND notebooks.user_id = ?", noteID, userID).
		First(&note).Error

	if err != nil {
		log.Error().Err(err).Str("noteID", noteID).Msg("Note not found")
		return map[string]string{"error": "Note not found or access denied"}
	}

	log.Info().Str("noteID", noteID).Str("noteName", note.Name).Msg("Note found, attempting update")

	// Update note's content using Select to force update
	result := db.DB.Model(&note).Select("Content").Updates(models.Notes{Content: content})
	if result.Error != nil {
		log.Error().Err(result.Error).Str("noteID", noteID).Msg("Failed to update note")
		return map[string]string{"error": "Failed to update note"}
	}

	log.Info().Str("noteID", noteID).Int64("rowsAffected", result.RowsAffected).Msg("Database update executed")

	// Reload to get updated timestamp
	err = db.DB.Preload("Chapter.Notebook").First(&note, "id = ?", note.ID).Error
	if err != nil {
		log.Error().Err(err).Msg("Failed to reload note after update")
	}

	log.Info().Str("noteId", note.ID).Int("contentSize", len(content)).Msg("Note content updated successfully via AI tool")

	return map[string]any{
		"success":      true,
		"message":      "Note content updated successfully!",
		"noteId":       note.ID,
		"noteName":     note.Name,
		"chapterId":    note.Chapter.ID,
		"chapterName":  note.Chapter.Name,
		"notebookId":   note.Chapter.Notebook.ID,
		"notebookName": note.Chapter.Notebook.Name,
		"contentSize":  len(content),
		"updatedAt":    note.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// generateNoteVideo creates video data for a note
func generateNoteVideo(userID uint, noteID string) any {
	// Verify note belongs to user
	var note models.Notes
	err := db.DB.Preload("Chapter.Notebook").
		Joins("JOIN chapters ON notes.chapter_id = chapters.id").
		Joins("JOIN notebooks ON chapters.notebook_id = notebooks.id").
		Where("notes.id = ? AND notebooks.user_id = ?", noteID, userID).
		First(&note).Error

	if err != nil {
		log.Error().Err(err).Str("noteID", noteID).Msg("Note not found")
		return map[string]string{"error": "Note not found or access denied"}
	}

	// Extract text from JSON content if needed
	extractedContent := extractTextFromJSONContent(note.Content)

	// Truncate content if too long
	truncatedContent := extractedContent
	if len(extractedContent) > 500 {
		if lastSpace := strings.LastIndex(extractedContent[:500], " "); lastSpace > 0 {
			truncatedContent = extractedContent[:lastSpace] + "..."
		} else {
			truncatedContent = extractedContent[:500] + "..."
		}
	}

	// Generate video data from note content
	videoData := map[string]interface{}{
		"title":            note.Name,
		"content":          truncatedContent,
		"durationInFrames": 180, // 6 seconds at 30fps
		"fps":              30,
		"theme":            "light",
	}

	// Convert video data to JSON string
	videoDataJSON, err := json.Marshal(videoData)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal video data")
		return map[string]string{"error": "Failed to generate video data"}
	}

	// Update note with video data
	result := db.DB.Model(&note).Select("VideoData", "HasVideo").Updates(models.Notes{
		VideoData: string(videoDataJSON),
		HasVideo:  true,
	})

	if result.Error != nil {
		log.Error().Err(result.Error).Str("noteID", noteID).Msg("Failed to update note with video data")
		return map[string]string{"error": "Failed to save video data"}
	}

	log.Info().Str("noteID", noteID).Str("noteName", note.Name).Msg("Video generated successfully via AI tool")

	return map[string]any{
		"success":      true,
		"message":      "Video generated successfully! Refresh the note to see it.",
		"noteId":       note.ID,
		"noteName":     note.Name,
		"chapterId":    note.Chapter.ID,
		"chapterName":  note.Chapter.Name,
		"notebookId":   note.Chapter.Notebook.ID,
		"notebookName": note.Chapter.Notebook.Name,
		"hasVideo":     true,
	}
}

// deleteNoteVideo removes video data from a note
func deleteNoteVideo(userID uint, noteID string) any {
	// Verify note belongs to user
	var note models.Notes
	err := db.DB.Preload("Chapter.Notebook").
		Joins("JOIN chapters ON notes.chapter_id = chapters.id").
		Joins("JOIN notebooks ON chapters.notebook_id = notebooks.id").
		Where("notes.id = ? AND notebooks.user_id = ?", noteID, userID).
		First(&note).Error

	if err != nil {
		log.Error().Err(err).Str("noteID", noteID).Msg("Note not found")
		return map[string]string{"error": "Note not found or access denied"}
	}

	// Remove video data - use map to force update empty strings
	result := db.DB.Model(&note).Select("VideoData", "HasVideo").Updates(map[string]interface{}{
		"video_data": "",
		"has_video":  false,
	})

	if result.Error != nil {
		log.Error().Err(result.Error).Str("noteID", noteID).Msg("Failed to remove video data")
		return map[string]string{"error": "Failed to remove video data"}
	}

	log.Info().Str("noteID", noteID).Str("noteName", note.Name).Msg("Video removed successfully via AI tool")

	return map[string]any{
		"success":      true,
		"message":      "Video removed successfully! Refresh the note to see the changes.",
		"noteId":       note.ID,
		"noteName":     note.Name,
		"chapterId":    note.Chapter.ID,
		"chapterName":  note.Chapter.Name,
		"notebookId":   note.Chapter.Notebook.ID,
		"notebookName": note.Chapter.Notebook.Name,
		"hasVideo":     false,
	}
}

// extractTextFromJSONContent extracts text from ProseMirror JSON content
func extractTextFromJSONContent(content string) string {
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

// DumpHandler dumps the last messages to a JSON file
func DumpHandler(c *gin.Context) {
	data, _ := json.MarshalIndent(lastMessages, "", "  ")
	err := os.WriteFile("dump.json", data, 0644)
	if err != nil {
		log.Error().Err(err).Msg("Failed to write dump file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to dump messages"})
		return
	}
	log.Info().Msg("Dumped messages to dump.json")
	c.JSON(http.StatusOK, gin.H{"message": "Dumped to dump.json"})
}

// ChatHandler handles the chat API endpoint with multi-provider support
func ChatHandler(c *gin.Context) {
	ctx := c.Request.Context()

	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse chat request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get user ID from session
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get user's API key for the requested provider
	apiKey, err := getUserAPIKey(userID, req.Provider)
	if err != nil {
		log.Error().Err(err).Str("provider", req.Provider).Uint("userID", userID).Msg("Failed to get user API key")
		c.JSON(http.StatusBadRequest, gin.H{"error": "API key not configured for this provider. Please set up your API key in settings."})
		return
	}

	// Define tools for notes app
	tools := []aisdk.Tool{
		{
			Name:        "searchNotes",
			Description: "Search through all notes content and titles. Returns matching notes with their content, chapter, and notebook information.",
			Schema: aisdk.Schema{
				Required: []string{"query"},
				Properties: map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "The search query to find in notes",
					},
				},
			},
		},
		{
			Name:        "listNotebooks",
			Description: "List all notebooks for the current user. Returns notebook IDs, names, and chapter counts.",
			Schema: aisdk.Schema{
				Properties: map[string]any{},
			},
		},
		{
			Name:        "listChapters",
			Description: "List all chapters in a specific notebook. Returns chapter IDs, names, and note counts.",
			Schema: aisdk.Schema{
				Required: []string{"notebookId"},
				Properties: map[string]any{
					"notebookId": map[string]any{
						"type":        "string",
						"description": "The ID of the notebook to list chapters from",
					},
				},
			},
		},
		{
			Name:        "getNoteContent",
			Description: "Get the full content of a specific note by its ID. Returns the note's title, content, and metadata.",
			Schema: aisdk.Schema{
				Required: []string{"noteId"},
				Properties: map[string]any{
					"noteId": map[string]any{
						"type":        "string",
						"description": "The ID of the note to retrieve",
					},
				},
			},
		},
		{
			Name:        "listNotesInChapter",
			Description: "List all notes in a specific chapter. Returns note IDs, names, and preview of content.",
			Schema: aisdk.Schema{
				Required: []string{"chapterId"},
				Properties: map[string]any{
					"chapterId": map[string]any{
						"type":        "string",
						"description": "The ID of the chapter to list notes from",
					},
				},
			},
		},
		{
			Name:        "createNote",
			Description: "Create a new note in a specific chapter. IMPORTANT: When the user asks you to create a note with content, you should: 1) First call this tool with just the title (content is optional, defaults to empty), which will create the note and allow navigation to it. 2) Then immediately call updateNoteContent with the full markdown content to populate the note. This two-step process allows the user to see the note being created and filled in real-time. If the user doesn't specify a chapter, list available chapters first.",
			Schema: aisdk.Schema{
				Required: []string{"chapterId", "title"},
				Properties: map[string]any{
					"chapterId": map[string]any{
						"type":        "string",
						"description": "The ID of the chapter to create the note in",
					},
					"title": map[string]any{
						"type":        "string",
						"description": "The title of the note",
					},
					"content": map[string]any{
						"type":        "string",
						"description": "Optional: The initial markdown content of the note. Can be left empty and populated later with updateNoteContent.",
					},
				},
			},
		},
		{
			Name:        "moveNote",
			Description: "Move a note to a different chapter. Use this when the user wants to reorganize their notes.",
			Schema: aisdk.Schema{
				Required: []string{"noteId", "targetChapterId"},
				Properties: map[string]any{
					"noteId": map[string]any{
						"type":        "string",
						"description": "The ID of the note to move",
					},
					"targetChapterId": map[string]any{
						"type":        "string",
						"description": "The ID of the chapter to move the note to",
					},
				},
			},
		},
		{
			Name:        "moveChapter",
			Description: "Move an entire chapter (with all its notes) to a different notebook. Use this when the user wants to reorganize chapters between notebooks.",
			Schema: aisdk.Schema{
				Required: []string{"chapterId", "targetNotebookId"},
				Properties: map[string]any{
					"chapterId": map[string]any{
						"type":        "string",
						"description": "The ID of the chapter to move",
					},
					"targetNotebookId": map[string]any{
						"type":        "string",
						"description": "The ID of the notebook to move the chapter to",
					},
				},
			},
		},
		{
			Name:        "renameNote",
			Description: "Rename a note. Use this when the user wants to change a note's title.",
			Schema: aisdk.Schema{
				Required: []string{"noteId", "newName"},
				Properties: map[string]any{
					"noteId": map[string]any{
						"type":        "string",
						"description": "The ID of the note to rename",
					},
					"newName": map[string]any{
						"type":        "string",
						"description": "The new name for the note",
					},
				},
			},
		},
		{
			Name:        "deleteNote",
			Description: "Delete a note permanently. Use this when the user explicitly wants to remove a note.",
			Schema: aisdk.Schema{
				Required: []string{"noteId"},
				Properties: map[string]any{
					"noteId": map[string]any{
						"type":        "string",
						"description": "The ID of the note to delete",
					},
				},
			},
		},
		{
			Name:        "updateNoteContent",
			Description: "Update and save the content of an existing note. REQUIRED: You must call this tool whenever the user asks to update, modify, edit, or change a note's content. First call getNoteContent to read the current content, then call this tool with the complete updated markdown content to save it permanently. This is the ONLY way to save changes to a note.",
			Schema: aisdk.Schema{
				Required: []string{"noteId", "content"},
				Properties: map[string]any{
					"noteId": map[string]any{
						"type":        "string",
						"description": "The ID of the note to update",
					},
					"content": map[string]any{
						"type":        "string",
						"description": "The complete new markdown content for the note. This will replace the entire note content.",
					},
				},
			},
		},
		{
			Name:        "generateNoteVideo",
			Description: "Generate a short explanatory video for an existing note based on its title and content. The video will be created automatically and can be viewed in the note editor. Use this when the user wants to create a video explanation for their note content.",
			Schema: aisdk.Schema{
				Required: []string{"noteId"},
				Properties: map[string]any{
					"noteId": map[string]any{
						"type":        "string",
						"description": "The ID of the note to generate a video for",
					},
				},
			},
		},
		{
			Name:        "deleteNoteVideo",
			Description: "Remove the video from a note. Use this when the user wants to delete an existing video from their note.",
			Schema: aisdk.Schema{
				Required: []string{"noteId"},
				Properties: map[string]any{
					"noteId": map[string]any{
						"type":        "string",
						"description": "The ID of the note to remove the video from",
					},
				},
			},
		},
	}

	// Tool handler
	handleToolCall := func(toolCall aisdk.ToolCall) any {
		log.Info().Str("toolName", toolCall.Name).Str("toolCallId", toolCall.ID).Msg("handleToolCall invoked")
		result := handleNotesToolCall(toolCall, userID)
		log.Info().Str("toolName", toolCall.Name).Interface("result", result).Msg("handleToolCall completed")
		return result
	}

	// Set data stream headers
	aisdk.WriteDataStreamHeaders(c.Writer)

	// Add system message if not present
	if len(req.Messages) == 0 || req.Messages[0].Role != "system" {
		req.Messages = append([]aisdk.Message{{
			Role:    "system",
			Content: "You are a helpful AI assistant integrated into a notes application. You have access to tools that can search, list, retrieve, create, update, move, rename, delete notes and chapters, and manage videos.\n\nAvailable tools:\n- searchNotes: Search through all notes by content or title\n- listNotebooks: List all notebooks\n- listChapters: List chapters in a notebook\n- listNotesInChapter: List notes in a chapter\n- getNoteContent: Get the full content of a specific note\n- createNote: Create a new note with markdown content in a chapter\n- moveNote: Move a note to a different chapter\n- moveChapter: Move an entire chapter (with all its notes) to a different notebook\n- renameNote: Rename a note\n- updateNoteContent: Update the content of an existing note\n- deleteNote: Delete a note permanently\n- generateNoteVideo: Generate a short explanatory video for a note based on its content\n- deleteNoteVideo: Remove a video from a note\n\nWhen managing notes and chapters:\n1. For create/move operations: If the user doesn't specify which chapter/notebook, list available options first\n2. For moving chapters: Use moveChapter to move entire chapters between notebooks in one operation\n3. For delete operations: Confirm the user really wants to delete before executing\n4. For rename operations: Keep the name concise and descriptive\n5. When creating/updating content: Generate high-quality markdown with proper formatting, then IMMEDIATELY call the appropriate tool (createNote or updateNoteContent) to save it\n6. IMPORTANT: If user asks to update/modify/edit note content, you MUST call getNoteContent first to read current content, then call updateNoteContent with the new content to save it. Never just describe what to write - always actually save it using the tool.\n7. For videos: Use generateNoteVideo when users want to create explanatory videos for their notes. Videos are generated automatically from note title and content.\n\nAlways provide a clear, helpful text response after using tools. Be conversational and helpful.",
		}}, req.Messages...)
	}

	// Main streaming loop (handles tool calls)
	for {
		var stream aisdk.DataStream

		switch req.Provider {
		case "openai":
			messages, err := aisdk.MessagesToOpenAI(req.Messages)
			if err != nil {
				log.Error().Err(err).Msg("Failed to convert messages for OpenAI")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare messages"})
				return
			}

			reasoningEffort := openai.ReasoningEffort("")
			if req.Thinking {
				reasoningEffort = openai.ReasoningEffortMedium
			}

			client := getOpenAIClient(apiKey)
			stream = aisdk.OpenAIToDataStream(client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
				Model:               req.Model,
				Messages:            messages,
				ReasoningEffort:     reasoningEffort,
				Tools:               aisdk.ToolsToOpenAI(tools),
				MaxCompletionTokens: openai.Int(16384),
			}))

		case "anthropic":
			messages, system, err := aisdk.MessagesToAnthropic(req.Messages)
			if err != nil {
				log.Error().Err(err).Msg("Failed to convert messages for Anthropic")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare messages"})
				return
			}

			thinking := anthropic.ThinkingConfigParamUnion{}
			if req.Thinking {
				thinking = anthropic.ThinkingConfigParamOfEnabled(2048)
			}

			client := getAnthropicClient(apiKey)
			stream = aisdk.AnthropicToDataStream(client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
				Model:     anthropic.Model(req.Model),
				Messages:  messages,
				System:    system,
				MaxTokens: 16384,
				Thinking:  thinking,
				Tools:     aisdk.ToolsToAnthropic(tools),
			}))

		case "google":
			googleClient, err := getGoogleClient(ctx, apiKey)
			if err != nil {
				log.Error().Err(err).Msg("Google client not initialized")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Google client not configured"})
				return
			}

			messages, err := aisdk.MessagesToGoogle(req.Messages)
			if err != nil {
				log.Error().Err(err).Msg("Failed to convert messages for Google")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare messages"})
				return
			}

			var thinkingConfig *genai.ThinkingConfig
			if req.Thinking {
				thinkingConfig = &genai.ThinkingConfig{
					IncludeThoughts: true,
				}
			}

			googleTools, err := aisdk.ToolsToGoogle(tools)
			if err != nil {
				log.Error().Err(err).Msg("Failed to convert tools for Google")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare tools"})
				return
			}

			stream = aisdk.GoogleToDataStream(googleClient.Models.GenerateContentStream(ctx, req.Model, messages, &genai.GenerateContentConfig{
				Tools:          googleTools,
				ThinkingConfig: thinkingConfig,
			}))

		default:
			log.Error().Str("provider", req.Provider).Msg("Invalid provider")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider"})
			return
		}

		if stream == nil {
			log.Error().Msg("Failed to create stream")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create stream"})
			return
		}

		// Setup accumulator and tool calling
		var acc aisdk.DataStreamAccumulator
		stream = stream.WithToolCalling(handleToolCall)
		stream = stream.WithAccumulator(&acc)

		// Pipe the stream to the response writer
		err := stream.Pipe(c.Writer)
		if err != nil {
			log.Error().Err(err).Msg("Error piping AI response stream")

			// Try to extract error message from the error
			errorMsg := err.Error()
			userFriendlyMsg := errorMsg

			// Check if it's an API key error
			if strings.Contains(errorMsg, "401") || strings.Contains(errorMsg, "Unauthorized") || strings.Contains(errorMsg, "invalid_api_key") || strings.Contains(errorMsg, "Incorrect API key") {
				userFriendlyMsg = "Invalid API key. Please check your API key in Profile settings and ensure it's correct."
			} else if strings.Contains(errorMsg, "insufficient_quota") {
				userFriendlyMsg = "API quota exceeded. Please check your API usage limits."
			} else if strings.Contains(errorMsg, "rate_limit") {
				userFriendlyMsg = "Rate limit exceeded. Please try again in a moment."
			} else if strings.Contains(errorMsg, "model_not_found") {
				userFriendlyMsg = "The selected model is not available. Please try a different model."
			} else {
				// Generic error message
				userFriendlyMsg = "An error occurred while processing your request. Please try again."
			}

			// Return JSON error response instead of trying to write to stream
			c.JSON(http.StatusInternalServerError, gin.H{"error": userFriendlyMsg})
			return
		}

		// Update messages with accumulated content
		req.Messages = append(req.Messages, acc.Messages()...)
		lastMessages = req.Messages[:]

		// Log finish reason for debugging
		log.Info().Str("finishReason", string(acc.FinishReason())).Int("messageCount", len(acc.Messages())).Msg("Stream completed")

		// Continue if tool calls need to be processed
		if acc.FinishReason() == aisdk.FinishReasonToolCalls {
			log.Info().Msg("Tool calls detected, continuing loop")
			continue
		}

		log.Info().Msg("Streaming loop finished")
		break
	}
}
