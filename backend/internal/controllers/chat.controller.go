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
		content, ok := toolCall.Args["content"].(string)
		if !ok {
			return map[string]string{"error": "Invalid content parameter"}
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
		noteID, ok := toolCall.Args["noteId"].(string)
		if !ok {
			return map[string]string{"error": "Invalid noteId parameter"}
		}
		content, ok := toolCall.Args["content"].(string)
		if !ok {
			return map[string]string{"error": "Invalid content parameter"}
		}
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
		"chapterName":  chapter.Name,
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

	// Update note's content using Select to force update
	result := db.DB.Model(&note).Select("Content").Updates(models.Notes{Content: content})
	if result.Error != nil {
		log.Error().Err(result.Error).Msg("Failed to update note")
		return map[string]string{"error": "Failed to update note"}
	}

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
		"contentSize":  len(content),
		"chapterName":  note.Chapter.Name,
		"notebookName": note.Chapter.Notebook.Name,
		"updatedAt":    note.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
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
			Description: "Create a new note with markdown content in a specific chapter. The content should be well-formatted markdown. If the user doesn't specify a chapter, list available chapters first.",
			Schema: aisdk.Schema{
				Required: []string{"chapterId", "title", "content"},
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
						"description": "The markdown content of the note. Should be well-formatted with proper headings, lists, code blocks, etc.",
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
			Description: "Update the content of an existing note. Use this when the user wants to modify or add to a note's content.",
			Schema: aisdk.Schema{
				Required: []string{"noteId", "content"},
				Properties: map[string]any{
					"noteId": map[string]any{
						"type":        "string",
						"description": "The ID of the note to update",
					},
					"content": map[string]any{
						"type":        "string",
						"description": "The new markdown content for the note",
					},
				},
			},
		},
	}

	// Tool handler
	handleToolCall := func(toolCall aisdk.ToolCall) any {
		return handleNotesToolCall(toolCall, userID)
	}

	// Set data stream headers
	aisdk.WriteDataStreamHeaders(c.Writer)

	// Add system message if not present
	if len(req.Messages) == 0 || req.Messages[0].Role != "system" {
		req.Messages = append([]aisdk.Message{{
			Role:    "system",
			Content: "You are a helpful AI assistant integrated into a notes application. You have access to tools that can search, list, retrieve, create, update, move, rename, and delete notes and chapters.\n\nAvailable tools:\n- searchNotes: Search through all notes by content or title\n- listNotebooks: List all notebooks\n- listChapters: List chapters in a notebook\n- listNotesInChapter: List notes in a chapter\n- getNoteContent: Get the full content of a specific note\n- createNote: Create a new note with markdown content in a chapter\n- moveNote: Move a note to a different chapter\n- moveChapter: Move an entire chapter (with all its notes) to a different notebook\n- renameNote: Rename a note\n- updateNoteContent: Update the content of an existing note\n- deleteNote: Delete a note permanently\n\nWhen managing notes and chapters:\n1. For create/move operations: If the user doesn't specify which chapter/notebook, list available options first\n2. For moving chapters: Use moveChapter to move entire chapters between notebooks in one operation\n3. For delete operations: Confirm the user really wants to delete before executing\n4. For rename operations: Keep the name concise and descriptive\n5. When creating/updating content: Generate high-quality markdown with proper formatting\n\nAlways provide a clear, helpful text response after using tools. Be conversational and helpful.",
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
				MaxCompletionTokens: openai.Int(2048),
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
				MaxTokens: 4096,
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

		// Continue if tool calls need to be processed
		if acc.FinishReason() == aisdk.FinishReasonToolCalls {
			continue
		}

		break
	}
}
