-- Create datasets table
CREATE TABLE datasets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    file_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_size BIGINT NOT NULL DEFAULT 0,
    mime_type VARCHAR(100) NOT NULL,
    row_count INTEGER DEFAULT 0,
    column_count INTEGER DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'processing',
    uploaded_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX idx_datasets_project_id ON datasets(project_id);
CREATE INDEX idx_datasets_uploaded_by ON datasets(uploaded_by);
CREATE INDEX idx_datasets_status ON datasets(status);
CREATE INDEX idx_datasets_created_at ON datasets(created_at);

-- Add updated_at trigger
CREATE TRIGGER update_datasets_updated_at
    BEFORE UPDATE ON datasets
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
