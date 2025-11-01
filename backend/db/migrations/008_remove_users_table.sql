-- Migration 008: Remove users table and use clerk_user_id directly
-- This migration eliminates the users table and uses Clerk as the single source of truth

-- Step 1: Add clerk_user_id columns to all related tables
ALTER TABLE notebooks ADD COLUMN IF NOT EXISTS clerk_user_id TEXT;
ALTER TABLE calendars ADD COLUMN IF NOT EXISTS clerk_user_id TEXT;
ALTER TABLE meeting_recordings ADD COLUMN IF NOT EXISTS clerk_user_id TEXT;
ALTER TABLE ai_credentials ADD COLUMN IF NOT EXISTS clerk_user_id TEXT;
ALTER TABLE calendar_o_auth_states ADD COLUMN IF NOT EXISTS clerk_user_id TEXT;

-- Step 2: Migrate existing data (copy user_id to clerk_user_id based on users table)
-- WARNING: This will fail if you have real data. Only works if users table is empty or test data.
-- For production, you'd need to map user.id -> user.clerk_user_id first

-- Step 3: Drop foreign key constraints
ALTER TABLE notebooks DROP CONSTRAINT IF EXISTS fk_users_notebooks;
ALTER TABLE calendars DROP CONSTRAINT IF EXISTS calendars_user_id_fkey;
ALTER TABLE meeting_recordings DROP CONSTRAINT IF EXISTS fk_meeting_recordings_user;
ALTER TABLE ai_credentials DROP CONSTRAINT IF EXISTS fk_ai_credentials_user;
ALTER TABLE calendar_o_auth_states DROP CONSTRAINT IF EXISTS calendar_o_auth_states_user_id_fkey;

-- Step 4: Drop old user_id columns
ALTER TABLE notebooks DROP COLUMN IF EXISTS user_id;
ALTER TABLE calendars DROP COLUMN IF EXISTS user_id;
ALTER TABLE meeting_recordings DROP COLUMN IF EXISTS user_id;
ALTER TABLE ai_credentials DROP COLUMN IF EXISTS user_id;
ALTER TABLE calendar_o_auth_states DROP COLUMN IF EXISTS user_id;

-- Step 5: Make clerk_user_id NOT NULL
ALTER TABLE notebooks ALTER COLUMN clerk_user_id SET NOT NULL;
ALTER TABLE calendars ALTER COLUMN clerk_user_id SET NOT NULL;
ALTER TABLE meeting_recordings ALTER COLUMN clerk_user_id SET NOT NULL;
ALTER TABLE ai_credentials ALTER COLUMN clerk_user_id SET NOT NULL;
ALTER TABLE calendar_o_auth_states ALTER COLUMN clerk_user_id SET NOT NULL;

-- Step 6: Create indexes on clerk_user_id for performance
CREATE INDEX IF NOT EXISTS idx_notebooks_clerk_user_id ON notebooks(clerk_user_id);
CREATE INDEX IF NOT EXISTS idx_calendars_clerk_user_id ON calendars(clerk_user_id);
CREATE INDEX IF NOT EXISTS idx_meeting_recordings_clerk_user_id ON meeting_recordings(clerk_user_id);
CREATE INDEX IF NOT EXISTS idx_ai_credentials_clerk_user_id ON ai_credentials(clerk_user_id);
CREATE INDEX IF NOT EXISTS idx_calendar_o_auth_states_clerk_user_id ON calendar_o_auth_states(clerk_user_id);

-- Step 7: Drop the users table
DROP TABLE IF EXISTS users CASCADE;

-- Note: Onboarding status will now be stored in Clerk user metadata
-- Use clerk.UpdateUser() with publicMetadata: { onboardingCompleted: true, onboardingType: "..." }

