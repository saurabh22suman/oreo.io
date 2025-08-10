-- Migration: Add project collaboration support
-- Creates project_members table for multi-user project access

CREATE TABLE project_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL CHECK (role IN ('owner', 'admin', 'collaborator', 'viewer')),
    invited_by UUID REFERENCES users(id),
    invited_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    joined_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'declined', 'removed')),
    permissions JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Ensure one role per user per project
    UNIQUE(project_id, user_id)
);

-- Create indexes for efficient queries
CREATE INDEX idx_project_members_project_id ON project_members(project_id);
CREATE INDEX idx_project_members_user_id ON project_members(user_id);
CREATE INDEX idx_project_members_role ON project_members(role);
CREATE INDEX idx_project_members_status ON project_members(status);

-- Create trigger to automatically update updated_at
CREATE OR REPLACE FUNCTION update_project_members_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_project_members_updated_at
    BEFORE UPDATE ON project_members
    FOR EACH ROW
    EXECUTE FUNCTION update_project_members_updated_at();

-- Add project owner to project_members automatically when project is created
CREATE OR REPLACE FUNCTION add_project_owner_to_members()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO project_members (project_id, user_id, role, status, joined_at)
    VALUES (NEW.id, NEW.owner_id, 'owner', 'accepted', CURRENT_TIMESTAMP);
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER add_project_owner_to_members
    AFTER INSERT ON projects
    FOR EACH ROW
    EXECUTE FUNCTION add_project_owner_to_members();

-- Create view for easier project access queries
CREATE VIEW project_access AS
SELECT 
    p.id as project_id,
    p.name as project_name,
    p.description as project_description,
    p.owner_id,
    pm.user_id,
    pm.role,
    pm.status,
    pm.joined_at,
    u.name as user_name,
    u.email as user_email
FROM projects p
JOIN project_members pm ON p.id = pm.project_id
JOIN users u ON pm.user_id = u.id
WHERE pm.status = 'accepted';
