package dto

import "time"

// GraphNode represents a node in the graph visualization
type GraphNode struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	ChapterName  string            `json:"chapterName,omitempty"`
	NotebookName string            `json:"notebookName,omitempty"`
	CreatedAt    time.Time         `json:"createdAt"`
	UpdatedAt    time.Time         `json:"updatedAt"`
	LinkCount    int               `json:"linkCount"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// GraphLink represents a link between two nodes in the graph
type GraphLink struct {
	ID       string `json:"id"`
	Source   string `json:"source"`
	Target   string `json:"target"`
	LinkType string `json:"linkType"`
}

// GraphData represents the complete graph structure
type GraphData struct {
	Nodes []GraphNode `json:"nodes"`
	Links []GraphLink `json:"links"`
}

// CreateNoteLinkRequest represents the request body for creating a note link
type CreateNoteLinkRequest struct {
	SourceNoteID string `json:"sourceNoteId" binding:"required"`
	TargetNoteID string `json:"targetNoteId" binding:"required"`
	LinkType     string `json:"linkType"`
}

// UpdateNoteLinkRequest represents the request body for updating a note link
type UpdateNoteLinkRequest struct {
	LinkType string `json:"linkType" binding:"required"`
}

