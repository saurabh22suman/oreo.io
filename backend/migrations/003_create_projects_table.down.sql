-- Drop trigger first
DROP TRIGGER IF EXISTS trigger_projects_updated_at ON projects;

-- Drop function
DROP FUNCTION IF EXISTS update_projects_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_projects_created_at;
DROP INDEX IF EXISTS idx_projects_owner_id;

-- Drop table
DROP TABLE IF EXISTS projects;
