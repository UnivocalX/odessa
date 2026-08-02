CREATE TABLE IF NOT EXISTS scan_origins (
    id          BIGSERIAL PRIMARY KEY,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,
    origin_id   BIGINT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'pending',
    results     JSONB NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_scan_origins_origin_id ON scan_origins(origin_id);
CREATE INDEX IF NOT EXISTS idx_scan_origins_deleted_at ON scan_origins(deleted_at);
