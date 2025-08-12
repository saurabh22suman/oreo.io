-- Create data submissions table for append requests
CREATE TABLE IF NOT EXISTS data_submissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dataset_id UUID NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
    submitted_by UUID NOT NULL REFERENCES users(id),
    file_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_size BIGINT NOT NULL,
    row_count INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, under_review, approved, rejected, applied
    validation_results JSONB, -- Schema and business rule validation results
    admin_notes TEXT,
    reviewed_by UUID REFERENCES users(id),
    reviewed_at TIMESTAMP,
    submitted_at TIMESTAMP NOT NULL DEFAULT NOW(),
    applied_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_data_submissions_dataset_id ON data_submissions(dataset_id);
CREATE INDEX IF NOT EXISTS idx_data_submissions_status ON data_submissions(status);
CREATE INDEX IF NOT EXISTS idx_data_submissions_submitted_by ON data_submissions(submitted_by);
CREATE INDEX IF NOT EXISTS idx_data_submissions_reviewed_by ON data_submissions(reviewed_by);

-- Create data submission staging table to store the actual data before approval
CREATE TABLE IF NOT EXISTS data_submission_staging (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    submission_id UUID NOT NULL REFERENCES data_submissions(id) ON DELETE CASCADE,
    row_index INTEGER NOT NULL,
    data JSONB NOT NULL,
    validation_status VARCHAR(50) DEFAULT 'valid', -- valid, invalid, warning
    validation_errors JSONB, -- Specific validation errors for this row
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes for staging table
CREATE INDEX IF NOT EXISTS idx_data_submission_staging_submission_id ON data_submission_staging(submission_id);
CREATE INDEX IF NOT EXISTS idx_data_submission_staging_validation_status ON data_submission_staging(validation_status);

-- Create business rules table to define validation rules for datasets
CREATE TABLE IF NOT EXISTS dataset_business_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dataset_id UUID NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
    rule_name VARCHAR(255) NOT NULL,
    rule_type VARCHAR(50) NOT NULL, -- field_validation, cross_field, custom_sql, range_check, etc.
    rule_config JSONB NOT NULL, -- Configuration for the rule (field names, conditions, etc.)
    error_message TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    priority INTEGER NOT NULL DEFAULT 100, -- Lower numbers = higher priority
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes for business rules
CREATE INDEX IF NOT EXISTS idx_dataset_business_rules_dataset_id ON dataset_business_rules(dataset_id);
CREATE INDEX IF NOT EXISTS idx_dataset_business_rules_is_active ON dataset_business_rules(is_active);
