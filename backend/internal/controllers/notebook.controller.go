package controllers

import (
	"backend/db"
	"backend/internal/middleware"
	"backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func GetUserNotebooks(c *gin.Context) {
	// Get authenticated user ID from Clerk
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var notebooks []models.Notebook

	// Get organization ID from query param (optional)
	orgID := c.Query("organizationId")

	var query *gorm.DB

	if orgID != "" {
		// Get organization notebooks - verify membership first
		role, isMember, err := middleware.GetOrgMemberRole(c.Request.Context(), orgID, clerkUserID)
		if err != nil || !isMember {
			log.Warn().Str("org_id", orgID).Str("user_id", clerkUserID).Msg("User not authorized for org notebooks")
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not a member of this organization"})
			return
		}
		log.Debug().Str("org_id", orgID).Str("user_id", clerkUserID).Str("role", role).Msg("Fetching org notebooks")

		// For org notebooks, fetch ALL notebooks in the org (not just user's own)
		query = db.DB.Where("organization_id = ?", orgID)
	} else {
		// Get personal notebooks only (null organization_id, owned by this user)
		query = db.DB.Where("clerk_user_id = ? AND organization_id IS NULL", clerkUserID)
	}

	// Get notebooks with chapters and notes, but exclude large content fields from notes
	if err := query.
		Preload("Chapters", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Preload("Chapters.Files", func(db *gorm.DB) *gorm.DB {
			// Only select metadata fields, exclude large content fields
			return db.Select("id, name, chapter_id, organization_id, is_public, has_video, meeting_recording_id, created_at, updated_at").
				Order("created_at DESC")
		}).
		Find(&notebooks).Error; err != nil {
		log.Print("Error fetching notebooks for user: ", clerkUserID, " Error: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, notebooks)
}

func GetNotebookById(c *gin.Context) {
	// Get authenticated user ID from Clerk
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var notebook models.Notebook
	id := c.Param("id")

	// Find notebook without preloads
	if err := db.DB.Where("id = ?", id).First(&notebook).Error; err != nil {
		log.Print("Notebook not found with id: ", id, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Notebook not found"})
		return
	}

	// Check authorization efficiently
	hasAccess, err := middleware.CheckNotebookAccess(c.Request.Context(), db.DB, id, clerkUserID)
	if err != nil || !hasAccess {
		log.Warn().Str("notebook_id", id).Str("user_id", clerkUserID).Msg("User not authorized for notebook")
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to access this notebook"})
		return
	}

	// Load chapters without notes
	var chapters []models.Chapter
	db.DB.Where("notebook_id = ?", id).Find(&chapters)
	notebook.Chapters = chapters

	c.JSON(http.StatusOK, notebook)
}

func CreateNotebook(c *gin.Context) {
	// Get authenticated user ID from Clerk
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var notebook models.Notebook

	if err := c.ShouldBindJSON(&notebook); err != nil {
		log.Print("Missing data to create a notebook", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set the Clerk user ID from authenticated session (security)
	notebook.ClerkUserID = clerkUserID

	// If organizationId is provided, verify membership
	if notebook.OrganizationID != nil && *notebook.OrganizationID != "" {
		_, isMember, err := middleware.GetOrgMemberRole(c.Request.Context(), *notebook.OrganizationID, clerkUserID)
		if err != nil || !isMember {
			log.Warn().Str("org_id", *notebook.OrganizationID).Str("user_id", clerkUserID).Msg("User not authorized to create notebook in org")
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not a member of this organization"})
			return
		}
		log.Info().Str("org_id", *notebook.OrganizationID).Str("user_id", clerkUserID).Msg("Creating org notebook")
	}

	if err := db.DB.Create(&notebook).Error; err != nil {
		log.Print("Error creating Notebook in db: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, notebook)
}

func DeleteNotebook(c *gin.Context) {
	// Get authenticated user ID from Clerk
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id := c.Param("id")

	// Check authorization efficiently
	hasAccess, err := middleware.CheckNotebookAccess(c.Request.Context(), db.DB, id, clerkUserID)
	if err != nil {
		log.Print("Notebook not found with id: ", id, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Notebook not found"})
		return
	}
	if !hasAccess {
		log.Warn().Str("notebook_id", id).Str("user_id", clerkUserID).Msg("User not authorized to delete notebook")
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to delete this notebook"})
		return
	}

	// Delete the notebook
	if err := db.DB.Delete(&models.Notebook{}, "id = ?", id).Error; err != nil {
		log.Print("Error deleting Notebook with id: ", id, " Error: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notebook deleted successfully"})
}

func UpdateNotebook(c *gin.Context) {
	// Get authenticated user ID from Clerk
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id := c.Param("id")

	// Check authorization efficiently
	hasAccess, err := middleware.CheckNotebookAccess(c.Request.Context(), db.DB, id, clerkUserID)
	if err != nil {
		log.Print("Notebook not found with id: ", id, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Notebook not found"})
		return
	}
	if !hasAccess {
		log.Warn().Str("notebook_id", id).Str("user_id", clerkUserID).Msg("User not authorized to update notebook")
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to update this notebook"})
		return
	}

	// Bind the update data from request body
	var updateData models.Notebook
	if err := c.ShouldBindJSON(&updateData); err != nil {
		log.Print("Invalid update data for notebook: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current notebook to preserve protected fields
	var notebook models.Notebook
	if err := db.DB.Where("id = ?", id).First(&notebook).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notebook"})
		return
	}

	// Prevent changing clerk_user_id and organization_id through update
	updateData.ClerkUserID = notebook.ClerkUserID
	updateData.OrganizationID = notebook.OrganizationID

	// Update the notebook
	if err := db.DB.Model(&notebook).Updates(updateData).Error; err != nil {
		log.Print("Error updating Notebook with id: ", id, " Error: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, notebook)
}
