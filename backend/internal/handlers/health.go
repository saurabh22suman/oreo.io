package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	Services  Services  `json:"services"`
}

// Services represents the status of external services
type Services struct {
	Database DatabaseStatus `json:"database"`
	Redis    RedisStatus    `json:"redis"`
}

// DatabaseStatus represents database health status
type DatabaseStatus struct {
	Status   string `json:"status"`
	Response string `json:"response_time,omitempty"`
}

// RedisStatus represents Redis health status
type RedisStatus struct {
	Status   string `json:"status"`
	Response string `json:"response_time,omitempty"`
}

// HealthCheck returns the overall health status
func HealthCheck(db *sql.DB, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Check database
		dbStatus := checkDatabase(db)
		
		// Check Redis
		redisStatus := checkRedis(rdb)

		// Determine overall status
		status := "healthy"
		if dbStatus.Status != "healthy" || redisStatus.Status != "healthy" {
			status = "unhealthy"
		}

		response := HealthResponse{
			Status:    status,
			Timestamp: time.Now(),
			Version:   "1.0.0", // TODO: Get from build info
			Services: Services{
				Database: dbStatus,
				Redis:    redisStatus,
			},
		}

		// Set appropriate HTTP status code
		statusCode := http.StatusOK
		if status == "unhealthy" {
			statusCode = http.StatusServiceUnavailable
		}

		// Add response time header
		c.Header("X-Response-Time", time.Since(start).String())

		c.JSON(statusCode, response)
	}
}

// DatabaseHealthCheck returns database-specific health status
func DatabaseHealthCheck(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		status := checkDatabase(db)
		
		statusCode := http.StatusOK
		if status.Status != "healthy" {
			statusCode = http.StatusServiceUnavailable
		}

		c.Header("X-Response-Time", time.Since(start).String())
		c.JSON(statusCode, status)
	}
}

// RedisHealthCheck returns Redis-specific health status
func RedisHealthCheck(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		status := checkRedis(rdb)
		
		statusCode := http.StatusOK
		if status.Status != "healthy" {
			statusCode = http.StatusServiceUnavailable
		}

		c.Header("X-Response-Time", time.Since(start).String())
		c.JSON(statusCode, status)
	}
}

// checkDatabase performs database health check
func checkDatabase(db *sql.DB) DatabaseStatus {
	start := time.Now()
	
	if db == nil {
		return DatabaseStatus{
			Status: "unhealthy",
		}
	}
	
	if err := db.Ping(); err != nil {
		return DatabaseStatus{
			Status: "unhealthy",
		}
	}

	// Test with a simple query
	var result int
	if err := db.QueryRow("SELECT 1").Scan(&result); err != nil {
		return DatabaseStatus{
			Status: "unhealthy",
		}
	}

	return DatabaseStatus{
		Status:   "healthy",
		Response: time.Since(start).String(),
	}
}

// checkRedis performs Redis health check
func checkRedis(rdb *redis.Client) RedisStatus {
	start := time.Now()
	
	if rdb == nil {
		return RedisStatus{
			Status: "unhealthy",
		}
	}
	
	ctx := context.Background()
	
	if err := rdb.Ping(ctx).Err(); err != nil {
		return RedisStatus{
			Status: "unhealthy",
		}
	}

	return RedisStatus{
		Status:   "healthy",
		Response: time.Since(start).String(),
	}
}
