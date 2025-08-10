package models

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Project represents a project in the system
type Project struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	OwnerID     uuid.UUID `json:"owner_id" db:"owner_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// CreateProjectRequest represents the request to create a new project
type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=255"`
	Description string `json:"description" binding:"max=1000"`
}

// UpdateProjectRequest represents the request to update a project
type UpdateProjectRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=1000"`
}

// Validate validates the create project request
func (req *CreateProjectRequest) Validate() error {
	// Trim whitespace
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)

	// Check required fields
	if req.Name == "" {
		return errors.New("project name is required")
	}

	// Check name length
	if len(req.Name) < 1 || len(req.Name) > 255 {
		return errors.New("project name must be between 1 and 255 characters")
	}

	// Check description length
	if len(req.Description) > 1000 {
		return errors.New("project description must be less than 1000 characters")
	}

	return nil
}

// ToProject converts a CreateProjectRequest to a Project
func (req *CreateProjectRequest) ToProject(ownerID uuid.UUID) *Project {
	return &Project{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     ownerID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// Validate validates the update project request
func (req *UpdateProjectRequest) Validate() error {
	// Check name if provided
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if len(name) < 1 || len(name) > 255 {
			return errors.New("project name must be between 1 and 255 characters")
		}
		*req.Name = name
	}

	// Check description if provided
	if req.Description != nil {
		description := strings.TrimSpace(*req.Description)
		if len(description) > 1000 {
			return errors.New("project description must be less than 1000 characters")
		}
		*req.Description = description
	}

	return nil
}

// HasUpdates checks if the update request has any actual updates
func (req *UpdateProjectRequest) HasUpdates() bool {
	return req.Name != nil || req.Description != nil
}
