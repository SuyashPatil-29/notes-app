-- Migration: Add task assignments table
-- This migration adds support for assigning tasks to organization members
-- Supports many-to-many relationships between tasks and users

-- Create task_assignments table
CREATE TABLE task_assignments (
    id VARCHAR(255) PRIMARY KEY,
    task_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better query performance
CREATE INDEX idx_task_assignments_task_id ON task_assignments(task_id);
CREATE INDEX idx_task_assignments_user_id ON task_assignments(user_id);
CREATE INDEX idx_task_assignments_task_user ON task_assignments(task_id, user_id);

-- Add foreign key constraint to tasks table
ALTER TABLE task_assignments ADD CONSTRAINT fk_task_assignments_task_id 
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE;

-- Add unique constraint to prevent duplicate assignments
ALTER TABLE task_assignments ADD CONSTRAINT uk_task_assignments_task_user 
    UNIQUE (task_id, user_id);

-- Add comments for documentation
COMMENT ON TABLE task_assignments IS 'Many-to-many relationship between tasks and assigned users';
COMMENT ON COLUMN task_assignments.task_id IS 'Reference to the assigned task';
COMMENT ON COLUMN task_assignments.user_id IS 'Clerk user ID of the assigned user';