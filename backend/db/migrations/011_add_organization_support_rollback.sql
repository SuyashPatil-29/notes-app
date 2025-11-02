-- Rollback Migration: Remove organization support
-- Remove organization_id columns and indexes

-- Drop composite indexes
DROP INDEX IF EXISTS idx_notes_chapter_org;
DROP INDEX IF EXISTS idx_chapters_notebook_org;
DROP INDEX IF EXISTS idx_notebooks_user_org;

-- Drop organization_id indexes
DROP INDEX IF EXISTS idx_notes_organization_id;
DROP INDEX IF EXISTS idx_chapters_organization_id;
DROP INDEX IF EXISTS idx_notebooks_organization_id;

-- Remove organization_id columns
ALTER TABLE notes DROP COLUMN IF EXISTS organization_id;
ALTER TABLE chapters DROP COLUMN IF EXISTS organization_id;
ALTER TABLE notebooks DROP COLUMN IF EXISTS organization_id;

