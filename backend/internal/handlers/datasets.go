package handlers

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/tealeg/xlsx/v3"

	"github.com/saurabh22suman/oreo.io/internal/models"
	"github.com/saurabh22suman/oreo.io/internal/repository"
)

// DatasetHandlers contains dataset-related handlers
type DatasetHandlers struct {
	datasetRepo *repository.DatasetRepository
}

// NewDatasetHandlers creates new dataset handlers
func NewDatasetHandlers(db *sqlx.DB) *DatasetHandlers {
	return &DatasetHandlers{
		datasetRepo: repository.NewDatasetRepository(db),
	}
}

// UploadDataset handles file upload for datasets
func (h *DatasetHandlers) UploadDataset() gin.HandlerFunc {
	return func(c *gin.Context) {
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

		// Get project ID from form
		projectIDStr := c.PostForm("project_id")
		if projectIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
			return
		}

		projectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
			return
		}

		// Check if user has access to upload to this project
		hasAccess, err := h.datasetRepo.CheckProjectAccess(projectID, userUUID)
		if err != nil {
			log.Printf("Error checking project access: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify project access"})
			return
		}

		if !hasAccess {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to upload to this project"})
			return
		}

		// Get file from form
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
			return
		}
		defer file.Close()

		// Validate file type
		if !isValidFileType(header.Filename) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid file type. Only CSV and Excel files are supported",
			})
			return
		}

		// Validate file size (50MB limit)
		const maxFileSize = 50 * 1024 * 1024 // 50MB
		if header.Size > maxFileSize {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "File size exceeds 50MB limit",
			})
			return
		}

		// Get optional dataset metadata
		name := c.PostForm("name")
		if name == "" {
			name = strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename))
		}
		description := c.PostForm("description")

		// Create dataset record
		dataset := &models.Dataset{
			ID:          uuid.New(),
			ProjectID:   projectID,
			Name:        name,
			Description: description,
			FileName:    header.Filename,
			FileSize:    header.Size,
			MimeType:    header.Header.Get("Content-Type"),
			Status:      models.DatasetStatusProcessing,
			UploadedBy:  userUUID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Save file to uploads directory
		uploadDir := "uploads"
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			log.Printf("Error creating upload directory: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
			return
		}

		filename := fmt.Sprintf("%s_%s", dataset.ID.String(), header.Filename)
		filepath := filepath.Join(uploadDir, filename)
		dataset.FilePath = filepath

		// Save file to disk
		out, err := os.Create(filepath)
		if err != nil {
			log.Printf("Error creating file: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}
		defer out.Close()

		_, err = io.Copy(out, file)
		if err != nil {
			log.Printf("Error copying file: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}

		// Process file to get row and column count
		rowCount, columnCount, err := h.processFile(filepath, header.Filename)
		if err != nil {
			log.Printf("Error processing file: %v", err)
			dataset.Status = models.DatasetStatusError
		} else {
			dataset.RowCount = rowCount
			dataset.ColumnCount = columnCount
			dataset.Status = models.DatasetStatusReady
		}

		// Save dataset to database
		if err := h.datasetRepo.Create(dataset); err != nil {
			log.Printf("Error creating dataset: %v", err)
			// Clean up uploaded file
			os.Remove(filepath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save dataset"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Dataset uploaded successfully",
			"dataset": dataset,
		})
	}
}

// GetDatasets returns datasets for a project
func (h *DatasetHandlers) GetDatasets() gin.HandlerFunc {
	return func(c *gin.Context) {
		projectIDStr := c.Param("project_id")
		projectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
			return
		}

		datasets, err := h.datasetRepo.GetByProjectID(projectID)
		if err != nil {
			log.Printf("Error fetching datasets: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch datasets"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"datasets": datasets,
			"count":    len(datasets),
		})
	}
}

// GetUserDatasets returns all datasets uploaded by the authenticated user
func (h *DatasetHandlers) GetUserDatasets() gin.HandlerFunc {
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

		datasets, err := h.datasetRepo.GetByUserID(userUUID)
		if err != nil {
			log.Printf("Error fetching user datasets: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch datasets"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"datasets": datasets,
			"count":    len(datasets),
		})
	}
}

// DeleteDataset deletes a dataset
func (h *DatasetHandlers) DeleteDataset() gin.HandlerFunc {
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

		datasetIDStr := c.Param("id")
		datasetID, err := uuid.Parse(datasetIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dataset ID"})
			return
		}

		// Get dataset to find file path
		dataset, err := h.datasetRepo.GetByID(datasetID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Dataset not found"})
			return
		}

		// Delete from database
		if err := h.datasetRepo.Delete(datasetID, userUUID); err != nil {
			log.Printf("Error deleting dataset: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete dataset"})
			return
		}

		// Delete file from disk
		if err := os.Remove(dataset.FilePath); err != nil {
			log.Printf("Warning: Failed to delete file %s: %v", dataset.FilePath, err)
		}

		c.JSON(http.StatusOK, gin.H{"message": "Dataset deleted successfully"})
	}
}

// Helper functions

func isValidFileType(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".csv" || ext == ".xlsx" || ext == ".xls"
}

func (h *DatasetHandlers) processFile(filePath, filename string) (int, int, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".csv":
		return h.processCSV(filePath)
	case ".xlsx", ".xls":
		return h.processExcel(filePath)
	default:
		return 0, 0, fmt.Errorf("unsupported file type: %s", ext)
	}
}

func (h *DatasetHandlers) processCSV(filePath string) (int, int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return 0, 0, err
	}

	if len(records) == 0 {
		return 0, 0, nil
	}

	rowCount := len(records) - 1 // Subtract header row
	columnCount := len(records[0])

	return rowCount, columnCount, nil
}

func (h *DatasetHandlers) processExcel(filePath string) (int, int, error) {
	workbook, err := xlsx.OpenFile(filePath)
	if err != nil {
		return 0, 0, err
	}

	if len(workbook.Sheets) == 0 {
		return 0, 0, nil
	}

	sheet := workbook.Sheets[0]  // Use first sheet
	rowCount := sheet.MaxRow - 1 // Subtract header row
	columnCount := sheet.MaxCol

	if rowCount < 0 {
		rowCount = 0
	}

	return rowCount, columnCount, nil
}
