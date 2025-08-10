package middleware

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/saurabh22suman/oreo.io/internal/services"
	"golang.org/x/time/rate"
)

// RateLimit implements a simple rate limiting middleware
func RateLimit() gin.HandlerFunc {
	// Get rate limit configuration from environment
	requestsStr := os.Getenv("RATE_LIMIT_REQUESTS")
	windowStr := os.Getenv("RATE_LIMIT_WINDOW")

	requests := 100 // default
	if r, err := strconv.Atoi(requestsStr); err == nil {
		requests = r
	}

	window := time.Minute // default
	if w, err := time.ParseDuration(windowStr); err == nil {
		window = w
	}

	// Create rate limiter
	limiter := rate.NewLimiter(rate.Every(window/time.Duration(requests)), requests)

	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests, please try again later",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireAuth middleware for protecting endpoints
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement JWT authentication when auth service is available
		// For now, just return unauthorized
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Authentication required - coming soon",
		})
		c.Abort()
	}
}

// RequireAuthWithService middleware for protecting endpoints using AuthService
func RequireAuthWithService(authService services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		// Check for Bearer token format
		tokenParts := strings.SplitN(authHeader, " ", 2)
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		token := tokenParts[1]
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token is required",
			})
			c.Abort()
			return
		}

		// Get user from token
		ctx := context.Background()
		user, err := authService.GetUserFromToken(ctx, token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Set user in context
		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Next()
	}
}
