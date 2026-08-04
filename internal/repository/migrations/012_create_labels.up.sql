CREATE TABLE IF NOT EXISTS labels (
    id          BIGSERIAL PRIMARY KEY,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,

    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT ''
);

CREATE INDEX  IF NOT EXISTS idx_labels_deleted_at ON labels(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_labels_name ON labels(name);

CREATE TABLE IF NOT EXISTS blob_labels (
    id          BIGSERIAL PRIMARY KEY,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,

    label_id    BIGINT NOT NULL REFERENCES labels(id) ON DELETE CASCADE,
    blob_id     BIGINT NOT NULL REFERENCES blobs(id)  ON DELETE CASCADE,
    value       TEXT NOT NULL DEFAULT ''
);

CREATE INDEX  IF NOT EXISTS idx_blob_labels_deleted_at    ON blob_labels(deleted_at);
CREATE INDEX  IF NOT EXISTS idx_blob_labels_blob_id       ON blob_labels(blob_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_blob_labels_label_blob ON blob_labels(label_id, blob_id) WHERE deleted_at IS NULL;
