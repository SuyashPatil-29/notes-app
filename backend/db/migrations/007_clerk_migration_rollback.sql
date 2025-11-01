-- Rollback Migration 007: Clerk Authentication Migration
-- This script reverts the Clerk authentication changes

-- WARNING: This will remove all Clerk user ID mappings
-- Make sure you have backed up your data before running this

-- Drop the clerk_user_id column
-- Note: Commenting out by default to prevent accidental data loss
-- Uncomment if you really want to rollback

-- ALTER TABLE users DROP COLUMN IF EXISTS clerk_user_id;

-- Drop AI credentials table
-- DROP TABLE IF EXISTS ai_credentials CASCADE;

-- Note: To restore Gothic/Goth authentication, you would also need to:
-- 1. Restore the old auth.go, middleware/auth.go, and main.go files
-- 2. Restore go.mod with gothic/goth dependencies
-- 3. Restore frontend auth hooks and components
-- 4. Update environment variables

