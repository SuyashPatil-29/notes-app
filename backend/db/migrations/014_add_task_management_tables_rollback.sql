-- Rollback migration: Remove task management tables
-- This rollback script removes the task management functionality

-- Drop foreign key constraints first
ALTER TABLE tasks DROP CONSTRAINT IF EXISTS fk_tasks_task_board_id;

-- Drop check constraints
ALTER TABLE tasks DROP CONSTRAINT IF EXISTS chk_tasks_status;
ALTER TABLE tasks DROP CONSTRAINT IF EXISTS chk_tasks_priority;

-- Drop indexes
DROP INDEX IF EXISTS idx_task_boards_note_id;
DROP INDEX IF EXISTS idx_task_boards_clerk_user_id;
DROP INDEX IF EXISTS idx_task_boards_organization_id;
DROP INDEX IF EXISTS idx_task_boards_is_standalone;

DROP INDEX IF EXISTS idx_tasks_task_board_id;
DROP INDEX IF EXISTS idx_tasks_status;
DROP INDEX IF EXISTS idx_tasks_priority;
DROP INDEX IF EXISTS idx_tasks_organization_id;
DROP INDEX IF EXISTS idx_tasks_position;

-- Drop tables
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS task_boards;