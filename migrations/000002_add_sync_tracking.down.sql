-- Remove sync tracking columns and revert to original schema
-- Drop sync jobs table
DROP TABLE IF EXISTS sync_jobs;

-- Drop indexes first
DROP INDEX IF EXISTS idx_sync_attempts ON items;
DROP INDEX IF EXISTS idx_content_hash ON items;
-- Note: idx_last_synced_at will be removed when we rename column back

-- Remove added columns
ALTER TABLE items DROP COLUMN IF EXISTS last_sync_error;
ALTER TABLE items DROP COLUMN IF EXISTS sync_attempts;
ALTER TABLE items DROP COLUMN IF EXISTS content_hash;

-- Rename last_synced_at back to synced_at
ALTER TABLE items CHANGE COLUMN last_synced_at synced_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;