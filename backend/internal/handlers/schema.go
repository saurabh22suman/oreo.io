package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/saurabh22suman/oreo.io/internal/models"
	"github.com/saurabh22suman/oreo.io/internal/repository"
	"github.com/saurabh22suman/oreo.io/internal/services"
)

// SchemaHandlers contains schema-related handlers
type SchemaHandlers struct {
	schemaRepo        *repository.SchemaRepository
	inferenceService  *services.SchemaInferenceService
}

// NewSchemaHandlers creates new schema handlers
func NewSchemaHandlers(db *sqlx.DB) *SchemaHandlers {
	return &SchemaHandlers{
		schemaRepo:       repository.NewSchemaRepository(db),
		inferenceService: services.NewSchemaInferenceService(),
	}
}

// CreateSchema creates a new dataset schema
func (h *SchemaHandlers) CreateSchema() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		userUUID, ok := userID.(uuid.UUID)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
			return
		}

		var req models.CreateSchemaRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Check if user has access to the dataset
		hasAccess, err := h.schemaRepo.CheckDatasetAccess(req.DatasetID, userUUID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify dataset access"})
			return
		}

		if !hasAccess {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to modify this dataset"})
			return
		}

		// Create schema object
		schema := &models.DatasetSchema{
			ID:          uuid.New(),
			DatasetID:   req.DatasetID,
			Name:        req.Name,
			Description: req.Description,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Create fields
		for i, fieldReq := range req.Fields {
			field := models.SchemaField{
				ID:           uuid.New(),
				SchemaID:     schema.ID,
				Name:         fieldReq.Name,
				DisplayName:  fieldReq.DisplayName,
				DataType:     fieldReq.DataType,
				IsRequired:   fieldReq.IsRequired,
				IsUnique:     fieldReq.IsUnique,
				DefaultValue: fieldReq.DefaultValue,
				Position:     fieldReq.Position,
				Validation:   fieldReq.Validation,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}

			if field.DisplayName == "" {
				field.DisplayName = field.Name
			}

			if field.Position == 0 {
				field.Position = i + 1
			}

			schema.Fields = append(schema.Fields, field)
		}

		// Save to database
		err = h.schemaRepo.CreateSchema(schema)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create schema"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"schema":  schema,
			"message": "Schema created successfully",
		})
	}
}

// GetSchema retrieves schema for a dataset
func (h *SchemaHandlers) GetSchema() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("[DEBUG] GetSchema: Starting request")
		
		userID, exists := c.Get("user_id")
		if !exists {
			log.Printf("[ERROR] GetSchema: User not authenticated")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		userUUID, ok := userID.(uuid.UUID)
		if !ok {
			log.Printf("[ERROR] GetSchema: Invalid user ID type")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
			return
		}

		datasetIDStr := c.Param("dataset_id")
		log.Printf("[DEBUG] GetSchema: Dataset ID param: %s", datasetIDStr)
		
		datasetID, err := uuid.Parse(datasetIDStr)
		if err != nil {
			log.Printf("[ERROR] GetSchema: Invalid dataset ID format: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dataset ID"})
			return
		}

		log.Printf("[DEBUG] GetSchema: User %s requesting schema for dataset %s", userUUID, datasetID)

		// Check access
		hasAccess, err := h.schemaRepo.CheckDatasetAccess(datasetID, userUUID)
		if err != nil {
			log.Printf("[ERROR] GetSchema: Error checking dataset access for dataset %s, user %s: %v", datasetID, userUUID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify dataset access"})
			return
		}

		if !hasAccess {
			log.Printf("[ERROR] GetSchema: User %s does not have access to dataset %s", userUUID, datasetID)
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view this dataset"})
			return
		}

		log.Printf("[DEBUG] GetSchema: Access verified, fetching schema...")

		schema, err := h.schemaRepo.GetSchemaByDatasetID(datasetID)
		if err != nil {
			log.Printf("[ERROR] GetSchema: Schema not found for dataset %s: %v", datasetID, err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Schema not found"})
			return
		}

		log.Printf("[DEBUG] GetSchema: Successfully fetched schema for dataset %s", datasetID)
		c.JSON(http.StatusOK, gin.H{"schema": schema})
	}
}

// UpdateSchema updates an existing schema
func (h *SchemaHandlers) UpdateSchema() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		userUUID, ok := userID.(uuid.UUID)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
			return
		}

		schemaIDStr := c.Param("schema_id")
		schemaID, err := uuid.Parse(schemaIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid schema ID"})
			return
		}

		var req models.UpdateSchemaRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get existing schema to check access
		existingSchema, err := h.schemaRepo.GetSchemaByDatasetID(uuid.UUID{}) // We need to get by schema ID instead
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Schema not found"})
			return
		}

		// Check access
		hasAccess, err := h.schemaRepo.CheckDatasetAccess(existingSchema.DatasetID, userUUID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify dataset access"})
			return
		}

		if !hasAccess {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to modify this dataset"})
			return
		}

		// Update schema
		existingSchema.Name = req.Name
		existingSchema.Description = req.Description
		existingSchema.UpdatedAt = time.Now()

		// Update fields
		existingSchema.Fields = []models.SchemaField{}
		for _, fieldReq := range req.Fields {
			field := models.SchemaField{
				ID:           fieldReq.ID,
				SchemaID:     schemaID,
				Name:         fieldReq.Name,
				DisplayName:  fieldReq.DisplayName,
				DataType:     fieldReq.DataType,
				IsRequired:   fieldReq.IsRequired,
				IsUnique:     fieldReq.IsUnique,
				DefaultValue: fieldReq.DefaultValue,
				Position:     fieldReq.Position,
				Validation:   fieldReq.Validation,
				UpdatedAt:    time.Now(),
			}

			if field.DisplayName == "" {
				field.DisplayName = field.Name
			}

			existingSchema.Fields = append(existingSchema.Fields, field)
		}

		err = h.schemaRepo.UpdateSchema(existingSchema)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update schema"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"schema":  existingSchema,
			"message": "Schema updated successfully",
		})
	}
}

// DeleteSchema deletes a schema
func (h *SchemaHandlers) DeleteSchema() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Add proper authorization check
		schemaIDStr := c.Param("schema_id")
		schemaID, err := uuid.Parse(schemaIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid schema ID"})
			return
		}

		err = h.schemaRepo.DeleteSchema(schemaID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete schema"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Schema deleted successfully"})
	}
}

// GetDatasetData retrieves paginated dataset data with maximum 1000 rows
func (h *SchemaHandlers) GetDatasetData() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("[DEBUG] GetDatasetData: Starting request")
		
		userID, exists := c.Get("user_id")
		if !exists {
			log.Printf("[ERROR] GetDatasetData: User not authenticated")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		userUUID, ok := userID.(uuid.UUID)
		if !ok {
			log.Printf("[ERROR] GetDatasetData: Invalid user ID type")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
			return
		}

		datasetIDStr := c.Param("dataset_id")
		log.Printf("[DEBUG] GetDatasetData: Dataset ID param: %s", datasetIDStr)
		
		datasetID, err := uuid.Parse(datasetIDStr)
		if err != nil {
			log.Printf("[ERROR] GetDatasetData: Invalid dataset ID format: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dataset ID"})
			return
		}

		// Parse pagination parameters with strict limits
		page := 1
		pageSize := 50 // Default page size
		maxRows := 1000 // Maximum rows to display

		if pageStr := c.Query("page"); pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				page = p
			}
		}

		if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
			if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
				pageSize = ps
			}
		}

		// Ensure we don't exceed max rows limit
		maxPage := maxRows / pageSize
		if page > maxPage {
			page = maxPage
		}

		log.Printf("[DEBUG] GetDatasetData: User %s requesting data for dataset %s (page=%d, pageSize=%d)", userUUID, datasetID, page, pageSize)

		// Check access
		hasAccess, err := h.schemaRepo.CheckDatasetAccess(datasetID, userUUID)
		if err != nil {
			log.Printf("[ERROR] GetDatasetData: Error checking dataset access for user %s, dataset %s: %v", userUUID, datasetID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify dataset access"})
			return
		}

		if !hasAccess {
			log.Printf("[ERROR] GetDatasetData: User %s does not have access to dataset %s", userUUID, datasetID)
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view this dataset"})
			return
		}

		log.Printf("[DEBUG] GetDatasetData: Access verified, fetching data...")

		// Get data with row limit
		result, err := h.schemaRepo.GetDatasetDataWithLimit(datasetID, page, pageSize, maxRows)
		if err != nil {
			log.Printf("[ERROR] GetDatasetData: Error getting dataset data for dataset %s: %v", datasetID, err)
			// Return empty result instead of error for missing data
			result = &models.DataPreviewResponse{
				Data:       []map[string]interface{}{},
				Schema:     nil,
				TotalRows:  0,
				Page:       page,
				PageSize:   pageSize,
				TotalPages: 0,
			}
			log.Printf("[DEBUG] GetDatasetData: Returning empty result due to error")
		} else {
			log.Printf("[DEBUG] GetDatasetData: Successfully fetched %d rows for dataset %s", len(result.Data), datasetID)
		}

		c.JSON(http.StatusOK, result)
	}
}

// UpdateDatasetData updates a specific row of dataset data
func (h *SchemaHandlers) UpdateDatasetData() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		userUUID, ok := userID.(uuid.UUID)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
			return
		}

		datasetIDStr := c.Param("dataset_id")
		datasetID, err := uuid.Parse(datasetIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dataset ID"})
			return
		}

		var req models.UpdateDataRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Check access
		hasAccess, err := h.schemaRepo.CheckDatasetAccess(datasetID, userUUID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify dataset access"})
			return
		}

		if !hasAccess {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to modify this dataset"})
			return
		}

		// TODO: Add schema validation here

		// Update data
		err = h.schemaRepo.UpdateDatasetData(datasetID, req.RowIndex, req.Data, userUUID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update dataset data"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Data updated successfully"})
	}
}

// DeleteDatasetData deletes a specific row of dataset data
func (h *SchemaHandlers) DeleteDatasetData() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		userUUID, ok := userID.(uuid.UUID)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
			return
		}

		datasetIDStr := c.Param("dataset_id")
		datasetID, err := uuid.Parse(datasetIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dataset ID"})
			return
		}

		rowIndexStr := c.Param("row_index")
		rowIndex, err := strconv.Atoi(rowIndexStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid row index"})
			return
		}

		// Check access
		hasAccess, err := h.schemaRepo.CheckDatasetAccess(datasetID, userUUID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify dataset access"})
			return
		}

		if !hasAccess {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to modify this dataset"})
			return
		}

		// Delete data
		err = h.schemaRepo.DeleteDatasetData(datasetID, rowIndex)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete dataset data"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Data deleted successfully"})
	}
}

// QueryDatasetData executes a SQL query on dataset data
func (h *SchemaHandlers) QueryDatasetData() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		userUUID, ok := userID.(uuid.UUID)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
			return
		}

		datasetIDStr := c.Param("dataset_id")
		datasetID, err := uuid.Parse(datasetIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dataset ID"})
			return
		}

		// Check access
		hasAccess, err := h.schemaRepo.CheckDatasetAccess(datasetID, userUUID)
		if err != nil {
			log.Printf("Error checking dataset access: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify dataset access"})
			return
		}

		if !hasAccess {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to query this dataset"})
			return
		}

		// Parse request body
		var queryReq struct {
			Query    string `json:"query" binding:"required"`
			PageSize int    `json:"page_size,omitempty"`
		}

		if err := c.ShouldBindJSON(&queryReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query request"})
			return
		}

		// Set default and max page size
		pageSize := queryReq.PageSize
		if pageSize <= 0 {
			pageSize = 100
		}
		if pageSize > 1000 {
			pageSize = 1000 // Hard limit
		}

		// Execute query
		result, err := h.schemaRepo.QueryDatasetData(datasetID, queryReq.Query, pageSize)
		if err != nil {
			log.Printf("Error executing query: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Query execution failed: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

// InferSchema automatically infers schema from dataset data
func (h *SchemaHandlers) InferSchema() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("[DEBUG] InferSchema: Starting schema inference request")

		// Get user ID from auth middleware
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		userUUID, ok := userID.(uuid.UUID)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
			return
		}

		// Get dataset ID from URL
		datasetIDStr := c.Param("dataset_id")
		datasetID, err := uuid.Parse(datasetIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dataset ID"})
			return
		}

		log.Printf("[DEBUG] InferSchema: User %s requesting inference for dataset %s", userUUID, datasetID)

		// Check if user has access to this dataset
		hasAccess, err := h.schemaRepo.CheckDatasetAccess(datasetID, userUUID)
		if err != nil {
			log.Printf("[ERROR] InferSchema: Error checking dataset access: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify dataset access"})
			return
		}

		if !hasAccess {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to access this dataset"})
			return
		}

		// Get dataset information
		dataset, err := h.schemaRepo.GetDatasetByID(datasetID)
		if err != nil {
			log.Printf("[ERROR] InferSchema: Error fetching dataset: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch dataset information"})
			return
		}

		// Get dataset data for analysis
		headers, rows, err := h.schemaRepo.GetDatasetDataForInference(datasetID, 1000) // Analyze first 1000 rows
		if err != nil {
			log.Printf("[ERROR] InferSchema: Error fetching dataset data: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch dataset data for analysis"})
			return
		}

		if len(headers) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Dataset has no data to analyze"})
			return
		}

		log.Printf("[DEBUG] InferSchema: Analyzing %d columns and %d rows", len(headers), len(rows))

		// Perform schema inference
		inferredSchema, err := h.inferenceService.InferSchemaFromData(headers, rows, dataset.Name)
		if err != nil {
			log.Printf("[ERROR] InferSchema: Error during inference: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to infer schema: " + err.Error()})
			return
		}

		log.Printf("[DEBUG] InferSchema: Successfully inferred schema with confidence %.2f", inferredSchema.Confidence)

		c.JSON(http.StatusOK, gin.H{
			"inferred_schema": inferredSchema,
			"message":        "Schema inference completed successfully",
		})
	}
}
