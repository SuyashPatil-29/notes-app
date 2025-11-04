-- Migration: Add task management tables
-- This migration adds support for AI-generated tasks and Kanban boards
-- Supports both note-associated tasks and standalone task boards

-- Create task_boards table
CREATE TABLE task_boards (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    note_id VARCHAR(255),
    clerk_user_id VARCHAR(255) NOT NULL,
    organization_id VARCHAR(255),
    is_standalone BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create tasks table
CREATE TABLE tasks (
    id VARCHAR(255) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) DEFAULT 'backlog',
    priority VARCHAR(50) DEFAULT 'medium',
    task_board_id VARCHAR(255) NOT NULL,
    position INTEGER DEFAULT 0,
    organization_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better query performance
CREATE INDEX idx_task_boards_note_id ON task_boards(note_id);
CREATE INDEX idx_task_boards_clerk_user_id ON task_boards(clerk_user_id);
CREATE INDEX idx_task_boards_organization_id ON task_boards(organization_id);
CREATE INDEX idx_task_boards_is_standalone ON task_boards(is_standalone);

CREATE INDEX idx_tasks_task_board_id ON tasks(task_board_id);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_priority ON tasks(priority);
CREATE INDEX idx_tasks_organization_id ON tasks(organization_id);
CREATE INDEX idx_tasks_position ON tasks(position);

-- Add foreign key constraints
ALTER TABLE tasks ADD CONSTRAINT fk_tasks_task_board_id 
    FOREIGN KEY (task_board_id) REFERENCES task_boards(id) ON DELETE CASCADE;

-- Add check constraints for valid values
ALTER TABLE tasks ADD CONSTRAINT chk_tasks_status 
    CHECK (status IN ('backlog', 'todo', 'in_progress', 'done'));

ALTER TABLE tasks ADD CONSTRAINT chk_tasks_priority 
    CHECK (priority IN ('low', 'medium', 'high'));

-- Add comments for documentation
COMMENT ON TABLE task_boards IS 'Stores task boards for both note-associated and standalone Kanban boards';
COMMENT ON COLUMN task_boards.note_id IS 'Optional reference to notes table for note-associated task boards';
COMMENT ON COLUMN task_boards.clerk_user_id IS 'Clerk user ID of the task board owner';
COMMENT ON COLUMN task_boards.organization_id IS 'Optional organization ID for multi-tenant support';
COMMENT ON COLUMN task_boards.is_standalone IS 'True for standalone boards, false for note-associated boards';

COMMENT ON TABLE tasks IS 'Individual tasks within task boards';
COMMENT ON COLUMN tasks.status IS 'Task status: backlog, todo, in_progress, or done';
COMMENT ON COLUMN tasks.priority IS 'Task priority: low, medium, or high';
COMMENT ON COLUMN tasks.position IS 'Position within the task board column for ordering';
COMMENT ON COLUMN tasks.organization_id IS 'Optional organization ID inherited from task board';