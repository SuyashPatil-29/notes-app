package controllers

import (
	"backend/internal/middleware"
	"backend/internal/models/dto"
	"backend/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// CreateNoteLink creates a new bidirectional link between notes
func CreateNoteLink(c *gin.Context) {
	// Initialize service (lazy initialization to ensure DB is ready)
	noteLinkService := services.NewNoteLinkService()
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req dto.CreateNoteLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get organization ID from context
	var organizationID *string
	if orgID, exists := middleware.GetOrganizationID(c); exists && orgID != "" {
		organizationID = &orgID
	}

	// Set default link type if not provided
	if req.LinkType == "" {
		req.LinkType = "references"
	}

	link, err := noteLinkService.CreateNoteLink(
		req.SourceNoteID,
		req.TargetNoteID,
		req.LinkType,
		clerkUserID,
		organizationID,
	)

	if err != nil {
		log.Error().Err(err).Msg("Failed to create note link")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, link)
}

// DeleteNoteLink deletes a note link by ID
func DeleteNoteLink(c *gin.Context) {
	// Initialize service (lazy initialization to ensure DB is ready)
	noteLinkService := services.NewNoteLinkService()
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	linkID := c.Param("id")

	err := noteLinkService.DeleteNoteLink(linkID, clerkUserID)
	if err != nil {
		log.Error().Err(err).Str("link_id", linkID).Msg("Failed to delete note link")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note link deleted successfully"})
}

// GetNoteLinksByNoteID retrieves all links for a specific note
func GetNoteLinksByNoteID(c *gin.Context) {
	// Initialize service (lazy initialization to ensure DB is ready)
	noteLinkService := services.NewNoteLinkService()
	_, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	noteID := c.Param("id")

	links, err := noteLinkService.GetNoteLinksByNoteID(noteID)
	if err != nil {
		log.Error().Err(err).Str("note_id", noteID).Msg("Failed to get note links")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, links)
}

// UpdateNoteLink updates a note link's type
func UpdateNoteLink(c *gin.Context) {
	// Initialize service (lazy initialization to ensure DB is ready)
	noteLinkService := services.NewNoteLinkService()
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	linkID := c.Param("id")

	var req dto.UpdateNoteLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	link, err := noteLinkService.UpdateNoteLink(linkID, req.LinkType, clerkUserID)
	if err != nil {
		log.Error().Err(err).Str("link_id", linkID).Msg("Failed to update note link")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, link)
}

// GetAllLinks retrieves all links with optional organization filter
func GetAllLinks(c *gin.Context) {
	// Initialize service (lazy initialization to ensure DB is ready)
	noteLinkService := services.NewNoteLinkService()
	_, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var organizationID *string
	if orgID, exists := middleware.GetOrganizationID(c); exists && orgID != "" {
		organizationID = &orgID
	}

	links, err := noteLinkService.GetAllLinks(organizationID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get all note links")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, links)
}
