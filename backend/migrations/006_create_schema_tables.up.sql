-- Create dataset schemas table
CREATE TABLE dataset_schemas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dataset_id UUID NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create schema fields table
CREATE TABLE schema_fields (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schema_id UUID NOT NULL REFERENCES dataset_schemas(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255),
    data_type VARCHAR(50) NOT NULL CHECK (data_type IN ('string', 'number', 'date', 'boolean', 'email', 'url')),
    is_required BOOLEAN DEFAULT FALSE,
    is_unique BOOLEAN DEFAULT FALSE,
    default_value TEXT,
    position INTEGER NOT NULL,
    validation JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(schema_id, name),
    UNIQUE(schema_id, position)
);

-- Create dataset data table
CREATE TABLE dataset_data (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dataset_id UUID NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
    row_index INTEGER NOT NULL,
    data JSONB NOT NULL,
    version INTEGER DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID NOT NULL REFERENCES users(id),
    UNIQUE(dataset_id, row_index)
);

-- Create indexes for better performance
CREATE INDEX idx_dataset_schemas_dataset_id ON dataset_schemas(dataset_id);
CREATE INDEX idx_schema_fields_schema_id ON schema_fields(schema_id);
CREATE INDEX idx_schema_fields_position ON schema_fields(schema_id, position);
CREATE INDEX idx_dataset_data_dataset_id ON dataset_data(dataset_id);
CREATE INDEX idx_dataset_data_row_index ON dataset_data(dataset_id, row_index);
CREATE INDEX idx_dataset_data_data_gin ON dataset_data USING GIN (data);

-- Create update triggers
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_dataset_schemas_updated_at BEFORE UPDATE ON dataset_schemas FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_schema_fields_updated_at BEFORE UPDATE ON schema_fields FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_dataset_data_updated_at BEFORE UPDATE ON dataset_data FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
