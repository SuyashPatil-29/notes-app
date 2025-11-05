package services

import (
	"backend/db"
	"backend/internal/models"
	"backend/internal/models/dto"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type NoteLinkService struct {
	db *gorm.DB
}

// NewNoteLinkService creates a new note link service instance
func NewNoteLinkService() *NoteLinkService {
	return &NoteLinkService{
		db: db.DB,
	}
}

// CreateNoteLink creates a bidirectional link between two notes
func (s *NoteLinkService) CreateNoteLink(sourceNoteID, targetNoteID, linkType, createdBy string, organizationID *string) (*models.NoteLink, error) {
	// Validation
	if sourceNoteID == "" || targetNoteID == "" {
		return nil, fmt.Errorf("source and target note IDs are required")
	}

	if sourceNoteID == targetNoteID {
		return nil, fmt.Errorf("cannot create a link from a note to itself")
	}

	// Use default link type if not specified
	if linkType == "" {
		linkType = models.LinkTypeReferences
	}

	// Validate link type
	validTypes := models.ValidLinkTypes()
	isValid := false
	for _, vt := range validTypes {
		if vt == linkType {
			isValid = true
			break
		}
	}
	if !isValid {
		return nil, fmt.Errorf("invalid link type: %s", linkType)
	}

	// Check if both notes exist and are accessible
	var sourceNote, targetNote models.Notes
	if err := s.db.Where("id = ?", sourceNoteID).First(&sourceNote).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("source note not found")
		}
		return nil, err
	}

	if err := s.db.Where("id = ?", targetNoteID).First(&targetNote).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("target note not found")
		}
		return nil, err
	}

	// Check for duplicate link
	var existing models.NoteLink
	err := s.db.Where("source_note_id = ? AND target_note_id = ? AND link_type = ?",
		sourceNoteID, targetNoteID, linkType).First(&existing).Error

	if err == nil {
		// Link already exists
		return &existing, nil
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Create the link
	link := &models.NoteLink{
		ID:             uuid.New().String(),
		SourceNoteID:   sourceNoteID,
		TargetNoteID:   targetNoteID,
		LinkType:       linkType,
		CreatedBy:      createdBy,
		OrganizationID: organizationID,
	}

	if err := s.db.Create(link).Error; err != nil {
		log.Error().
			Err(err).
			Str("source_note_id", sourceNoteID).
			Str("target_note_id", targetNoteID).
			Msg("Failed to create note link")
		return nil, err
	}

	orgIDStr := "nil"
	if organizationID != nil {
		orgIDStr = *organizationID
	}

	log.Info().
		Str("link_id", link.ID).
		Str("source_note_id", sourceNoteID).
		Str("target_note_id", targetNoteID).
		Str("link_type", linkType).
		Str("organization_id", orgIDStr).
		Msg("Note link created successfully")

	return link, nil
}

// DeleteNoteLink deletes a note link by ID
func (s *NoteLinkService) DeleteNoteLink(linkID, userID string) error {
	if linkID == "" {
		return fmt.Errorf("link ID is required")
	}

	var link models.NoteLink
	if err := s.db.Where("id = ?", linkID).First(&link).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("note link not found")
		}
		return err
	}

	// Delete the link
	if err := s.db.Delete(&link).Error; err != nil {
		log.Error().
			Err(err).
			Str("link_id", linkID).
			Msg("Failed to delete note link")
		return err
	}

	log.Info().
		Str("link_id", linkID).
		Str("user_id", userID).
		Msg("Note link deleted successfully")

	return nil
}

// GetNoteLinksByNoteID retrieves all links for a specific note
func (s *NoteLinkService) GetNoteLinksByNoteID(noteID string) ([]models.NoteLink, error) {
	if noteID == "" {
		return nil, fmt.Errorf("note ID is required")
	}

	var links []models.NoteLink
	err := s.db.
		Preload("SourceNote").
		Preload("TargetNote").
		Where("source_note_id = ? OR target_note_id = ?", noteID, noteID).
		Find(&links).Error

	if err != nil {
		log.Error().
			Err(err).
			Str("note_id", noteID).
			Msg("Failed to get note links")
		return nil, err
	}

	return links, nil
}

// UpdateNoteLink updates a note link's type
func (s *NoteLinkService) UpdateNoteLink(linkID, linkType, userID string) (*models.NoteLink, error) {
	if linkID == "" {
		return nil, fmt.Errorf("link ID is required")
	}

	// Validate link type
	validTypes := models.ValidLinkTypes()
	isValid := false
	for _, vt := range validTypes {
		if vt == linkType {
			isValid = true
			break
		}
	}
	if !isValid {
		return nil, fmt.Errorf("invalid link type: %s", linkType)
	}

	var link models.NoteLink
	if err := s.db.Where("id = ?", linkID).First(&link).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("note link not found")
		}
		return nil, err
	}

	// Update the link type
	link.LinkType = linkType
	if err := s.db.Save(&link).Error; err != nil {
		log.Error().
			Err(err).
			Str("link_id", linkID).
			Msg("Failed to update note link")
		return nil, err
	}

	log.Info().
		Str("link_id", linkID).
		Str("new_link_type", linkType).
		Msg("Note link updated successfully")

	return &link, nil
}

// GetAllLinks retrieves all links with optional organization filter
func (s *NoteLinkService) GetAllLinks(organizationID *string) ([]models.NoteLink, error) {
	var links []models.NoteLink
	query := s.db.Preload("SourceNote").Preload("TargetNote")

	if organizationID != nil {
		query = query.Where("organization_id = ?", *organizationID)
	}

	err := query.Find(&links).Error
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to get all note links")
		return nil, err
	}

	return links, nil
}

// GetGraphData generates graph visualization data
func (s *NoteLinkService) GetGraphData(organizationID *string, searchQuery string) (*dto.GraphData, error) {
	log.Info().
		Interface("organizationID", organizationID).
		Str("searchQuery", searchQuery).
		Msg("GetGraphData called")

	// Get all links for the organization
	links, err := s.GetAllLinks(organizationID)
	if err != nil {
		return nil, err
	}

	log.Info().
		Int("linkCount", len(links)).
		Msg("Fetched links for graph")

	// Collect unique note IDs
	noteIDMap := make(map[string]bool)
	for _, link := range links {
		noteIDMap[link.SourceNoteID] = true
		noteIDMap[link.TargetNoteID] = true
	}

	// Fetch all notes with their notebooks and chapters
	noteIDs := make([]string, 0, len(noteIDMap))
	for id := range noteIDMap {
		noteIDs = append(noteIDs, id)
	}

	log.Info().
		Int("uniqueNoteCount", len(noteIDs)).
		Strs("noteIDs", noteIDs).
		Msg("Collected note IDs from links")

	var notes []models.Notes
	query := s.db.Preload("Chapter.Notebook").Preload("Chapter")

	if organizationID != nil {
		query = query.Where("organization_id = ?", *organizationID)
	}

	if searchQuery != "" {
		query = query.Where("name LIKE ? OR content LIKE ?",
			"%"+searchQuery+"%", "%"+searchQuery+"%")
	}

	if len(noteIDs) > 0 {
		err = query.Where("id IN ?", noteIDs).Find(&notes).Error
	} else {
		log.Warn().Msg("No note IDs to fetch - returning empty graph")
		return &dto.GraphData{Nodes: []dto.GraphNode{}, Links: []dto.GraphLink{}}, nil
	}

	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to fetch notes for graph")
		return nil, err
	}

	log.Info().
		Int("notesFound", len(notes)).
		Msg("Fetched notes for graph")

	// Count links for each note
	linkCountMap := make(map[string]int)
	for _, link := range links {
		linkCountMap[link.SourceNoteID]++
		linkCountMap[link.TargetNoteID]++
	}

	// Build graph nodes
	graphNodes := make([]dto.GraphNode, 0, len(notes))
	for _, note := range notes {
		node := dto.GraphNode{
			ID:        note.ID,
			Name:      note.Name,
			CreatedAt: note.CreatedAt,
			UpdatedAt: note.UpdatedAt,
			LinkCount: linkCountMap[note.ID],
			Metadata:  make(map[string]string),
		}

		// Access notebook through chapter relationship
		if note.Chapter.Notebook.Name != "" {
			node.NotebookName = note.Chapter.Notebook.Name
		}

		// Access chapter name
		if note.Chapter.Name != "" {
			node.ChapterName = note.Chapter.Name
		}

		graphNodes = append(graphNodes, node)
	}

	// Build graph links
	graphLinks := make([]dto.GraphLink, 0, len(links))
	for _, link := range links {
		graphLinks = append(graphLinks, dto.GraphLink{
			ID:       link.ID,
			Source:   link.SourceNoteID,
			Target:   link.TargetNoteID,
			LinkType: link.LinkType,
		})
	}

	return &dto.GraphData{
		Nodes: graphNodes,
		Links: graphLinks,
	}, nil
}
