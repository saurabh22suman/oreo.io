package services

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/saurabh22suman/oreo.io/internal/models"
)

type SchemaInferenceService struct{}

type InferredField struct {
	Name         string                 `json:"name"`
	DisplayName  string                 `json:"display_name"`
	DataType     models.SchemaFieldType `json:"data_type"`
	IsRequired   bool                   `json:"is_required"`
	Constraints  map[string]interface{} `json:"constraints,omitempty"`
	Pattern      string                 `json:"pattern,omitempty"`
	Confidence   float64                `json:"confidence"` // 0.0 to 1.0
	SampleValues []string               `json:"sample_values,omitempty"`
}

type InferredSchema struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Fields      []InferredField `json:"fields"`
	RowCount    int             `json:"row_count"`
	Confidence  float64         `json:"overall_confidence"`
}

// Common patterns for field detection
var (
	emailPattern    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	phonePattern    = regexp.MustCompile(`^\+?[\d\s\-\(\)]{7,15}$`)
	urlPattern      = regexp.MustCompile(`^https?://[^\s]+$`)
	datePatterns    = []*regexp.Regexp{
		regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`),         // YYYY-MM-DD
		regexp.MustCompile(`^\d{2}/\d{2}/\d{4}$`),         // MM/DD/YYYY
		regexp.MustCompile(`^\d{2}-\d{2}-\d{4}$`),         // MM-DD-YYYY
		regexp.MustCompile(`^\d{4}/\d{2}/\d{2}$`),         // YYYY/MM/DD
	}
	timePatterns = []*regexp.Regexp{
		regexp.MustCompile(`^\d{2}:\d{2}:\d{2}$`),         // HH:MM:SS
		regexp.MustCompile(`^\d{2}:\d{2}$`),               // HH:MM
	}
	uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
)

func NewSchemaInferenceService() *SchemaInferenceService {
	return &SchemaInferenceService{}
}

// InferSchemaFromData analyzes data and infers schema with confidence scores
func (s *SchemaInferenceService) InferSchemaFromData(headers []string, rows [][]string, datasetName string) (*InferredSchema, error) {
	log.Printf("[DEBUG] InferSchemaFromData: Starting inference for dataset '%s' with %d columns and %d rows", datasetName, len(headers), len(rows))

	fields := make([]InferredField, len(headers))
	totalConfidence := 0.0

	// Analyze each column
	for i, header := range headers {
		field := s.analyzeColumn(header, s.extractColumn(rows, i))
		fields[i] = field
		totalConfidence += field.Confidence
	}

	// Calculate overall confidence
	overallConfidence := totalConfidence / float64(len(headers))

	schema := &InferredSchema{
		Name:        generateSchemaName(datasetName),
		Description: fmt.Sprintf("Auto-inferred schema for dataset '%s'", datasetName),
		Fields:      fields,
		RowCount:    len(rows),
		Confidence:  overallConfidence,
	}

	log.Printf("[DEBUG] InferSchemaFromData: Completed inference with overall confidence %.2f", overallConfidence)
	return schema, nil
}

// analyzeColumn performs deep analysis on a single column
func (s *SchemaInferenceService) analyzeColumn(header string, values []string) InferredField {
	log.Printf("[DEBUG] analyzeColumn: Analyzing column '%s' with %d values", header, len(values))

	field := InferredField{
		Name:        sanitizeFieldName(header),
		DisplayName: header,
		IsRequired:  false,
		Constraints: make(map[string]interface{}),
	}

	// Remove empty values for analysis
	nonEmptyValues := make([]string, 0, len(values))
	emptyCount := 0
	
	for _, val := range values {
		trimmed := strings.TrimSpace(val)
		if trimmed != "" {
			nonEmptyValues = append(nonEmptyValues, trimmed)
		} else {
			emptyCount++
		}
	}

	// Calculate required field confidence
	if len(values) > 0 {
		requiredConfidence := float64(len(nonEmptyValues)) / float64(len(values))
		field.IsRequired = requiredConfidence > 0.9 // Required if >90% of values are non-empty
	}

	// Store sample values (up to 5)
	sampleCount := min(5, len(nonEmptyValues))
	field.SampleValues = make([]string, sampleCount)
	copy(field.SampleValues, nonEmptyValues[:sampleCount])

	if len(nonEmptyValues) == 0 {
		field.DataType = models.FieldTypeString
		field.Confidence = 0.1 // Low confidence for empty columns
		return field
	}

	// Analyze data types with confidence scoring
	typeAnalysis := s.analyzeDataTypes(nonEmptyValues)
	field.DataType = typeAnalysis.PrimaryType
	field.Confidence = typeAnalysis.Confidence
	field.Pattern = typeAnalysis.Pattern

	// Add constraints based on data type
	s.addConstraints(&field, nonEmptyValues, typeAnalysis)

	log.Printf("[DEBUG] analyzeColumn: Column '%s' inferred as %s with confidence %.2f", header, field.DataType, field.Confidence)
	return field
}

type TypeAnalysis struct {
	PrimaryType models.SchemaFieldType
	Confidence  float64
	Pattern     string
	Constraints map[string]interface{}
}

// analyzeDataTypes performs statistical analysis of data types
func (s *SchemaInferenceService) analyzeDataTypes(values []string) TypeAnalysis {
	if len(values) == 0 {
		return TypeAnalysis{
			PrimaryType: models.FieldTypeString,
			Confidence:  0.1,
		}
	}

	// Count matches for each type
	typeScores := map[models.SchemaFieldType]int{
		models.FieldTypeString:   0,
		models.FieldTypeNumber:   0,
		models.FieldTypeBoolean:  0,
		models.FieldTypeDate:     0,
		models.FieldTypeDateTime: 0,
		models.FieldTypeEmail:    0,
		models.FieldTypeURL:      0,
		models.FieldTypeUUID:     0,
	}

	patterns := make(map[string]int)
	
	for _, value := range values {
		// Test each type
		if s.isNumber(value) {
			typeScores[models.FieldTypeNumber]++
		}
		if s.isBoolean(value) {
			typeScores[models.FieldTypeBoolean]++
		}
		if s.isEmail(value) {
			typeScores[models.FieldTypeEmail]++
		}
		if s.isURL(value) {
			typeScores[models.FieldTypeURL]++
		}
		if s.isUUID(value) {
			typeScores[models.FieldTypeUUID]++
		}
		
		// Date/time analysis
		if datePattern := s.isDate(value); datePattern != "" {
			typeScores[models.FieldTypeDate]++
			patterns[datePattern]++
		}
		if timePattern := s.isDateTime(value); timePattern != "" {
			typeScores[models.FieldTypeDateTime]++
			patterns[timePattern]++
		}
		
		// Always count as string (fallback)
		typeScores[models.FieldTypeString]++
	}

	// Find the type with highest score (excluding string)
	var bestType models.SchemaFieldType = models.FieldTypeString
	var bestScore int = 0
	var confidence float64 = 0.1

	for dataType, score := range typeScores {
		if dataType != models.FieldTypeString && score > bestScore {
			bestType = dataType
			bestScore = score
		}
	}

	// Calculate confidence based on how many values match the type
	if bestScore > 0 {
		confidence = float64(bestScore) / float64(len(values))
		
		// Require high confidence for non-string types
		if confidence < 0.8 {
			bestType = models.FieldTypeString
			confidence = 0.7 // Medium confidence for string fallback
		}
	}

	// Find most common pattern
	var bestPattern string
	var bestPatternCount int
	for pattern, count := range patterns {
		if count > bestPatternCount {
			bestPattern = pattern
			bestPatternCount = count
		}
	}

	return TypeAnalysis{
		PrimaryType: bestType,
		Confidence:  confidence,
		Pattern:     bestPattern,
	}
}

// Type checking helper functions
func (s *SchemaInferenceService) isNumber(value string) bool {
	_, err := strconv.ParseFloat(value, 64)
	return err == nil
}

func (s *SchemaInferenceService) isBoolean(value string) bool {
	lower := strings.ToLower(value)
	return lower == "true" || lower == "false" || lower == "yes" || lower == "no" || 
		   lower == "1" || lower == "0" || lower == "y" || lower == "n"
}

func (s *SchemaInferenceService) isEmail(value string) bool {
	return emailPattern.MatchString(value)
}

func (s *SchemaInferenceService) isURL(value string) bool {
	return urlPattern.MatchString(value)
}

func (s *SchemaInferenceService) isUUID(value string) bool {
	return uuidPattern.MatchString(strings.ToLower(value))
}

func (s *SchemaInferenceService) isDate(value string) string {
	for i, pattern := range datePatterns {
		if pattern.MatchString(value) {
			// Try to parse to validate it's a real date
			formats := []string{"2006-01-02", "01/02/2006", "01-02-2006", "2006/01/02"}
			if i < len(formats) {
				if _, err := time.Parse(formats[i], value); err == nil {
					return formats[i]
				}
			}
		}
	}
	return ""
}

func (s *SchemaInferenceService) isDateTime(value string) string {
	// Check for datetime patterns
	datetimeFormats := []string{
		"2006-01-02 15:04:05",
		"01/02/2006 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
	}
	
	for _, format := range datetimeFormats {
		if _, err := time.Parse(format, value); err == nil {
			return format
		}
	}
	return ""
}

// addConstraints adds appropriate constraints based on data analysis
func (s *SchemaInferenceService) addConstraints(field *InferredField, values []string, analysis TypeAnalysis) {
	switch field.DataType {
	case models.FieldTypeNumber:
		s.addNumberConstraints(field, values)
	case models.FieldTypeString:
		s.addStringConstraints(field, values)
	case models.FieldTypeDate, models.FieldTypeDateTime:
		if analysis.Pattern != "" {
			field.Constraints["format"] = analysis.Pattern
		}
	}
}

func (s *SchemaInferenceService) addNumberConstraints(field *InferredField, values []string) {
	var numbers []float64
	for _, value := range values {
		if num, err := strconv.ParseFloat(value, 64); err == nil {
			numbers = append(numbers, num)
		}
	}

	if len(numbers) > 0 {
		min, max := numbers[0], numbers[0]
		for _, num := range numbers {
			if num < min {
				min = num
			}
			if num > max {
				max = num
			}
		}
		
		field.Constraints["min"] = min
		field.Constraints["max"] = max
		
		// Check if all numbers are integers
		allIntegers := true
		for _, num := range numbers {
			if num != float64(int64(num)) {
				allIntegers = false
				break
			}
		}
		field.Constraints["integer"] = allIntegers
	}
}

func (s *SchemaInferenceService) addStringConstraints(field *InferredField, values []string) {
	if len(values) > 0 {
		minLen, maxLen := len(values[0]), len(values[0])
		for _, value := range values {
			length := len(value)
			if length < minLen {
				minLen = length
			}
			if length > maxLen {
				maxLen = length
			}
		}
		
		field.Constraints["min_length"] = minLen
		field.Constraints["max_length"] = maxLen
	}
}

// Utility functions
func (s *SchemaInferenceService) extractColumn(rows [][]string, columnIndex int) []string {
	column := make([]string, len(rows))
	for i, row := range rows {
		if columnIndex < len(row) {
			column[i] = row[columnIndex]
		}
	}
	return column
}

func sanitizeFieldName(name string) string {
	// Convert to lowercase, replace spaces and special chars with underscores
	sanitized := strings.ToLower(name)
	sanitized = regexp.MustCompile(`[^a-z0-9_]`).ReplaceAllString(sanitized, "_")
	sanitized = regexp.MustCompile(`_+`).ReplaceAllString(sanitized, "_")
	sanitized = strings.Trim(sanitized, "_")
	
	if sanitized == "" {
		sanitized = "field"
	}
	
	return sanitized
}

func generateSchemaName(datasetName string) string {
	return fmt.Sprintf("%s_schema", sanitizeFieldName(datasetName))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
