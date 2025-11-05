-- Rollback note_links table creation
DROP INDEX IF EXISTS idx_note_links_created_by;
DROP INDEX IF EXISTS idx_note_links_org;
DROP INDEX IF EXISTS idx_note_links_target;
DROP INDEX IF EXISTS idx_note_links_source;
DROP TABLE IF EXISTS note_links;

