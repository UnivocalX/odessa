CREATE TABLE IF NOT EXISTS locations (
    id          BIGSERIAL PRIMARY KEY,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,

    blob_id     BIGINT NOT NULL REFERENCES blobs(id) ON DELETE CASCADE,
    uri         TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_locations_deleted_at
    ON locations(deleted_at);

CREATE INDEX IF NOT EXISTS idx_locations_blob_id
    ON locations(blob_id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_locations_blob_id_uri
    ON locations(blob_id, uri);