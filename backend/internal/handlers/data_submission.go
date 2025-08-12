package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/saurabh22suman/oreo.io/internal/models"
	"github.com/saurabh22suman/oreo.io/internal/repository"
	"github.com/saurabh22suman/oreo.io/internal/services"
)

type DataSubmissionHandlers struct {
	submissionRepo  *repository.DataSubmissionRepository
	schemaRepo      *repository.SchemaRepository
	validationSvc   *services.ValidationService
}

func NewDataSubmissionHandlers(
	submissionRepo *repository.DataSubmissionRepository,
	schemaRepo *repository.SchemaRepository,
	validationSvc *services.ValidationService,
) *DataSubmissionHandlers {
	return &DataSubmissionHandlers{
		submissionRepo: submissionRepo,
		schemaRepo:     schemaRepo,
		validationSvc:  validationSvc,
	}
}

// SubmitDataForAppend handles uploading data for appending to existing dataset
func (h *DataSubmissionHandlers) SubmitDataForAppend() gin.HandlerFunc {
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

		// Get dataset ID from URL params
		datasetIDStr := c.Param("dataset_id")
		datasetID, err := uuid.Parse(datasetIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dataset ID"})
			return
		}

		// Check if user has access to this dataset
		hasAccess, err := h.submissionRepo.CheckDatasetAccess(datasetID, userUUID)
		if err != nil {
			log.Printf("Error checking dataset access: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify dataset access"})
			return
		}

		if !hasAccess {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to submit data to this dataset"})
			return
		}

		// Get file from form
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
			return
		}
		defer file.Close()

		// Validate file type (only CSV for now)
		if !isValidCSVFile(header.Filename) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid file type. Only CSV files are supported for data append",
			})
			return
		}

		// Validate file size (10MB limit for append operations)
		const maxFileSize = 10 * 1024 * 1024 // 10MB
		if header.Size > maxFileSize {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "File size exceeds 10MB limit for data append",
			})
			return
		}

		// Create submission record
		submission := &models.DataSubmission{
			ID:          uuid.New(),
			DatasetID:   datasetID,
			SubmittedBy: userUUID,
			FileName:    header.Filename,
			FileSize:    header.Size,
			Status:      models.DataSubmissionStatusPending,
			SubmittedAt: time.Now(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Save file to submissions directory
		submissionDir := "submissions"
		if err := os.MkdirAll(submissionDir, 0755); err != nil {
			log.Printf("Error creating submission directory: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create submission directory"})
			return
		}

		filename := fmt.Sprintf("%s_%s", submission.ID.String(), header.Filename)
		filepath := filepath.Join(submissionDir, filename)
		submission.FilePath = filepath

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

		// Validate the data against schema and business rules
		validationResult, stagingData, err := h.validationSvc.ValidateDataSubmission(filepath, datasetID)
		if err != nil {
			log.Printf("Error validating submission: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate submission"})
			return
		}

		// Store validation results
		validationJSON, _ := json.Marshal(validationResult)
		validationRawMessage := json.RawMessage(validationJSON)
		submission.ValidationResults = &validationRawMessage
		submission.RowCount = validationResult.TotalRows

		// Save submission to database
		if err := h.submissionRepo.CreateSubmission(submission); err != nil {
			log.Printf("Error creating submission: %v", err)
			os.Remove(filepath) // Clean up uploaded file
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save submission"})
			return
		}

		// Save staging data
		for _, stagingRow := range stagingData {
			stagingRow.SubmissionID = submission.ID
		}

		if err := h.submissionRepo.CreateStagingData(stagingData); err != nil {
			log.Printf("Error saving staging data: %v", err)
			// Don't fail the entire submission, but log the error
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":           "Data submission created successfully",
			"submission":        submission,
			"validation_result": validationResult,
		})
	}
}

// GetDataSubmissions retrieves submissions for a dataset
func (h *DataSubmissionHandlers) GetDataSubmissions() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get dataset ID from URL params
		datasetIDStr := c.Param("dataset_id")
		datasetID, err := uuid.Parse(datasetIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dataset ID"})
			return
		}

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

		// Check if user has access to this dataset
		hasAccess, err := h.submissionRepo.CheckDatasetAccess(datasetID, userUUID)
		if err != nil {
			log.Printf("Error checking dataset access: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify dataset access"})
			return
		}

		if !hasAccess {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view submissions for this dataset"})
			return
		}

		submissions, err := h.submissionRepo.GetSubmissionsByDataset(datasetID)
		if err != nil {
			log.Printf("Error getting submissions: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve submissions"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"submissions": submissions,
			"count":       len(submissions),
		})
	}
}

// GetSubmissionDetails retrieves detailed information about a submission including staging data
func (h *DataSubmissionHandlers) GetSubmissionDetails() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get submission ID from URL params
		submissionIDStr := c.Param("submission_id")
		submissionID, err := uuid.Parse(submissionIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid submission ID"})
			return
		}

		// Get pagination parameters
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
		
		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 50
		}

		offset := (page - 1) * pageSize

		// Get submission details
		submission, err := h.submissionRepo.GetSubmissionWithDetails(submissionID)
		if err != nil {
			log.Printf("Error getting submission details: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve submission details"})
			return
		}

		// Get user ID and check access
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

		// Check if user has access to this dataset
		hasAccess, err := h.submissionRepo.CheckDatasetAccess(submission.DatasetID, userUUID)
		if err != nil {
			log.Printf("Error checking dataset access: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify dataset access"})
			return
		}

		if !hasAccess {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view this submission"})
			return
		}

		// Get staging data
		stagingData, err := h.submissionRepo.GetStagingData(submissionID, pageSize, offset)
		if err != nil {
			log.Printf("Error getting staging data: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve staging data"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"submission":   submission,
			"staging_data": stagingData,
			"pagination": gin.H{
				"page":      page,
				"page_size": pageSize,
				"total":     submission.RowCount,
			},
		})
	}
}

// UpdateStagingData handles live editing of staging data
func (h *DataSubmissionHandlers) UpdateStagingData() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get staging data ID from URL params
		stagingIDStr := c.Param("staging_id")
		stagingID, err := uuid.Parse(stagingIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid staging data ID"})
			return
		}

		var updateRequest struct {
			Data map[string]interface{} `json:"data" binding:"required"`
		}

		if err := c.ShouldBindJSON(&updateRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// TODO: Add validation logic here to validate the updated data
		// For now, we'll assume it's valid
		dataJSON, _ := json.Marshal(updateRequest.Data)
		validationErrors := json.RawMessage("[]")

		err = h.submissionRepo.UpdateStagingDataRow(stagingID, dataJSON, models.ValidationStatusValid, &validationErrors)
		if err != nil {
			log.Printf("Error updating staging data: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update staging data"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Staging data updated successfully",
		})
	}
}

// Admin endpoints

// GetPendingSubmissions retrieves all pending submissions for admin review
func (h *DataSubmissionHandlers) GetPendingSubmissions() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID and check admin privileges
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

		// Check if user is admin
		isAdmin, err := h.submissionRepo.IsUserAdmin(userUUID)
		if err != nil {
			log.Printf("Error checking admin status: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify admin status"})
			return
		}

		if !isAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin privileges required"})
			return
		}

		submissions, err := h.submissionRepo.GetPendingSubmissions()
		if err != nil {
			log.Printf("Error getting pending submissions: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve pending submissions"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"submissions": submissions,
			"count":       len(submissions),
		})
	}
}

// ReviewSubmission handles admin review of a submission
func (h *DataSubmissionHandlers) ReviewSubmission() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get submission ID from URL params
		submissionIDStr := c.Param("submission_id")
		submissionID, err := uuid.Parse(submissionIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid submission ID"})
			return
		}

		// Get user ID and check admin privileges
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

		// Check if user is admin
		isAdmin, err := h.submissionRepo.IsUserAdmin(userUUID)
		if err != nil {
			log.Printf("Error checking admin status: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify admin status"})
			return
		}

		if !isAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin privileges required"})
			return
		}

		var reviewRequest models.UpdateDataSubmissionRequest
		if err := c.ShouldBindJSON(&reviewRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Update submission status
		err = h.submissionRepo.UpdateSubmissionStatus(submissionID, reviewRequest.Status, reviewRequest.AdminNotes, userUUID)
		if err != nil {
			log.Printf("Error updating submission status: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update submission status"})
			return
		}

		// If approved, apply the data to the target dataset
		if reviewRequest.Status == models.DataSubmissionStatusApproved {
			submission, err := h.submissionRepo.GetSubmission(submissionID)
			if err != nil {
				log.Printf("Error getting submission for approval: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve submission"})
				return
			}

			err = h.submissionRepo.ApplyStagingDataToDataset(submissionID, submission.DatasetID, userUUID)
			if err != nil {
				log.Printf("Error applying data to dataset: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to apply data to dataset"})
				return
			}

			// Mark submission as applied
			err = h.submissionRepo.MarkSubmissionApplied(submissionID)
			if err != nil {
				log.Printf("Error marking submission as applied: %v", err)
				// Don't fail the request, just log the error
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Submission review completed successfully",
		})
	}
}

// Business Rules endpoints

// CreateBusinessRule creates a new business rule for a dataset
func (h *DataSubmissionHandlers) CreateBusinessRule() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get dataset ID from URL params
		datasetIDStr := c.Param("dataset_id")
		datasetID, err := uuid.Parse(datasetIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dataset ID"})
			return
		}

		// Get user ID
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

		var ruleRequest struct {
			RuleName     string                     `json:"rule_name" binding:"required"`
			RuleType     string                     `json:"rule_type" binding:"required"`
			RuleConfig   models.BusinessRuleConfig  `json:"rule_config" binding:"required"`
			ErrorMessage string                     `json:"error_message" binding:"required"`
			Priority     int                        `json:"priority"`
		}

		if err := c.ShouldBindJSON(&ruleRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Create business rule
		configJSON, _ := json.Marshal(ruleRequest.RuleConfig)
		rule := &models.DatasetBusinessRule{
			ID:           uuid.New(),
			DatasetID:    datasetID,
			RuleName:     ruleRequest.RuleName,
			RuleType:     ruleRequest.RuleType,
			RuleConfig:   configJSON,
			ErrorMessage: ruleRequest.ErrorMessage,
			IsActive:     true,
			Priority:     ruleRequest.Priority,
			CreatedBy:    userUUID,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		if err := h.submissionRepo.CreateBusinessRule(rule); err != nil {
			log.Printf("Error creating business rule: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create business rule"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Business rule created successfully",
			"rule":    rule,
		})
	}
}

// GetBusinessRules retrieves business rules for a dataset
func (h *DataSubmissionHandlers) GetBusinessRules() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get dataset ID from URL params
		datasetIDStr := c.Param("dataset_id")
		datasetID, err := uuid.Parse(datasetIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dataset ID"})
			return
		}

		rules, err := h.submissionRepo.GetBusinessRules(datasetID)
		if err != nil {
			log.Printf("Error getting business rules: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve business rules"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"rules": rules,
			"count": len(rules),
		})
	}
}

// Helper functions

func isValidCSVFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".csv"
}
