CREATE TABLE IF NOT EXISTS datasets (
    id          BIGSERIAL PRIMARY KEY,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,

    name        VARCHAR(64)  NOT NULL,
    description VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_datasets_deleted_at ON datasets(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_datasets_name ON datasets(name);

CREATE TABLE IF NOT EXISTS dataset_versions (
    id          BIGSERIAL PRIMARY KEY,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,

    dataset_id  BIGINT NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
    commit      VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_dataset_versions_deleted_at ON dataset_versions(deleted_at);
CREATE INDEX IF NOT EXISTS idx_dataset_versions_dataset_id ON dataset_versions(dataset_id);

CREATE TABLE IF NOT EXISTS dataset_version_blobs (
    dataset_version_id BIGINT NOT NULL REFERENCES dataset_versions(id) ON DELETE CASCADE,
    blob_id            BIGINT NOT NULL REFERENCES blobs(id) ON DELETE CASCADE,
    PRIMARY KEY (dataset_version_id, blob_id)
);

CREATE INDEX IF NOT EXISTS idx_dataset_version_blobs_blob_id ON dataset_version_blobs(blob_id);
