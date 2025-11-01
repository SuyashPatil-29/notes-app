-- Migration 010: Clean up old user_id columns that were left behind
-- These should have been removed in migration 008 but some remained

-- Drop foreign key constraints first (if they exist)
ALTER TABLE notebooks DROP CONSTRAINT IF EXISTS fk_users_notebooks;
ALTER TABLE calendars DROP CONSTRAINT IF EXISTS fk_users_calendars;
ALTER TABLE meeting_recordings DROP CONSTRAINT IF EXISTS fk_meeting_recordings_user;
ALTER TABLE ai_credentials DROP CONSTRAINT IF EXISTS fk_ai_credentials_user;
ALTER TABLE calendar_o_auth_states DROP CONSTRAINT IF EXISTS fk_calendar_o_auth_states_user;

-- Drop old user_id columns
ALTER TABLE notebooks DROP COLUMN IF EXISTS user_id;
ALTER TABLE calendars DROP COLUMN IF EXISTS user_id;
ALTER TABLE meeting_recordings DROP COLUMN IF EXISTS user_id;
ALTER TABLE ai_credentials DROP COLUMN IF EXISTS user_id;
ALTER TABLE calendar_o_auth_states DROP COLUMN IF EXISTS user_id;

-- Drop old indexes
DROP INDEX IF EXISTS idx_notebooks_user_id;
DROP INDEX IF EXISTS idx_calendars_user_id;
DROP INDEX IF EXISTS idx_meeting_recordings_user_id;
DROP INDEX IF EXISTS idx_ai_credentials_user_id;
DROP INDEX IF EXISTS idx_calendar_o_auth_states_user_id;

