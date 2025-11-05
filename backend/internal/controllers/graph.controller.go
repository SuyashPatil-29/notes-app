package controllers

import (
	"backend/internal/middleware"
	"backend/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// GetGraphData retrieves graph visualization data
func GetGraphData(c *gin.Context) {
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	log.Info().Str("user_id", clerkUserID).Msg("Fetching graph data")

	var organizationID *string
	if orgID, exists := middleware.GetOrganizationID(c); exists && orgID != "" {
		organizationID = &orgID
	}

	// Get optional search query
	searchQuery := c.Query("q")

	// Initialize service (lazy initialization to ensure DB is ready)
	graphService := services.NewNoteLinkService()
	graphData, err := graphService.GetGraphData(organizationID, searchQuery)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get graph data")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Info().
		Int("node_count", len(graphData.Nodes)).
		Int("link_count", len(graphData.Links)).
		Msg("Graph data retrieved successfully")

	c.JSON(http.StatusOK, graphData)
}
