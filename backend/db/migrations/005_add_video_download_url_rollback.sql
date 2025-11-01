-- Rollback Migration: Remove video_download_url from meeting_recordings
-- Date: 2025-11-01
-- Description: Rollback the addition of video_download_url column

-- Remove video_download_url column from meeting_recordings table
ALTER TABLE meeting_recordings DROP COLUMN IF EXISTS video_download_url;

