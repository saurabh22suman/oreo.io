package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

// NewConnection creates a new PostgreSQL database connection
func NewConnection() (*sql.DB, error) {
	// First try DATABASE_URL if available
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		db, err := sql.Open("postgres", databaseURL)
		if err != nil {
			return nil, fmt.Errorf("failed to open database connection with DATABASE_URL: %w", err)
		}

		// Configure connection pool
		maxConnections, _ := strconv.Atoi(os.Getenv("DB_MAX_CONNECTIONS"))
		if maxConnections == 0 {
			maxConnections = 25
		}

		maxIdleConnections, _ := strconv.Atoi(os.Getenv("DB_MAX_IDLE_CONNECTIONS"))
		if maxIdleConnections == 0 {
			maxIdleConnections = 5
		}

		db.SetMaxOpenConns(maxConnections)
		db.SetMaxIdleConns(maxIdleConnections)
		db.SetConnMaxLifetime(time.Hour)

		// Test the connection
		if err := db.Ping(); err != nil {
			return nil, fmt.Errorf("failed to ping database: %w", err)
		}

		return db, nil
	}

	// Fallback to individual environment variables
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSL_MODE")

	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "5432"
	}
	if sslmode == "" {
		sslmode = "disable"
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	maxConnections, _ := strconv.Atoi(os.Getenv("DB_MAX_CONNECTIONS"))
	if maxConnections == 0 {
		maxConnections = 25
	}

	maxIdleConnections, _ := strconv.Atoi(os.Getenv("DB_MAX_IDLE_CONNECTIONS"))
	if maxIdleConnections == 0 {
		maxIdleConnections = 5
	}

	db.SetMaxOpenConns(maxConnections)
	db.SetMaxIdleConns(maxIdleConnections)
	db.SetConnMaxLifetime(time.Hour)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// NewTestConnection creates a new test database connection
func NewTestConnection() (*sql.DB, error) {
	host := os.Getenv("TEST_DB_HOST")
	port := os.Getenv("TEST_DB_PORT")
	user := os.Getenv("TEST_DB_USER")
	password := os.Getenv("TEST_DB_PASSWORD")
	dbname := os.Getenv("TEST_DB_NAME")
	sslmode := os.Getenv("TEST_DB_SSL_MODE")

	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "5433"
	}
	if sslmode == "" {
		sslmode = "disable"
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open test database connection: %w", err)
	}

	// Configure connection pool for testing
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(time.Minute * 30)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping test database: %w", err)
	}

	return db, nil
}

// NewRedisConnection creates a new Redis connection
func NewRedisConnection() (*redis.Client, error) {
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	password := os.Getenv("REDIS_PASSWORD")
	db := os.Getenv("REDIS_DB")

	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "6379"
	}

	dbNum, err := strconv.Atoi(db)
	if err != nil {
		dbNum = 0
	}

	poolSize, _ := strconv.Atoi(os.Getenv("REDIS_POOL_SIZE"))
	if poolSize == 0 {
		poolSize = 10
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       dbNum,
		PoolSize: poolSize,
	})

	// Test the connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	return rdb, nil
}
