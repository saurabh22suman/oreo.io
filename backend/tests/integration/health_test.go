package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// HealthResponse represents the response from health endpoints
type HealthResponse struct {
	Status    string `json:"status"`
	Database  string `json:"database"`
	Redis     string `json:"redis"`
	Timestamp string `json:"timestamp"`
	Type      string `json:"type,omitempty"`
}

func TestHealthEndpoints(t *testing.T) {
	t.Run("GET /health", func(t *testing.T) {
		resp, err := http.Get(testBaseURL + "/health")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var health HealthResponse
		err = json.NewDecoder(resp.Body).Decode(&health)
		require.NoError(t, err)

		assert.Equal(t, "healthy", health.Status)
		assert.NotEmpty(t, health.Database)
		assert.NotEmpty(t, health.Redis)
		assert.NotEmpty(t, health.Timestamp)
	})

	t.Run("GET /health/db", func(t *testing.T) {
		resp, err := http.Get(testBaseURL + "/health/db")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var health HealthResponse
		err = json.NewDecoder(resp.Body).Decode(&health)
		require.NoError(t, err)

		assert.Equal(t, "healthy", health.Status)
		// Health/db endpoint might return different structure
	})

	t.Run("GET /health/redis", func(t *testing.T) {
		resp, err := http.Get(testBaseURL + "/health/redis")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var health HealthResponse
		err = json.NewDecoder(resp.Body).Decode(&health)
		require.NoError(t, err)

		assert.Equal(t, "healthy", health.Status)
		// Health/redis endpoint might return different structure
	})
}
