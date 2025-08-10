-- Revert google_id constraint fix
-- Drop the partial unique index
DROP INDEX IF EXISTS users_google_id_unique_idx;

-- Add back the original unique constraint
ALTER TABLE users ADD CONSTRAINT users_google_id_key UNIQUE (google_id);
