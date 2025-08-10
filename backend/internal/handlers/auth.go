package handlers

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/saurabh22suman/oreo.io/internal/models"
	"github.com/saurabh22suman/oreo.io/internal/services"
)

type AuthHandlers struct {
	authService services.AuthService
}

func NewAuthHandlers(authService services.AuthService) *AuthHandlers {
	return &AuthHandlers{
		authService: authService,
	}
}

// LoginRequest for standalone handlers (duplicates models.LoginRequest for flexibility)
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type AuthResponse struct {
	User         models.PublicUser `json:"user"`
	AccessToken  string            `json:"access_token"`
	RefreshToken string            `json:"refresh_token"`
}

// Register creates a new user account
func Register() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.CreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request body",
				"details": err.Error(),
			})
			return
		}

		// Validate required fields
		if req.Email == "" || req.Name == "" || req.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Email, name, and password are required",
			})
			return
		}

		// TODO: Use actual auth service when available
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Authentication service not available",
		})
	}
}

// Login authenticates a user
func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request body",
				"details": err.Error(),
			})
			return
		}

		// TODO: Use actual auth service when available
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Authentication service not available",
		})
	}
}

// RefreshToken generates a new access token
func RefreshToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RefreshTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request body",
				"details": err.Error(),
			})
			return
		}

		// TODO: Use actual auth service when available
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Authentication service not available",
		})
	}
}

// Logout handles user logout
func Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement user logout
		c.JSON(http.StatusNotImplemented, gin.H{
			"message": "User logout endpoint - coming soon",
			"status":  "not_implemented",
		})
	}
}

// GetCurrentUser returns the current authenticated user
func GetCurrentUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not authenticated",
			})
			return
		}

		userModel, ok := user.(*models.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid user context",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user": userModel,
		})
	}
}

// RegisterWithService creates a new user account using the auth service
func (h *AuthHandlers) RegisterWithService() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("RegisterWithService called - START")
		var req models.CreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("RegisterWithService: JSON binding error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request body",
				"details": err.Error(),
			})
			return
		}

		log.Printf("RegisterWithService: Request data: %+v", req)
		ctx := context.Background()
		authResp, err := h.authService.Register(ctx, &req)
		if err != nil {
			log.Printf("RegisterWithService: Auth service error: %v", err)
			// Check for user already exists error
			if strings.Contains(err.Error(), "already exists") {
				c.JSON(http.StatusConflict, gin.H{
					"error": "An account with this email address already exists",
				})
				return
			}

			// Check for validation errors
			if strings.Contains(err.Error(), "validation failed") {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Invalid user data provided",
					"details": err.Error(),
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to register user. Please try again later.",
			})
			return
		}

		log.Printf("RegisterWithService: Success - User created: %+v", authResp.User)
		c.JSON(http.StatusCreated, AuthResponse{
			User:         authResp.User,
			AccessToken:  authResp.Tokens.AccessToken,
			RefreshToken: authResp.Tokens.RefreshToken,
		})
	}
}

// LoginWithService authenticates a user using the auth service
func (h *AuthHandlers) LoginWithService() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request body",
				"details": err.Error(),
			})
			return
		}

		loginReq := &models.LoginRequest{
			Email:    req.Email,
			Password: req.Password,
		}

		ctx := context.Background()
		authResp, err := h.authService.Login(ctx, loginReq)
		if err != nil {
			// Check for authentication errors (invalid credentials or user not found)
			if strings.Contains(err.Error(), "invalid email or password") {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Invalid email or password. Please check your credentials and try again.",
				})
				return
			}

			// Check for other authentication-related errors
			if strings.Contains(err.Error(), "failed to get user") {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Invalid email or password. Please check your credentials and try again.",
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Authentication failed. Please try again later.",
			})
			return
		}

		c.JSON(http.StatusOK, AuthResponse{
			User:         authResp.User,
			AccessToken:  authResp.Tokens.AccessToken,
			RefreshToken: authResp.Tokens.RefreshToken,
		})
	}
}

// RefreshTokenWithService generates a new access token using the auth service
func (h *AuthHandlers) RefreshTokenWithService() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RefreshTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request body",
				"details": err.Error(),
			})
			return
		}

		ctx := context.Background()
		newAccessToken, err := h.authService.RefreshToken(ctx, req.RefreshToken)
		if err != nil {
			if err.Error() == "invalid refresh token" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Invalid refresh token",
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to refresh token",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"access_token": newAccessToken,
		})
	}
}
