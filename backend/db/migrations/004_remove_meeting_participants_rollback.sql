-- Recreate meeting_participants table (rollback)
CREATE TABLE IF NOT EXISTS meeting_participants (
    id SERIAL PRIMARY KEY,
    meeting_recording_id VARCHAR(255) NOT NULL,
    participant_id INTEGER NOT NULL,
    name VARCHAR(255),
    is_host BOOLEAN DEFAULT FALSE,
    platform VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_meeting_participants_recording
        FOREIGN KEY (meeting_recording_id)
        REFERENCES meeting_recordings(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_meeting_participants_recording_id ON meeting_participants(meeting_recording_id);

