package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/saurabh22suman/oreo.io/internal/models"
	"github.com/saurabh22suman/oreo.io/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService implements the AuthService interface for testing
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, req *models.CreateUserRequest) (*services.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.AuthResponse), args.Error(1)
}

func (m *MockAuthService) Login(ctx context.Context, req *models.LoginRequest) (*services.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.AuthResponse), args.Error(1)
}

func (m *MockAuthService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	args := m.Called(ctx, refreshToken)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) GetUserFromToken(ctx context.Context, token string) (*models.User, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func TestAuthHandlers_RegisterWithService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful registration", func(t *testing.T) {
		mockService := new(MockAuthService)
		handlers := NewAuthHandlers(mockService)

		testUser := models.PublicUser{
			ID:    uuid.New(),
			Email: "test@example.com",
			Name:  "Test User",
		}

		expectedResp := &services.AuthResponse{
			User: testUser,
			Tokens: services.TokenPair{
				AccessToken:  "access-token",
				RefreshToken: "refresh-token",
			},
		}

		mockService.On("Register", mock.Anything, mock.AnythingOfType("*models.CreateUserRequest")).Return(expectedResp, nil)

		router := gin.New()
		router.POST("/register", handlers.RegisterWithService())

		reqBody := models.CreateUserRequest{
			Email:    "test@example.com",
			Name:     "Test User",
			Password: "password123",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response AuthResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, testUser, response.User)
		assert.Equal(t, "access-token", response.AccessToken)
		assert.Equal(t, "refresh-token", response.RefreshToken)

		mockService.AssertExpectations(t)
	})

	t.Run("email already exists", func(t *testing.T) {
		mockService := new(MockAuthService)
		handlers := NewAuthHandlers(mockService)

		mockService.On("Register", mock.Anything, mock.AnythingOfType("*models.CreateUserRequest")).Return(nil, errors.New("email already exists"))

		router := gin.New()
		router.POST("/register", handlers.RegisterWithService())

		reqBody := models.CreateUserRequest{
			Email:    "existing@example.com",
			Name:     "Test User",
			Password: "password123",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Email already exists", response["error"])

		mockService.AssertExpectations(t)
	})
}

func TestAuthHandlers_LoginWithService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful login", func(t *testing.T) {
		mockService := new(MockAuthService)
		handlers := NewAuthHandlers(mockService)

		testUser := models.PublicUser{
			ID:    uuid.New(),
			Email: "test@example.com",
			Name:  "Test User",
		}

		expectedResp := &services.AuthResponse{
			User: testUser,
			Tokens: services.TokenPair{
				AccessToken:  "access-token",
				RefreshToken: "refresh-token",
			},
		}

		mockService.On("Login", mock.Anything, mock.AnythingOfType("*models.LoginRequest")).Return(expectedResp, nil)

		router := gin.New()
		router.POST("/login", handlers.LoginWithService())

		reqBody := LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response AuthResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, testUser, response.User)
		assert.Equal(t, "access-token", response.AccessToken)
		assert.Equal(t, "refresh-token", response.RefreshToken)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		mockService := new(MockAuthService)
		handlers := NewAuthHandlers(mockService)

		mockService.On("Login", mock.Anything, mock.AnythingOfType("*models.LoginRequest")).Return(nil, errors.New("invalid credentials"))

		router := gin.New()
		router.POST("/login", handlers.LoginWithService())

		reqBody := LoginRequest{
			Email:    "test@example.com",
			Password: "wrong-password",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid email or password", response["error"])

		mockService.AssertExpectations(t)
	})
}

func TestAuthHandlers_RefreshTokenWithService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful refresh", func(t *testing.T) {
		mockService := new(MockAuthService)
		handlers := NewAuthHandlers(mockService)

		mockService.On("RefreshToken", mock.Anything, "valid-refresh-token").Return("new-access-token", nil)

		router := gin.New()
		router.POST("/refresh", handlers.RefreshTokenWithService())

		reqBody := RefreshTokenRequest{
			RefreshToken: "valid-refresh-token",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/refresh", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "new-access-token", response["access_token"])

		mockService.AssertExpectations(t)
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		mockService := new(MockAuthService)
		handlers := NewAuthHandlers(mockService)

		mockService.On("RefreshToken", mock.Anything, "invalid-refresh-token").Return("", errors.New("invalid refresh token"))

		router := gin.New()
		router.POST("/refresh", handlers.RefreshTokenWithService())

		reqBody := RefreshTokenRequest{
			RefreshToken: "invalid-refresh-token",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/refresh", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid refresh token", response["error"])

		mockService.AssertExpectations(t)
	})
}
