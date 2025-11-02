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

	// Get all notebooks with chapters and notes preloaded
	if err := query.Preload("Chapters.Files").Preload("Chapters").Find(&notebooks).Error; err != nil {
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

	// Find notebook first
	if err := db.DB.Where("id = ?", id).First(&notebook).Error; err != nil {
		log.Print("Notebook not found with id: ", id, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Notebook not found"})
		return
	}

	// Check authorization: personal notebook or org member
	if notebook.OrganizationID != nil && *notebook.OrganizationID != "" {
		// Organization notebook - verify membership
		_, isMember, err := middleware.GetOrgMemberRole(c.Request.Context(), *notebook.OrganizationID, clerkUserID)
		if err != nil || !isMember {
			log.Warn().Str("notebook_id", id).Str("org_id", *notebook.OrganizationID).Str("user_id", clerkUserID).Msg("User not authorized for org notebook")
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to access this notebook"})
			return
		}
	} else {
		// Personal notebook - verify ownership
		if notebook.ClerkUserID != clerkUserID {
			log.Warn().Str("notebook_id", id).Str("user_id", clerkUserID).Msg("User not authorized for personal notebook")
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to access this notebook"})
			return
		}
	}

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

	var notebook models.Notebook
	id := c.Param("id")

	// Find notebook first
	if err := db.DB.Where("id = ?", id).First(&notebook).Error; err != nil {
		log.Print("Notebook not found with id: ", id, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Notebook not found"})
		return
	}

	// Check authorization
	if notebook.OrganizationID != nil && *notebook.OrganizationID != "" {
		// Organization notebook - any member can delete (for now - could restrict to admins)
		_, isMember, err := middleware.GetOrgMemberRole(c.Request.Context(), *notebook.OrganizationID, clerkUserID)
		if err != nil || !isMember {
			log.Warn().Str("notebook_id", id).Str("org_id", *notebook.OrganizationID).Str("user_id", clerkUserID).Msg("User not authorized to delete org notebook")
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to delete this notebook"})
			return
		}
	} else {
		// Personal notebook - verify ownership
		if notebook.ClerkUserID != clerkUserID {
			log.Warn().Str("notebook_id", id).Str("user_id", clerkUserID).Msg("User not authorized to delete personal notebook")
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to delete this notebook"})
			return
		}
	}

	if err := db.DB.Delete(&notebook).Error; err != nil {
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
	var notebook models.Notebook

	// Find notebook first
	if err := db.DB.Where("id = ?", id).First(&notebook).Error; err != nil {
		log.Print("Notebook not found with id: ", id, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Notebook not found"})
		return
	}

	// Check authorization
	if notebook.OrganizationID != nil && *notebook.OrganizationID != "" {
		// Organization notebook - any member can edit
		_, isMember, err := middleware.GetOrgMemberRole(c.Request.Context(), *notebook.OrganizationID, clerkUserID)
		if err != nil || !isMember {
			log.Warn().Str("notebook_id", id).Str("org_id", *notebook.OrganizationID).Str("user_id", clerkUserID).Msg("User not authorized to update org notebook")
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to update this notebook"})
			return
		}
	} else {
		// Personal notebook - verify ownership
		if notebook.ClerkUserID != clerkUserID {
			log.Warn().Str("notebook_id", id).Str("user_id", clerkUserID).Msg("User not authorized to update personal notebook")
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to update this notebook"})
			return
		}
	}

	// Bind the update data from request body
	var updateData models.Notebook
	if err := c.ShouldBindJSON(&updateData); err != nil {
		log.Print("Invalid update data for notebook: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
