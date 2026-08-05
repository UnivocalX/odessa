-- Keep one row per URI before adding a URI-only unique index.
-- We keep the most recently inserted row (highest id).
DELETE FROM locations older
USING locations newer
WHERE older.uri = newer.uri
  AND older.id < newer.id;

DROP INDEX IF EXISTS idx_locations_blob_id_uri;
CREATE UNIQUE INDEX IF NOT EXISTS idx_locations_uri ON locations(uri);
