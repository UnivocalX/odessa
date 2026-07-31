DROP TABLE IF EXISTS password_reset_tokens;
DROP TABLE IF EXISTS refresh_sessions;
ALTER TABLE users DROP COLUMN IF EXISTS disabled_at;