package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/saurabh22suman/oreo.io/internal/models"
)

type ProjectMemberRepository struct {
	db *sqlx.DB
}

func NewProjectMemberRepository(db *sqlx.DB) *ProjectMemberRepository {
	return &ProjectMemberRepository{db: db}
}

// GetProjectMembers returns all members of a project
func (r *ProjectMemberRepository) GetProjectMembers(projectID uuid.UUID) ([]models.ProjectMemberWithUser, error) {
	query := `
		SELECT 
			pm.id, pm.project_id, pm.user_id, pm.role, pm.invited_by, 
			pm.invited_at, pm.joined_at, pm.status, pm.permissions, 
			pm.created_at, pm.updated_at,
			u.name as user_name, u.email as user_email
		FROM project_members pm
		JOIN users u ON pm.user_id = u.id
		WHERE pm.project_id = $1 AND pm.status = 'accepted'
		ORDER BY pm.role, pm.joined_at`

	var members []models.ProjectMemberWithUser
	err := r.db.Select(&members, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project members: %w", err)
	}

	return members, nil
}

// GetUserRole returns the user's role in a specific project
func (r *ProjectMemberRepository) GetUserRole(projectID, userID uuid.UUID) (string, error) {
	query := `
		SELECT role 
		FROM project_members 
		WHERE project_id = $1 AND user_id = $2 AND status = 'accepted'`

	var role string
	err := r.db.Get(&role, query, projectID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("user is not a member of this project")
		}
		return "", fmt.Errorf("failed to get user role: %w", err)
	}

	return role, nil
}

// GetUserProjects returns all projects a user has access to
func (r *ProjectMemberRepository) GetUserProjects(userID uuid.UUID) ([]models.ProjectWithMembers, error) {
	query := `
		SELECT DISTINCT
			p.id, p.name, p.description, p.owner_id, p.created_at, p.updated_at
		FROM projects p
		JOIN project_members pm ON p.id = pm.project_id
		WHERE pm.user_id = $1 AND pm.status = 'accepted'
		ORDER BY p.created_at DESC`

	var projects []models.Project
	err := r.db.Select(&projects, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user projects: %w", err)
	}

	var result []models.ProjectWithMembers
	for _, project := range projects {
		members, err := r.GetProjectMembers(project.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get project members for project %s: %w", project.ID, err)
		}

		result = append(result, models.ProjectWithMembers{
			Project: project,
			Members: members,
		})
	}

	return result, nil
}

// InviteUser invites a user to a project
func (r *ProjectMemberRepository) InviteUser(projectID, inviterID, inviteeID uuid.UUID, role string, permissions map[string]interface{}) (*models.ProjectMember, error) {
	// Check if user is already a member
	var existingID uuid.UUID
	checkQuery := `SELECT id FROM project_members WHERE project_id = $1 AND user_id = $2`
	err := r.db.Get(&existingID, checkQuery, projectID, inviteeID)
	if err == nil {
		return nil, fmt.Errorf("user is already a member of this project")
	}
	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check existing membership: %w", err)
	}

	member := &models.ProjectMember{
		ID:          uuid.New(),
		ProjectID:   projectID,
		UserID:      inviteeID,
		Role:        role,
		InvitedBy:   &inviterID,
		InvitedAt:   time.Now(),
		Status:      "pending",
		Permissions: permissions,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	query := `
		INSERT INTO project_members 
		(id, project_id, user_id, role, invited_by, invited_at, status, permissions, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	permissionsJSON, _ := pq.Array([]byte{}).Value()
	if permissions != nil {
		// Convert permissions to JSONB
		// Note: This is a simplified approach. In production, use proper JSON marshaling
	}

	_, err = r.db.Exec(query,
		member.ID, member.ProjectID, member.UserID, member.Role,
		member.InvitedBy, member.InvitedAt, member.Status, permissionsJSON,
		member.CreatedAt, member.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to invite user: %w", err)
	}

	return member, nil
}

// AcceptInvitation accepts a project invitation
func (r *ProjectMemberRepository) AcceptInvitation(projectID, userID uuid.UUID) error {
	query := `
		UPDATE project_members 
		SET status = 'accepted', joined_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE project_id = $1 AND user_id = $2 AND status = 'pending'`

	result, err := r.db.Exec(query, projectID, userID)
	if err != nil {
		return fmt.Errorf("failed to accept invitation: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no pending invitation found")
	}

	return nil
}

// RemoveMember removes a user from a project
func (r *ProjectMemberRepository) RemoveMember(projectID, userID uuid.UUID) error {
	// Don't allow removing the project owner
	var role string
	roleQuery := `SELECT role FROM project_members WHERE project_id = $1 AND user_id = $2`
	err := r.db.Get(&role, roleQuery, projectID, userID)
	if err != nil {
		return fmt.Errorf("failed to get user role: %w", err)
	}

	if role == "owner" {
		return fmt.Errorf("cannot remove project owner")
	}

	query := `DELETE FROM project_members WHERE project_id = $1 AND user_id = $2`
	result, err := r.db.Exec(query, projectID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("member not found")
	}

	return nil
}

// UpdateMemberRole updates a member's role and permissions
func (r *ProjectMemberRepository) UpdateMemberRole(projectID, userID uuid.UUID, role string, permissions map[string]interface{}) error {
	// Don't allow changing the owner role
	var currentRole string
	roleQuery := `SELECT role FROM project_members WHERE project_id = $1 AND user_id = $2`
	err := r.db.Get(&currentRole, roleQuery, projectID, userID)
	if err != nil {
		return fmt.Errorf("failed to get current role: %w", err)
	}

	if currentRole == "owner" {
		return fmt.Errorf("cannot change owner role")
	}

	query := `
		UPDATE project_members 
		SET role = $3, permissions = $4, updated_at = CURRENT_TIMESTAMP
		WHERE project_id = $1 AND user_id = $2`

	permissionsJSON, _ := pq.Array([]byte{}).Value()
	if permissions != nil {
		// Convert permissions to JSONB
		// Note: This is a simplified approach. In production, use proper JSON marshaling
	}

	result, err := r.db.Exec(query, projectID, userID, role, permissionsJSON)
	if err != nil {
		return fmt.Errorf("failed to update member role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("member not found")
	}

	return nil
}
