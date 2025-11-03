-- Rollback: Remove Yjs collaboration tables

DROP INDEX IF EXISTS idx_yjs_updates_created_at;
DROP INDEX IF EXISTS idx_yjs_updates_note_id;
DROP TABLE IF EXISTS yjs_updates;

DROP INDEX IF EXISTS idx_yjs_documents_note_id;
DROP TABLE IF EXISTS yjs_documents;

