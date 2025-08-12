package services

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/saurabh22suman/oreo.io/internal/models"
)

type ValidationService struct {
	schemaRepo         SchemaRepositoryInterface
	submissionRepo     DataSubmissionRepositoryInterface
}

func NewValidationService(schemaRepo SchemaRepositoryInterface, submissionRepo DataSubmissionRepositoryInterface) *ValidationService {
	return &ValidationService{
		schemaRepo:     schemaRepo,
		submissionRepo: submissionRepo,
	}
}

// hasValidationRules checks if a FieldValidation struct has any validation rules set
func (v *ValidationService) hasValidationRules(validation models.FieldValidation) bool {
	return validation.MinLength != nil || validation.MaxLength != nil ||
		validation.MinValue != nil || validation.MaxValue != nil ||
		validation.Pattern != nil || len(validation.Options) > 0 ||
		validation.Format != nil
}

type SchemaRepositoryInterface interface {
	GetSchemaByDatasetID(datasetID uuid.UUID) (*models.DatasetSchema, error)
}

type DataSubmissionRepositoryInterface interface {
	GetBusinessRules(datasetID uuid.UUID) ([]*models.DatasetBusinessRule, error)
}

// ValidateDataSubmission validates an uploaded file against dataset schema and business rules
func (v *ValidationService) ValidateDataSubmission(filePath string, datasetID uuid.UUID) (*models.ValidationResult, []*models.DataSubmissionStaging, error) {
	// Load dataset schema
	schema, err := v.schemaRepo.GetSchemaByDatasetID(datasetID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load schema: %w", err)
	}

	// Load business rules
	businessRules, err := v.submissionRepo.GetBusinessRules(datasetID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load business rules: %w", err)
	}

	// Parse CSV file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	
	// Read header
	headers, err := reader.Read()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read headers: %w", err)
	}

	// Validate headers against schema
	headerValidation := v.validateHeaders(headers, schema)
	if !headerValidation.IsValid {
		return headerValidation, nil, nil
	}

	// Read and validate data rows
	validationResult := &models.ValidationResult{
		IsValid:            true,
		TotalRows:          0,
		ValidRows:          0,
		InvalidRows:        0,
		WarningRows:        0,
		SchemaErrors:       []models.DataValidationError{},
		BusinessRuleErrors: []models.DataValidationError{},
		FieldStats:         make(map[string]models.FieldStats),
	}

	var stagingData []*models.DataSubmissionStaging
	var allRowData []map[string]interface{}

	// Initialize field stats
	for _, field := range schema.Fields {
		validationResult.FieldStats[field.Name] = models.FieldStats{
			TotalValues:   0,
			UniqueValues:  0,
			NullValues:    0,
			InvalidValues: 0,
		}
	}

	rowIndex := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read row %d: %w", rowIndex, err)
		}

		validationResult.TotalRows++

		// Convert row to map
		rowData := make(map[string]interface{})
		for i, header := range headers {
			if i < len(record) {
				rowData[header] = record[i]
			} else {
				rowData[header] = ""
			}
		}

		// Validate row against schema
		rowValidation := v.validateRowAgainstSchema(rowData, schema, rowIndex)
		validationResult.SchemaErrors = append(validationResult.SchemaErrors, rowValidation.Errors...)

		// Update field statistics
		v.updateFieldStats(rowData, schema, validationResult.FieldStats)

		// Store row data for business rule validation
		allRowData = append(allRowData, rowData)

		// Create staging data
		dataJSON, _ := json.Marshal(rowData)
		validationErrors, _ := json.Marshal(rowValidation.Errors)
		
		validationStatus := models.ValidationStatusValid
		if len(rowValidation.Errors) > 0 {
			validationStatus = models.ValidationStatusInvalid
			validationResult.InvalidRows++
		} else {
			validationResult.ValidRows++
		}

		validationErrorsJSON := json.RawMessage(validationErrors)
		stagingRow := &models.DataSubmissionStaging{
			ID:               uuid.New(),
			RowIndex:         rowIndex,
			Data:             dataJSON,
			ValidationStatus: validationStatus,
			ValidationErrors: &validationErrorsJSON,
			CreatedAt:        time.Now(),
		}

		stagingData = append(stagingData, stagingRow)
		rowIndex++
	}

	// Validate business rules across all data
	businessRuleErrors := v.validateBusinessRules(allRowData, businessRules)
	validationResult.BusinessRuleErrors = businessRuleErrors

	// Update validation status based on business rule errors
	for _, err := range businessRuleErrors {
		if err.RowIndex < len(stagingData) {
			currentErrors := []models.DataValidationError{}
			if stagingData[err.RowIndex].ValidationErrors != nil {
				json.Unmarshal(*stagingData[err.RowIndex].ValidationErrors, &currentErrors)
			}
			currentErrors = append(currentErrors, err)
			
			updatedErrors, _ := json.Marshal(currentErrors)
			updatedErrorsJSON := json.RawMessage(updatedErrors)
			stagingData[err.RowIndex].ValidationErrors = &updatedErrorsJSON
			
			if stagingData[err.RowIndex].ValidationStatus == models.ValidationStatusValid {
				stagingData[err.RowIndex].ValidationStatus = models.ValidationStatusInvalid
				validationResult.ValidRows--
				validationResult.InvalidRows++
			}
		}
	}

	// Calculate unique values for field stats
	v.calculateUniqueValues(allRowData, validationResult.FieldStats)

	// Overall validation status
	validationResult.IsValid = validationResult.InvalidRows == 0

	return validationResult, stagingData, nil
}

// validateHeaders checks if uploaded headers match schema fields
func (v *ValidationService) validateHeaders(headers []string, schema *models.DatasetSchema) *models.ValidationResult {
	result := &models.ValidationResult{
		IsValid:            true,
		SchemaErrors:       []models.DataValidationError{},
		BusinessRuleErrors: []models.DataValidationError{},
	}

	schemaFields := make(map[string]bool)
	for _, field := range schema.Fields {
		schemaFields[field.Name] = true
	}

	// Check for missing required fields
	for _, field := range schema.Fields {
		found := false
		for _, header := range headers {
			if header == field.Name {
				found = true
				break
			}
		}
		if !found {
			result.SchemaErrors = append(result.SchemaErrors, models.DataValidationError{
				RowIndex:    -1, // Header validation
				FieldName:   field.Name,
				ErrorType:   "missing_field",
				Message:     fmt.Sprintf("Required field '%s' is missing from uploaded data", field.Name),
			})
			result.IsValid = false
		}
	}

	// Check for unexpected fields
	for _, header := range headers {
		if !schemaFields[header] {
			result.SchemaErrors = append(result.SchemaErrors, models.DataValidationError{
				RowIndex:    -1, // Header validation
				FieldName:   header,
				ErrorType:   "unexpected_field",
				Message:     fmt.Sprintf("Field '%s' is not defined in the dataset schema", header),
			})
		}
	}

	return result
}

// validateRowAgainstSchema validates a single row against the schema
func (v *ValidationService) validateRowAgainstSchema(rowData map[string]interface{}, schema *models.DatasetSchema, rowIndex int) *rowValidationResult {
	result := &rowValidationResult{
		Errors: []models.DataValidationError{},
	}

	for _, field := range schema.Fields {
		value, exists := rowData[field.Name]
		
		// Check required fields
		if field.IsRequired && (!exists || value == "" || value == nil) {
			result.Errors = append(result.Errors, models.DataValidationError{
				RowIndex:    rowIndex,
				FieldName:   field.Name,
				ErrorType:   "required_field",
				Message:     fmt.Sprintf("Required field '%s' cannot be empty", field.Name),
				ActualValue: fmt.Sprintf("%v", value),
			})
			continue
		}

		// Skip validation for empty optional fields
		if !exists || value == "" || value == nil {
			continue
		}

		// Validate data type
		if err := v.validateDataType(value, field, rowIndex); err != nil {
			result.Errors = append(result.Errors, *err)
		}

		// Validate field-specific rules from validation config
		if v.hasValidationRules(field.Validation) {
			if errs := v.validateFieldRules(value, field, rowIndex); len(errs) > 0 {
				result.Errors = append(result.Errors, errs...)
			}
		}
	}

	return result
}

type rowValidationResult struct {
	Errors []models.DataValidationError
}

// validateDataType validates the data type of a field value
func (v *ValidationService) validateDataType(value interface{}, field models.SchemaField, rowIndex int) *models.DataValidationError {
	valueStr := fmt.Sprintf("%v", value)
	
	switch field.DataType {
	case "number":
		if _, err := strconv.ParseFloat(valueStr, 64); err != nil {
			return &models.DataValidationError{
				RowIndex:      rowIndex,
				FieldName:     field.Name,
				ErrorType:     "invalid_data_type",
				Message:       fmt.Sprintf("Field '%s' must be a number", field.Name),
				ActualValue:   valueStr,
				ExpectedValue: "number",
			}
		}
	case "boolean":
		lowerValue := strings.ToLower(valueStr)
		if lowerValue != "true" && lowerValue != "false" && lowerValue != "1" && lowerValue != "0" {
			return &models.DataValidationError{
				RowIndex:      rowIndex,
				FieldName:     field.Name,
				ErrorType:     "invalid_data_type",
				Message:       fmt.Sprintf("Field '%s' must be a boolean (true/false)", field.Name),
				ActualValue:   valueStr,
				ExpectedValue: "true/false",
			}
		}
	case "date":
		// Try common date formats
		formats := []string{
			"2006-01-02",
			"2006-01-02 15:04:05",
			"01/02/2006",
			"02-01-2006",
		}
		
		valid := false
		for _, format := range formats {
			if _, err := time.Parse(format, valueStr); err == nil {
				valid = true
				break
			}
		}
		
		if !valid {
			return &models.DataValidationError{
				RowIndex:      rowIndex,
				FieldName:     field.Name,
				ErrorType:     "invalid_data_type",
				Message:       fmt.Sprintf("Field '%s' must be a valid date", field.Name),
				ActualValue:   valueStr,
				ExpectedValue: "YYYY-MM-DD or MM/DD/YYYY",
			}
		}
	case "email":
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(valueStr) {
			return &models.DataValidationError{
				RowIndex:      rowIndex,
				FieldName:     field.Name,
				ErrorType:     "invalid_data_type",
				Message:       fmt.Sprintf("Field '%s' must be a valid email address", field.Name),
				ActualValue:   valueStr,
				ExpectedValue: "valid email format",
			}
		}
	}

	return nil
}

// validateFieldRules validates field-specific validation rules
func (v *ValidationService) validateFieldRules(value interface{}, field models.SchemaField, rowIndex int) []models.DataValidationError {
	var errors []models.DataValidationError
	valueStr := fmt.Sprintf("%v", value)
	
	validation := field.Validation

	// String length validation
	if field.DataType == "string" {
		if validation.MinLength != nil && len(valueStr) < *validation.MinLength {
			errors = append(errors, models.DataValidationError{
				RowIndex:      rowIndex,
				FieldName:     field.Name,
				ErrorType:     "min_length",
				Message:       fmt.Sprintf("Field '%s' must be at least %d characters", field.Name, *validation.MinLength),
				ActualValue:   valueStr,
				ExpectedValue: fmt.Sprintf("min %d chars", *validation.MinLength),
			})
		}
		if validation.MaxLength != nil && len(valueStr) > *validation.MaxLength {
			errors = append(errors, models.DataValidationError{
				RowIndex:      rowIndex,
				FieldName:     field.Name,
				ErrorType:     "max_length",
				Message:       fmt.Sprintf("Field '%s' must be at most %d characters", field.Name, *validation.MaxLength),
				ActualValue:   valueStr,
				ExpectedValue: fmt.Sprintf("max %d chars", *validation.MaxLength),
			})
		}
	}

	// Numeric range validation
	if field.DataType == "number" {
		if floatVal, err := strconv.ParseFloat(valueStr, 64); err == nil {
			if validation.MinValue != nil && floatVal < *validation.MinValue {
				errors = append(errors, models.DataValidationError{
					RowIndex:      rowIndex,
					FieldName:     field.Name,
					ErrorType:     "min_value",
					Message:       fmt.Sprintf("Field '%s' must be at least %f", field.Name, *validation.MinValue),
					ActualValue:   valueStr,
					ExpectedValue: fmt.Sprintf("min %f", *validation.MinValue),
				})
			}
			if validation.MaxValue != nil && floatVal > *validation.MaxValue {
				errors = append(errors, models.DataValidationError{
					RowIndex:      rowIndex,
					FieldName:     field.Name,
					ErrorType:     "max_value",
					Message:       fmt.Sprintf("Field '%s' must be at most %f", field.Name, *validation.MaxValue),
					ActualValue:   valueStr,
					ExpectedValue: fmt.Sprintf("max %f", *validation.MaxValue),
				})
			}
		}
	}

	// Pattern validation
	if validation.Pattern != nil {
		if matched, _ := regexp.MatchString(*validation.Pattern, valueStr); !matched {
			errors = append(errors, models.DataValidationError{
				RowIndex:      rowIndex,
				FieldName:     field.Name,
				ErrorType:     "pattern",
				Message:       fmt.Sprintf("Field '%s' does not match required pattern", field.Name),
				ActualValue:   valueStr,
				ExpectedValue: *validation.Pattern,
			})
		}
	}

	// Options validation (enum)
	if len(validation.Options) > 0 {
		valid := false
		for _, option := range validation.Options {
			if valueStr == option {
				valid = true
				break
			}
		}
		if !valid {
			errors = append(errors, models.DataValidationError{
				RowIndex:      rowIndex,
				FieldName:     field.Name,
				ErrorType:     "invalid_option",
				Message:       fmt.Sprintf("Field '%s' must be one of: %s", field.Name, strings.Join(validation.Options, ", ")),
				ActualValue:   valueStr,
				ExpectedValue: strings.Join(validation.Options, ", "),
			})
		}
	}

	return errors
}

// validateBusinessRules validates data against business rules
func (v *ValidationService) validateBusinessRules(allRowData []map[string]interface{}, rules []*models.DatasetBusinessRule) []models.DataValidationError {
	var errors []models.DataValidationError

	for _, rule := range rules {
		switch rule.RuleType {
		case models.RuleTypeUnique:
			errors = append(errors, v.validateUniqueRule(allRowData, rule)...)
		case models.RuleTypeRangeCheck:
			errors = append(errors, v.validateRangeRule(allRowData, rule)...)
		case models.RuleTypeCrossField:
			errors = append(errors, v.validateCrossFieldRule(allRowData, rule)...)
		}
	}

	return errors
}

// validateUniqueRule validates uniqueness constraints
func (v *ValidationService) validateUniqueRule(allRowData []map[string]interface{}, rule *models.DatasetBusinessRule) []models.DataValidationError {
	var errors []models.DataValidationError
	
	var config models.BusinessRuleConfig
	if err := json.Unmarshal(rule.RuleConfig, &config); err != nil {
		return errors
	}

	seen := make(map[string][]int)
	
	for rowIndex, rowData := range allRowData {
		if value, exists := rowData[config.FieldName]; exists && value != "" {
			valueStr := fmt.Sprintf("%v", value)
			seen[valueStr] = append(seen[valueStr], rowIndex)
		}
	}

	// Report duplicates
	for value, indices := range seen {
		if len(indices) > 1 {
			for i := 1; i < len(indices); i++ { // Skip first occurrence
				errors = append(errors, models.DataValidationError{
					RowIndex:    indices[i],
					FieldName:   config.FieldName,
					ErrorType:   "duplicate_value",
					Message:     rule.ErrorMessage,
					ActualValue: value,
				})
			}
		}
	}

	return errors
}

// validateRangeRule validates range constraints
func (v *ValidationService) validateRangeRule(allRowData []map[string]interface{}, rule *models.DatasetBusinessRule) []models.DataValidationError {
	var errors []models.DataValidationError
	
	var config models.BusinessRuleConfig
	if err := json.Unmarshal(rule.RuleConfig, &config); err != nil {
		return errors
	}

	for rowIndex, rowData := range allRowData {
		if value, exists := rowData[config.FieldName]; exists && value != "" {
			if numValue, err := strconv.ParseFloat(fmt.Sprintf("%v", value), 64); err == nil {
				valid := true
				
				if config.MinValue != nil {
					if minVal, ok := config.MinValue.(float64); ok && numValue < minVal {
						valid = false
					}
				}
				
				if config.MaxValue != nil {
					if maxVal, ok := config.MaxValue.(float64); ok && numValue > maxVal {
						valid = false
					}
				}

				if !valid {
					errors = append(errors, models.DataValidationError{
						RowIndex:    rowIndex,
						FieldName:   config.FieldName,
						ErrorType:   "range_violation",
						Message:     rule.ErrorMessage,
						ActualValue: fmt.Sprintf("%v", value),
					})
				}
			}
		}
	}

	return errors
}

// validateCrossFieldRule validates relationships between fields
func (v *ValidationService) validateCrossFieldRule(allRowData []map[string]interface{}, rule *models.DatasetBusinessRule) []models.DataValidationError {
	var errors []models.DataValidationError
	
	var config models.BusinessRuleConfig
	if err := json.Unmarshal(rule.RuleConfig, &config); err != nil {
		return errors
	}

	// This is a simplified implementation - in practice, you'd parse and evaluate the condition
	for rowIndex, rowData := range allRowData {
		if !v.evaluateCrossFieldCondition(rowData, config) {
			errors = append(errors, models.DataValidationError{
				RowIndex:    rowIndex,
				FieldName:   strings.Join(config.Fields, ", "),
				ErrorType:   "cross_field_violation",
				Message:     rule.ErrorMessage,
				ActualValue: "condition failed",
			})
		}
	}

	return errors
}

// evaluateCrossFieldCondition evaluates cross-field conditions (simplified)
func (v *ValidationService) evaluateCrossFieldCondition(rowData map[string]interface{}, config models.BusinessRuleConfig) bool {
	// This is a very basic implementation
	// In a production system, you'd want a proper expression parser
	
	if len(config.Fields) < 2 {
		return true
	}

	// Example: "field1 > field2"
	if strings.Contains(config.Condition, ">") {
		parts := strings.Split(config.Condition, ">")
		if len(parts) == 2 {
			field1 := strings.TrimSpace(parts[0])
			field2 := strings.TrimSpace(parts[1])
			
			val1, _ := strconv.ParseFloat(fmt.Sprintf("%v", rowData[field1]), 64)
			val2, _ := strconv.ParseFloat(fmt.Sprintf("%v", rowData[field2]), 64)
			
			return val1 > val2
		}
	}

	return true // Default to valid if condition can't be evaluated
}

// updateFieldStats updates field statistics during validation
func (v *ValidationService) updateFieldStats(rowData map[string]interface{}, schema *models.DatasetSchema, fieldStats map[string]models.FieldStats) {
	for _, field := range schema.Fields {
		stats := fieldStats[field.Name]
		stats.TotalValues++

		value, exists := rowData[field.Name]
		if !exists || value == "" || value == nil {
			stats.NullValues++
		}

		fieldStats[field.Name] = stats
	}
}

// calculateUniqueValues calculates unique value counts for field statistics
func (v *ValidationService) calculateUniqueValues(allRowData []map[string]interface{}, fieldStats map[string]models.FieldStats) {
	uniqueValues := make(map[string]map[string]bool)
	
	// Initialize maps
	for fieldName := range fieldStats {
		uniqueValues[fieldName] = make(map[string]bool)
	}

	// Count unique values
	for _, rowData := range allRowData {
		for fieldName := range fieldStats {
			if value, exists := rowData[fieldName]; exists && value != "" && value != nil {
				uniqueValues[fieldName][fmt.Sprintf("%v", value)] = true
			}
		}
	}

	// Update stats
	for fieldName, stats := range fieldStats {
		stats.UniqueValues = len(uniqueValues[fieldName])
		fieldStats[fieldName] = stats
	}
}
