CREATE TABLE IF NOT EXISTS blobs (
    id          BIGSERIAL PRIMARY KEY,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,
    hash        TEXT NOT NULL,
    mime_type   TEXT NOT NULL DEFAULT '',
    size        BIGINT NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_blobs_deleted_at ON blobs(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_blobs_hash ON blobs(hash);