package repository

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/saurabh22suman/oreo.io/internal/models"
)

// ProjectRepository handles project database operations
type ProjectRepository struct {
	db *sqlx.DB
}

// NewProjectRepository creates a new project repository
func NewProjectRepository(db *sqlx.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// Create creates a new project
func (r *ProjectRepository) Create(project *models.Project) error {
	query := `
		INSERT INTO projects (id, name, description, owner_id, created_at, updated_at)
		VALUES (:id, :name, :description, :owner_id, :created_at, :updated_at)`

	_, err := r.db.NamedExec(query, project)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	return nil
}

// GetByID retrieves a project by ID
func (r *ProjectRepository) GetByID(id uuid.UUID) (*models.Project, error) {
	var project models.Project
	query := `SELECT id, name, description, owner_id, created_at, updated_at FROM projects WHERE id = $1`

	err := r.db.Get(&project, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project not found")
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return &project, nil
}

// GetByOwnerID retrieves all projects owned by a user
func (r *ProjectRepository) GetByOwnerID(ownerID uuid.UUID) ([]*models.Project, error) {
	var projects []*models.Project
	query := `
		SELECT id, name, description, owner_id, created_at, updated_at 
		FROM projects 
		WHERE owner_id = $1 
		ORDER BY created_at DESC`

	err := r.db.Select(&projects, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects for owner: %w", err)
	}

	return projects, nil
}

// Update updates a project
func (r *ProjectRepository) Update(id uuid.UUID, updates *models.UpdateProjectRequest) (*models.Project, error) {
	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if updates.Name != nil {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, *updates.Name)
		argIndex++
	}

	if updates.Description != nil {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *updates.Description)
		argIndex++
	}

	if len(setParts) == 0 {
		// No updates to perform, just return the current project
		return r.GetByID(id)
	}

	// Add updated_at
	setParts = append(setParts, fmt.Sprintf("updated_at = CURRENT_TIMESTAMP"))

	// Build the query
	query := fmt.Sprintf(
		"UPDATE projects SET %s WHERE id = $%d RETURNING id, name, description, owner_id, created_at, updated_at",
		fmt.Sprintf("%s", setParts[0]),
		argIndex,
	)

	// Join all SET parts
	if len(setParts) > 1 {
		setClause := ""
		for i, part := range setParts {
			if i > 0 {
				setClause += ", "
			}
			setClause += part
		}
		query = fmt.Sprintf(
			"UPDATE projects SET %s WHERE id = $%d RETURNING id, name, description, owner_id, created_at, updated_at",
			setClause,
			argIndex,
		)
	}

	args = append(args, id)

	var project models.Project
	err := r.db.Get(&project, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project not found")
		}
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return &project, nil
}

// Delete deletes a project
func (r *ProjectRepository) Delete(id uuid.UUID, ownerID uuid.UUID) error {
	query := `DELETE FROM projects WHERE id = $1 AND owner_id = $2`

	result, err := r.db.Exec(query, id, ownerID)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("project not found or not owned by user")
	}

	return nil
}

// Exists checks if a project exists and is owned by the user
func (r *ProjectRepository) Exists(id uuid.UUID, ownerID uuid.UUID) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM projects WHERE id = $1 AND owner_id = $2`

	err := r.db.Get(&count, query, id, ownerID)
	if err != nil {
		return false, fmt.Errorf("failed to check if project exists: %w", err)
	}

	return count > 0, nil
}
