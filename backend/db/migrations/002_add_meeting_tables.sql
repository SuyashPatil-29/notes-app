-- Migration: Add meeting recording and participant tables
-- Date: 2024-11-01
-- Description: Add support for Recall.ai meeting transcription integration

-- Create meeting_recordings table
CREATE TABLE IF NOT EXISTS meeting_recordings (
    id VARCHAR(255) PRIMARY KEY,
    user_id INTEGER NOT NULL,
    bot_id VARCHAR(255) NOT NULL UNIQUE,
    meeting_url TEXT NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    recall_recording_id VARCHAR(255),
    transcript_download_url TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP,
    
    CONSTRAINT fk_meeting_recordings_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create indexes for meeting_recordings
CREATE INDEX IF NOT EXISTS idx_meeting_recordings_user_id ON meeting_recordings(user_id);
CREATE INDEX IF NOT EXISTS idx_meeting_recordings_bot_id ON meeting_recordings(bot_id);
CREATE INDEX IF NOT EXISTS idx_meeting_recordings_status ON meeting_recordings(status);

-- Create meeting_participants table
CREATE TABLE IF NOT EXISTS meeting_participants (
    id SERIAL PRIMARY KEY,
    meeting_recording_id VARCHAR(255) NOT NULL,
    participant_id INTEGER NOT NULL,
    name VARCHAR(255),
    is_host BOOLEAN DEFAULT FALSE,
    platform VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW(),
    
    CONSTRAINT fk_meeting_participants_recording FOREIGN KEY (meeting_recording_id) REFERENCES meeting_recordings(id) ON DELETE CASCADE
);

-- Create indexes for meeting_participants
CREATE INDEX IF NOT EXISTS idx_meeting_participants_recording_id ON meeting_participants(meeting_recording_id);

-- Add meeting-related columns to notes table
ALTER TABLE notes ADD COLUMN IF NOT EXISTS meeting_recording_id VARCHAR(255);
ALTER TABLE notes ADD COLUMN IF NOT EXISTS ai_summary TEXT;
ALTER TABLE notes ADD COLUMN IF NOT EXISTS transcript_raw TEXT;

-- Add foreign key constraint for meeting_recording_id
ALTER TABLE notes ADD CONSTRAINT IF NOT EXISTS fk_notes_meeting_recording 
    FOREIGN KEY (meeting_recording_id) REFERENCES meeting_recordings(id) ON DELETE SET NULL;