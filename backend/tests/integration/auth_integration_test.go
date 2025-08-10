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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	baseURL = "http://localhost:8080"
	timeout = 30 * time.Second
)

// TestUser represents a test user for registration/login
type TestUser struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents the response from auth endpoints
type AuthResponse struct {
	User struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// UserResponse represents the response from /auth/me endpoint
type UserResponse struct {
	User struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"user"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

// HTTPClient creates a configured HTTP client
func HTTPClient() *http.Client {
	return &http.Client{
		Timeout: timeout,
	}
}

// WaitForServer waits for the server to be ready
func WaitForServer(t *testing.T) {
	client := HTTPClient()

	for i := 0; i < 30; i++ {
		resp, err := client.Get(baseURL + "/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(2 * time.Second)
	}

	t.Fatal("Server did not become ready within timeout")
}

// MakeRequest makes an HTTP request and returns the response
func MakeRequest(t *testing.T, method, url string, body interface{}, headers map[string]string) (*http.Response, []byte) {
	client := HTTPClient()

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err)
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, baseURL+url, bodyReader)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	require.NoError(t, err)

	responseBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	return resp, responseBody
}

// TestMain sets up the test environment
func TestMain(m *testing.M) {
	// Wait for services to be ready
	fmt.Println("Waiting for services to be ready...")

	// Run tests
	code := m.Run()

	os.Exit(code)
}

// TestHealthEndpoints tests all health check endpoints
func TestHealthEndpoints(t *testing.T) {
	WaitForServer(t)

	tests := []struct {
		name     string
		endpoint string
		expected int
	}{
		{
			name:     "Main health check",
			endpoint: "/health",
			expected: 200,
		},
		{
			name:     "Database health check",
			endpoint: "/health/db",
			expected: 200,
		},
		{
			name:     "Redis health check",
			endpoint: "/health/redis",
			expected: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := MakeRequest(t, "GET", tt.endpoint, nil, nil)

			t.Logf("Response status: %d", resp.StatusCode)
			t.Logf("Response body: %s", string(body))

			assert.Equal(t, tt.expected, resp.StatusCode, "Health check should return expected status")
		})
	}
}

// TestUserRegistration tests user registration functionality
func TestUserRegistration(t *testing.T) {
	WaitForServer(t)

	testUser := TestUser{
		Name:     "Integration Test User",
		Email:    fmt.Sprintf("test_%d@example.com", time.Now().Unix()),
		Password: "testpassword123",
	}

	t.Run("Successful registration", func(t *testing.T) {
		resp, body := MakeRequest(t, "POST", "/api/v1/auth/register", testUser, nil)

		t.Logf("Registration response status: %d", resp.StatusCode)
		t.Logf("Registration response body: %s", string(body))

		if resp.StatusCode != 201 {
			var errorResp ErrorResponse
			err := json.Unmarshal(body, &errorResp)
			if err == nil {
				t.Logf("Error response: %+v", errorResp)
			}
		}

		assert.Equal(t, 201, resp.StatusCode, "Registration should succeed")

		var authResp AuthResponse
		err := json.Unmarshal(body, &authResp)
		require.NoError(t, err)

		assert.NotEmpty(t, authResp.AccessToken)
		assert.NotEmpty(t, authResp.RefreshToken)
		assert.Equal(t, testUser.Email, authResp.User.Email)
		assert.Equal(t, testUser.Name, authResp.User.Name)
	})

	t.Run("Duplicate email registration", func(t *testing.T) {
		// Try to register the same user again
		resp, body := MakeRequest(t, "POST", "/api/v1/auth/register", testUser, nil)

		t.Logf("Duplicate registration response status: %d", resp.StatusCode)
		t.Logf("Duplicate registration response body: %s", string(body))

		assert.Equal(t, 400, resp.StatusCode, "Duplicate registration should fail")

		var errorResp ErrorResponse
		err := json.Unmarshal(body, &errorResp)
		require.NoError(t, err)

		assert.False(t, errorResp.Success)
		assert.Contains(t, errorResp.Message, "already exists")
	})

	t.Run("Invalid registration data", func(t *testing.T) {
		invalidUser := TestUser{
			Name:     "",
			Email:    "invalid-email",
			Password: "123",
		}

		resp, body := MakeRequest(t, "POST", "/api/v1/auth/register", invalidUser, nil)

		t.Logf("Invalid registration response status: %d", resp.StatusCode)
		t.Logf("Invalid registration response body: %s", string(body))

		assert.Equal(t, 400, resp.StatusCode, "Invalid registration should fail")
	})
}

// TestUserLogin tests user login functionality
func TestUserLogin(t *testing.T) {
	WaitForServer(t)

	// First, register a test user
	testUser := TestUser{
		Name:     "Login Test User",
		Email:    fmt.Sprintf("login_%d@example.com", time.Now().Unix()),
		Password: "loginpassword123",
	}

	// Register the user
	resp, _ := MakeRequest(t, "POST", "/api/v1/auth/register", testUser, nil)
	require.Equal(t, 201, resp.StatusCode, "Registration should succeed before login test")

	t.Run("Successful login", func(t *testing.T) {
		loginReq := LoginRequest{
			Email:    testUser.Email,
			Password: testUser.Password,
		}

		resp, body := MakeRequest(t, "POST", "/api/v1/auth/login", loginReq, nil)

		t.Logf("Login response status: %d", resp.StatusCode)
		t.Logf("Login response body: %s", string(body))

		if resp.StatusCode != 200 {
			var errorResp ErrorResponse
			err := json.Unmarshal(body, &errorResp)
			if err == nil {
				t.Logf("Error response: %+v", errorResp)
			}
		}

		assert.Equal(t, 200, resp.StatusCode, "Login should succeed")

		var authResp AuthResponse
		err := json.Unmarshal(body, &authResp)
		require.NoError(t, err)

		assert.NotEmpty(t, authResp.AccessToken)
		assert.NotEmpty(t, authResp.RefreshToken)
		assert.Equal(t, testUser.Email, authResp.User.Email)
	})

	t.Run("Invalid credentials", func(t *testing.T) {
		loginReq := LoginRequest{
			Email:    testUser.Email,
			Password: "wrongpassword",
		}

		resp, body := MakeRequest(t, "POST", "/api/v1/auth/login", loginReq, nil)

		t.Logf("Invalid login response status: %d", resp.StatusCode)
		t.Logf("Invalid login response body: %s", string(body))

		assert.Equal(t, 401, resp.StatusCode, "Invalid login should fail")

		var errorResp ErrorResponse
		err := json.Unmarshal(body, &errorResp)
		require.NoError(t, err)

		assert.False(t, errorResp.Success)
	})

	t.Run("Non-existent user", func(t *testing.T) {
		loginReq := LoginRequest{
			Email:    "nonexistent@example.com",
			Password: "password123",
		}

		resp, body := MakeRequest(t, "POST", "/api/v1/auth/login", loginReq, nil)

		t.Logf("Non-existent user login response status: %d", resp.StatusCode)
		t.Logf("Non-existent user login response body: %s", string(body))

		assert.Equal(t, 401, resp.StatusCode, "Non-existent user login should fail")
	})
}

// TestAuthenticatedEndpoints tests endpoints that require authentication
func TestAuthenticatedEndpoints(t *testing.T) {
	WaitForServer(t)

	// Register and login to get a token
	testUser := TestUser{
		Name:     "Auth Test User",
		Email:    fmt.Sprintf("auth_%d@example.com", time.Now().Unix()),
		Password: "authpassword123",
	}

	// Register
	resp, body := MakeRequest(t, "POST", "/api/v1/auth/register", testUser, nil)
	require.Equal(t, 201, resp.StatusCode, "Registration should succeed")

	var authResp AuthResponse
	err := json.Unmarshal(body, &authResp)
	require.NoError(t, err)

	accessToken := authResp.AccessToken
	require.NotEmpty(t, accessToken)

	t.Run("Get current user with valid token", func(t *testing.T) {
		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}

		resp, body := MakeRequest(t, "GET", "/api/v1/auth/me", nil, headers)

		t.Logf("Get me response status: %d", resp.StatusCode)
		t.Logf("Get me response body: %s", string(body))

		assert.Equal(t, 200, resp.StatusCode, "Get current user should succeed with valid token")

		var userResp UserResponse
		err := json.Unmarshal(body, &userResp)
		require.NoError(t, err)

		assert.Equal(t, testUser.Email, userResp.User.Email)
	})

	t.Run("Get current user without token", func(t *testing.T) {
		resp, body := MakeRequest(t, "GET", "/api/v1/auth/me", nil, nil)

		t.Logf("Get me without token response status: %d", resp.StatusCode)
		t.Logf("Get me without token response body: %s", string(body))

		assert.Equal(t, 401, resp.StatusCode, "Get current user should fail without token")
	})

	t.Run("Get current user with invalid token", func(t *testing.T) {
		headers := map[string]string{
			"Authorization": "Bearer invalid_token",
		}

		resp, body := MakeRequest(t, "GET", "/api/v1/auth/me", nil, headers)

		t.Logf("Get me with invalid token response status: %d", resp.StatusCode)
		t.Logf("Get me with invalid token response body: %s", string(body))

		assert.Equal(t, 401, resp.StatusCode, "Get current user should fail with invalid token")
	})
}

// TestDatabaseConnection tests that the backend can connect to the database
func TestDatabaseConnection(t *testing.T) {
	WaitForServer(t)

	t.Run("Database health check", func(t *testing.T) {
		resp, body := MakeRequest(t, "GET", "/health/db", nil, nil)

		t.Logf("DB health response status: %d", resp.StatusCode)
		t.Logf("DB health response body: %s", string(body))

		assert.Equal(t, 200, resp.StatusCode, "Database should be healthy")

		// Check that it's not using mock database
		assert.NotContains(t, string(body), "mock", "Should not be using mock database in integration test")
	})
}

// TestRedisConnection tests that the backend can connect to Redis
func TestRedisConnection(t *testing.T) {
	WaitForServer(t)

	t.Run("Redis health check", func(t *testing.T) {
		resp, body := MakeRequest(t, "GET", "/health/redis", nil, nil)

		t.Logf("Redis health response status: %d", resp.StatusCode)
		t.Logf("Redis health response body: %s", string(body))

		assert.Equal(t, 200, resp.StatusCode, "Redis should be healthy")
	})
}

// TestCORSHeaders tests that CORS headers are properly set
func TestCORSHeaders(t *testing.T) {
	WaitForServer(t)

	t.Run("OPTIONS request should include CORS headers", func(t *testing.T) {
		client := HTTPClient()
		req, err := http.NewRequest("OPTIONS", baseURL+"/api/v1/auth/register", nil)
		require.NoError(t, err)

		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "Content-Type")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		t.Logf("CORS OPTIONS response status: %d", resp.StatusCode)
		t.Logf("CORS headers: %+v", resp.Header)

		assert.Equal(t, 204, resp.StatusCode, "OPTIONS request should succeed")
		assert.NotEmpty(t, resp.Header.Get("Access-Control-Allow-Origin"))
		assert.NotEmpty(t, resp.Header.Get("Access-Control-Allow-Methods"))
		assert.NotEmpty(t, resp.Header.Get("Access-Control-Allow-Headers"))
	})
}
