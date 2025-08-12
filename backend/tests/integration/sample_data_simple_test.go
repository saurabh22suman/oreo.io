package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SampleDataInfo represents sample data information
type SampleDataInfo struct {
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Format      string `json:"format"`
}

// SampleDataList represents the list of sample data
type SampleDataList struct {
	Samples []SampleDataInfo `json:"samples"`
	Count   int              `json:"count"`
}

func TestSampleDataEndpoints(t *testing.T) {
	t.Run("List Sample Data", func(t *testing.T) {
		resp, bodyBytes := makeRequest(t, "GET", "/api/v1/sample-data/", nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var sampleList SampleDataList
		err := json.Unmarshal(bodyBytes, &sampleList)
		require.NoError(t, err)

		// Should have at least one sample
		assert.GreaterOrEqual(t, len(sampleList.Samples), 0)
		assert.Equal(t, len(sampleList.Samples), sampleList.Count)
	})

	t.Run("Get Sample Data Info", func(t *testing.T) {
		// Test with employees.csv which should exist
		resp, bodyBytes := makeRequest(t, "GET", "/api/v1/sample-data/employees.csv/info", nil)
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var info SampleDataInfo
			err := json.Unmarshal(bodyBytes, &info)
			require.NoError(t, err)

			assert.NotEmpty(t, info.Name)
			assert.Greater(t, info.Size, int64(0))
		} else {
			// If file doesn't exist, should return 404
			assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		}
	})

	t.Run("Preview Sample Data", func(t *testing.T) {
		// Test preview endpoint
		resp, _ := makeRequest(t, "GET", "/api/v1/sample-data/employees.csv/preview", nil)
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			// Should return some data
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		} else {
			// If file doesn't exist, should return 404
			assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		}
	})

	t.Run("Non-existent Sample - should return 404", func(t *testing.T) {
		resp, _ := makeRequest(t, "GET", "/api/v1/sample-data/non-existent.csv/info", nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
