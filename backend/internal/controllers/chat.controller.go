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
	"backend/internal/services"
	"backend/internal/types"
	internalutils "backend/internal/utils"

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
	Provider       string  `json:"provider"`
	Model          string  `json:"model"`
	Thinking       bool    `json:"thinking"`
	OrganizationID *string `json:"organizationId,omitempty"` // Optional organization context
}

var (
	lastMessages []aisdk.Message
)

// Helper function to check if user has access to a notebook (personal or org) - chat version
func userCanAccessNotebookChat(ctx context.Context, notebook *models.Notebook, clerkUserID string) bool {
	if notebook.OrganizationID != nil && *notebook.OrganizationID != "" {
		// Organization notebook - verify membership
		_, isMember, err := middleware.GetOrgMemberRole(ctx, *notebook.OrganizationID, clerkUserID)
		return err == nil && isMember
	}
	// Personal notebook - verify ownership
	return notebook.ClerkUserID == clerkUserID
}

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

// getUserAPIKey retrieves the user's API key for a specific provider using the new APIKeyResolver
func getUserAPIKey(clerkUserID string, provider string) (string, error) {
	return getUserAPIKeyWithOrg(clerkUserID, nil, provider)
}

// getUserAPIKeyWithOrg retrieves the user's API key for a specific provider with organization context
func getUserAPIKeyWithOrg(clerkUserID string, organizationID *string, provider string) (string, error) {
	resolver := services.NewAPIKeyResolver()
	result, err := resolver.GetAPIKey(clerkUserID, organizationID, provider)
	if err != nil {
		return "", err
	}
	return result.APIKey, nil
}

// handleNotesToolCall handles tool calls for notes-related operations
func handleNotesToolCall(toolCall aisdk.ToolCall, clerkUserID string, organizationID *string) any {
	switch toolCall.Name {
	case "searchNotes":
		query, ok := toolCall.Args["query"].(string)
		if !ok {
			return map[string]string{"error": "Invalid query parameter"}
		}
		return searchNotes(clerkUserID, organizationID, query)

	case "listNotebooks":
		return listNotebooks(clerkUserID, organizationID)

	case "listChapters":
		notebookID, ok := toolCall.Args["notebookId"].(string)
		if !ok {
			return map[string]string{"error": "Invalid notebookId parameter"}
		}
		return listChapters(clerkUserID, notebookID)

	case "getNoteContent":
		noteID, ok := toolCall.Args["noteId"].(string)
		if !ok {
			return map[string]string{"error": "Invalid noteId parameter"}
		}
		return getNoteContent(clerkUserID, noteID)

	case "listNotesInChapter":
		chapterID, ok := toolCall.Args["chapterId"].(string)
		if !ok {
			return map[string]string{"error": "Invalid chapterId parameter"}
		}
		return listNotesInChapter(clerkUserID, chapterID)

	case "createNotebook":
		name, ok := toolCall.Args["name"].(string)
		if !ok {
			return map[string]string{"error": "Invalid name parameter"}
		}
		return createNotebook(clerkUserID, organizationID, name)

	case "createChapter":
		notebookID, ok := toolCall.Args["notebookId"].(string)
		if !ok {
			return map[string]string{"error": "Invalid notebookId parameter"}
		}
		name, ok := toolCall.Args["name"].(string)
		if !ok {
			return map[string]string{"error": "Invalid name parameter"}
		}
		return createChapter(clerkUserID, notebookID, name)

	case "renameNotebook":
		notebookID, ok := toolCall.Args["notebookId"].(string)
		if !ok {
			return map[string]string{"error": "Invalid notebookId parameter"}
		}
		newName, ok := toolCall.Args["newName"].(string)
		if !ok {
			return map[string]string{"error": "Invalid newName parameter"}
		}
		return renameNotebook(clerkUserID, notebookID, newName)

	case "renameChapter":
		chapterID, ok := toolCall.Args["chapterId"].(string)
		if !ok {
			return map[string]string{"error": "Invalid chapterId parameter"}
		}
		newName, ok := toolCall.Args["newName"].(string)
		if !ok {
			return map[string]string{"error": "Invalid newName parameter"}
		}
		return renameChapter(clerkUserID, chapterID, newName)

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
		return createNote(clerkUserID, chapterID, title, content)

	case "moveNote":
		noteID, ok := toolCall.Args["noteId"].(string)
		if !ok {
			return map[string]string{"error": "Invalid noteId parameter"}
		}
		targetChapterID, ok := toolCall.Args["targetChapterId"].(string)
		if !ok {
			return map[string]string{"error": "Invalid targetChapterId parameter"}
		}
		return moveNote(clerkUserID, noteID, targetChapterID)

	case "moveChapter":
		chapterID, ok := toolCall.Args["chapterId"].(string)
		if !ok {
			return map[string]string{"error": "Invalid chapterId parameter"}
		}
		targetNotebookID, ok := toolCall.Args["targetNotebookId"].(string)
		if !ok {
			return map[string]string{"error": "Invalid targetNotebookId parameter"}
		}
		return moveChapter(clerkUserID, chapterID, targetNotebookID)

	case "generateNoteVideo":
		noteID, ok := toolCall.Args["noteId"].(string)
		if !ok {
			return map[string]string{"error": "Invalid noteId parameter"}
		}
		return generateNoteVideo(clerkUserID, noteID)

	case "deleteNoteVideo":
		noteID, ok := toolCall.Args["noteId"].(string)
		if !ok {
			return map[string]string{"error": "Invalid noteId parameter"}
		}
		return deleteNoteVideo(clerkUserID, noteID)

	case "renameNote":
		noteID, ok := toolCall.Args["noteId"].(string)
		if !ok {
			return map[string]string{"error": "Invalid noteId parameter"}
		}
		newName, ok := toolCall.Args["newName"].(string)
		if !ok {
			return map[string]string{"error": "Invalid newName parameter"}
		}
		return renameNote(clerkUserID, noteID, newName)

	case "deleteNote":
		noteID, ok := toolCall.Args["noteId"].(string)
		if !ok {
			return map[string]string{"error": "Invalid noteId parameter"}
		}
		return deleteNote(clerkUserID, noteID)

	case "updateNoteContent":
		noteID, ok := toolCall.Args["noteId"].(string)
		if !ok {
			return map[string]string{"error": "Invalid noteId parameter"}
		}
		content, ok := toolCall.Args["content"].(string)
		if !ok {
			return map[string]string{"error": "Invalid content parameter"}
		}
		return updateNoteContent(clerkUserID, noteID, content)

	default:
		return map[string]string{"error": "Unknown tool: " + toolCall.Name}
	}
}

// getPreviewText extracts preview text from note content (handles JSON or plain text)
func getPreviewText(content string, maxLength int) string {
	if content == "" {
		return "Empty note"
	}

	// Check if content is JSON (ProseMirror format)
	if strings.HasPrefix(strings.TrimSpace(content), "{\"type\":\"doc\"") {
		// Try to parse as ProseMirror JSON and extract text
		var doc map[string]interface{}
		if err := json.Unmarshal([]byte(content), &doc); err == nil {
			text := extractTextFromProseMirror(doc)
			if text == "" {
				return "Empty note"
			}
			if len(text) > maxLength {
				return text[:maxLength] + "..."
			}
			return text
		}
	}

	// Plain text content
	if len(content) > maxLength {
		return content[:maxLength] + "..."
	}
	return content
}

// extractTextFromProseMirror recursively extracts plain text from ProseMirror JSON
func extractTextFromProseMirror(node map[string]interface{}) string {
	var textParts []string

	// If this node has text content, extract it
	if text, ok := node["text"].(string); ok {
		return text
	}

	// If this node has content array, recurse through it
	if content, ok := node["content"].([]interface{}); ok {
		for _, child := range content {
			if childMap, ok := child.(map[string]interface{}); ok {
				childText := extractTextFromProseMirror(childMap)
				if childText != "" {
					textParts = append(textParts, childText)
				}
			}
		}
	}

	return strings.Join(textParts, " ")
}

// searchNotes searches for notes by query in title and content
func searchNotes(clerkUserID string, organizationID *string, query string) any {
	var allNotes []models.Notes

	searchQuery := "%" + strings.ToLower(query) + "%"

	if organizationID != nil && *organizationID != "" {
		// Search in organization notebooks
		err := db.DB.Preload("Chapter.Notebook").
			Joins("JOIN chapters ON notes.chapter_id = chapters.id").
			Joins("JOIN notebooks ON chapters.notebook_id = notebooks.id").
			Where("notebooks.organization_id = ? AND (LOWER(notes.name) LIKE ? OR LOWER(notes.content) LIKE ?)",
				*organizationID, searchQuery, searchQuery).
			Limit(10).
			Find(&allNotes).Error

		if err != nil {
			log.Error().Err(err).Msg("Failed to search organization notes")
			return map[string]string{"error": "Failed to search notes"}
		}
	} else {
		// Search in personal notebooks (organization_id IS NULL)
		err := db.DB.Preload("Chapter.Notebook").
			Joins("JOIN chapters ON notes.chapter_id = chapters.id").
			Joins("JOIN notebooks ON chapters.notebook_id = notebooks.id").
			Where("notebooks.clerk_user_id = ? AND notebooks.organization_id IS NULL AND (LOWER(notes.name) LIKE ? OR LOWER(notes.content) LIKE ?)",
				clerkUserID, searchQuery, searchQuery).
			Limit(10).
			Find(&allNotes).Error

		if err != nil {
			log.Error().Err(err).Msg("Failed to search personal notes")
			return map[string]string{"error": "Failed to search notes"}
		}
	}

	if len(allNotes) == 0 {
		return map[string]any{
			"message": "No notes found matching the query",
			"query":   query,
			"count":   0,
		}
	}

	// Format results
	results := make([]map[string]any, len(allNotes))
	for i, note := range allNotes {
		results[i] = map[string]any{
			"id":           note.ID,
			"name":         note.Name,
			"content":      note.Content,
			"preview":      getPreviewText(note.Content, 150),
			"chapterId":    note.Chapter.ID,
			"chapterName":  note.Chapter.Name,
			"notebookId":   note.Chapter.Notebook.ID,
			"notebookName": note.Chapter.Notebook.Name,
			"updatedAt":    note.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return map[string]any{
		"query":   query,
		"count":   len(allNotes),
		"results": results,
	}
}

// listNotebooks lists all notebooks for a user
func listNotebooks(clerkUserID string, organizationID *string) any {
	var notebooks []models.Notebook

	if organizationID != nil && *organizationID != "" {
		// List organization notebooks
		err := db.DB.Preload("Chapters").
			Where("organization_id = ?", *organizationID).
			Order("updated_at DESC").
			Find(&notebooks).Error

		if err != nil {
			log.Error().Err(err).Msg("Failed to list organization notebooks")
			return map[string]string{"error": "Failed to list notebooks"}
		}

		if len(notebooks) == 0 {
			return map[string]any{
				"message":   "No organization notebooks found. Create one to get started!",
				"count":     0,
				"notebooks": []any{},
			}
		}
	} else {
		// List personal notebooks (organization_id IS NULL)
		err := db.DB.Preload("Chapters").
			Where("clerk_user_id = ? AND organization_id IS NULL", clerkUserID).
			Order("updated_at DESC").
			Find(&notebooks).Error

		if err != nil {
			log.Error().Err(err).Msg("Failed to list personal notebooks")
			return map[string]string{"error": "Failed to list notebooks"}
		}

		if len(notebooks) == 0 {
			return map[string]any{
				"message":   "No personal notebooks found. Create one to get started!",
				"count":     0,
				"notebooks": []any{},
			}
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
func listChapters(clerkUserID string, notebookID string) any {
	// Find notebook first
	var notebook models.Notebook
	err := db.DB.Where("id = ?", notebookID).First(&notebook).Error
	if err != nil {
		log.Error().Err(err).Msg("Notebook not found")
		return map[string]string{"error": "Notebook not found"}
	}

	if !userCanAccessNotebookChat(context.Background(), &notebook, clerkUserID) {
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
func getNoteContent(clerkUserID string, noteID string) any {
	var note models.Notes

	// Get note with relationships
	err := db.DB.Preload("Chapter.Notebook").Where("id = ?", noteID).First(&note).Error
	if err != nil {
		log.Error().Err(err).Msg("Note not found")
		return map[string]string{"error": "Note not found"}
	}

	if !userCanAccessNotebookChat(context.Background(), &note.Chapter.Notebook, clerkUserID) {
		return map[string]string{"error": "Note not found or access denied"}
	}

	return map[string]any{
		"id":           note.ID,
		"name":         note.Name,
		"content":      note.Content,
		"chapterId":    note.Chapter.ID,
		"chapterName":  note.Chapter.Name,
		"notebookId":   note.Chapter.Notebook.ID,
		"notebookName": note.Chapter.Notebook.Name,
		"createdAt":    note.CreatedAt.Format("2006-01-02 15:04:05"),
		"updatedAt":    note.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// listNotesInChapter lists all notes in a chapter
func listNotesInChapter(clerkUserID string, chapterID string) any {
	// Get chapter with notebook
	var chapter models.Chapter
	err := db.DB.Preload("Notebook").Where("id = ?", chapterID).First(&chapter).Error
	if err != nil {
		log.Error().Err(err).Msg("Chapter not found")
		return map[string]string{"error": "Chapter not found"}
	}

	if !userCanAccessNotebookChat(context.Background(), &chapter.Notebook, clerkUserID) {
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
		results[i] = map[string]any{
			"id":        note.ID,
			"name":      note.Name,
			"content":   note.Content,
			"preview":   getPreviewText(note.Content, 150),
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
func createNote(clerkUserID string, chapterID string, title string, content string) any {
	// Get chapter with notebook
	var chapter models.Chapter
	err := db.DB.Preload("Notebook").Where("id = ?", chapterID).First(&chapter).Error
	if err != nil {
		log.Error().Err(err).Msg("Chapter not found")
		return map[string]string{"error": "Chapter not found"}
	}

	if !userCanAccessNotebookChat(context.Background(), &chapter.Notebook, clerkUserID) {
		return map[string]string{"error": "Chapter not found or access denied"}
	}

	// Convert markdown to TipTap JSON format
	tiptapContent, err := internalutils.MarkdownToTipTap(content)
	if err != nil {
		log.Error().Err(err).Str("chapterID", chapterID).Msg("Failed to convert markdown to TipTap JSON")
		return map[string]string{"error": "Failed to convert content format"}
	}

	// Create the note with inherited organization_id
	note := models.Notes{
		Name:           title,
		Content:        tiptapContent,
		ChapterID:      chapterID,
		OrganizationID: chapter.OrganizationID,
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
func moveChapter(clerkUserID string, chapterID string, targetNotebookID string) any {
	// Get chapter with notebook
	var chapter models.Chapter
	err := db.DB.Preload("Notebook").Where("id = ?", chapterID).First(&chapter).Error
	if err != nil {
		log.Error().Err(err).Msg("Chapter not found")
		return map[string]string{"error": "Chapter not found"}
	}

	if !userCanAccessNotebookChat(context.Background(), &chapter.Notebook, clerkUserID) {
		return map[string]string{"error": "Chapter not found or access denied"}
	}

	// Get target notebook and check authorization
	var targetNotebook models.Notebook
	err = db.DB.Where("id = ?", targetNotebookID).First(&targetNotebook).Error
	if err != nil {
		log.Error().Err(err).Msg("Target notebook not found")
		return map[string]string{"error": "Target notebook not found"}
	}

	if !userCanAccessNotebookChat(context.Background(), &targetNotebook, clerkUserID) {
		return map[string]string{"error": "Target notebook not found or access denied"}
	}

	oldNotebookName := chapter.Notebook.Name

	// Count notes in chapter
	var noteCount int64
	db.DB.Model(&models.Notes{}).Where("chapter_id = ?", chapterID).Count(&noteCount)

	// Update chapter's notebook and inherit target notebook's organization_id using explicit Updates
	updateData := map[string]interface{}{
		"notebook_id":     targetNotebookID,
		"organization_id": targetNotebook.OrganizationID,
	}

	result := db.DB.Model(&models.Chapter{}).Where("id = ?", chapterID).Select("notebook_id", "organization_id").Updates(updateData)
	if result.Error != nil {
		log.Error().Err(result.Error).Msg("Failed to move chapter")
		return map[string]string{"error": "Failed to move chapter"}
	}

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
func moveNote(clerkUserID string, noteID string, targetChapterID string) any {
	// Get note with relationships
	var note models.Notes
	err := db.DB.Preload("Chapter.Notebook").Where("id = ?", noteID).First(&note).Error
	if err != nil {
		log.Error().Err(err).Msg("Note not found")
		return map[string]string{"error": "Note not found"}
	}

	if !userCanAccessNotebookChat(context.Background(), &note.Chapter.Notebook, clerkUserID) {
		return map[string]string{"error": "Note not found or access denied"}
	}

	// Get target chapter and check authorization
	var targetChapter models.Chapter
	err = db.DB.Preload("Notebook").Where("id = ?", targetChapterID).First(&targetChapter).Error
	if err != nil {
		log.Error().Err(err).Msg("Target chapter not found")
		return map[string]string{"error": "Target chapter not found"}
	}

	if !userCanAccessNotebookChat(context.Background(), &targetChapter.Notebook, clerkUserID) {
		return map[string]string{"error": "Target chapter not found or access denied"}
	}

	oldChapterName := note.Chapter.Name
	oldNotebookName := note.Chapter.Notebook.Name

	// Update note's chapter and inherit target chapter's organization_id using explicit Updates
	updateData := map[string]interface{}{
		"chapter_id":      targetChapterID,
		"organization_id": targetChapter.OrganizationID,
	}

	result := db.DB.Model(&models.Notes{}).Where("id = ?", noteID).Select("chapter_id", "organization_id").Updates(updateData)
	if result.Error != nil {
		log.Error().Err(result.Error).Msg("Failed to move note")
		return map[string]string{"error": "Failed to move note"}
	}

	return map[string]any{
		"success":      true,
		"message":      "Note moved successfully!",
		"noteId":       note.ID,
		"noteName":     note.Name,
		"fromChapter":  oldChapterName,
		"fromNotebook": oldNotebookName,
		"toChapter":    targetChapter.Name,
		"toNotebook":   targetChapter.Notebook.Name,
		"chapterId":    targetChapter.ID,
		"notebookId":   targetChapter.Notebook.ID,
	}
}

// renameNote renames a note
func renameNote(clerkUserID string, noteID string, newName string) any {
	// Get note with relationships
	var note models.Notes
	err := db.DB.Preload("Chapter.Notebook").Where("id = ?", noteID).First(&note).Error
	if err != nil {
		log.Error().Err(err).Msg("Note not found")
		return map[string]string{"error": "Note not found"}
	}

	if !userCanAccessNotebookChat(context.Background(), &note.Chapter.Notebook, clerkUserID) {
		return map[string]string{"error": "Note not found or access denied"}
	}

	oldName := note.Name

	result := db.DB.Model(&note).Select("Name").Updates(models.Notes{Name: newName})
	if result.Error != nil {
		log.Error().Err(result.Error).Msg("Failed to rename note")
		return map[string]string{"error": "Failed to rename note"}
	}

	return map[string]any{
		"success":      true,
		"message":      "Note renamed successfully!",
		"noteId":       note.ID,
		"oldName":      oldName,
		"newName":      newName,
		"chapterId":    note.Chapter.ID,
		"chapterName":  note.Chapter.Name,
		"notebookId":   note.Chapter.Notebook.ID,
		"notebookName": note.Chapter.Notebook.Name,
	}
}

// deleteNote deletes a note
func deleteNote(clerkUserID string, noteID string) any {
	// Get note with relationships
	var note models.Notes
	err := db.DB.Preload("Chapter.Notebook").Where("id = ?", noteID).First(&note).Error
	if err != nil {
		log.Error().Err(err).Msg("Note not found")
		return map[string]string{"error": "Note not found"}
	}

	if !userCanAccessNotebookChat(context.Background(), &note.Chapter.Notebook, clerkUserID) {
		return map[string]string{"error": "Note not found or access denied"}
	}

	noteName := note.Name
	chapterName := note.Chapter.Name
	notebookName := note.Chapter.Notebook.Name

	err = db.DB.Delete(&note).Error
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete note")
		return map[string]string{"error": "Failed to delete note"}
	}

	return map[string]any{
		"success":      true,
		"message":      "Note deleted successfully!",
		"noteName":     noteName,
		"chapterName":  chapterName,
		"notebookName": notebookName,
	}
}

// updateNoteContent updates the content of a note
func updateNoteContent(clerkUserID string, noteID string, content string) any {
	var note models.Notes
	err := db.DB.Preload("Chapter.Notebook").Where("id = ?", noteID).First(&note).Error
	if err != nil {
		log.Error().Err(err).Str("noteID", noteID).Msg("Note not found")
		return map[string]string{"error": "Note not found"}
	}

	if !userCanAccessNotebookChat(context.Background(), &note.Chapter.Notebook, clerkUserID) {
		return map[string]string{"error": "Note not found or access denied"}
	}

	tiptapContent, err := internalutils.MarkdownToTipTap(content)
	if err != nil {
		log.Error().Err(err).Str("noteID", noteID).Msg("Failed to convert markdown to TipTap JSON")
		return map[string]string{"error": "Failed to convert content format"}
	}

	result := db.DB.Model(&note).Updates(models.Notes{Content: tiptapContent})
	if result.Error != nil {
		log.Error().Err(result.Error).Str("noteID", noteID).Msg("Failed to update note")
		return map[string]string{"error": "Failed to update note"}
	}

	yjsService := services.NewYjsService(db.DB)
	_ = yjsService.DeleteYjsDocument(noteID)

	err = db.DB.Preload("Chapter.Notebook").First(&note, "id = ?", note.ID).Error
	if err != nil {
		log.Error().Err(err).Msg("Failed to reload note after update")
	}

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

// generateNoteVideo creates video data for a note using AI
func generateNoteVideo(clerkUserID string, noteID string) any {
	// Get note with relationships
	var note models.Notes
	err := db.DB.Preload("Chapter.Notebook").Where("id = ?", noteID).First(&note).Error
	if err != nil {
		log.Error().Err(err).Str("noteID", noteID).Msg("Note not found")
		return map[string]string{"error": "Note not found"}
	}

	if !userCanAccessNotebookChat(context.Background(), &note.Chapter.Notebook, clerkUserID) {
		return map[string]string{"error": "Note not found or access denied"}
	}

	videoData, err := GenerateVideoDataWithAI(clerkUserID, note.Name, note.Content)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate AI video, using fallback")
		// Fallback to simple generation
		extractedContent := ExtractTextFromJSON(note.Content)

		// Truncate content if too long
		truncatedContent := extractedContent
		if len(extractedContent) > 500 {
			if lastSpace := strings.LastIndex(extractedContent[:500], " "); lastSpace > 0 {
				truncatedContent = extractedContent[:lastSpace] + "..."
			} else {
				truncatedContent = extractedContent[:500] + "..."
			}
		}

		videoData = map[string]interface{}{
			"title":            note.Name,
			"content":          truncatedContent,
			"durationInFrames": 180,
			"fps":              30,
			"theme":            "light",
		}
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

	return map[string]any{
		"success":      true,
		"message":      "AI-powered video generated successfully! Refresh the note to see it.",
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
func deleteNoteVideo(clerkUserID string, noteID string) any {
	// Get note with relationships
	var note models.Notes
	err := db.DB.Preload("Chapter.Notebook").Where("id = ?", noteID).First(&note).Error
	if err != nil {
		log.Error().Err(err).Str("noteID", noteID).Msg("Note not found")
		return map[string]string{"error": "Note not found"}
	}

	if !userCanAccessNotebookChat(context.Background(), &note.Chapter.Notebook, clerkUserID) {
		return map[string]string{"error": "Note not found or access denied"}
	}

	result := db.DB.Model(&note).Select("VideoData", "HasVideo").Updates(map[string]interface{}{
		"video_data": "",
		"has_video":  false,
	})

	if result.Error != nil {
		log.Error().Err(result.Error).Str("noteID", noteID).Msg("Failed to remove video data")
		return map[string]string{"error": "Failed to remove video data"}
	}

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

// createNotebook creates a new notebook
func createNotebook(clerkUserID string, organizationID *string, name string) any {
	// Create the notebook with appropriate organization context
	notebook := models.Notebook{
		Name:           name,
		ClerkUserID:    clerkUserID,
		OrganizationID: organizationID,
	}

	err := db.DB.Create(&notebook).Error
	if err != nil {
		log.Error().Err(err).Msg("Failed to create notebook")
		return map[string]string{"error": "Failed to create notebook"}
	}

	return map[string]any{
		"success":    true,
		"message":    "Notebook created successfully!",
		"notebookId": notebook.ID,
		"name":       notebook.Name,
		"createdAt":  notebook.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// createChapter creates a new chapter in a notebook
func createChapter(clerkUserID string, notebookID string, name string) any {
	// Get notebook
	var notebook models.Notebook
	err := db.DB.Where("id = ?", notebookID).First(&notebook).Error
	if err != nil {
		log.Error().Err(err).Msg("Notebook not found")
		return map[string]string{"error": "Notebook not found"}
	}

	if !userCanAccessNotebookChat(context.Background(), &notebook, clerkUserID) {
		return map[string]string{"error": "Notebook not found or access denied"}
	}

	chapter := models.Chapter{
		Name:           name,
		NotebookID:     notebookID,
		OrganizationID: notebook.OrganizationID,
	}

	err = db.DB.Create(&chapter).Error
	if err != nil {
		log.Error().Err(err).Msg("Failed to create chapter")
		return map[string]string{"error": "Failed to create chapter"}
	}

	return map[string]any{
		"success":      true,
		"message":      "Chapter created successfully!",
		"chapterId":    chapter.ID,
		"chapterName":  chapter.Name,
		"notebookId":   notebook.ID,
		"notebookName": notebook.Name,
		"createdAt":    chapter.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// renameNotebook renames a notebook
func renameNotebook(clerkUserID string, notebookID string, newName string) any {
	// Get notebook
	var notebook models.Notebook
	err := db.DB.Where("id = ?", notebookID).First(&notebook).Error
	if err != nil {
		log.Error().Err(err).Msg("Notebook not found")
		return map[string]string{"error": "Notebook not found"}
	}

	if !userCanAccessNotebookChat(context.Background(), &notebook, clerkUserID) {
		return map[string]string{"error": "Notebook not found or access denied"}
	}

	oldName := notebook.Name

	result := db.DB.Model(&notebook).Select("Name").Updates(models.Notebook{Name: newName})
	if result.Error != nil {
		log.Error().Err(result.Error).Msg("Failed to rename notebook")
		return map[string]string{"error": "Failed to rename notebook"}
	}

	return map[string]any{
		"success":    true,
		"message":    "Notebook renamed successfully!",
		"notebookId": notebook.ID,
		"oldName":    oldName,
		"newName":    newName,
	}
}

// renameChapter renames a chapter
func renameChapter(clerkUserID string, chapterID string, newName string) any {
	// Get chapter with notebook
	var chapter models.Chapter
	err := db.DB.Preload("Notebook").Where("id = ?", chapterID).First(&chapter).Error
	if err != nil {
		log.Error().Err(err).Msg("Chapter not found")
		return map[string]string{"error": "Chapter not found"}
	}

	if !userCanAccessNotebookChat(context.Background(), &chapter.Notebook, clerkUserID) {
		return map[string]string{"error": "Chapter not found or access denied"}
	}

	oldName := chapter.Name

	result := db.DB.Model(&chapter).Select("Name").Updates(models.Chapter{Name: newName})
	if result.Error != nil {
		log.Error().Err(result.Error).Msg("Failed to rename chapter")
		return map[string]string{"error": "Failed to rename chapter"}
	}

	return map[string]any{
		"success":      true,
		"message":      "Chapter renamed successfully!",
		"chapterId":    chapter.ID,
		"oldName":      oldName,
		"newName":      newName,
		"notebookName": chapter.Notebook.Name,
	}
}

// GenerateRequest represents the request body for the generate endpoint
type GenerateRequest struct {
	Prompt  string `json:"prompt"`
	Option  string `json:"option"`
	Command string `json:"command"`
}

// GenerateHandler handles AI text generation requests (improve, fix, continue, etc.)
func GenerateHandler(c *gin.Context) {
	ctx := c.Request.Context()

	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse generate request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Debug logging
	log.Info().
		Str("prompt", req.Prompt).
		Str("option", req.Option).
		Str("command", req.Command).
		Int("promptLen", len(req.Prompt)).
		Msg("Generate request received")

	// Validate prompt
	if strings.TrimSpace(req.Prompt) == "" && req.Option != "continue" {
		log.Warn().Str("option", req.Option).Msg("Empty prompt received")
		c.JSON(http.StatusBadRequest, gin.H{"error": "No text selected. Please select some text to edit."})
		return
	}

	// Get user ID from session
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get user's OpenAI API key (defaulting to OpenAI for text generation)
	apiKey, err := getUserAPIKey(clerkUserID, "openai")
	if err != nil {
		log.Error().Err(err).Str("clerk_user_id", clerkUserID).Msg("Failed to get user API key")
		c.JSON(http.StatusBadRequest, gin.H{"error": "OpenAI API key not configured. Please set up your API key in settings."})
		return
	}

	// Construct system and user messages based on option
	var systemMessage, userMessage string

	switch req.Option {
	case "continue":
		systemMessage = "You are an AI writing assistant that continues existing text based on context from prior text. " +
			"Give more weight/priority to the later characters than the beginning ones. " +
			"Limit your response to no more than 200 characters, but make sure to construct complete sentences. " +
			"Use Markdown formatting when appropriate. " +
			"Return ONLY the continuation text without any preamble, explanation, or closing remarks."
		userMessage = req.Prompt

	case "improve":
		systemMessage = "You are an AI writing assistant that improves existing text. " +
			"Limit your response to no more than 200 characters, but make sure to construct complete sentences. " +
			"Use Markdown formatting when appropriate. " +
			"Return ONLY the improved text without any preamble like 'Here is...' or closing remarks. " +
			"Do not add any conversational filler - just return the enhanced version of the text directly."
		userMessage = req.Prompt

	case "shorter":
		systemMessage = "You are an AI writing assistant that shortens existing text. " +
			"Use Markdown formatting when appropriate. " +
			"Return ONLY the shortened text without any preamble like 'Here is...' or closing remarks. " +
			"Do not add any conversational filler - just return the condensed version directly."
		userMessage = req.Prompt

	case "longer":
		systemMessage = "You are an AI writing assistant that lengthens existing text. " +
			"Use Markdown formatting when appropriate. " +
			"Return ONLY the expanded text without any preamble like 'Certainly!' or 'Here is...' or closing remarks like 'This elaboration...'. " +
			"Do not add any conversational filler or meta-commentary - just return the lengthened version of the text directly."
		userMessage = req.Prompt

	case "fix":
		systemMessage = "You are an AI writing assistant that fixes grammar and spelling errors in existing text. " +
			"Limit your response to no more than 200 characters, but make sure to construct complete sentences. " +
			"Use Markdown formatting when appropriate. " +
			"Return ONLY the corrected text without any preamble like 'Here is...' or closing remarks. " +
			"Do not add any conversational filler - just return the fixed version directly."
		userMessage = req.Prompt

	case "zap":
		systemMessage = "You are an AI writing assistant that generates text based on a prompt. " +
			"You take an input from the user and a command for manipulating the text. " +
			"Use Markdown formatting when appropriate. " +
			"Return ONLY the generated/edited text without any preamble or explanation. " +
			"Do not add any conversational filler - just return the result directly."
		userMessage = fmt.Sprintf("For this text: %s. You have to respect the command: %s", req.Prompt, req.Command)

	default:
		log.Error().Str("option", req.Option).Msg("Invalid option")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid option"})
		return
	}

	log.Info().
		Str("systemMessage", systemMessage).
		Str("userMessage", userMessage).
		Msg("Messages prepared for OpenAI")

	// Set data stream headers
	aisdk.WriteDataStreamHeaders(c.Writer)

	// Add additional headers to prevent buffering in production (nginx, reverse proxies, etc.)
	c.Header("X-Accel-Buffering", "no")
	c.Header("Cache-Control", "no-cache, no-transform")
	c.Header("Connection", "keep-alive")

	// Create OpenAI messages directly using the SDK
	client := getOpenAIClient(apiKey)

	// Build messages in OpenAI format directly
	openaiMessages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(systemMessage),
		openai.UserMessage(userMessage),
	}

	log.Info().
		Int("messageCount", len(openaiMessages)).
		Msg("OpenAI messages prepared")

	stream := aisdk.OpenAIToDataStream(client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Model:               openai.ChatModelGPT4oMini,
		Messages:            openaiMessages,
		MaxCompletionTokens: openai.Int(4096),
		Temperature:         openai.Float(0.7),
		TopP:                openai.Float(1),
		FrequencyPenalty:    openai.Float(0),
		PresencePenalty:     openai.Float(0),
	}))

	if stream == nil {
		log.Error().Msg("Failed to create stream")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create stream"})
		return
	}

	// Pipe the stream to the response writer
	err = stream.Pipe(c.Writer)
	if err != nil {
		log.Error().Err(err).Msg("Error piping AI response stream")

		// Try to extract error message
		errorMsg := err.Error()
		userFriendlyMsg := errorMsg

		// Check if it's an API key error
		if strings.Contains(errorMsg, "401") || strings.Contains(errorMsg, "Unauthorized") || strings.Contains(errorMsg, "invalid_api_key") || strings.Contains(errorMsg, "Incorrect API key") {
			userFriendlyMsg = "Invalid API key. Please check your API key in Profile settings and ensure it's correct."
		} else if strings.Contains(errorMsg, "insufficient_quota") {
			userFriendlyMsg = "API quota exceeded. Please check your API usage limits."
		} else if strings.Contains(errorMsg, "rate_limit") {
			userFriendlyMsg = "Rate limit exceeded. Please try again in a moment."
		} else {
			userFriendlyMsg = "An error occurred while processing your request. Please try again."
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": userFriendlyMsg})
		return
	}

	log.Info().Str("option", req.Option).Msg("Generate request completed successfully")
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
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get user's API key for the requested provider with organization context
	apiKey, err := getUserAPIKeyWithOrg(clerkUserID, req.OrganizationID, req.Provider)
	if err != nil {
		log.Error().Err(err).Str("provider", req.Provider).Str("clerk_user_id", clerkUserID).Msg("Failed to get user API key")

		// Provide context-aware error message
		var errorMsg string
		var suggestion string

		if req.OrganizationID != nil && *req.OrganizationID != "" {
			errorMsg = "No API key configured for this provider in your organization or personal settings"
			suggestion = "Ask your organization admin to configure API keys in Organization Settings, or set up your personal API key in Profile settings"
		} else {
			errorMsg = "API key not configured for this provider"
			suggestion = "Please set up your API key in Profile settings"
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":       "API_KEY_NOT_CONFIGURED",
				"message":    errorMsg,
				"suggestion": suggestion,
			},
		})
		return
	}

	// Define tools for Atlas
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
			Name:        "createNotebook",
			Description: "Create a new notebook for organizing notes. Use this when reorganizing the structure or when the user needs a new category.",
			Schema: aisdk.Schema{
				Required: []string{"name"},
				Properties: map[string]any{
					"name": map[string]any{
						"type":        "string",
						"description": "The name of the new notebook",
					},
				},
			},
		},
		{
			Name:        "createChapter",
			Description: "Create a new chapter in a notebook. Use this when reorganizing content or when a new sub-category is needed.",
			Schema: aisdk.Schema{
				Required: []string{"notebookId", "name"},
				Properties: map[string]any{
					"notebookId": map[string]any{
						"type":        "string",
						"description": "The ID of the notebook to create the chapter in",
					},
					"name": map[string]any{
						"type":        "string",
						"description": "The name of the new chapter",
					},
				},
			},
		},
		{
			Name:        "renameNotebook",
			Description: "Rename an existing notebook. Use this when reorganizing or improving the organizational structure.",
			Schema: aisdk.Schema{
				Required: []string{"notebookId", "newName"},
				Properties: map[string]any{
					"notebookId": map[string]any{
						"type":        "string",
						"description": "The ID of the notebook to rename",
					},
					"newName": map[string]any{
						"type":        "string",
						"description": "The new name for the notebook",
					},
				},
			},
		},
		{
			Name:        "renameChapter",
			Description: "Rename an existing chapter. Use this when reorganizing or improving the organizational structure.",
			Schema: aisdk.Schema{
				Required: []string{"chapterId", "newName"},
				Properties: map[string]any{
					"chapterId": map[string]any{
						"type":        "string",
						"description": "The ID of the chapter to rename",
					},
					"newName": map[string]any{
						"type":        "string",
						"description": "The new name for the chapter",
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
		return handleNotesToolCall(toolCall, clerkUserID, req.OrganizationID)
	}

	// Set data stream headers
	aisdk.WriteDataStreamHeaders(c.Writer)

	// Add additional headers to prevent buffering in production (nginx, reverse proxies, etc.)
	c.Header("X-Accel-Buffering", "no")
	c.Header("Cache-Control", "no-cache, no-transform")
	c.Header("Connection", "keep-alive")

	// Add system message if not present
	if len(req.Messages) == 0 || req.Messages[0].Role != "system" {
		contextInfo := "personal workspace"
		if req.OrganizationID != nil && *req.OrganizationID != "" {
			contextInfo = "organization workspace"
		}

		req.Messages = append([]aisdk.Message{{
			Role:    "system",
			Content: fmt.Sprintf("You are a helpful AI assistant integrated into Atlas, a knowledge management application. You are currently operating in the user's %s. You have access to tools that can search, list, retrieve, create, update, move, rename, delete notes, chapters, and notebooks, and manage videos within this context.\n\nIMPORTANT: All operations will be scoped to the current workspace context (%s). You will only see and interact with notes, chapters, and notebooks that belong to this workspace.\n\nAvailable tools:\n- searchNotes: Search through all notes by content or title in the current workspace\n- listNotebooks: List all notebooks in the current workspace\n- listChapters: List chapters in a notebook\n- listNotesInChapter: List notes in a chapter\n- getNoteContent: Get the full content of a specific note\n- createNotebook: Create a new notebook in the current workspace\n- createChapter: Create a new chapter in a notebook\n- createNote: Create a new note with markdown content in a chapter\n- moveNote: Move a note to a different chapter\n- moveChapter: Move an entire chapter (with all its notes) to a different notebook\n- renameNotebook: Rename a notebook\n- renameChapter: Rename a chapter\n- renameNote: Rename a note\n- updateNoteContent: Update the content of an existing note\n- deleteNote: Delete a note permanently\n- generateNoteVideo: Generate a short explanatory video for a note based on its content\n- deleteNoteVideo: Remove a video from a note\n\nWhen managing notes and chapters:\n1. For create/move operations: If the user doesn't specify which chapter/notebook, list available options first\n2. For moving chapters: Use moveChapter to move entire chapters between notebooks in one operation\n3. For delete operations: Confirm the user really wants to delete before executing\n4. For rename operations: Keep the name concise and descriptive\n5. When creating/updating content: Generate high-quality markdown with proper formatting, then IMMEDIATELY call the appropriate tool (createNote or updateNoteContent) to save it\n6. IMPORTANT: If user asks to update/modify/edit note content, you MUST call getNoteContent first to read current content, then call updateNoteContent with the new content to save it. Never just describe what to write - always actually save it using the tool.\n7. For videos: Use generateNoteVideo when users want to create explanatory videos for their notes. Videos are generated automatically from note title and content.\n\nREORGANIZATION CAPABILITY:\nYou have the ability to intelligently reorganize the entire notes structure within the current workspace. When asked to reorganize:\n1. Use listNotebooks to see all notebooks in the current workspace\n2. For each notebook, use listChapters to see chapters\n3. For each chapter, use listNotesInChapter and getNoteContent to understand the content\n4. Analyze the content and determine better organizational structure\n5. Create new notebooks/chapters as needed using createNotebook and createChapter\n6. Move notes and chapters to their optimal locations using moveNote and moveChapter\n7. Rename notebooks, chapters, and notes for better clarity using renameNotebook, renameChapter, and renameNote\n8. Provide a summary of all changes made\n\nWhen reorganizing, think about:\n- Thematic grouping (similar topics together)\n- Logical hierarchy (general to specific)\n- Clear, descriptive names\n- Reducing clutter and improving discoverability\n\nAlways provide a clear, helpful text response after using tools. Be conversational and helpful.", contextInfo, contextInfo),
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

			// Get appropriate API error based on the error message
			apiErr := types.GetAPIKeyErrorFromMessage(err)

			// Add context-specific suggestions
			if apiErr.Code == types.ErrorCodeAPIKeyInvalid {
				if req.OrganizationID != nil && *req.OrganizationID != "" {
					apiErr = apiErr.WithSuggestion("Check your organization's API key settings or contact your admin")
				} else {
					apiErr = apiErr.WithSuggestion("Please check your API key in Profile settings and ensure it's correct")
				}
			}

			// Return structured JSON error response
			c.JSON(apiErr.HTTPStatus, gin.H{
				"error": gin.H{
					"code":       apiErr.Code,
					"message":    apiErr.Message,
					"details":    apiErr.Details,
					"suggestion": apiErr.Suggestion,
				},
			})
			return
		}

		req.Messages = append(req.Messages, acc.Messages()...)
		lastMessages = req.Messages[:]

		if acc.FinishReason() == aisdk.FinishReasonToolCalls {
			continue
		}

		break
	}
}
