-- Drop data append related tables
DROP INDEX IF EXISTS idx_dataset_business_rules_is_active;
DROP INDEX IF EXISTS idx_dataset_business_rules_dataset_id;
DROP TABLE IF EXISTS dataset_business_rules;

DROP INDEX IF EXISTS idx_data_submission_staging_validation_status;
DROP INDEX IF EXISTS idx_data_submission_staging_submission_id;
DROP TABLE IF EXISTS data_submission_staging;

DROP INDEX IF EXISTS idx_data_submissions_reviewed_by;
DROP INDEX IF EXISTS idx_data_submissions_submitted_by;
DROP INDEX IF EXISTS idx_data_submissions_status;
DROP INDEX IF EXISTS idx_data_submissions_dataset_id;
DROP TABLE IF EXISTS data_submissions;
