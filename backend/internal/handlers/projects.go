package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/saurabh22suman/oreo.io/internal/models"
	"github.com/saurabh22suman/oreo.io/internal/repository"
)

// ProjectHandlers contains project-related handlers
type ProjectHandlers struct {
	projectRepo *repository.ProjectRepository
}

// NewProjectHandlers creates new project handlers
func NewProjectHandlers(db *sqlx.DB) *ProjectHandlers {
	log.Printf("Creating new ProjectHandlers with db: %+v", db)
	handlers := &ProjectHandlers{
		projectRepo: repository.NewProjectRepository(db),
	}
	log.Printf("Created ProjectHandlers: %+v", handlers)
	return handlers
}

// GetProjects returns all projects for the authenticated user
func (h *ProjectHandlers) GetProjects() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("ProjectHandlers.GetProjects called - NEW HANDLER IS WORKING!")
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

		// Get projects from repository
		projects, err := h.projectRepo.GetByOwnerID(userUUID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to retrieve projects",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"projects": projects,
			"count":    len(projects),
		})
	}
}

// CreateProject creates a new project
func (h *ProjectHandlers) CreateProject() gin.HandlerFunc {
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

		// Parse request body
		var req models.CreateProjectRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request data",
				"details": err.Error(),
			})
			return
		}

		// Validate request
		if err := req.Validate(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation failed",
				"details": err.Error(),
			})
			return
		}

		// Create project model
		project := req.ToProject(userUUID)

		// Save to database
		if err := h.projectRepo.Create(project); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to create project",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Project created successfully",
			"project": project,
		})
	}
}

// GetProject returns a specific project
func (h *ProjectHandlers) GetProject() gin.HandlerFunc {
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

		// Parse project ID from URL
		projectIDStr := c.Param("id")
		projectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid project ID",
				"details": err.Error(),
			})
			return
		}

		// Check if project exists and is owned by user
		exists, err = h.projectRepo.Exists(projectID, userUUID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to check project ownership",
				"details": err.Error(),
			})
			return
		}

		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}

		// Get project
		project, err := h.projectRepo.GetByID(projectID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to retrieve project",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"project": project})
	}
}

// UpdateProject updates an existing project
func (h *ProjectHandlers) UpdateProject() gin.HandlerFunc {
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

		// Parse project ID from URL
		projectIDStr := c.Param("id")
		projectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid project ID",
				"details": err.Error(),
			})
			return
		}

		// Check if project exists and is owned by user
		exists, err = h.projectRepo.Exists(projectID, userUUID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to check project ownership",
				"details": err.Error(),
			})
			return
		}

		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}

		// Parse request body
		var req models.UpdateProjectRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request data",
				"details": err.Error(),
			})
			return
		}

		// Validate request
		if err := req.Validate(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation failed",
				"details": err.Error(),
			})
			return
		}

		// Check if there are any updates
		if !req.HasUpdates() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No updates provided"})
			return
		}

		// Update project
		project, err := h.projectRepo.Update(projectID, &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to update project",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Project updated successfully",
			"project": project,
		})
	}
}

// DeleteProject deletes a project
func (h *ProjectHandlers) DeleteProject() gin.HandlerFunc {
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

		// Parse project ID from URL
		projectIDStr := c.Param("id")
		projectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid project ID",
				"details": err.Error(),
			})
			return
		}

		// Delete project
		if err := h.projectRepo.Delete(projectID, userUUID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to delete project",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Project deleted successfully"})
	}
}
