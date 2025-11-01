-- Rollback Migration: Remove meeting recording and participant tables
-- Date: 2024-11-01
-- Description: Rollback Recall.ai meeting transcription integration

-- Remove foreign key constraint from notes table
ALTER TABLE notes DROP CONSTRAINT IF EXISTS fk_notes_meeting_recording;

-- Remove meeting-related columns from notes table
ALTER TABLE notes DROP COLUMN IF EXISTS meeting_recording_id;
ALTER TABLE notes DROP COLUMN IF EXISTS ai_summary;
ALTER TABLE notes DROP COLUMN IF EXISTS transcript_raw;

-- Drop meeting_participants table
DROP TABLE IF EXISTS meeting_participants;

-- Drop meeting_recordings table
DROP TABLE IF EXISTS meeting_recordings;