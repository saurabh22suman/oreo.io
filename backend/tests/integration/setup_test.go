package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	testBaseURL = "http://localhost:8080"
	testClient  = &http.Client{Timeout: 30 * time.Second}
)

// TestMain sets up and tears down the test environment
func TestMain(m *testing.M) {
	// Ensure the server is ready before running tests
	if err := waitForServer(); err != nil {
		fmt.Printf("Server is not ready: %v\n", err)
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup any test data
	cleanupTestData()

	os.Exit(code)
}

// waitForServer waits for the server to be ready
func waitForServer() error {
	maxAttempts := 30
	for i := 0; i < maxAttempts; i++ {
		resp, err := testClient.Get(testBaseURL + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("server not ready after 30 attempts")
}

// cleanupTestData removes any test data created during tests
func cleanupTestData() {
	// This function can be expanded to clean up specific test data
	// For now, it's a placeholder
}

// TestUser represents a test user for authentication
type TestUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// getTestUser returns a unique test user for each test run
func getTestUser() TestUser {
	timestamp := time.Now().UnixNano()
	return TestUser{
		Email:    fmt.Sprintf("testuser%d@example.com", timestamp),
		Password: "testpassword123",
		Name:     "Test User",
	}
}

// registerTestUser registers a test user and returns the user data
func registerTestUser(t *testing.T, user TestUser) map[string]interface{} {
	registerReq := map[string]interface{}{
		"email":    user.Email,
		"password": user.Password,
		"name":     user.Name,
	}

	jsonData, err := json.Marshal(registerReq)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", testBaseURL+"/api/v1/auth/register", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := testClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	return result
}

// loginTestUser logs in a test user and returns the access token
func loginTestUser(t *testing.T, user TestUser) string {
	loginReq := map[string]interface{}{
		"email":    user.Email,
		"password": user.Password,
	}

	jsonData, err := json.Marshal(loginReq)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", testBaseURL+"/api/v1/auth/login", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := testClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	accessToken, ok := result["access_token"].(string)
	require.True(t, ok, "access_token should be a string")
	require.NotEmpty(t, accessToken)

	return accessToken
}

// createTestUserAndLogin creates a test user and returns the access token
func createTestUserAndLogin(t *testing.T) (TestUser, string) {
	user := getTestUser()
	registerTestUser(t, user)
	token := loginTestUser(t, user)
	return user, token
}

// makeAuthenticatedRequest makes an HTTP request with authentication
func makeAuthenticatedRequest(t *testing.T, method, url string, body interface{}, token string) (*http.Response, []byte) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		require.NoError(t, err)
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, testBaseURL+url, reqBody)
	require.NoError(t, err)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := testClient.Do(req)
	require.NoError(t, err)

	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	// Create a new response with the body for the caller
	resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	return resp, bodyBytes
}

// makeRequest makes an HTTP request without authentication
func makeRequest(t *testing.T, method, url string, body interface{}) (*http.Response, []byte) {
	return makeAuthenticatedRequest(t, method, url, body, "")
}
