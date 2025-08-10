package models

import (
	"time"

	"github.com/google/uuid"
)

// ProjectMember represents a user's membership in a project
type ProjectMember struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	ProjectID   uuid.UUID              `json:"project_id" db:"project_id"`
	UserID      uuid.UUID              `json:"user_id" db:"user_id"`
	Role        string                 `json:"role" db:"role"` // owner, admin, collaborator, viewer
	InvitedBy   *uuid.UUID             `json:"invited_by,omitempty" db:"invited_by"`
	InvitedAt   time.Time              `json:"invited_at" db:"invited_at"`
	JoinedAt    *time.Time             `json:"joined_at,omitempty" db:"joined_at"`
	Status      string                 `json:"status" db:"status"` // pending, accepted, declined, removed
	Permissions map[string]interface{} `json:"permissions" db:"permissions"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// ProjectMemberWithUser includes user information
type ProjectMemberWithUser struct {
	ProjectMember
	UserName  string `json:"user_name" db:"user_name"`
	UserEmail string `json:"user_email" db:"user_email"`
}

// InviteUserRequest represents a request to invite a user to a project
type InviteUserRequest struct {
	Email       string                 `json:"email" binding:"required,email"`
	Role        string                 `json:"role" binding:"required"`
	Permissions map[string]interface{} `json:"permissions,omitempty"`
}

// UpdateMemberRoleRequest represents a request to update a member's role
type UpdateMemberRoleRequest struct {
	Role        string                 `json:"role" binding:"required"`
	Permissions map[string]interface{} `json:"permissions,omitempty"`
}

// ProjectWithMembers includes project information with member details
type ProjectWithMembers struct {
	Project
	Members []ProjectMemberWithUser `json:"members"`
}

// ValidateRole checks if the role is valid
func (r *InviteUserRequest) ValidateRole() bool {
	validRoles := map[string]bool{
		"admin":        true,
		"collaborator": true,
		"viewer":       true,
	}
	return validRoles[r.Role]
}

// ValidateRole checks if the role is valid for updates
func (r *UpdateMemberRoleRequest) ValidateRole() bool {
	validRoles := map[string]bool{
		"admin":        true,
		"collaborator": true,
		"viewer":       true,
	}
	return validRoles[r.Role]
}

// CanManageMembers checks if a user role can manage other members
func CanManageMembers(role string) bool {
	return role == "owner" || role == "admin"
}

// CanEditProject checks if a user role can edit the project
func CanEditProject(role string) bool {
	return role == "owner" || role == "admin" || role == "collaborator"
}

// CanViewProject checks if a user role can view the project
func CanViewProject(role string) bool {
	return role == "owner" || role == "admin" || role == "collaborator" || role == "viewer"
}
