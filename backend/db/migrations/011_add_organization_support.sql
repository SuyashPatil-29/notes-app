-- Migration: Add organization support
-- Add organization_id columns to notebooks, chapters, and notes tables

-- Add organization_id to notebooks table
ALTER TABLE notebooks ADD COLUMN organization_id VARCHAR(255) NULL;

-- Add organization_id to chapters table
ALTER TABLE chapters ADD COLUMN organization_id VARCHAR(255) NULL;

-- Add organization_id to notes table
ALTER TABLE notes ADD COLUMN organization_id VARCHAR(255) NULL;

-- Create indexes for better query performance
CREATE INDEX idx_notebooks_organization_id ON notebooks(organization_id);
CREATE INDEX idx_chapters_organization_id ON chapters(organization_id);
CREATE INDEX idx_notes_organization_id ON notes(organization_id);

-- Create composite indexes for common queries (user + org context)
CREATE INDEX idx_notebooks_user_org ON notebooks(clerk_user_id, organization_id);
CREATE INDEX idx_chapters_notebook_org ON chapters(notebook_id, organization_id);
CREATE INDEX idx_notes_chapter_org ON notes(chapter_id, organization_id);

