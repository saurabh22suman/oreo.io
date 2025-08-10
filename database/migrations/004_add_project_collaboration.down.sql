-- Rollback migration: Remove project collaboration support

DROP VIEW IF EXISTS project_access;
DROP TRIGGER IF EXISTS add_project_owner_to_members ON projects;
DROP FUNCTION IF EXISTS add_project_owner_to_members();
DROP TRIGGER IF EXISTS update_project_members_updated_at ON project_members;
DROP FUNCTION IF EXISTS update_project_members_updated_at();
DROP INDEX IF EXISTS idx_project_members_status;
DROP INDEX IF EXISTS idx_project_members_role;
DROP INDEX IF EXISTS idx_project_members_user_id;
DROP INDEX IF EXISTS idx_project_members_project_id;
DROP TABLE IF EXISTS project_members;
