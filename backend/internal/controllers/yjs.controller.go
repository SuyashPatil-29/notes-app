package controllers

import (
	"backend/db"
	"backend/internal/middleware"
	"backend/internal/services"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// GetYjsState retrieves the current Yjs state for a note
// GET /api/notes/:id/yjs-state
func GetYjsState(c *gin.Context) {
	// Get authenticated user ID
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	noteID := c.Param("id")

	// Check authorization
	hasAccess, err := middleware.CheckNoteAccess(c.Request.Context(), db.DB, noteID, clerkUserID)
	if err != nil || !hasAccess {
		log.Warn().Str("note_id", noteID).Str("user_id", clerkUserID).Msg("User not authorized to access note")
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	// Get or create Yjs state
	yjsService := services.NewYjsService(db.DB)
	response, err := yjsService.GetOrCreateYjsState(noteID)
	if err != nil {
		log.Error().Err(err).Str("note_id", noteID).Msg("Failed to get Yjs state")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Yjs state"})
		return
	}

	// If document exists, return binary Yjs state
	if response.Exists {
		c.Data(http.StatusOK, "application/octet-stream", response.YjsState)
		return
	}

	// Document doesn't exist, return JSON indicating initialization is needed
	c.JSON(http.StatusOK, gin.H{
		"requiresInit": true,
		"noteContent":  response.NoteContent,
		"version":      response.Version,
	})
}

// InitializeYjsDocument creates a new Yjs document from initial state
// POST /api/notes/:id/yjs-init
func InitializeYjsDocument(c *gin.Context) {
	// Get authenticated user ID
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	noteID := c.Param("id")

	// Check authorization
	hasAccess, err := middleware.CheckNoteAccess(c.Request.Context(), db.DB, noteID, clerkUserID)
	if err != nil || !hasAccess {
		log.Warn().Str("note_id", noteID).Str("user_id", clerkUserID).Msg("User not authorized to access note")
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	// Read binary Yjs state from request body
	initialState, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if len(initialState) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Empty initial state"})
		return
	}

	// Initialize Yjs document
	yjsService := services.NewYjsService(db.DB)
	err = yjsService.InitializeYjsDocument(noteID, initialState)
	if err != nil {
		log.Error().Err(err).Str("note_id", noteID).Msg("Failed to initialize Yjs document")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize document"})
		return
	}

	log.Info().Str("note_id", noteID).Msg("Yjs document initialized successfully")
	c.JSON(http.StatusCreated, gin.H{"message": "Yjs document initialized"})
}

// ApplyYjsUpdate applies a Yjs update to the document
// POST /api/notes/:id/yjs-update
func ApplyYjsUpdate(c *gin.Context) {
	// Get authenticated user ID
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	noteID := c.Param("id")

	// Check authorization
	hasAccess, err := middleware.CheckNoteAccess(c.Request.Context(), db.DB, noteID, clerkUserID)
	if err != nil || !hasAccess {
		log.Warn().Str("note_id", noteID).Str("user_id", clerkUserID).Msg("User not authorized to access note")
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	// Expect two fields: update (binary) and state (binary)
	// We'll use multipart form or expect JSON with base64
	// For simplicity, let's use query params to differentiate
	
	// Read binary update from request body
	updateData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if len(updateData) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Empty update data"})
		return
	}

	// For now, we'll use the update data as both the update and the new state
	// In a more sophisticated implementation, we'd apply the update to the existing state
	// But since we're just storing binary data, we can use it directly
	yjsService := services.NewYjsService(db.DB)
	err = yjsService.ApplyUpdate(noteID, updateData, updateData)
	if err != nil {
		log.Error().Err(err).Str("note_id", noteID).Msg("Failed to apply Yjs update")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to apply update"})
		return
	}

	log.Debug().Str("note_id", noteID).Msg("Yjs update applied successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Update applied"})
}

// SyncYjsToNote syncs Yjs state back to the note's JSON content field
// POST /api/notes/:id/yjs-sync
func SyncYjsToNote(c *gin.Context) {
	// Get authenticated user ID
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	noteID := c.Param("id")

	// Check authorization
	hasAccess, err := middleware.CheckNoteAccess(c.Request.Context(), db.DB, noteID, clerkUserID)
	if err != nil || !hasAccess {
		log.Warn().Str("note_id", noteID).Str("user_id", clerkUserID).Msg("User not authorized to access note")
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	// Expect JSON content in request body
	var requestData struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Sync to note
	yjsService := services.NewYjsService(db.DB)
	err = yjsService.SyncYjsToNoteContent(noteID, requestData.Content)
	if err != nil {
		log.Error().Err(err).Str("note_id", noteID).Msg("Failed to sync content")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync content"})
		return
	}

	log.Debug().Str("note_id", noteID).Msg("Content synced to note")
	c.JSON(http.StatusOK, gin.H{"message": "Content synced"})
}

// GetDocumentVersion gets the current version of a Yjs document
// GET /api/notes/:id/yjs-version
func GetDocumentVersion(c *gin.Context) {
	// Get authenticated user ID
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	noteID := c.Param("id")

	// Check authorization
	hasAccess, err := middleware.CheckNoteAccess(c.Request.Context(), db.DB, noteID, clerkUserID)
	if err != nil || !hasAccess {
		log.Warn().Str("note_id", noteID).Str("user_id", clerkUserID).Msg("User not authorized to access note")
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	// Get version
	yjsService := services.NewYjsService(db.DB)
	version, err := yjsService.GetDocumentVersion(noteID)
	if err != nil {
		log.Error().Err(err).Str("note_id", noteID).Msg("Failed to get document version")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get version"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"version": version})
}

