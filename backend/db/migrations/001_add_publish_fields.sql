-- Migration: Add publishing fields to support public content
-- This migration adds IsPublic boolean fields to notebooks, chapters, and notes tables
-- GORM AutoMigrate will handle this automatically, but this file documents the schema changes

-- Add is_public column to notebooks table
ALTER TABLE notebooks ADD COLUMN IF NOT EXISTS is_public BOOLEAN DEFAULT false;

-- Add is_public column to chapters table
ALTER TABLE chapters ADD COLUMN IF NOT EXISTS is_public BOOLEAN DEFAULT false;

-- Add is_public column to notes table
ALTER TABLE notes ADD COLUMN IF NOT EXISTS is_public BOOLEAN DEFAULT false;

-- Note: Since this project uses GORM AutoMigrate, these columns will be added automatically
-- when the application starts. This file serves as documentation of the schema changes.
