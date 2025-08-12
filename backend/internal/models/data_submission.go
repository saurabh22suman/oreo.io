package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// DataSubmission represen// DataValidationError represents a specific validation error during data submission
type DataValidationError struct {
	RowIndex      int    `json:"row_index"`
	FieldName     string `json:"field_name"`
	ErrorType     string `json:"error_type"`
	Message       string `json:"message"`
	ActualValue   string `json:"actual_value"`
	ExpectedValue string `json:"expected_value,omitempty"`
}

// DataSubmission represents a request to append data to an existing dataset
type DataSubmission struct {
	ID                uuid.UUID              `json:"id" db:"id"`
	DatasetID         uuid.UUID              `json:"dataset_id" db:"dataset_id"`
	SubmittedBy       uuid.UUID              `json:"submitted_by" db:"submitted_by"`
	FileName          string                 `json:"file_name" db:"file_name"`
	FilePath          string                 `json:"file_path" db:"file_path"`
	FileSize          int64                  `json:"file_size" db:"file_size"`
	RowCount          int                    `json:"row_count" db:"row_count"`
	Status            string                 `json:"status" db:"status"`
	ValidationResults *json.RawMessage       `json:"validation_results" db:"validation_results"`
	AdminNotes        *string                `json:"admin_notes" db:"admin_notes"`
	ReviewedBy        *uuid.UUID             `json:"reviewed_by" db:"reviewed_by"`
	ReviewedAt        *time.Time             `json:"reviewed_at" db:"reviewed_at"`
	SubmittedAt       time.Time              `json:"submitted_at" db:"submitted_at"`
	AppliedAt         *time.Time             `json:"applied_at" db:"applied_at"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
}

// DataSubmissionWithDetails includes additional details for display
type DataSubmissionWithDetails struct {
	DataSubmission
	DatasetName      string `json:"dataset_name" db:"dataset_name"`
	ProjectName      string `json:"project_name" db:"project_name"`
	SubmitterName    string `json:"submitter_name" db:"submitter_name"`
	SubmitterEmail   string `json:"submitter_email" db:"submitter_email"`
	ReviewerName     *string `json:"reviewer_name" db:"reviewer_name"`
}

// DataSubmissionStaging represents staged data before approval
type DataSubmissionStaging struct {
	ID               uuid.UUID        `json:"id" db:"id"`
	SubmissionID     uuid.UUID        `json:"submission_id" db:"submission_id"`
	RowIndex         int              `json:"row_index" db:"row_index"`
	Data             json.RawMessage  `json:"data" db:"data"`
	ValidationStatus string           `json:"validation_status" db:"validation_status"`
	ValidationErrors *json.RawMessage `json:"validation_errors" db:"validation_errors"`
	CreatedAt        time.Time        `json:"created_at" db:"created_at"`
}

// DatasetBusinessRule represents validation rules for datasets
type DatasetBusinessRule struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	DatasetID    uuid.UUID       `json:"dataset_id" db:"dataset_id"`
	RuleName     string          `json:"rule_name" db:"rule_name"`
	RuleType     string          `json:"rule_type" db:"rule_type"`
	RuleConfig   json.RawMessage `json:"rule_config" db:"rule_config"`
	ErrorMessage string          `json:"error_message" db:"error_message"`
	IsActive     bool            `json:"is_active" db:"is_active"`
	Priority     int             `json:"priority" db:"priority"`
	CreatedBy    uuid.UUID       `json:"created_by" db:"created_by"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at" db:"updated_at"`
}

// DataSubmissionStatus constants
const (
	DataSubmissionStatusPending     = "pending"
	DataSubmissionStatusUnderReview = "under_review"
	DataSubmissionStatusApproved    = "approved"
	DataSubmissionStatusRejected    = "rejected"
	DataSubmissionStatusApplied     = "applied"
)

// ValidationStatus constants for staging data
const (
	ValidationStatusValid   = "valid"
	ValidationStatusInvalid = "invalid"
	ValidationStatusWarning = "warning"
)

// Business rule types
const (
	RuleTypeFieldValidation = "field_validation"
	RuleTypeCrossField      = "cross_field"
	RuleTypeCustomSQL       = "custom_sql"
	RuleTypeRangeCheck      = "range_check"
	RuleTypeUnique          = "unique"
	RuleTypeRequired        = "required"
)

// CreateDataSubmissionRequest represents the request to submit new data
type CreateDataSubmissionRequest struct {
	DatasetID uuid.UUID `json:"dataset_id" binding:"required"`
	FileName  string    `json:"file_name" binding:"required"`
}

// UpdateDataSubmissionRequest represents admin update to submission
type UpdateDataSubmissionRequest struct {
	Status     string  `json:"status" binding:"required,oneof=under_review approved rejected"`
	AdminNotes *string `json:"admin_notes"`
}

// ValidationResult represents the result of validating a data submission
type ValidationResult struct {
	IsValid            bool                   `json:"is_valid"`
	TotalRows          int                    `json:"total_rows"`
	ValidRows          int                    `json:"valid_rows"`
	InvalidRows        int                    `json:"invalid_rows"`
	WarningRows        int                    `json:"warning_rows"`
	SchemaErrors       []DataValidationError  `json:"schema_errors"`
	BusinessRuleErrors []DataValidationError  `json:"business_rule_errors"`
	FieldStats         map[string]FieldStats  `json:"field_stats"`
}

// FieldStats represents statistics for a field during validation
type FieldStats struct {
	TotalValues   int `json:"total_values"`
	UniqueValues  int `json:"unique_values"`
	NullValues    int `json:"null_values"`
	InvalidValues int `json:"invalid_values"`
}

// BusinessRuleConfig represents configuration for different rule types
type BusinessRuleConfig struct {
	// For field validation rules
	FieldName    string      `json:"field_name,omitempty"`
	DataType     string      `json:"data_type,omitempty"`
	MinValue     interface{} `json:"min_value,omitempty"`
	MaxValue     interface{} `json:"max_value,omitempty"`
	Pattern      string      `json:"pattern,omitempty"`
	AllowedValues []string   `json:"allowed_values,omitempty"`
	
	// For cross-field validation
	Fields       []string    `json:"fields,omitempty"`
	Condition    string      `json:"condition,omitempty"`
	
	// For custom SQL validation  
	Query        string      `json:"query,omitempty"`
	Parameters   []string    `json:"parameters,omitempty"`
}
