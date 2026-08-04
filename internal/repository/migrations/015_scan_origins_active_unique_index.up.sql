DROP INDEX IF EXISTS idx_scan_origins_origin_id;

CREATE UNIQUE INDEX IF NOT EXISTS idx_scan_origins_active_origin
ON scan_origins (origin_id)
WHERE status IN ('pending', 'in_progress') AND deleted_at IS NULL;