-- Migration 007: Clerk Authentication Migration
-- This migration updates the users table for Clerk authentication

-- Add clerk_user_id column if it doesn't exist (already exists from GORM auto-migration)
-- ALTER TABLE users ADD COLUMN IF NOT EXISTS clerk_user_id TEXT;

-- Create unique index on clerk_user_id if it doesn't exist (already exists)
-- CREATE UNIQUE INDEX IF NOT EXISTS idx_users_clerk_user_id ON users(clerk_user_id);

-- Note: We cannot make clerk_user_id NOT NULL immediately because existing users don't have Clerk IDs
-- Users will get Clerk IDs assigned on their first login after the migration
-- New users created through Clerk will have this field populated immediately

-- Optional: If you want to remove users without Clerk IDs after a grace period, uncomment:
-- DELETE FROM users WHERE clerk_user_id IS NULL;
-- ALTER TABLE users ALTER COLUMN clerk_user_id SET NOT NULL;

-- Create AI credentials table if it doesn't exist
CREATE TABLE IF NOT EXISTS ai_credentials (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    provider TEXT NOT NULL,
    key_cipher TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_ai_credentials_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT unique_user_provider UNIQUE (user_id, provider)
);

-- Create index on user_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_ai_credentials_user_id ON ai_credentials(user_id);

