package middleware

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
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
		// TODO: Implement JWT authentication
		// For now, just return unauthorized
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Authentication required - coming soon",
		})
		c.Abort()
	}
}
