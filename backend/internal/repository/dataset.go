package repository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/saurabh22suman/oreo.io/internal/models"
)

// DatasetRepository handles dataset data operations
type DatasetRepository struct {
	db *sqlx.DB
}

// NewDatasetRepository creates a new dataset repository
func NewDatasetRepository(db *sqlx.DB) *DatasetRepository {
	return &DatasetRepository{db: db}
}

// Create creates a new dataset
func (r *DatasetRepository) Create(dataset *models.Dataset) error {
	query := `
		INSERT INTO datasets (id, project_id, name, description, file_name, file_path, 
			file_size, mime_type, row_count, column_count, status, uploaded_by, created_at, updated_at)
		VALUES (:id, :project_id, :name, :description, :file_name, :file_path, 
			:file_size, :mime_type, :row_count, :column_count, :status, :uploaded_by, :created_at, :updated_at)`

	_, err := r.db.NamedExec(query, dataset)
	return err
}

// GetByID retrieves a dataset by ID
func (r *DatasetRepository) GetByID(id uuid.UUID) (*models.Dataset, error) {
	var dataset models.Dataset
	query := `SELECT * FROM datasets WHERE id = $1`

	err := r.db.Get(&dataset, query, id)
	if err != nil {
		return nil, err
	}

	return &dataset, nil
}

// GetByProjectID retrieves all datasets for a project
func (r *DatasetRepository) GetByProjectID(projectID uuid.UUID) ([]models.Dataset, error) {
	var datasets []models.Dataset
	query := `
		SELECT * FROM datasets 
		WHERE project_id = $1 
		ORDER BY created_at DESC`

	err := r.db.Select(&datasets, query, projectID)
	if err != nil {
		return nil, err
	}

	return datasets, nil
}

// GetByUserID retrieves all datasets uploaded by a user
func (r *DatasetRepository) GetByUserID(userID uuid.UUID) ([]models.DatasetWithProject, error) {
	var datasets []models.DatasetWithProject
	query := `
		SELECT d.*, p.name as project_name
		FROM datasets d
		JOIN projects p ON d.project_id = p.id
		WHERE d.uploaded_by = $1
		ORDER BY d.created_at DESC`

	err := r.db.Select(&datasets, query, userID)
	if err != nil {
		return nil, err
	}

	return datasets, nil
}

// Update updates a dataset
func (r *DatasetRepository) Update(id uuid.UUID, updates *models.UpdateDatasetRequest) (*models.Dataset, error) {
	// Update the dataset
	updateQuery := `
		UPDATE datasets 
		SET name = $1, description = $2, updated_at = $3
		WHERE id = $4`

	_, err := r.db.Exec(updateQuery, updates.Name, updates.Description, time.Now(), id)
	if err != nil {
		return nil, fmt.Errorf("failed to update dataset: %w", err)
	}

	// Return the updated dataset
	return r.GetByID(id)
}

// UpdateStatus updates the status of a dataset
func (r *DatasetRepository) UpdateStatus(id uuid.UUID, status string, rowCount, columnCount int) error {
	query := `
		UPDATE datasets 
		SET status = $1, row_count = $2, column_count = $3, updated_at = $4
		WHERE id = $5`

	_, err := r.db.Exec(query, status, rowCount, columnCount, time.Now(), id)
	return err
}

// Delete deletes a dataset
func (r *DatasetRepository) Delete(id uuid.UUID, userID uuid.UUID) error {
	query := `DELETE FROM datasets WHERE id = $1 AND uploaded_by = $2`

	result, err := r.db.Exec(query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete dataset: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("dataset not found or access denied")
	}

	return nil
}

// CheckProjectAccess verifies if a user has access to upload to a project
func (r *DatasetRepository) CheckProjectAccess(projectID, userID uuid.UUID) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM projects 
		WHERE id = $1 AND owner_id = $2`

	err := r.db.Get(&count, query, projectID, userID)
	if err != nil {
		return false, err
	}

	// TODO: Also check project_members table for collaborator access
	// For now, only project owner can upload
	return count > 0, nil
}
