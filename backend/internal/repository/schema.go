package repository

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/saurabh22suman/oreo.io/internal/models"
)

// SchemaRepository handles database operations for schemas
type SchemaRepository struct {
	db *sqlx.DB
}

// NewSchemaRepository creates a new schema repository
func NewSchemaRepository(db *sqlx.DB) *SchemaRepository {
	return &SchemaRepository{db: db}
}

// CreateSchema creates a new dataset schema
func (r *SchemaRepository) CreateSchema(schema *models.DatasetSchema) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert schema
	query := `
		INSERT INTO dataset_schemas (id, dataset_id, name, description, created_at, updated_at)
		VALUES (:id, :dataset_id, :name, :description, :created_at, :updated_at)`
	
	_, err = tx.NamedExec(query, schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	// Insert fields
	for _, field := range schema.Fields {
		fieldQuery := `
			INSERT INTO schema_fields (id, schema_id, name, display_name, data_type, is_required, is_unique, 
				default_value, position, validation, created_at, updated_at)
			VALUES (:id, :schema_id, :name, :display_name, :data_type, :is_required, :is_unique, 
				:default_value, :position, :validation, :created_at, :updated_at)`
		
		// Convert validation to JSON
		validationJSON, err := json.Marshal(field.Validation)
		if err != nil {
			return fmt.Errorf("failed to marshal validation: %w", err)
		}

		params := map[string]interface{}{
			"id":            field.ID,
			"schema_id":     field.SchemaID,
			"name":          field.Name,
			"display_name":  field.DisplayName,
			"data_type":     field.DataType,
			"is_required":   field.IsRequired,
			"is_unique":     field.IsUnique,
			"default_value": field.DefaultValue,
			"position":      field.Position,
			"validation":    validationJSON,
			"created_at":    field.CreatedAt,
			"updated_at":    field.UpdatedAt,
		}

		_, err = tx.NamedExec(fieldQuery, params)
		if err != nil {
			return fmt.Errorf("failed to create schema field: %w", err)
		}
	}

	return tx.Commit()
}

// GetSchemaByDatasetID retrieves schema for a dataset
func (r *SchemaRepository) GetSchemaByDatasetID(datasetID uuid.UUID) (*models.DatasetSchema, error) {
	schema := &models.DatasetSchema{}
	
	// Get schema
	query := `SELECT id, dataset_id, name, description, created_at, updated_at 
			  FROM dataset_schemas WHERE dataset_id = $1`
	
	err := r.db.Get(schema, query, datasetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema: %w", err)
	}

	// Get fields
	fieldsQuery := `
		SELECT id, schema_id, name, display_name, data_type, is_required, is_unique, 
			   default_value, position, validation, created_at, updated_at
		FROM schema_fields 
		WHERE schema_id = $1 
		ORDER BY position`
	
	rows, err := r.db.Query(fieldsQuery, schema.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema fields: %w", err)
	}
	defer rows.Close()

	var fields []models.SchemaField
	for rows.Next() {
		field := models.SchemaField{}
		var validationJSON []byte
		
		err := rows.Scan(
			&field.ID, &field.SchemaID, &field.Name, &field.DisplayName,
			&field.DataType, &field.IsRequired, &field.IsUnique,
			&field.DefaultValue, &field.Position, &validationJSON,
			&field.CreatedAt, &field.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan field: %w", err)
		}

		// Unmarshal validation JSON
		if len(validationJSON) > 0 {
			err = json.Unmarshal(validationJSON, &field.Validation)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal validation: %w", err)
			}
		}

		fields = append(fields, field)
	}

	schema.Fields = fields
	return schema, nil
}

// UpdateSchema updates an existing schema
func (r *SchemaRepository) UpdateSchema(schema *models.DatasetSchema) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update schema
	query := `
		UPDATE dataset_schemas 
		SET name = :name, description = :description, updated_at = :updated_at
		WHERE id = :id`
	
	_, err = tx.NamedExec(query, schema)
	if err != nil {
		return fmt.Errorf("failed to update schema: %w", err)
	}

	// Delete existing fields
	_, err = tx.Exec("DELETE FROM schema_fields WHERE schema_id = $1", schema.ID)
	if err != nil {
		return fmt.Errorf("failed to delete existing fields: %w", err)
	}

	// Insert updated fields
	for _, field := range schema.Fields {
		fieldQuery := `
			INSERT INTO schema_fields (id, schema_id, name, display_name, data_type, is_required, is_unique, 
				default_value, position, validation, created_at, updated_at)
			VALUES (:id, :schema_id, :name, :display_name, :data_type, :is_required, :is_unique, 
				:default_value, :position, :validation, :created_at, :updated_at)`
		
		validationJSON, err := json.Marshal(field.Validation)
		if err != nil {
			return fmt.Errorf("failed to marshal validation: %w", err)
		}

		params := map[string]interface{}{
			"id":            field.ID,
			"schema_id":     field.SchemaID,
			"name":          field.Name,
			"display_name":  field.DisplayName,
			"data_type":     field.DataType,
			"is_required":   field.IsRequired,
			"is_unique":     field.IsUnique,
			"default_value": field.DefaultValue,
			"position":      field.Position,
			"validation":    validationJSON,
			"created_at":    field.CreatedAt,
			"updated_at":    field.UpdatedAt,
		}

		_, err = tx.NamedExec(fieldQuery, params)
		if err != nil {
			return fmt.Errorf("failed to create schema field: %w", err)
		}
	}

	return tx.Commit()
}

// DeleteSchema deletes a schema and all its fields
func (r *SchemaRepository) DeleteSchema(schemaID uuid.UUID) error {
	query := `DELETE FROM dataset_schemas WHERE id = $1`
	_, err := r.db.Exec(query, schemaID)
	if err != nil {
		return fmt.Errorf("failed to delete schema: %w", err)
	}
	return nil
}

// GetDatasetData retrieves paginated data for a dataset
func (r *SchemaRepository) GetDatasetData(datasetID uuid.UUID, page, pageSize int) (*models.DataPreviewResponse, error) {
	offset := (page - 1) * pageSize
	
	// Get total count
	var totalRows int
	countQuery := `SELECT COUNT(*) FROM dataset_data WHERE dataset_id = $1`
	err := r.db.Get(&totalRows, countQuery, datasetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Get data
	dataQuery := `
		SELECT row_index, data 
		FROM dataset_data 
		WHERE dataset_id = $1 
		ORDER BY row_index 
		LIMIT $2 OFFSET $3`
	
	rows, err := r.db.Query(dataQuery, datasetID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset data: %w", err)
	}
	defer rows.Close()

	var data []map[string]interface{}
	for rows.Next() {
		var rowIndex int
		var dataJSON []byte
		
		err := rows.Scan(&rowIndex, &dataJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan data row: %w", err)
		}

		var rowData map[string]interface{}
		err = json.Unmarshal(dataJSON, &rowData)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal data: %w", err)
		}

		// Add row index to data
		rowData["_row_index"] = rowIndex
		data = append(data, rowData)
	}

	// Get schema
	schema, err := r.GetSchemaByDatasetID(datasetID)
	if err != nil {
		// Schema might not exist yet, that's okay
		schema = nil
	}

	totalPages := (totalRows + pageSize - 1) / pageSize

	return &models.DataPreviewResponse{
		Data:       data,
		Schema:     schema,
		TotalRows:  totalRows,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetDatasetDataWithLimit retrieves dataset data with a maximum row limit
func (r *SchemaRepository) GetDatasetDataWithLimit(datasetID uuid.UUID, page, pageSize, maxRows int) (*models.DataPreviewResponse, error) {
	// Calculate the maximum offset we can allow
	offset := (page - 1) * pageSize
	if offset >= maxRows {
		// Return empty result if beyond limit
		return &models.DataPreviewResponse{
			Data:       []map[string]interface{}{},
			Schema:     nil,
			TotalRows:  maxRows,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: maxRows / pageSize,
		}, nil
	}

	// Adjust page size if it would exceed the limit
	remainingRows := maxRows - offset
	if pageSize > remainingRows {
		pageSize = remainingRows
	}

	// Get count query with limit
	countQuery := `SELECT LEAST(COUNT(*), $2) FROM dataset_data WHERE dataset_id = $1`
	var totalRows int
	err := r.db.Get(&totalRows, countQuery, datasetID, maxRows)
	if err != nil {
		return nil, fmt.Errorf("failed to get data count: %w", err)
	}

	// Get data with limit
	dataQuery := `
		SELECT row_index, data 
		FROM dataset_data 
		WHERE dataset_id = $1 
		ORDER BY row_index 
		LIMIT $2 OFFSET $3`
	
	rows, err := r.db.Query(dataQuery, datasetID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get data: %w", err)
	}
	defer rows.Close()

	var data []map[string]interface{}
	for rows.Next() {
		var rowIndex int
		var dataJSON []byte
		
		err := rows.Scan(&rowIndex, &dataJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan data row: %w", err)
		}

		var rowData map[string]interface{}
		err = json.Unmarshal(dataJSON, &rowData)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal data: %w", err)
		}

		// Add row index to data
		rowData["_row_index"] = rowIndex
		data = append(data, rowData)
	}

	// Get schema
	schema, err := r.GetSchemaByDatasetID(datasetID)
	if err != nil {
		// Schema might not exist yet, that's okay
		schema = nil
	}

	// Calculate total pages based on limited rows
	limitedTotalRows := totalRows
	if limitedTotalRows > maxRows {
		limitedTotalRows = maxRows
	}
	totalPages := (limitedTotalRows + pageSize - 1) / pageSize

	return &models.DataPreviewResponse{
		Data:       data,
		Schema:     schema,
		TotalRows:  limitedTotalRows,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// QueryDatasetData executes a SQL-like query on dataset data
func (r *SchemaRepository) QueryDatasetData(datasetID uuid.UUID, sqlQuery string, pageSize int) (*models.DataPreviewResponse, error) {
	// For security, we'll implement a simple WHERE clause parser
	// This is a simplified version - in production, use a proper SQL parser
	
	// Start with base query
	baseQuery := `
		SELECT row_index, data 
		FROM dataset_data 
		WHERE dataset_id = $1`
	
	var args []interface{}
	args = append(args, datasetID)
	
	// Very basic WHERE clause support - just search in JSON data
	// This is simplified and should be enhanced for production
	finalQuery := baseQuery
	if sqlQuery != "" {
		// Simple LIKE search in JSON data
		finalQuery += ` AND data::text ILIKE $2`
		args = append(args, "%"+sqlQuery+"%")
	}
	
	finalQuery += ` ORDER BY row_index LIMIT $` + fmt.Sprintf("%d", len(args)+1)
	args = append(args, pageSize)

	// Get count first
	countQuery := `SELECT COUNT(*) FROM dataset_data WHERE dataset_id = $1`
	countArgs := []interface{}{datasetID}
	if sqlQuery != "" {
		countQuery += ` AND data::text ILIKE $2`
		countArgs = append(countArgs, "%"+sqlQuery+"%")
	}

	var totalRows int
	err := r.db.Get(&totalRows, countQuery, countArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to get count: %w", err)
	}

	// Execute main query
	rows, err := r.db.Query(finalQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var data []map[string]interface{}
	for rows.Next() {
		var rowIndex int
		var dataJSON []byte
		
		err := rows.Scan(&rowIndex, &dataJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan data row: %w", err)
		}

		var rowData map[string]interface{}
		err = json.Unmarshal(dataJSON, &rowData)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal data: %w", err)
		}

		// Add row index to data
		rowData["_row_index"] = rowIndex
		data = append(data, rowData)
	}

	// Get schema
	schema, err := r.GetSchemaByDatasetID(datasetID)
	if err != nil {
		schema = nil
	}

	return &models.DataPreviewResponse{
		Data:       data,
		Schema:     schema,
		TotalRows:  totalRows,
		Page:       1,
		PageSize:   pageSize,
		TotalPages: (totalRows + pageSize - 1) / pageSize,
	}, nil
}

// BulkInsertDatasetData inserts multiple rows of CSV data
func (r *SchemaRepository) BulkInsertDatasetData(datasetID uuid.UUID, headers []string, rows [][]string, userID uuid.UUID) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare the insert statement
	query := `
		INSERT INTO dataset_data (dataset_id, row_index, data, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $4)`

	for i, row := range rows {
		// Create a map from headers to row values
		data := make(map[string]interface{})
		for j, header := range headers {
			if j < len(row) {
				data[header] = row[j]
			} else {
				data[header] = "" // Handle missing values
			}
		}

		// Marshal to JSON
		dataJSON, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal data for row %d: %w", i, err)
		}

		// Insert the row (row_index starts from 0)
		_, err = tx.Exec(query, datasetID, i, dataJSON, userID)
		if err != nil {
			return fmt.Errorf("failed to insert data for row %d: %w", i, err)
		}
	}

	return tx.Commit()
}

// UpdateDatasetData updates or inserts a data row
func (r *SchemaRepository) UpdateDatasetData(datasetID uuid.UUID, rowIndex int, data map[string]interface{}, userID uuid.UUID) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	query := `
		INSERT INTO dataset_data (dataset_id, row_index, data, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $4)
		ON CONFLICT (dataset_id, row_index)
		DO UPDATE SET 
			data = EXCLUDED.data,
			version = dataset_data.version + 1,
			updated_by = EXCLUDED.updated_by,
			updated_at = NOW()`
	
	_, err = r.db.Exec(query, datasetID, rowIndex, dataJSON, userID)
	if err != nil {
		return fmt.Errorf("failed to update dataset data: %w", err)
	}

	return nil
}

// DeleteDatasetData deletes a data row
func (r *SchemaRepository) DeleteDatasetData(datasetID uuid.UUID, rowIndex int) error {
	query := `DELETE FROM dataset_data WHERE dataset_id = $1 AND row_index = $2`
	_, err := r.db.Exec(query, datasetID, rowIndex)
	if err != nil {
		return fmt.Errorf("failed to delete dataset data: %w", err)
	}
	return nil
}

// CheckDatasetAccess checks if user has access to dataset
func (r *SchemaRepository) CheckDatasetAccess(datasetID, userID uuid.UUID) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM datasets d 
		JOIN projects p ON d.project_id = p.id 
		WHERE d.id = $1 AND (p.owner_id = $2 OR EXISTS (
			SELECT 1 FROM project_members pm 
			WHERE pm.project_id = p.id AND pm.user_id = $2
		))`
	
	var count int
	err := r.db.Get(&count, query, datasetID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check dataset access: %w", err)
	}
	
	return count > 0, nil
}

// GetDatasetByID retrieves dataset information by ID
func (r *SchemaRepository) GetDatasetByID(datasetID uuid.UUID) (*models.Dataset, error) {
	query := `SELECT id, project_id, name, description, file_name, file_path, file_size, 
			  mime_type, row_count, column_count, status, uploaded_by, created_at, updated_at 
			  FROM datasets WHERE id = $1`
	
	var dataset models.Dataset
	err := r.db.Get(&dataset, query, datasetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset: %w", err)
	}
	
	return &dataset, nil
}

// GetDatasetDataForInference retrieves dataset headers and sample data for schema inference
func (r *SchemaRepository) GetDatasetDataForInference(datasetID uuid.UUID, maxRows int) ([]string, [][]string, error) {
	// Get sample data rows
	dataQuery := `
		SELECT data 
		FROM dataset_data 
		WHERE dataset_id = $1 
		ORDER BY row_index 
		LIMIT $2
	`
	
	var rawDataRows [][]byte
	err := r.db.Select(&rawDataRows, dataQuery, datasetID, maxRows)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get dataset data: %w", err)
	}
	
	if len(rawDataRows) == 0 {
		return nil, nil, fmt.Errorf("no data found in dataset")
	}
	
	// Parse first row to get headers
	var firstRowData map[string]interface{}
	err = json.Unmarshal(rawDataRows[0], &firstRowData)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse first row data: %w", err)
	}
	
	// Extract headers from the first row
	var headers []string
	for key := range firstRowData {
		headers = append(headers, key)
	}
	
	// If no headers found, return empty
	if len(headers) == 0 {
		return nil, nil, fmt.Errorf("no columns found in dataset")
	}
	
	// Convert all rows to string matrix
	rows := make([][]string, len(rawDataRows))
	for i, rawRow := range rawDataRows {
		var rowData map[string]interface{}
		err = json.Unmarshal(rawRow, &rowData)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse row %d: %w", i, err)
		}
		
		row := make([]string, len(headers))
		for j, header := range headers {
			if value, exists := rowData[header]; exists && value != nil {
				row[j] = fmt.Sprintf("%v", value)
			} else {
				row[j] = ""
			}
		}
		rows[i] = row
	}
	
	return headers, rows, nil
}
