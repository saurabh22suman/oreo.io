package models

import (
	"time"
	"github.com/google/uuid"
)

// SchemaFieldType represents the data type of a schema field
type SchemaFieldType string

const (
	FieldTypeString   SchemaFieldType = "string"
	FieldTypeNumber   SchemaFieldType = "number"
	FieldTypeBoolean  SchemaFieldType = "boolean"
	FieldTypeDate     SchemaFieldType = "date"
	FieldTypeDateTime SchemaFieldType = "datetime"
	FieldTypeEmail    SchemaFieldType = "email"
	FieldTypeURL      SchemaFieldType = "url"
	FieldTypeUUID     SchemaFieldType = "uuid"
)

// DatasetSchema represents the schema definition for a dataset
type DatasetSchema struct {
	ID          uuid.UUID      `json:"id" db:"id"`
	DatasetID   uuid.UUID      `json:"dataset_id" db:"dataset_id"`
	Name        string         `json:"name" db:"name"`
	Description string         `json:"description" db:"description"`
	Fields      []SchemaField  `json:"fields"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
}

// SchemaField represents a field definition in a dataset schema
type SchemaField struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	SchemaID     uuid.UUID       `json:"schema_id" db:"schema_id"`
	Name         string          `json:"name" db:"name"`
	DisplayName  string          `json:"display_name" db:"display_name"`
	DataType     string          `json:"data_type" db:"data_type"` // Will store string values from SchemaFieldType
	IsRequired   bool            `json:"is_required" db:"is_required"`
	IsUnique     bool            `json:"is_unique" db:"is_unique"`
	DefaultValue *string         `json:"default_value" db:"default_value"`
	Position     int             `json:"position" db:"position"`
	Validation   FieldValidation `json:"validation"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at" db:"updated_at"`
}

// FieldValidation represents validation rules for a schema field
type FieldValidation struct {
	MinLength   *int     `json:"min_length,omitempty"`
	MaxLength   *int     `json:"max_length,omitempty"`
	MinValue    *float64 `json:"min_value,omitempty"`
	MaxValue    *float64 `json:"max_value,omitempty"`
	Pattern     *string  `json:"pattern,omitempty"`
	Options     []string `json:"options,omitempty"` // For enum/select fields
	Format      *string  `json:"format,omitempty"`  // date format, etc.
}

// DatasetData represents the actual data rows in a dataset
type DatasetData struct {
	ID        uuid.UUID              `json:"id" db:"id"`
	DatasetID uuid.UUID              `json:"dataset_id" db:"dataset_id"`
	RowIndex  int                    `json:"row_index" db:"row_index"`
	Data      map[string]interface{} `json:"data" db:"data"` // JSONB column
	Version   int                    `json:"version" db:"version"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt time.Time              `json:"updated_at" db:"updated_at"`
	CreatedBy uuid.UUID              `json:"created_by" db:"created_by"`
	UpdatedBy uuid.UUID              `json:"updated_by" db:"updated_by"`
}

// CreateSchemaRequest represents the request to create a new schema
type CreateSchemaRequest struct {
	DatasetID   uuid.UUID             `json:"dataset_id" binding:"required"`
	Name        string                `json:"name" binding:"required"`
	Description string                `json:"description"`
	Fields      []CreateFieldRequest  `json:"fields" binding:"required"`
}

// CreateFieldRequest represents the request to create a new field
type CreateFieldRequest struct {
	Name         string          `json:"name" binding:"required"`
	DisplayName  string          `json:"display_name"`
	DataType     string          `json:"data_type" binding:"required"`
	IsRequired   bool            `json:"is_required"`
	IsUnique     bool            `json:"is_unique"`
	DefaultValue *string         `json:"default_value"`
	Position     int             `json:"position"`
	Validation   FieldValidation `json:"validation"`
}

// UpdateSchemaRequest represents the request to update a schema
type UpdateSchemaRequest struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Fields      []UpdateFieldRequest  `json:"fields"`
}

// UpdateFieldRequest represents the request to update a field
type UpdateFieldRequest struct {
	ID           uuid.UUID       `json:"id"`
	Name         string          `json:"name"`
	DisplayName  string          `json:"display_name"`
	DataType     string          `json:"data_type"`
	IsRequired   bool            `json:"is_required"`
	IsUnique     bool            `json:"is_unique"`
	DefaultValue *string         `json:"default_value"`
	Position     int             `json:"position"`
	Validation   FieldValidation `json:"validation"`
}

// DataPreviewRequest represents request for data preview
type DataPreviewRequest struct {
	Page     int `json:"page" form:"page"`
	PageSize int `json:"page_size" form:"page_size"`
}

// DataPreviewResponse represents the response for data preview
type DataPreviewResponse struct {
	Data        []map[string]interface{} `json:"data"`
	Schema      *DatasetSchema           `json:"schema"`
	TotalRows   int                      `json:"total"`
	Page        int                      `json:"page"`
	PageSize    int                      `json:"page_size"`
	TotalPages  int                      `json:"total_pages"`
}

// UpdateDataRequest represents request to update dataset data
type UpdateDataRequest struct {
	RowIndex int                    `json:"row_index" binding:"required"`
	Data     map[string]interface{} `json:"data" binding:"required"`
}

// ValidationError represents a schema validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   interface{} `json:"value"`
}

// SchemaValidationResult represents the result of schema validation
type SchemaValidationResult struct {
	IsValid bool              `json:"is_valid"`
	Errors  []ValidationError `json:"errors"`
}
