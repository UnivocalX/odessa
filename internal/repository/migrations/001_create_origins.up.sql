CREATE TABLE IF NOT EXISTS origins (
    id          BIGSERIAL PRIMARY KEY,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,
    uri         TEXT NOT NULL UNIQUE
);

CREATE INDEX IF NOT EXISTS idx_origins_deleted_at ON origins(deleted_at);
