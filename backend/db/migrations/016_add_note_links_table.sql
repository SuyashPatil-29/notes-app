-- Create note_links table for storing relationships between notes
CREATE TABLE IF NOT EXISTS note_links (
    id VARCHAR(255) PRIMARY KEY,
    source_note_id VARCHAR(255) NOT NULL,
    target_note_id VARCHAR(255) NOT NULL,
    link_type VARCHAR(50) NOT NULL DEFAULT 'references',
    organization_id VARCHAR(255),
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (source_note_id) REFERENCES notes(id) ON DELETE CASCADE,
    FOREIGN KEY (target_note_id) REFERENCES notes(id) ON DELETE CASCADE,
    UNIQUE(source_note_id, target_note_id, link_type)
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_note_links_source ON note_links(source_note_id);
CREATE INDEX IF NOT EXISTS idx_note_links_target ON note_links(target_note_id);
CREATE INDEX IF NOT EXISTS idx_note_links_org ON note_links(organization_id);
CREATE INDEX IF NOT EXISTS idx_note_links_created_by ON note_links(created_by);

