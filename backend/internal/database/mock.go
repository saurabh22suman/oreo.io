package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

// MockDB represents a mock database connection for development
type MockDB struct {
	connected bool
}

func (m *MockDB) Close() error {
	m.connected = false
	log.Println("Mock database connection closed")
	return nil
}

func (m *MockDB) Ping() error {
	if !m.connected {
		return fmt.Errorf("mock database not connected")
	}
	return nil
}

// MockRedis represents a mock Redis connection for development
type MockRedis struct {
	connected bool
}

func (m *MockRedis) Close() error {
	m.connected = false
	log.Println("Mock Redis connection closed")
	return nil
}

func (m *MockRedis) Ping(ctx context.Context) *redis.StatusCmd {
	// Create a mock status command that always returns success
	cmd := redis.NewStatusCmd(ctx)
	if m.connected {
		cmd.SetVal("PONG")
	} else {
		cmd.SetErr(fmt.Errorf("mock redis not connected"))
	}
	return cmd
}

// NewConnectionWithFallback creates a database connection with fallback to mock for development
func NewConnectionWithFallback() (interface{}, error) {
	if os.Getenv("ENVIRONMENT") == "development" && os.Getenv("USE_MOCK_DB") == "true" {
		log.Println("Using mock database for development")
		return &MockDB{connected: true}, nil
	}
	
	// Try actual database connection
	db, err := NewConnection()
	if err != nil {
		log.Printf("Failed to connect to real database: %v", err)
		log.Println("Falling back to mock database for development")
		return &MockDB{connected: true}, nil
	}
	
	return db, nil
}

// NewRedisConnectionWithFallback creates a Redis connection with fallback to mock for development
func NewRedisConnectionWithFallback() (interface{}, error) {
	if os.Getenv("ENVIRONMENT") == "development" && os.Getenv("USE_MOCK_REDIS") == "true" {
		log.Println("Using mock Redis for development")
		return &MockRedis{connected: true}, nil
	}
	
	// Try actual Redis connection
	redis, err := NewRedisConnection()
	if err != nil {
		log.Printf("Failed to connect to real Redis: %v", err)
		log.Println("Falling back to mock Redis for development")
		return &MockRedis{connected: true}, nil
	}
	
	return redis, nil
}
