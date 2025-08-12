package integration

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Project represents a project
type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedBy   string `json:"created_by"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ProjectResponse represents the response from project endpoints
type ProjectResponse struct {
	Project Project `json:"project"`
}

// ProjectsResponse represents the response from GET /projects
type ProjectsResponse struct {
	Projects []Project `json:"projects"`
	Count    int       `json:"count"`
}

func TestProjectFlow(t *testing.T) {
	// Create and login a user for this test
	user, token := createTestUserAndLogin(t)
	var createdProjectID string

	t.Run("Create Project", func(t *testing.T) {
		createReq := map[string]interface{}{
			"name":        "Test Project " + time.Now().Format("20060102150405"),
			"description": "This is a test project for integration testing",
		}

		resp, bodyBytes := makeAuthenticatedRequest(t, "POST", "/api/v1/projects", createReq, token)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var projectResp ProjectResponse
		err := json.Unmarshal(bodyBytes, &projectResp)
		require.NoError(t, err)

		project := projectResp.Project
		assert.NotEmpty(t, project.ID)
		assert.Equal(t, createReq["name"], project.Name)
		assert.Equal(t, createReq["description"], project.Description)
		assert.NotEmpty(t, project.CreatedAt)

		createdProjectID = project.ID
	})

	t.Run("Get User Projects", func(t *testing.T) {
		resp, bodyBytes := makeAuthenticatedRequest(t, "GET", "/api/v1/projects", nil, token)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var projectsResp ProjectsResponse
		err := json.Unmarshal(bodyBytes, &projectsResp)
		require.NoError(t, err)

		assert.Greater(t, len(projectsResp.Projects), 0)
		assert.Equal(t, len(projectsResp.Projects), projectsResp.Count)

		// Check if our created project is in the list
		found := false
		for _, project := range projectsResp.Projects {
			if project.ID == createdProjectID {
				found = true
				break
			}
		}
		assert.True(t, found, "Created project should be in the list")
	})

	t.Run("Get Single Project", func(t *testing.T) {
		resp, bodyBytes := makeAuthenticatedRequest(t, "GET", "/api/v1/projects/"+createdProjectID, nil, token)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var projectResp ProjectResponse
		err := json.Unmarshal(bodyBytes, &projectResp)
		require.NoError(t, err)

		project := projectResp.Project
		assert.Equal(t, createdProjectID, project.ID)
		assert.NotEmpty(t, project.Name)
		assert.NotEmpty(t, project.Description)
	})

	t.Run("Update Project", func(t *testing.T) {
		updateReq := map[string]interface{}{
			"name":        "Updated Test Project",
			"description": "This project has been updated",
		}

		resp, bodyBytes := makeAuthenticatedRequest(t, "PUT", "/api/v1/projects/"+createdProjectID, updateReq, token)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var projectResp ProjectResponse
		err := json.Unmarshal(bodyBytes, &projectResp)
		require.NoError(t, err)

		project := projectResp.Project
		assert.Equal(t, createdProjectID, project.ID)
		assert.Equal(t, updateReq["name"], project.Name)
		assert.Equal(t, updateReq["description"], project.Description)
	})

	t.Run("Delete Project", func(t *testing.T) {
		resp, _ := makeAuthenticatedRequest(t, "DELETE", "/api/v1/projects/"+createdProjectID, nil, token)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify project is deleted
		getResp, _ := makeAuthenticatedRequest(t, "GET", "/api/v1/projects/"+createdProjectID, nil, token)
		defer getResp.Body.Close()

		assert.Equal(t, http.StatusNotFound, getResp.StatusCode)
	})

	t.Run("Unauthorized Access - should return 401", func(t *testing.T) {
		resp, _ := makeRequest(t, "GET", "/api/v1/projects", nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	_ = user // Use the user variable to avoid unused variable error
}
