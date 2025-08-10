package models

import (
	"time"

	"github.com/google/uuid"
)

// Dataset represents a data file uploaded to a project
type Dataset struct {
	ID          uuid.UUID `json:"id" db:"id"`
	ProjectID   uuid.UUID `json:"project_id" db:"project_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	FileName    string    `json:"file_name" db:"file_name"`
	FilePath    string    `json:"file_path" db:"file_path"`
	FileSize    int64     `json:"file_size" db:"file_size"`
	MimeType    string    `json:"mime_type" db:"mime_type"`
	RowCount    int       `json:"row_count" db:"row_count"`
	ColumnCount int       `json:"column_count" db:"column_count"`
	Status      string    `json:"status" db:"status"` // "processing", "ready", "error"
	UploadedBy  uuid.UUID `json:"uploaded_by" db:"uploaded_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// DatasetWithProject includes project information
type DatasetWithProject struct {
	Dataset
	ProjectName string `json:"project_name" db:"project_name"`
}

// CreateDatasetRequest represents the request to create a new dataset
type CreateDatasetRequest struct {
	ProjectID   uuid.UUID `json:"project_id" binding:"required"`
	Name        string    `json:"name" binding:"required,min=1,max=255"`
	Description string    `json:"description" binding:"max=1000"`
}

// UpdateDatasetRequest represents the request to update a dataset
type UpdateDatasetRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=255"`
	Description string `json:"description" binding:"max=1000"`
}

// DatasetStatus constants
const (
	DatasetStatusProcessing = "processing"
	DatasetStatusReady      = "ready"
	DatasetStatusError      = "error"
)
