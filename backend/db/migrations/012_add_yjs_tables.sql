-- Migration: Add Yjs collaboration tables
-- Description: Create tables for storing Yjs CRDT document state and update history

-- Create yjs_documents table for storing current Yjs binary state
CREATE TABLE IF NOT EXISTS yjs_documents (
    id VARCHAR(255) PRIMARY KEY,
    note_id VARCHAR(255) NOT NULL UNIQUE,
    yjs_state BYTEA NOT NULL,
    version INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (note_id) REFERENCES notes(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_yjs_documents_note_id ON yjs_documents(note_id);

-- Create yjs_updates table for incremental updates (for version history)
CREATE TABLE IF NOT EXISTS yjs_updates (
    id SERIAL PRIMARY KEY,
    note_id VARCHAR(255) NOT NULL,
    update_data BYTEA NOT NULL,
    clock INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (note_id) REFERENCES notes(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_yjs_updates_note_id ON yjs_updates(note_id);
CREATE INDEX IF NOT EXISTS idx_yjs_updates_created_at ON yjs_updates(created_at);

