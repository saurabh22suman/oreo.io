-- Drop triggers
DROP TRIGGER IF EXISTS update_dataset_data_updated_at ON dataset_data;
DROP TRIGGER IF EXISTS update_schema_fields_updated_at ON schema_fields;
DROP TRIGGER IF EXISTS update_dataset_schemas_updated_at ON dataset_schemas;

-- Drop indexes
DROP INDEX IF EXISTS idx_dataset_data_data_gin;
DROP INDEX IF EXISTS idx_dataset_data_row_index;
DROP INDEX IF EXISTS idx_dataset_data_dataset_id;
DROP INDEX IF EXISTS idx_schema_fields_position;
DROP INDEX IF EXISTS idx_schema_fields_schema_id;
DROP INDEX IF EXISTS idx_dataset_schemas_dataset_id;

-- Drop tables
DROP TABLE IF EXISTS dataset_data;
DROP TABLE IF EXISTS schema_fields;
DROP TABLE IF EXISTS dataset_schemas;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();
