-- Migration: Add video_download_url to meeting_recordings
-- Date: 2025-11-01
-- Description: Add support for storing video recording download URL from Recall.ai

-- Add video_download_url column to meeting_recordings table
ALTER TABLE meeting_recordings ADD COLUMN IF NOT EXISTS video_download_url TEXT;

-- Add comment to clarify the purpose
COMMENT ON COLUMN meeting_recordings.video_download_url IS 'URL to download the mixed video recording from Recall.ai (MP4 format)';

