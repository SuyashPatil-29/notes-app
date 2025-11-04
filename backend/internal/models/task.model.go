package models

import (
	"time"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

type Task struct {
	ID             string           `json:"id" gorm:"primaryKey;type:varchar(255)"`
	Title          string           `json:"title"`
	Description    string           `json:"description" gorm:"type:text"`
	Status         string           `json:"status" gorm:"default:'backlog'"`  // "backlog", "todo", "in_progress", "done"
	Priority       string           `json:"priority" gorm:"default:'medium'"` // "low", "medium", "high"
	TaskBoardID    string           `json:"taskBoardId" gorm:"type:varchar(255);index"`
	Position       int              `json:"position" gorm:"default:0"`
	OrganizationID *string          `json:"organizationId,omitempty" gorm:"type:varchar(255);index"`
	TaskBoard      TaskBoard        `json:"taskBoard" gorm:"foreignKey:TaskBoardID"`
	Assignments    []TaskAssignment `json:"assignments" gorm:"foreignKey:TaskID"`
	CreatedAt      time.Time        `json:"createdAt"`
	UpdatedAt      time.Time        `json:"updatedAt"`
}

// TaskAssignment represents the many-to-many relationship between tasks and users
type TaskAssignment struct {
	ID        string    `json:"id" gorm:"primaryKey;type:varchar(255)"`
	TaskID    string    `json:"taskId" gorm:"type:varchar(255);index"`
	UserID    string    `json:"userId" gorm:"type:varchar(255);index"`
	Task      Task      `json:"task" gorm:"foreignKey:TaskID"`
	CreatedAt time.Time `json:"createdAt"`
}

// BeforeCreate hook to generate CUID before creating a task
func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = cuid.New()
	}
	return nil
}

// BeforeCreate hook to generate CUID before creating a task assignment
func (ta *TaskAssignment) BeforeCreate(tx *gorm.DB) error {
	if ta.ID == "" {
		ta.ID = cuid.New()
	}
	return nil
}
