package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AuthRequest represents authentication requests
type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name,omitempty"`
}

// AuthResponse represents authentication responses
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
	Message      string `json:"message"`
}

// User represents a user
type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	GoogleID  string `json:"google_id,omitempty"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// RefreshRequest represents token refresh request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func TestAuthenticationFlow(t *testing.T) {
	// Create a unique user for this test
	user := getTestUser()
	var refreshToken string

	t.Run("Register New User", func(t *testing.T) {
		registerReq := AuthRequest{
			Email:    user.Email,
			Password: user.Password,
			Name:     user.Name,
		}

		resp, bodyBytes := makeRequest(t, "POST", "/api/v1/auth/register", registerReq)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var authResp AuthResponse
		err := json.Unmarshal(bodyBytes, &authResp)
		require.NoError(t, err)

		assert.NotEmpty(t, authResp.AccessToken)
		assert.NotEmpty(t, authResp.RefreshToken)
		assert.Equal(t, user.Email, authResp.User.Email)
		assert.Equal(t, user.Name, authResp.User.Name)
		assert.NotEmpty(t, authResp.User.ID)

		refreshToken = authResp.RefreshToken
	})

	t.Run("Login Existing User", func(t *testing.T) {
		loginReq := AuthRequest{
			Email:    user.Email,
			Password: user.Password,
		}

		resp, bodyBytes := makeRequest(t, "POST", "/api/v1/auth/login", loginReq)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var authResp AuthResponse
		err := json.Unmarshal(bodyBytes, &authResp)
		require.NoError(t, err)

		assert.NotEmpty(t, authResp.AccessToken)
		assert.NotEmpty(t, authResp.RefreshToken)
		assert.Equal(t, user.Email, authResp.User.Email)
		assert.Equal(t, user.Name, authResp.User.Name)
	})

	t.Run("Refresh Access Token", func(t *testing.T) {
		if refreshToken == "" {
			t.Skip("No refresh token available")
			return
		}

		refreshReq := RefreshRequest{
			RefreshToken: refreshToken,
		}

		resp, bodyBytes := makeRequest(t, "POST", "/api/v1/auth/refresh", refreshReq)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var authResp AuthResponse
		err := json.Unmarshal(bodyBytes, &authResp)
		require.NoError(t, err)

		assert.NotEmpty(t, authResp.AccessToken)
		// Note: refresh_token might be empty in response, which is okay
	})

	t.Run("Register with Duplicate Email - should return 409", func(t *testing.T) {
		duplicateReq := AuthRequest{
			Email:    user.Email, // Same email as before
			Password: "differentpassword",
			Name:     "Different Name",
		}

		resp, _ := makeRequest(t, "POST", "/api/v1/auth/register", duplicateReq)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})

	t.Run("Login with Wrong Password - should return 401", func(t *testing.T) {
		wrongPasswordReq := AuthRequest{
			Email:    user.Email,
			Password: "wrongpassword",
		}

		resp, _ := makeRequest(t, "POST", "/api/v1/auth/login", wrongPasswordReq)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Login with Non-existent Email - should return 401", func(t *testing.T) {
		nonExistentReq := AuthRequest{
			Email:    "nonexistent@example.com",
			Password: "anypassword",
		}

		resp, _ := makeRequest(t, "POST", "/api/v1/auth/login", nonExistentReq)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Access Protected Endpoint with Invalid Token - should return 401", func(t *testing.T) {
		resp, _ := makeAuthenticatedRequest(t, "GET", "/api/v1/projects", nil, "invalid_token")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
