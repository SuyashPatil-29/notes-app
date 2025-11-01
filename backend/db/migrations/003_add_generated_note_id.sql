-- Add generated_note_id column to meeting_recordings table
ALTER TABLE meeting_recordings 
ADD COLUMN generated_note_id VARCHAR(255),
ADD CONSTRAINT fk_meeting_recordings_generated_note 
    FOREIGN KEY (generated_note_id) 
    REFERENCES notes(id) 
    ON DELETE SET NULL;

-- Add index for performance
CREATE INDEX idx_meeting_recordings_generated_note_id ON meeting_recordings(generated_note_id);

