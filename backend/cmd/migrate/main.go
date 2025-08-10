package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	var direction string
	flag.StringVar(&direction, "direction", "up", "Migration direction: up or down")
	flag.Parse()

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Build database URL
	host := getEnvOrDefault("DB_HOST", "localhost")
	port := getEnvOrDefault("DB_PORT", "5432")
	user := getEnvOrDefault("DB_USER", "oreo_user")
	password := getEnvOrDefault("DB_PASSWORD", "oreo_password")
	dbname := getEnvOrDefault("DB_NAME", "oreo_db")
	sslmode := getEnvOrDefault("DB_SSL_MODE", "disable")

	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, dbname, sslmode)

	// Initialize migration
	m, err := migrate.New(
		"file://./database/migrations",
		databaseURL,
	)
	if err != nil {
		log.Fatalf("Failed to initialize migrate: %v", err)
	}

	defer func() {
		sourceErr, dbErr := m.Close()
		if sourceErr != nil {
			log.Printf("Source close error: %v", sourceErr)
		}
		if dbErr != nil {
			log.Printf("Database close error: %v", dbErr)
		}
	}()

	// Run migration
	switch direction {
	case "up":
		log.Println("Running migrations up...")
		if err := m.Up(); err != nil {
			if err == migrate.ErrNoChange {
				log.Println("No new migrations to apply")
			} else {
				log.Fatalf("Failed to run migrations up: %v", err)
			}
		} else {
			log.Println("Migrations applied successfully")
		}
	case "down":
		log.Println("Running migrations down...")
		if err := m.Down(); err != nil {
			if err == migrate.ErrNoChange {
				log.Println("No migrations to rollback")
			} else {
				log.Fatalf("Failed to run migrations down: %v", err)
			}
		} else {
			log.Println("Migrations rolled back successfully")
		}
	default:
		log.Fatalf("Invalid direction: %s. Use 'up' or 'down'", direction)
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
