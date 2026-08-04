DROP INDEX IF EXISTS idx_scan_origins_active_origin;

CREATE UNIQUE INDEX IF NOT EXISTS idx_scan_origins_origin_id ON scan_origins(origin_id);