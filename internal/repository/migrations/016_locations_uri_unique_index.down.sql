DROP INDEX IF EXISTS idx_locations_uri;
CREATE UNIQUE INDEX IF NOT EXISTS idx_locations_blob_id_uri ON locations(blob_id, uri);
