package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// SampleDataHandlers provides endpoints for accessing sample datasets
type SampleDataHandlers struct {
	sampleDataPath string
}

// NewSampleDataHandlers creates a new instance of sample data handlers
func NewSampleDataHandlers() *SampleDataHandlers {
	return &SampleDataHandlers{
		sampleDataPath: "./sample-data",
	}
}

// DatasetInfo represents metadata about a dataset
type DatasetInfo struct {
	Filename    string            `json:"filename"`
	Category    string            `json:"category"`
	Size        int64             `json:"size"`
	Rows        int               `json:"rows"`
	Columns     []string          `json:"columns"`
	SampleData  []map[string]string `json:"sample_data,omitempty"`
	DownloadURL string            `json:"download_url"`
	Description string            `json:"description,omitempty"`
}

// ListSampleDatasets returns a list of available sample datasets
func (h *SampleDataHandlers) ListSampleDatasets(c *gin.Context) {
	datasets := make(map[string][]DatasetInfo)
	
	categories := []string{"transportation", "users", "finance", "mixed"}
	
	for _, category := range categories {
		categoryPath := filepath.Join(h.sampleDataPath, category)
		files, err := os.ReadDir(categoryPath)
		if err != nil {
			continue
		}
		
		var datasetInfos []DatasetInfo
		for _, file := range files {
			if !file.IsDir() && filepath.Ext(file.Name()) == ".csv" {
				info, err := h.getDatasetInfo(category, file.Name())
				if err != nil {
					continue
				}
				datasetInfos = append(datasetInfos, *info)
			}
		}
		
		if len(datasetInfos) > 0 {
			datasets[category] = datasetInfos
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    datasets,
	})
}

// GetSampleDatasetInfo returns detailed metadata about a specific dataset
func (h *SampleDataHandlers) GetSampleDatasetInfo(c *gin.Context) {
	category := c.Param("category")
	filename := c.Param("filename")
	
	// Add .csv extension if not provided
	if !strings.HasSuffix(filename, ".csv") {
		filename += ".csv"
	}
	
	info, err := h.getDatasetInfo(category, filename)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Dataset not found: " + err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    info,
	})
}

// DownloadSampleDataset allows downloading a specific sample dataset
func (h *SampleDataHandlers) DownloadSampleDataset(c *gin.Context) {
	category := c.Param("category")
	filename := c.Param("filename")
	
	// Add .csv extension if not provided
	if !strings.HasSuffix(filename, ".csv") {
		filename += ".csv"
	}
	
	// Validate category
	validCategories := map[string]bool{
		"transportation": true, "users": true, "finance": true, "mixed": true,
	}
	
	if !validCategories[category] {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid category. Valid categories: transportation, users, finance, mixed",
		})
		return
	}
	
	// Construct file path
	filePath := filepath.Join(h.sampleDataPath, category, filename)
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "File not found",
		})
		return
	}
	
	// Serve the file
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "text/csv")
	c.File(filePath)
}

// PreviewSampleDataset returns a preview of the dataset (first few rows)
func (h *SampleDataHandlers) PreviewSampleDataset(c *gin.Context) {
	category := c.Param("category")
	filename := c.Param("filename")
	
	// Add .csv extension if not provided
	if !strings.HasSuffix(filename, ".csv") {
		filename += ".csv"
	}
	
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100 // Max limit for preview
	}
	
	filePath := filepath.Join(h.sampleDataPath, category, filename)
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "File not found",
		})
		return
	}
	
	// Read and parse CSV
	file, err := os.Open(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to open file",
		})
		return
	}
	defer file.Close()
	
	reader := csv.NewReader(file)
	
	// Read header
	header, err := reader.Read()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to read CSV header",
		})
		return
	}
	
	// Read data rows up to limit
	var rows []map[string]string
	for i := 0; i < limit; i++ {
		record, err := reader.Read()
		if err != nil {
			break // End of file or error
		}
		
		row := make(map[string]string)
		for j, value := range record {
			if j < len(header) {
				row[header[j]] = value
			}
		}
		rows = append(rows, row)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"filename": filename,
			"category": category,
			"columns":  header,
			"rows":     rows,
			"count":    len(rows),
		},
	})
}

// getDatasetInfo is a helper function to get dataset metadata
func (h *SampleDataHandlers) getDatasetInfo(category, filename string) (*DatasetInfo, error) {
	filePath := filepath.Join(h.sampleDataPath, category, filename)
	
	// Check if file exists
	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found")
	}
	
	// Count rows and get columns
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file")
	}
	defer file.Close()
	
	reader := csv.NewReader(file)
	
	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header")
	}
	
	// Count rows
	rowCount := 0
	for {
		_, err := reader.Read()
		if err != nil {
			break
		}
		rowCount++
	}
	
	// Get sample data (first 3 rows)
	file.Seek(0, 0)
	reader = csv.NewReader(file)
	reader.Read() // Skip header
	
	var sampleData []map[string]string
	for i := 0; i < 3; i++ {
		record, err := reader.Read()
		if err != nil {
			break
		}
		
		row := make(map[string]string)
		for j, value := range record {
			if j < len(header) {
				row[header[j]] = value
			}
		}
		sampleData = append(sampleData, row)
	}
	
	// Add description based on filename
	description := h.getDatasetDescription(filename)
	
	return &DatasetInfo{
		Filename:    filename,
		Category:    category,
		Size:        fileInfo.Size(),
		Rows:        rowCount,
		Columns:     header,
		SampleData:  sampleData,
		DownloadURL: fmt.Sprintf("/api/v1/sample-data/%s/%s/download", category, strings.TrimSuffix(filename, ".csv")),
		Description: description,
	}, nil
}

// getDatasetDescription returns a description for known datasets
func (h *SampleDataHandlers) getDatasetDescription(filename string) string {
	descriptions := map[string]string{
		"airlines_flights_data.csv": "Comprehensive flight booking data from various Indian airlines including pricing, routes, and booking details. Perfect for transportation analytics and price optimization studies.",
	}
	
	if desc, exists := descriptions[filename]; exists {
		return desc
	}
	
	return "Sample dataset for testing and development purposes."
}
