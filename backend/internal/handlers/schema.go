package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/saurabh22suman/oreo.io/internal/models"
	"github.com/saurabh22suman/oreo.io/internal/repository"
)

// SchemaHandlers contains schema-related handlers
type SchemaHandlers struct {
	schemaRepo *repository.SchemaRepository
}

// NewSchemaHandlers creates new schema handlers
func NewSchemaHandlers(db *sqlx.DB) *SchemaHandlers {
	return &SchemaHandlers{
		schemaRepo: repository.NewSchemaRepository(db),
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify dataset access"})
			return
		}

		if !hasAccess {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view this dataset"})
			return
		}

		schema, err := h.schemaRepo.GetSchemaByDatasetID(datasetID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Schema not found"})
			return
		}

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

// GetDatasetData retrieves paginated dataset data
func (h *SchemaHandlers) GetDatasetData() gin.HandlerFunc {
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

		// Parse pagination parameters
		page := 1
		pageSize := 50 // Default page size

		if pageStr := c.Query("page"); pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				page = p
			}
		}

		if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
			if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 1000 {
				pageSize = ps
			}
		}

		// Check access
		hasAccess, err := h.schemaRepo.CheckDatasetAccess(datasetID, userUUID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify dataset access"})
			return
		}

		if !hasAccess {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view this dataset"})
			return
		}

		// Get data
		result, err := h.schemaRepo.GetDatasetData(datasetID, page, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve dataset data"})
			return
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
