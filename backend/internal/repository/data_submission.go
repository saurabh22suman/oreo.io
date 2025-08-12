package repository

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/saurabh22suman/oreo.io/internal/models"
)

type DataSubmissionRepository struct {
	db *sqlx.DB
}

func NewDataSubmissionRepository(db *sqlx.DB) *DataSubmissionRepository {
	return &DataSubmissionRepository{db: db}
}

// CreateSubmission creates a new data submission request
func (r *DataSubmissionRepository) CreateSubmission(submission *models.DataSubmission) error {
	query := `
		INSERT INTO data_submissions (
			id, dataset_id, submitted_by, file_name, file_path, file_size, 
			row_count, status, validation_results, submitted_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err := r.db.Exec(query,
		submission.ID,
		submission.DatasetID,
		submission.SubmittedBy,
		submission.FileName,
		submission.FilePath,
		submission.FileSize,
		submission.RowCount,
		submission.Status,
		submission.ValidationResults,
		submission.SubmittedAt,
		submission.CreatedAt,
		submission.UpdatedAt,
	)

	return err
}

// GetSubmission retrieves a data submission by ID
func (r *DataSubmissionRepository) GetSubmission(id uuid.UUID) (*models.DataSubmission, error) {
	var submission models.DataSubmission
	query := `SELECT * FROM data_submissions WHERE id = $1`

	err := r.db.Get(&submission, query, id)
	if err != nil {
		return nil, err
	}

	return &submission, nil
}

// GetSubmissionWithDetails retrieves a submission with additional details
func (r *DataSubmissionRepository) GetSubmissionWithDetails(id uuid.UUID) (*models.DataSubmissionWithDetails, error) {
	var submission models.DataSubmissionWithDetails
	query := `
		SELECT 
			ds.*,
			d.name as dataset_name,
			p.name as project_name,
			u1.name as submitter_name,
			u1.email as submitter_email,
			u2.name as reviewer_name
		FROM data_submissions ds
		JOIN datasets d ON ds.dataset_id = d.id
		JOIN projects p ON d.project_id = p.id
		JOIN users u1 ON ds.submitted_by = u1.id
		LEFT JOIN users u2 ON ds.reviewed_by = u2.id
		WHERE ds.id = $1`

	err := r.db.Get(&submission, query, id)
	if err != nil {
		return nil, err
	}

	return &submission, nil
}

// GetSubmissionsByDataset retrieves all submissions for a dataset
func (r *DataSubmissionRepository) GetSubmissionsByDataset(datasetID uuid.UUID) ([]*models.DataSubmissionWithDetails, error) {
	var submissions []*models.DataSubmissionWithDetails
	query := `
		SELECT 
			ds.*,
			d.name as dataset_name,
			p.name as project_name,
			u1.name as submitter_name,
			u1.email as submitter_email,
			u2.name as reviewer_name
		FROM data_submissions ds
		JOIN datasets d ON ds.dataset_id = d.id
		JOIN projects p ON d.project_id = p.id
		JOIN users u1 ON ds.submitted_by = u1.id
		LEFT JOIN users u2 ON ds.reviewed_by = u2.id
		WHERE ds.dataset_id = $1
		ORDER BY ds.submitted_at DESC`

	rows, err := r.db.Query(query, datasetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var submission models.DataSubmissionWithDetails
		err := rows.Scan(
			&submission.ID, &submission.DatasetID, &submission.SubmittedBy,
			&submission.FileName, &submission.FilePath, &submission.FileSize,
			&submission.RowCount, &submission.Status, &submission.ValidationResults,
			&submission.AdminNotes, &submission.ReviewedBy, &submission.ReviewedAt,
			&submission.SubmittedAt, &submission.AppliedAt, &submission.CreatedAt,
			&submission.UpdatedAt, &submission.DatasetName, &submission.ProjectName,
			&submission.SubmitterName, &submission.SubmitterEmail, &submission.ReviewerName,
		)
		if err != nil {
			return nil, err
		}
		submissions = append(submissions, &submission)
	}

	return submissions, nil
}

// GetPendingSubmissions retrieves all pending submissions for admin review
func (r *DataSubmissionRepository) GetPendingSubmissions() ([]*models.DataSubmissionWithDetails, error) {
	var submissions []*models.DataSubmissionWithDetails
	query := `
		SELECT 
			ds.*,
			d.name as dataset_name,
			p.name as project_name,
			u1.name as submitter_name,
			u1.email as submitter_email,
			u2.name as reviewer_name
		FROM data_submissions ds
		JOIN datasets d ON ds.dataset_id = d.id
		JOIN projects p ON d.project_id = p.id
		JOIN users u1 ON ds.submitted_by = u1.id
		LEFT JOIN users u2 ON ds.reviewed_by = u2.id
		WHERE ds.status IN ($1, $2)
		ORDER BY ds.submitted_at ASC`

	rows, err := r.db.Query(query, models.DataSubmissionStatusPending, models.DataSubmissionStatusUnderReview)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var submission models.DataSubmissionWithDetails
		err := rows.Scan(
			&submission.ID, &submission.DatasetID, &submission.SubmittedBy,
			&submission.FileName, &submission.FilePath, &submission.FileSize,
			&submission.RowCount, &submission.Status, &submission.ValidationResults,
			&submission.AdminNotes, &submission.ReviewedBy, &submission.ReviewedAt,
			&submission.SubmittedAt, &submission.AppliedAt, &submission.CreatedAt,
			&submission.UpdatedAt, &submission.DatasetName, &submission.ProjectName,
			&submission.SubmitterName, &submission.SubmitterEmail, &submission.ReviewerName,
		)
		if err != nil {
			return nil, err
		}
		submissions = append(submissions, &submission)
	}

	return submissions, nil
}

// UpdateSubmissionStatus updates the status and admin review of a submission
func (r *DataSubmissionRepository) UpdateSubmissionStatus(id uuid.UUID, status string, adminNotes *string, reviewedBy uuid.UUID) error {
	query := `
		UPDATE data_submissions 
		SET status = $1, admin_notes = $2, reviewed_by = $3, reviewed_at = $4, updated_at = $5
		WHERE id = $6`

	now := time.Now()
	_, err := r.db.Exec(query, status, adminNotes, reviewedBy, now, now, id)
	return err
}

// MarkSubmissionApplied marks a submission as applied to the target dataset
func (r *DataSubmissionRepository) MarkSubmissionApplied(id uuid.UUID) error {
	query := `
		UPDATE data_submissions 
		SET status = $1, applied_at = $2, updated_at = $3
		WHERE id = $4`

	now := time.Now()
	_, err := r.db.Exec(query, models.DataSubmissionStatusApplied, now, now, id)
	return err
}

// DeleteSubmission deletes a submission and all its staging data
func (r *DataSubmissionRepository) DeleteSubmission(id uuid.UUID) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete staging data first
	_, err = tx.Exec("DELETE FROM data_submission_staging WHERE submission_id = $1", id)
	if err != nil {
		return err
	}

	// Delete submission
	_, err = tx.Exec("DELETE FROM data_submissions WHERE id = $1", id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// CreateStagingData creates staging data for a submission
func (r *DataSubmissionRepository) CreateStagingData(stagingData []*models.DataSubmissionStaging) error {
	if len(stagingData) == 0 {
		return nil
	}

	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO data_submission_staging (
			id, submission_id, row_index, data, validation_status, validation_errors, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	for _, data := range stagingData {
		_, err = tx.Exec(query,
			data.ID,
			data.SubmissionID,
			data.RowIndex,
			data.Data,
			data.ValidationStatus,
			data.ValidationErrors,
			data.CreatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetStagingData retrieves staging data for a submission
func (r *DataSubmissionRepository) GetStagingData(submissionID uuid.UUID, limit, offset int) ([]*models.DataSubmissionStaging, error) {
	var stagingData []*models.DataSubmissionStaging
	query := `
		SELECT * FROM data_submission_staging 
		WHERE submission_id = $1 
		ORDER BY row_index 
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(query, submissionID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var data models.DataSubmissionStaging
		err := rows.Scan(
			&data.ID, &data.SubmissionID, &data.RowIndex, &data.Data,
			&data.ValidationStatus, &data.ValidationErrors, &data.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		stagingData = append(stagingData, &data)
	}

	return stagingData, nil
}

// UpdateStagingDataRow updates a single row in staging data (for live editing)
func (r *DataSubmissionRepository) UpdateStagingDataRow(id uuid.UUID, data json.RawMessage, validationStatus string, validationErrors *json.RawMessage) error {
	query := `
		UPDATE data_submission_staging 
		SET data = $1, validation_status = $2, validation_errors = $3
		WHERE id = $4`

	_, err := r.db.Exec(query, data, validationStatus, validationErrors, id)
	return err
}

// ApplyStagingDataToDataset applies approved staging data to the target dataset
func (r *DataSubmissionRepository) ApplyStagingDataToDataset(submissionID uuid.UUID, datasetID uuid.UUID, userID uuid.UUID) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get the current max row index in the dataset
	var maxRowIndex sql.NullInt64
	err = tx.Get(&maxRowIndex, "SELECT MAX(row_index) FROM dataset_data WHERE dataset_id = $1", datasetID)
	if err != nil {
		return err
	}

	startIndex := 0
	if maxRowIndex.Valid {
		startIndex = int(maxRowIndex.Int64) + 1
	}

	// Copy valid staging data to dataset_data
	query := `
		INSERT INTO dataset_data (dataset_id, row_index, data, created_by, updated_by)
		SELECT $1, $2 + row_index, data, $3, $3
		FROM data_submission_staging 
		WHERE submission_id = $4 AND validation_status = $5
		ORDER BY row_index`

	_, err = tx.Exec(query, datasetID, startIndex, userID, submissionID, models.ValidationStatusValid)
	if err != nil {
		return err
	}

	// Update dataset row count
	_, err = tx.Exec(`
		UPDATE datasets 
		SET row_count = (SELECT COUNT(*) FROM dataset_data WHERE dataset_id = $1),
		    updated_at = NOW()
		WHERE id = $1`, datasetID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Business Rules methods

// CreateBusinessRule creates a new business rule for a dataset
func (r *DataSubmissionRepository) CreateBusinessRule(rule *models.DatasetBusinessRule) error {
	query := `
		INSERT INTO dataset_business_rules (
			id, dataset_id, rule_name, rule_type, rule_config, error_message,
			is_active, priority, created_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.db.Exec(query,
		rule.ID, rule.DatasetID, rule.RuleName, rule.RuleType, rule.RuleConfig,
		rule.ErrorMessage, rule.IsActive, rule.Priority, rule.CreatedBy,
		rule.CreatedAt, rule.UpdatedAt,
	)

	return err
}

// GetBusinessRules retrieves active business rules for a dataset
func (r *DataSubmissionRepository) GetBusinessRules(datasetID uuid.UUID) ([]*models.DatasetBusinessRule, error) {
	var rules []*models.DatasetBusinessRule
	query := `
		SELECT * FROM dataset_business_rules 
		WHERE dataset_id = $1 AND is_active = true 
		ORDER BY priority ASC, created_at ASC`

	rows, err := r.db.Query(query, datasetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var rule models.DatasetBusinessRule
		err := rows.Scan(
			&rule.ID, &rule.DatasetID, &rule.RuleName, &rule.RuleType,
			&rule.RuleConfig, &rule.ErrorMessage, &rule.IsActive, &rule.Priority,
			&rule.CreatedBy, &rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		rules = append(rules, &rule)
	}

	return rules, nil
}

// UpdateBusinessRule updates an existing business rule
func (r *DataSubmissionRepository) UpdateBusinessRule(rule *models.DatasetBusinessRule) error {
	query := `
		UPDATE dataset_business_rules 
		SET rule_name = $1, rule_type = $2, rule_config = $3, error_message = $4,
		    is_active = $5, priority = $6, updated_at = $7
		WHERE id = $8`

	_, err := r.db.Exec(query,
		rule.RuleName, rule.RuleType, rule.RuleConfig, rule.ErrorMessage,
		rule.IsActive, rule.Priority, time.Now(), rule.ID,
	)

	return err
}

// DeleteBusinessRule deletes a business rule
func (r *DataSubmissionRepository) DeleteBusinessRule(id uuid.UUID) error {
	query := `DELETE FROM dataset_business_rules WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

// CheckDatasetAccess verifies if user has access to the dataset
func (r *DataSubmissionRepository) CheckDatasetAccess(datasetID uuid.UUID, userID uuid.UUID) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM datasets d
		JOIN projects p ON d.project_id = p.id
		LEFT JOIN project_collaborators pc ON p.id = pc.project_id
		WHERE d.id = $1 AND (p.created_by = $2 OR pc.user_id = $2)`

	err := r.db.Get(&count, query, datasetID, userID)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// IsUserAdmin checks if user has admin privileges
func (r *DataSubmissionRepository) IsUserAdmin(userID uuid.UUID) (bool, error) {
	var role string
	query := `SELECT role FROM users WHERE id = $1`
	
	err := r.db.Get(&role, query, userID)
	if err != nil {
		return false, err
	}

	// Assuming 'admin' or 'super_admin' roles have admin privileges
	return role == "admin" || role == "super_admin", nil
}
