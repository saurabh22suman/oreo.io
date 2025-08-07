package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetProjects returns all projects for the authenticated user
func GetProjects() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement get projects
		c.JSON(http.StatusNotImplemented, gin.H{
			"message": "Get projects endpoint - coming soon",
			"status":  "not_implemented",
		})
	}
}

// CreateProject creates a new project
func CreateProject() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement create project
		c.JSON(http.StatusNotImplemented, gin.H{
			"message": "Create project endpoint - coming soon",
			"status":  "not_implemented",
		})
	}
}

// GetProject returns a specific project
func GetProject() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement get project
		c.JSON(http.StatusNotImplemented, gin.H{
			"message": "Get project endpoint - coming soon",
			"status":  "not_implemented",
		})
	}
}

// UpdateProject updates an existing project
func UpdateProject() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement update project
		c.JSON(http.StatusNotImplemented, gin.H{
			"message": "Update project endpoint - coming soon",
			"status":  "not_implemented",
		})
	}
}

// DeleteProject deletes a project
func DeleteProject() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement delete project
		c.JSON(http.StatusNotImplemented, gin.H{
			"message": "Delete project endpoint - coming soon",
			"status":  "not_implemented",
		})
	}
}
