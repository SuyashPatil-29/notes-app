package controllers

import (
	"backend/db"
	"backend/internal/middleware"
	"backend/internal/models"
	"backend/internal/services"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	anthropicoption "github.com/anthropics/anthropic-sdk-go/option"
	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go"
	openaioption "github.com/openai/openai-go/option"
	"github.com/rs/zerolog/log"
	"google.golang.org/genai"
)

func GetNoteById(c *gin.Context) {
	// Get authenticated user ID
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var note models.Notes
	id := c.Param("id")

	// Get note without preloads
	if err := db.DB.Where("id = ?", id).First(&note).Error; err != nil {
		log.Print("Note not found with id: ", id, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	// Check authorization efficiently
	hasAccess, err := middleware.CheckNoteAccess(c.Request.Context(), db.DB, id, clerkUserID)
	if err != nil || !hasAccess {
		log.Warn().Str("note_id", id).Str("user_id", clerkUserID).Msg("User not authorized to access note")
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	c.JSON(http.StatusOK, note)
}

func GetNotesByChapter(c *gin.Context) {
	// Get authenticated user ID
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	chapterID := c.Param("chapterId")

	// Parse pagination params with defaults
	page := 1
	pageSize := 50
	if pageParam := c.Query("page"); pageParam != "" {
		fmt.Sscanf(pageParam, "%d", &page)
	}
	if pageSizeParam := c.Query("page_size"); pageSizeParam != "" {
		fmt.Sscanf(pageSizeParam, "%d", &pageSize)
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// Check authorization efficiently
	hasAccess, err := middleware.CheckChapterAccess(c.Request.Context(), db.DB, chapterID, clerkUserID)
	if err != nil {
		log.Print("Chapter not found with id: ", chapterID, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Chapter not found"})
		return
	}
	if !hasAccess {
		log.Warn().Str("chapter_id", chapterID).Str("user_id", clerkUserID).Msg("User not authorized to access chapter notes")
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	// Get total count
	var total int64
	db.DB.Model(&models.Notes{}).Where("chapter_id = ?", chapterID).Count(&total)

	// Get notes without large content fields, with pagination
	var notes []models.Notes
	offset := (page - 1) * pageSize
	if err := db.DB.Select("id, name, chapter_id, organization_id, is_public, has_video, meeting_recording_id, created_at, updated_at").
		Where("chapter_id = ?", chapterID).
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&notes).Error; err != nil {
		log.Print("Error fetching notes for chapter: ", chapterID, " Error: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Build response with note list items
	response := make([]gin.H, len(notes))
	for i, note := range notes {
		response[i] = gin.H{
			"id":                 note.ID,
			"name":               note.Name,
			"chapterId":          note.ChapterID,
			"organizationId":     note.OrganizationID,
			"isPublic":           note.IsPublic,
			"hasVideo":           note.HasVideo,
			"meetingRecordingId": note.MeetingRecordingID,
			"createdAt":          note.CreatedAt,
			"updatedAt":          note.UpdatedAt,
		}
	}

	// Calculate pagination metadata
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	if totalPages == 0 {
		totalPages = 1
	}

	paginatedResponse := gin.H{
		"data":       response,
		"page":       page,
		"pageSize":   pageSize,
		"total":      total,
		"totalPages": totalPages,
		"hasNext":    page < totalPages,
		"hasPrev":    page > 1,
	}

	c.JSON(http.StatusOK, paginatedResponse)
}

func CreateNote(c *gin.Context) {
	// Get authenticated user ID
	clerkUserID, exists := middleware.GetClerkUserID(c)
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

	// Check authorization efficiently
	hasAccess, err := middleware.CheckChapterAccess(c.Request.Context(), db.DB, note.ChapterID, clerkUserID)
	if err != nil {
		log.Print("Chapter not found with id: ", note.ChapterID, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Chapter not found"})
		return
	}
	if !hasAccess {
		log.Warn().Str("chapter_id", note.ChapterID).Str("user_id", clerkUserID).Msg("User not authorized to create note in chapter")
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	// Get organization_id from parent chapter
	var chapter models.Chapter
	if err := db.DB.Select("organization_id").Where("id = ?", note.ChapterID).First(&chapter).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create note"})
		return
	}

	// Inherit organization_id from parent chapter
	note.OrganizationID = chapter.OrganizationID

	if err := db.DB.Create(&note).Error; err != nil {
		log.Print("Error creating note in db", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, note)
}

func DeleteNote(c *gin.Context) {
	// Get authenticated user ID
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id := c.Param("id")

	// Check authorization efficiently
	hasAccess, err := middleware.CheckNoteAccess(c.Request.Context(), db.DB, id, clerkUserID)
	if err != nil {
		log.Print("Note not found with id: ", id, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}
	if !hasAccess {
		log.Warn().Str("note_id", id).Str("user_id", clerkUserID).Msg("User not authorized to delete note")
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	if err := db.DB.Delete(&models.Notes{}, "id = ?", id).Error; err != nil {
		log.Print("Error deleting Note with id: ", id, " Error: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note deleted successfully"})
}

func UpdateNote(c *gin.Context) {
	// Get authenticated user ID
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id := c.Param("id")

	// Check authorization efficiently
	hasAccess, err := middleware.CheckNoteAccess(c.Request.Context(), db.DB, id, clerkUserID)
	if err != nil {
		log.Print("Note not found with id: ", id, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}
	if !hasAccess {
		log.Warn().Str("note_id", id).Str("user_id", clerkUserID).Msg("User not authorized to update note")
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

	// Get current note to preserve protected fields
	var note models.Notes
	if err := db.DB.Where("id = ?", id).First(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update note"})
		return
	}

	// Prevent changing chapter_id and organization_id through update (security)
	updateData.ChapterID = note.ChapterID
	updateData.OrganizationID = note.OrganizationID

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
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id := c.Param("id")

	// Check authorization for source note efficiently
	hasAccess, err := middleware.CheckNoteAccess(c.Request.Context(), db.DB, id, clerkUserID)
	if err != nil {
		log.Print("Note not found with id: ", id, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}
	if !hasAccess {
		log.Warn().Str("note_id", id).Str("user_id", clerkUserID).Msg("User not authorized to move note")
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	// Bind the move data from request body
	var moveData struct {
		ChapterID      string  `json:"chapter_id" binding:"required"`
		OrganizationID *string `json:"organization_id"`
	}
	if err := c.ShouldBindJSON(&moveData); err != nil {
		log.Print("Invalid move data for note: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log the received organization ID
	if moveData.OrganizationID != nil {
		log.Printf("DEBUG: Move note request with organization ID: %s", *moveData.OrganizationID)
	}

	// Check authorization for target chapter efficiently
	targetHasAccess, err := middleware.CheckChapterAccess(c.Request.Context(), db.DB, moveData.ChapterID, clerkUserID)
	if err != nil {
		log.Print("Target chapter not found with id: ", moveData.ChapterID, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Target chapter not found"})
		return
	}
	if !targetHasAccess {
		log.Warn().Str("chapter_id", moveData.ChapterID).Str("user_id", clerkUserID).Msg("User not authorized to move note to target chapter")
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized: Target chapter does not belong to user"})
		return
	}

	// Get target chapter's organization_id
	var targetChapter models.Chapter
	if err := db.DB.Select("organization_id").Where("id = ?", moveData.ChapterID).First(&targetChapter).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to move note"})
		return
	}

	// Log before update
	log.Printf("DEBUG: Before update - Note ID: %s, Target Chapter ID: %s", id, moveData.ChapterID)

	// Determine which organization ID to use
	var orgIDToUse *string
	if moveData.OrganizationID != nil && *moveData.OrganizationID != "" {
		// Use the provided organization ID from request
		orgIDToUse = moveData.OrganizationID
		log.Printf("DEBUG: Using organization ID from request: %s", *orgIDToUse)
	} else {
		// Fall back to target chapter's organization ID
		orgIDToUse = targetChapter.OrganizationID
		if orgIDToUse != nil {
			log.Printf("DEBUG: Using organization ID from target chapter: %s", *orgIDToUse)
		}
	}

	// Update the note's chapter_id and organization_id using Updates with Select
	updateData := map[string]interface{}{
		"chapter_id":      moveData.ChapterID,
		"organization_id": orgIDToUse,
	}

	log.Printf("DEBUG: About to update with data: %+v", updateData)

	// Use Model().Select().Updates() for explicit column updates
	result := db.DB.Model(&models.Notes{}).Where("id = ?", id).Select("chapter_id", "organization_id").Updates(updateData)
	if result.Error != nil {
		log.Printf("ERROR: Failed to update note - Error: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	log.Printf("DEBUG: GORM Update Result - Rows Affected: %d, Error: %v", result.RowsAffected, result.Error)

	if result.RowsAffected == 0 {
		log.Printf("WARNING: Update returned 0 rows affected for note ID: %s", id)
	}

	// Reload the note without preloads
	var updatedNote models.Notes
	if err := db.DB.Where("id = ?", id).First(&updatedNote).Error; err != nil {
		log.Printf("ERROR: Failed to reload note after move: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload note"})
		return
	}

	log.Printf("DEBUG: After reload - Note ID: %s, Chapter ID: %s (expected: %s)",
		updatedNote.ID, updatedNote.ChapterID, moveData.ChapterID)
	log.Printf("Note moved successfully. Note ID: %s, New Chapter ID: %s", updatedNote.ID, moveData.ChapterID)

	c.JSON(http.StatusOK, updatedNote)
}

func GenerateNoteVideo(c *gin.Context) {
	// Get authenticated user ID
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id := c.Param("id")

	// Check authorization efficiently
	hasAccess, err := middleware.CheckNoteAccess(c.Request.Context(), db.DB, id, clerkUserID)
	if err != nil {
		log.Print("Note not found with id: ", id, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}
	if !hasAccess {
		log.Warn().Str("note_id", id).Str("user_id", clerkUserID).Msg("User not authorized to generate video for note")
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	// Get note for video generation
	var note models.Notes
	if err := db.DB.Where("id = ?", id).First(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate video"})
		return
	}

	// Generate video data with AI based on note content
	log.Info().Str("note_id", id).Msg("Generating AI-powered video for note")
	videoData, err := GenerateVideoDataWithAI(clerkUserID, note.Name, note.Content)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate AI video, using fallback")
		// Fallback to simple generation
		extractedContent := ExtractTextFromJSON(note.Content)
		videoData = generateVideoData(note.Name, extractedContent)
	}

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

	// Reload the note without preloads
	if err := db.DB.Where("id = ?", id).First(&note).Error; err != nil {
		log.Print("Error reloading note after video generation: ", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Video generated successfully", "note": note})
}

func DeleteNoteVideo(c *gin.Context) {
	// Get authenticated user ID
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id := c.Param("id")
	var note models.Notes

	// Check authorization efficiently
	hasAccess, err := middleware.CheckNoteAccess(c.Request.Context(), db.DB, id, clerkUserID)
	if err != nil {
		log.Print("Note not found with id: ", id, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}
	if !hasAccess {
		log.Warn().Str("note_id", id).Str("user_id", clerkUserID).Msg("User not authorized to delete video for note")
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	// Remove video data from note - use Select to force update empty strings
	if err := db.DB.Model(&models.Notes{}).Where("id = ?", id).Select("VideoData", "HasVideo").Updates(map[string]interface{}{
		"video_data": "",
		"has_video":  false,
	}).Error; err != nil {
		log.Print("Error removing video data from note: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove video data"})
		return
	}

	// Reload the note without preloads
	if err := db.DB.Where("id = ?", id).First(&note).Error; err != nil {
		log.Print("Error reloading note after video deletion: ", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Video removed successfully", "note": note})
}

// VideoSlide represents a single slide/scene in the video
type VideoSlide struct {
	Type     string   `json:"type"`     // "title", "content", "list", "quote"
	Title    string   `json:"title"`    // Main heading
	Content  string   `json:"content"`  // Main text content
	Items    []string `json:"items"`    // For list type
	Duration int      `json:"duration"` // Duration in frames
}

// VideoStructure represents the complete AI-generated video structure
type VideoStructure struct {
	Title           string       `json:"title"`
	Slides          []VideoSlide `json:"slides"`
	TotalDuration   int          `json:"durationInFrames"`
	FPS             int          `json:"fps"`
	Theme           string       `json:"theme"`
	BackgroundStyle string       `json:"backgroundStyle"` // "gradient", "solid", "animated"
	TransitionStyle string       `json:"transitionStyle"` // "fade", "slide", "zoom"
}

// getUserAPIKeyForVideo retrieves the user's API key for video generation using the new APIKeyResolver
func getUserAPIKeyForVideo(clerkUserID string, provider string) (string, error) {
	// TODO: Add organization context support for video generation
	resolver := services.NewAPIKeyResolver()
	result, err := resolver.GetAPIKey(clerkUserID, nil, provider)
	if err != nil {
		return "", err
	}
	return result.APIKey, nil
}

// GenerateVideoDataWithAI uses AI to create a structured video from note content
func GenerateVideoDataWithAI(clerkUserID string, title, content string) (map[string]interface{}, error) {
	// Extract text from JSON content if needed
	extractedContent := ExtractTextFromJSON(content)

	// Try to get API keys in order of preference: OpenAI, Anthropic, Google
	providers := []string{"openai", "anthropic", "google"}
	var apiKey string
	var provider string

	for _, p := range providers {
		key, err := getUserAPIKeyForVideo(clerkUserID, p)
		if err == nil && key != "" {
			apiKey = key
			provider = p
			break
		}
	}

	if apiKey == "" {
		// Fallback to simple generation if no AI provider available
		log.Warn().Msg("No AI provider configured, using simple video generation")
		return generateVideoData(title, extractedContent), nil
	}

	log.Info().Str("provider", provider).Msg("Generating video with AI provider")

	// Create prompt for AI to generate video structure
	prompt := fmt.Sprintf(`You are a video content creator. Analyze the following note content and create an engaging video structure.

Note Title: %s
Note Content: %s

Create a video structure with multiple slides. Each slide should be concise and visually appealing.
Follow these guidelines:
1. Create 3-6 slides depending on content length
2. First slide should be a title slide
3. Break down complex content into digestible chunks
4. Use different slide types: "title", "content", "list", "quote"
5. Each slide should have 60-120 frames (2-4 seconds at 30fps)
6. Extract key points, quotes, or lists from the content

Respond with a JSON object in this exact format (no markdown, just raw JSON):
{
  "title": "Main video title",
  "slides": [
    {
      "type": "title",
      "title": "Introduction title",
      "content": "Brief subtitle or description",
      "items": [],
      "duration": 90
    },
    {
      "type": "content",
      "title": "Section heading",
      "content": "Main content text (2-3 sentences max)",
      "items": [],
      "duration": 120
    },
    {
      "type": "list",
      "title": "Key Points",
      "content": "",
      "items": ["Point 1", "Point 2", "Point 3"],
      "duration": 150
    }
  ],
  "backgroundStyle": "gradient",
  "transitionStyle": "slide"
}`, title, extractedContent[:min(len(extractedContent), 2000)])

	var responseText string
	var err error

	switch provider {
	case "openai":
		responseText, err = generateWithOpenAI(apiKey, prompt)
	case "anthropic":
		responseText, err = generateWithAnthropic(apiKey, prompt)
	case "google":
		responseText, err = generateWithGoogle(apiKey, prompt)
	default:
		return generateVideoData(title, extractedContent), nil
	}

	if err != nil {
		log.Error().Err(err).Msg("AI generation failed, falling back to simple generation")
		return generateVideoData(title, extractedContent), nil
	}

	// Parse AI response
	var videoStructure VideoStructure

	// Clean up response - remove markdown code blocks if present
	responseText = strings.TrimSpace(responseText)
	responseText = strings.TrimPrefix(responseText, "```json")
	responseText = strings.TrimPrefix(responseText, "```")
	responseText = strings.TrimSuffix(responseText, "```")
	responseText = strings.TrimSpace(responseText)

	if err := json.Unmarshal([]byte(responseText), &videoStructure); err != nil {
		log.Error().Err(err).Str("response", responseText).Msg("Failed to parse AI response, using fallback")
		return generateVideoData(title, extractedContent), nil
	}

	// Calculate total duration
	totalDuration := 0
	for _, slide := range videoStructure.Slides {
		totalDuration += slide.Duration
	}
	videoStructure.TotalDuration = totalDuration
	videoStructure.FPS = 30
	videoStructure.Theme = "light"

	// Convert to map for compatibility
	result, err := json.Marshal(videoStructure)
	if err != nil {
		return generateVideoData(title, extractedContent), nil
	}

	var resultMap map[string]interface{}
	json.Unmarshal(result, &resultMap)

	// Add durationInFrames for backward compatibility
	resultMap["durationInFrames"] = totalDuration
	resultMap["fps"] = 30

	return resultMap, nil
}

func generateWithOpenAI(apiKey, prompt string) (string, error) {
	client := openai.NewClient(openaioption.WithAPIKey(apiKey))

	ctx := context.Background()
	completion, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		Model: openai.ChatModelGPT4o,
	})

	if err != nil {
		return "", err
	}

	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return completion.Choices[0].Message.Content, nil
}

func generateWithAnthropic(apiKey, prompt string) (string, error) {
	client := anthropic.NewClient(anthropicoption.WithAPIKey(apiKey))

	ctx := context.Background()
	message, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model("claude-3-5-sonnet-20241022"),
		MaxTokens: 2000,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})

	if err != nil {
		return "", err
	}

	if len(message.Content) == 0 {
		return "", fmt.Errorf("no response from Anthropic")
	}

	return message.Content[0].Text, nil
}

func generateWithGoogle(apiKey, prompt string) (string, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return "", err
	}

	// Collect streamed response
	stream := client.Models.GenerateContentStream(ctx, "gemini-1.5-flash",
		[]*genai.Content{{
			Role: "user",
			Parts: []*genai.Part{{
				Text: prompt,
			}},
		}},
		nil)

	var fullResponse strings.Builder
	for chunk := range stream {
		for _, candidate := range chunk.Candidates {
			if candidate.Content != nil {
				for _, part := range candidate.Content.Parts {
					if part.Text != "" {
						fullResponse.WriteString(part.Text)
					}
				}
			}
		}
	}

	result := fullResponse.String()
	if result == "" {
		return "", fmt.Errorf("no response from Google")
	}

	return result, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ExtractTextFromJSON extracts text from ProseMirror JSON content
func ExtractTextFromJSON(content string) string {
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
	extractedContent := ExtractTextFromJSON(content)

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
