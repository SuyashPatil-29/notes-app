package controllers

import (
	"backend/db"
	"backend/internal/middleware"
	"backend/internal/models"
	"backend/internal/services"
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// CheckTaskBoardAccess is a wrapper to use the middleware function
func CheckTaskBoardAccess(ctx context.Context, db *gorm.DB, taskBoardID, clerkUserID string) (bool, error) {
	return middleware.CheckTaskBoardAccess(ctx, db, taskBoardID, clerkUserID)
}

// CheckTaskAccess is a wrapper to use the middleware function
func CheckTaskAccess(ctx context.Context, db *gorm.DB, taskID, clerkUserID string) (bool, error) {
	return middleware.CheckTaskAccess(ctx, db, taskID, clerkUserID)
}

// GetTaskBoard retrieves a task board with its tasks
func GetTaskBoard(c *gin.Context) {
	// Get authenticated user ID
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	boardID := c.Param("boardId")

	// Check authorization for task board
	hasAccess, err := CheckTaskBoardAccess(c.Request.Context(), db.DB, boardID, clerkUserID)
	if err != nil {
		log.Print("Task board not found with id: ", boardID, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Task board not found"})
		return
	}
	if !hasAccess {
		log.Warn().Str("board_id", boardID).Str("user_id", clerkUserID).Msg("User not authorized to access task board")
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	// Get task board with tasks and note relations
	var taskBoard models.TaskBoard
	if err := db.DB.
		Preload("Tasks.Assignments").
		Preload("Note").
		Preload("Note.Chapter").
		Preload("Note.Chapter.Notebook").
		Where("id = ?", boardID).
		First(&taskBoard).Error; err != nil {
		log.Print("Error fetching task board: ", boardID, " Error: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch task board"})
		return
	}

	c.JSON(http.StatusOK, taskBoard)
}

// CreateTaskBoard creates a new task board
func CreateTaskBoard(c *gin.Context) {
	// Get authenticated user ID
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var taskBoard models.TaskBoard
	if err := c.ShouldBindJSON(&taskBoard); err != nil {
		log.Print("Invalid task board data: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set the user ID
	taskBoard.ClerkUserID = clerkUserID

	// If this is a note-associated board, verify note access
	if taskBoard.NoteID != nil && *taskBoard.NoteID != "" {
		hasAccess, err := middleware.CheckNoteAccess(c.Request.Context(), db.DB, *taskBoard.NoteID, clerkUserID)
		if err != nil {
			log.Print("Note not found with id: ", *taskBoard.NoteID, " Error: ", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
			return
		}
		if !hasAccess {
			log.Warn().Str("note_id", *taskBoard.NoteID).Str("user_id", clerkUserID).Msg("User not authorized to create task board for note")
			c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
			return
		}

		// Get organization ID from note
		var note models.Notes
		if err := db.DB.Select("organization_id").Where("id = ?", *taskBoard.NoteID).First(&note).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task board"})
			return
		}
		taskBoard.OrganizationID = note.OrganizationID
		taskBoard.IsStandalone = false
	} else {
		// Standalone board - get organization from context if available
		if orgID := c.GetString("organization_id"); orgID != "" {
			taskBoard.OrganizationID = &orgID
		}
		taskBoard.IsStandalone = true
	}

	// Create the task board
	if err := db.DB.Create(&taskBoard).Error; err != nil {
		log.Print("Error creating task board: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task board"})
		return
	}

	c.JSON(http.StatusCreated, taskBoard)
}

// UpdateTaskBoard updates an existing task board
func UpdateTaskBoard(c *gin.Context) {
	// Get authenticated user ID
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	boardID := c.Param("boardId")

	// Check authorization for task board
	hasAccess, err := CheckTaskBoardAccess(c.Request.Context(), db.DB, boardID, clerkUserID)
	if err != nil {
		log.Print("Task board not found with id: ", boardID, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Task board not found"})
		return
	}
	if !hasAccess {
		log.Warn().Str("board_id", boardID).Str("user_id", clerkUserID).Msg("User not authorized to update task board")
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	// Bind update data
	var updateData models.TaskBoard
	if err := c.ShouldBindJSON(&updateData); err != nil {
		log.Print("Invalid update data for task board: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current task board to preserve protected fields
	var taskBoard models.TaskBoard
	if err := db.DB.Where("id = ?", boardID).First(&taskBoard).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task board"})
		return
	}

	// Prevent changing protected fields
	updateData.ClerkUserID = taskBoard.ClerkUserID
	updateData.OrganizationID = taskBoard.OrganizationID
	updateData.NoteID = taskBoard.NoteID
	updateData.IsStandalone = taskBoard.IsStandalone

	// Update the task board
	if err := db.DB.Model(&taskBoard).Updates(updateData).Error; err != nil {
		log.Print("Error updating task board: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task board"})
		return
	}

	c.JSON(http.StatusOK, taskBoard)
}

// DeleteTaskBoard deletes a task board and all its tasks
func DeleteTaskBoard(c *gin.Context) {
	// Get authenticated user ID
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	boardID := c.Param("boardId")

	// Check authorization for task board
	hasAccess, err := CheckTaskBoardAccess(c.Request.Context(), db.DB, boardID, clerkUserID)
	if err != nil {
		log.Print("Task board not found with id: ", boardID, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Task board not found"})
		return
	}
	if !hasAccess {
		log.Warn().Str("board_id", boardID).Str("user_id", clerkUserID).Msg("User not authorized to delete task board")
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	// Delete the task board (tasks will be deleted by cascade)
	if err := db.DB.Delete(&models.TaskBoard{}, "id = ?", boardID).Error; err != nil {
		log.Print("Error deleting task board: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task board"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task board deleted successfully"})
}

// GetUserTaskBoards retrieves all task boards for a user
func GetUserTaskBoards(c *gin.Context) {
	// Get authenticated user ID
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse pagination params
	page := 1
	pageSize := 50
	if pageParam := c.Query("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}
	if pageSizeParam := c.Query("page_size"); pageSizeParam != "" {
		if ps, err := strconv.Atoi(pageSizeParam); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	// Get organization ID from query parameter
	orgID := c.Query("organizationId")

	// Build query
	var query *gorm.DB

	// Filter by organization context
	if orgID != "" {
		// Verify organization membership before returning org task boards
		_, isMember, err := middleware.GetOrgMemberRole(c.Request.Context(), orgID, clerkUserID)
		if err != nil || !isMember {
			log.Warn().Str("org_id", orgID).Str("user_id", clerkUserID).Msg("User not authorized for org task boards")
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not a member of this organization"})
			return
		}
		// In organization context, return ALL task boards for the organization
		query = db.DB.Model(&models.TaskBoard{}).Where("task_boards.organization_id = ?", orgID)
	} else {
		// In personal context, return only user's personal task boards
		query = db.DB.Model(&models.TaskBoard{}).Where("clerk_user_id = ? AND task_boards.organization_id IS NULL", clerkUserID)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get task boards with task count
	var taskBoards []struct {
		models.TaskBoard
		TaskCount int `json:"taskCount"`
	}

	offset := (page - 1) * pageSize
	if err := query.
		Select("task_boards.*, COUNT(tasks.id) as task_count").
		Joins("LEFT JOIN tasks ON tasks.task_board_id = task_boards.id").
		Group("task_boards.id").
		Order("task_boards.created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&taskBoards).Error; err != nil {
		log.Print("Error fetching user task boards: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch task boards"})
		return
	}

	// Calculate pagination metadata
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	if totalPages == 0 {
		totalPages = 1
	}

	response := gin.H{
		"data":       taskBoards,
		"page":       page,
		"pageSize":   pageSize,
		"total":      total,
		"totalPages": totalPages,
		"hasNext":    page < totalPages,
		"hasPrev":    page > 1,
	}

	c.JSON(http.StatusOK, response)
}

// CreateTask creates a new task in a task board
func CreateTask(c *gin.Context) {
	// Get authenticated user ID
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	boardID := c.Param("boardId")

	// Check authorization for task board
	hasAccess, err := CheckTaskBoardAccess(c.Request.Context(), db.DB, boardID, clerkUserID)
	if err != nil {
		log.Print("Task board not found with id: ", boardID, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Task board not found"})
		return
	}
	if !hasAccess {
		log.Warn().Str("board_id", boardID).Str("user_id", clerkUserID).Msg("User not authorized to create task in board")
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		log.Print("Invalid task data: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set the task board ID
	task.TaskBoardID = boardID

	// Get organization ID from task board
	var taskBoard models.TaskBoard
	if err := db.DB.Select("organization_id").Where("id = ?", boardID).First(&taskBoard).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}
	task.OrganizationID = taskBoard.OrganizationID

	// Create the task
	if err := db.DB.Create(&task).Error; err != nil {
		log.Print("Error creating task: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// UpdateTask updates an existing task
func UpdateTask(c *gin.Context) {
	// Get authenticated user ID
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	taskID := c.Param("taskId")

	// Check authorization for task
	hasAccess, err := CheckTaskAccess(c.Request.Context(), db.DB, taskID, clerkUserID)
	if err != nil {
		log.Print("Task not found with id: ", taskID, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	if !hasAccess {
		log.Warn().Str("task_id", taskID).Str("user_id", clerkUserID).Msg("User not authorized to update task")
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	// Bind update data
	var updateData models.Task
	if err := c.ShouldBindJSON(&updateData); err != nil {
		log.Print("Invalid update data for task: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current task to preserve protected fields
	var task models.Task
	if err := db.DB.Where("id = ?", taskID).First(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	// Prevent changing protected fields
	updateData.TaskBoardID = task.TaskBoardID
	updateData.OrganizationID = task.OrganizationID

	// Update the task
	if err := db.DB.Model(&task).Updates(updateData).Error; err != nil {
		log.Print("Error updating task: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// DeleteTask deletes a task
func DeleteTask(c *gin.Context) {
	// Get authenticated user ID
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	taskID := c.Param("taskId")

	// Check authorization for task
	hasAccess, err := CheckTaskAccess(c.Request.Context(), db.DB, taskID, clerkUserID)
	if err != nil {
		log.Print("Task not found with id: ", taskID, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	if !hasAccess {
		log.Warn().Str("task_id", taskID).Str("user_id", clerkUserID).Msg("User not authorized to delete task")
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	// Delete the task
	if err := db.DB.Delete(&models.Task{}, "id = ?", taskID).Error; err != nil {
		log.Print("Error deleting task: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

// GetTasksForNote retrieves tasks associated with a note
func GetTasksForNote(c *gin.Context) {
	// Get authenticated user ID
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	noteID := c.Param("noteId")

	// Check authorization for note
	hasAccess, err := middleware.CheckNoteAccess(c.Request.Context(), db.DB, noteID, clerkUserID)
	if err != nil {
		log.Print("Note not found with id: ", noteID, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}
	if !hasAccess {
		log.Warn().Str("note_id", noteID).Str("user_id", clerkUserID).Msg("User not authorized to access note tasks")
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	// Get task board for note
	var taskBoard models.TaskBoard
	if err := db.DB.Preload("Tasks.Assignments").Where("note_id = ?", noteID).First(&taskBoard).Error; err != nil {
		// No task board exists for this note yet
		c.JSON(http.StatusOK, gin.H{"taskBoard": nil, "tasks": []models.Task{}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"taskBoard": taskBoard, "tasks": taskBoard.Tasks})
}

// GenerateTasksFromNote generates tasks from note content using AI
func GenerateTasksFromNote(c *gin.Context) {
	// Get authenticated user ID
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	noteID := c.Param("noteId")

	// Check authorization for note
	hasAccess, err := middleware.CheckNoteAccess(c.Request.Context(), db.DB, noteID, clerkUserID)
	if err != nil {
		log.Print("Note not found with id: ", noteID, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}
	if !hasAccess {
		log.Warn().Str("note_id", noteID).Str("user_id", clerkUserID).Msg("User not authorized to generate tasks for note")
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	// Check if task board already exists for this note
	var existingBoard models.TaskBoard
	if err := db.DB.Where("note_id = ?", noteID).First(&existingBoard).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Task board already exists for this note", "taskBoardId": existingBoard.ID})
		return
	}

	// Get the note content
	var note models.Notes
	if err := db.DB.Where("id = ?", noteID).First(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch note"})
		return
	}

	// Initialize AI service
	aiService := services.NewAIService()

	// Prepare task generation request
	taskRequest := services.TaskGenerationRequest{
		NoteTitle:   note.Name,
		NoteContent: note.Content,
		UserID:      clerkUserID,
		OrgID:       note.OrganizationID,
	}

	// Generate tasks using AI service
	taskResponse, err := aiService.GenerateTasksFromNote(c.Request.Context(), taskRequest)
	if err != nil {
		log.Error().Err(err).Str("note_id", noteID).Msg("Failed to generate tasks from note")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tasks from note content"})
		return
	}

	// Create task board with AI-generated name
	taskBoard := models.TaskBoard{
		Name:           taskResponse.BoardName,
		Description:    "AI-generated tasks from note content",
		NoteID:         &noteID,
		ClerkUserID:    clerkUserID,
		OrganizationID: note.OrganizationID,
		IsStandalone:   false,
	}

	// Start database transaction
	tx := db.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create the task board
	if err := tx.Create(&taskBoard).Error; err != nil {
		tx.Rollback()
		log.Error().Err(err).Str("note_id", noteID).Msg("Error creating task board")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task board"})
		return
	}

	// Create AI-generated tasks
	var createdTasks []models.Task
	for i, aiTask := range taskResponse.Tasks {
		task := models.Task{
			Title:          aiTask.Title,
			Description:    aiTask.Description,
			Status:         aiTask.Status,
			Priority:       aiTask.Priority,
			TaskBoardID:    taskBoard.ID,
			Position:       i + 1,
			OrganizationID: note.OrganizationID,
		}

		if err := tx.Create(&task).Error; err != nil {
			log.Error().Err(err).Str("task_title", task.Title).Msg("Error creating task")
			// Continue creating other tasks even if one fails
			continue
		}

		createdTasks = append(createdTasks, task)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		log.Error().Err(err).Str("note_id", noteID).Msg("Error committing task generation transaction")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save generated tasks"})
		return
	}

	// Set tasks on task board for response
	taskBoard.Tasks = createdTasks

	log.Info().
		Str("note_id", noteID).
		Str("task_board_id", taskBoard.ID).
		Int("tasks_created", len(createdTasks)).
		Msg("Successfully generated tasks from note content")

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Tasks generated successfully",
		"taskBoard": taskBoard,
	})
}

// AssignTaskToUsers assigns one or more users to a task
func AssignTaskToUsers(c *gin.Context) {
	// Get authenticated user ID
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	taskID := c.Param("taskId")

	// Check authorization for task
	hasAccess, err := CheckTaskAccess(c.Request.Context(), db.DB, taskID, clerkUserID)
	if err != nil {
		log.Print("Task not found with id: ", taskID, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	if !hasAccess {
		log.Warn().Str("task_id", taskID).Str("user_id", clerkUserID).Msg("User not authorized to assign task")
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse request body
	var assignmentRequest struct {
		UserIDs []string `json:"userIds" binding:"required"`
	}
	if err := c.ShouldBindJSON(&assignmentRequest); err != nil {
		log.Print("Invalid assignment data: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get task to verify organization context
	var task models.Task
	if err := db.DB.Where("id = ?", taskID).First(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign task"})
		return
	}

	// Verify all users belong to the same organization (if task has organization)
	if task.OrganizationID != nil {
		for _, userID := range assignmentRequest.UserIDs {
			_, isMember, err := middleware.GetOrgMemberRole(c.Request.Context(), *task.OrganizationID, userID)
			if err != nil || !isMember {
				log.Warn().Str("user_id", userID).Str("org_id", *task.OrganizationID).Msg("User not member of task organization")
				c.JSON(http.StatusBadRequest, gin.H{"error": "One or more users are not members of the task's organization"})
				return
			}
		}
	}

	// Start transaction
	tx := db.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Remove existing assignments for this task
	if err := tx.Where("task_id = ?", taskID).Delete(&models.TaskAssignment{}).Error; err != nil {
		tx.Rollback()
		log.Error().Err(err).Str("task_id", taskID).Msg("Error removing existing task assignments")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task assignments"})
		return
	}

	// Create new assignments
	var assignments []models.TaskAssignment
	for _, userID := range assignmentRequest.UserIDs {
		assignment := models.TaskAssignment{
			TaskID: taskID,
			UserID: userID,
		}
		if err := tx.Create(&assignment).Error; err != nil {
			tx.Rollback()
			log.Error().Err(err).Str("task_id", taskID).Str("user_id", userID).Msg("Error creating task assignment")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign task"})
			return
		}
		assignments = append(assignments, assignment)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		log.Error().Err(err).Str("task_id", taskID).Msg("Error committing task assignment transaction")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign task"})
		return
	}

	log.Info().
		Str("task_id", taskID).
		Int("assignments_created", len(assignments)).
		Msg("Successfully assigned task to users")

	c.JSON(http.StatusOK, gin.H{
		"message":     "Task assigned successfully",
		"assignments": assignments,
	})
}

// UnassignUserFromTask removes a user assignment from a task
func UnassignUserFromTask(c *gin.Context) {
	// Get authenticated user ID
	clerkUserID, exists := middleware.GetClerkUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	taskID := c.Param("taskId")
	userIDToUnassign := c.Param("userId")

	// Check authorization for task
	hasAccess, err := CheckTaskAccess(c.Request.Context(), db.DB, taskID, clerkUserID)
	if err != nil {
		log.Print("Task not found with id: ", taskID, " Error: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	if !hasAccess {
		log.Warn().Str("task_id", taskID).Str("user_id", clerkUserID).Msg("User not authorized to unassign task")
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	// Remove the assignment
	result := db.DB.Where("task_id = ? AND user_id = ?", taskID, userIDToUnassign).Delete(&models.TaskAssignment{})
	if result.Error != nil {
		log.Error().Err(result.Error).Str("task_id", taskID).Str("user_id", userIDToUnassign).Msg("Error removing task assignment")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unassign task"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Assignment not found"})
		return
	}

	log.Info().
		Str("task_id", taskID).
		Str("unassigned_user_id", userIDToUnassign).
		Msg("Successfully unassigned user from task")

	c.JSON(http.StatusOK, gin.H{"message": "User unassigned from task successfully"})
}
