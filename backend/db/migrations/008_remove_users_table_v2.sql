-- Migration 008: Remove users table and use clerk_user_id directly
-- This migration eliminates the users table and uses Clerk as the single source of truth

-- WARNING: This will delete all existing data since we can't migrate without clerk_user_id mapping
-- For production, you would need a proper data migration strategy

-- Step 1: Clear existing data (dev only)
TRUNCATE TABLE notebooks CASCADE;
TRUNCATE TABLE calendars CASCADE;
TRUNCATE TABLE meeting_recordings CASCADE;
TRUNCATE TABLE ai_credentials CASCADE;
TRUNCATE TABLE calendar_o_auth_states CASCADE;

-- Step 2: Drop foreign key constraints
ALTER TABLE notebooks DROP CONSTRAINT IF EXISTS fk_users_notebooks;
ALTER TABLE calendars DROP CONSTRAINT IF EXISTS calendars_user_id_fkey;
ALTER TABLE meeting_recordings DROP CONSTRAINT IF EXISTS fk_meeting_recordings_user;
ALTER TABLE ai_credentials DROP CONSTRAINT IF EXISTS fk_ai_credentials_user;
ALTER TABLE calendar_o_auth_states DROP CONSTRAINT IF EXISTS calendar_o_auth_states_user_id_fkey;

-- Step 3: Drop old user_id columns
ALTER TABLE notebooks DROP COLUMN IF EXISTS user_id;
ALTER TABLE calendars DROP COLUMN IF EXISTS user_id;
ALTER TABLE meeting_recordings DROP COLUMN IF EXISTS user_id;
ALTER TABLE ai_credentials DROP COLUMN IF EXISTS user_id;
ALTER TABLE calendar_o_auth_states DROP COLUMN IF EXISTS user_id;

-- Step 4: Add clerk_user_id columns
ALTER TABLE notebooks ADD COLUMN clerk_user_id TEXT NOT NULL DEFAULT '';
ALTER TABLE calendars ADD COLUMN clerk_user_id TEXT NOT NULL DEFAULT '';
ALTER TABLE meeting_recordings ADD COLUMN clerk_user_id TEXT NOT NULL DEFAULT '';
ALTER TABLE ai_credentials ADD COLUMN clerk_user_id TEXT NOT NULL DEFAULT '';
ALTER TABLE calendar_o_auth_states ADD COLUMN clerk_user_id TEXT NOT NULL DEFAULT '';

-- Step 5: Remove defaults (they were just for adding the column)
ALTER TABLE notebooks ALTER COLUMN clerk_user_id DROP DEFAULT;
ALTER TABLE calendars ALTER COLUMN clerk_user_id DROP DEFAULT;
ALTER TABLE meeting_recordings ALTER COLUMN clerk_user_id DROP DEFAULT;
ALTER TABLE ai_credentials ALTER COLUMN clerk_user_id DROP DEFAULT;
ALTER TABLE calendar_o_auth_states ALTER COLUMN clerk_user_id DROP DEFAULT;

-- Step 6: Create indexes on clerk_user_id for performance
CREATE INDEX IF NOT EXISTS idx_notebooks_clerk_user_id ON notebooks(clerk_user_id);
CREATE INDEX IF NOT EXISTS idx_calendars_clerk_user_id ON calendars(clerk_user_id);
CREATE INDEX IF NOT EXISTS idx_meeting_recordings_clerk_user_id ON meeting_recordings(clerk_user_id);
CREATE INDEX IF NOT EXISTS idx_ai_credentials_clerk_user_id ON ai_credentials(clerk_user_id);
CREATE INDEX IF NOT EXISTS idx_calendar_o_auth_states_clerk_user_id ON calendar_o_auth_states(clerk_user_id);

-- Step 7: Drop the users table
DROP TABLE IF EXISTS users CASCADE;

-- Note: Onboarding status will now be stored in Clerk user metadata
-- Use clerk.UpdateUser() with publicMetadata or unsafeMetadata

