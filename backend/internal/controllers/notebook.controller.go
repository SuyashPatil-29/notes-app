package controllers

import (
	"backend/db"
	"backend/internal/middleware"
	"backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func GetUserNotebooks(c *gin.Context) {
	// Get authenticated user ID from Clerk
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var notebooks []models.Notebook

	// Get all notebooks for the authenticated user with chapters and notes preloaded
	if err := db.DB.Where("clerk_user_id = ?", clerkUserID).Preload("Chapters.Files").Preload("Chapters").Find(&notebooks).Error; err != nil {
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

	// Find notebook and verify ownership
	if err := db.DB.Where("id = ? AND clerk_user_id = ?", id, clerkUserID).First(&notebook).Error; err != nil {
		log.Print("Notebook not found with id: ", id, " for user: ", clerkUserID, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Notebook not found"})
		return
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

	// Find notebook and verify ownership
	if err := db.DB.Where("id = ? AND clerk_user_id = ?", id, clerkUserID).First(&notebook).Error; err != nil {
		log.Print("Notebook not found with id: ", id, " for user: ", clerkUserID, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Notebook not found"})
		return
	}

	if err := db.DB.Delete(&notebook, id).Error; err != nil {
		log.Print("Error deleting Notebook with id: ", id, "Error: ", err)
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

	// Check if notebook exists and verify ownership
	if err := db.DB.Where("id = ? AND clerk_user_id = ?", id, clerkUserID).First(&notebook).Error; err != nil {
		log.Print("Notebook not found with id: ", id, " for user: ", clerkUserID, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Notebook not found"})
		return
	}

	// Bind the update data from request body
	var updateData models.Notebook
	if err := c.ShouldBindJSON(&updateData); err != nil {
		log.Print("Invalid update data for notebook: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Prevent changing clerk_user_id through update
	updateData.ClerkUserID = notebook.ClerkUserID

	// Update the notebook
	if err := db.DB.Model(&notebook).Updates(updateData).Error; err != nil {
		log.Print("Error updating Notebook with id: ", id, " Error: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, notebook)
}
