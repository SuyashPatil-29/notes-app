-- Rollback: Remove generated_note_id column from meeting_recordings table
ALTER TABLE meeting_recordings 
DROP CONSTRAINT IF EXISTS fk_meeting_recordings_generated_note,
DROP COLUMN IF EXISTS generated_note_id;

-- Drop index
DROP INDEX IF EXISTS idx_meeting_recordings_generated_note_id;

