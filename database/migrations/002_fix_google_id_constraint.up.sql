-- Fix google_id constraint to allow NULLs instead of empty strings
-- Drop the existing unique constraint
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_google_id_key;

-- Update existing empty string google_id values to NULL
UPDATE users SET google_id = NULL WHERE google_id = '';

-- Create a partial unique index that only applies to non-NULL values
CREATE UNIQUE INDEX users_google_id_unique_idx ON users (google_id) WHERE google_id IS NOT NULL;
