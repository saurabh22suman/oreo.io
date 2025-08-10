-- Drop constraints first
ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_auth_method;
ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_name_length;
ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_email_format;

-- Drop indexes
DROP INDEX IF EXISTS idx_users_created_at;
DROP INDEX IF EXISTS idx_users_google_id;
DROP INDEX IF EXISTS idx_users_email;

-- Drop table
DROP TABLE IF EXISTS users;
