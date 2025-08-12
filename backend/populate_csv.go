package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	// Database connection
	db, err := sqlx.Connect("postgres", "postgres://oreo_user:oreo_password@localhost:5432/oreo_db?sslmode=disable")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Specific dataset to populate
	datasetID := "b06bef23-4041-44f9-bb94-a2eeb27e00d2"
	filePath := "../uploads/b06bef23-4041-44f9-bb94-a2eeb27e00d2_airlines_flights_data.csv"

	log.Printf("Processing dataset: %s", datasetID)
	log.Printf("File path: %s", filePath)
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Fatalf("File does not exist: %s", filePath)
	}

	// Read CSV file
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read CSV: %v", err)
	}

	if len(records) == 0 {
		log.Fatal("CSV file is empty")
	}

	headers := records[0]
	dataRows := records[1:]

	log.Printf("CSV has %d headers and %d data rows", len(headers), len(dataRows))

	// Insert data into database
	if err := bulkInsertData(db, datasetID, headers, dataRows); err != nil {
		log.Fatalf("Failed to insert data: %v", err)
	}

	log.Printf("Successfully populated dataset %s with %d rows", datasetID, len(dataRows))
}

func bulkInsertData(db *sqlx.DB, datasetID string, headers []string, rows [][]string) error {
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO dataset_data (dataset_id, row_index, data, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $4)`

	// Use a real user ID from the database
	defaultUserID := "727c0441-d99e-4aa5-ab01-b7bc257a31d6"

	for i, row := range rows {
		// Create a map from headers to row values
		data := make(map[string]interface{})
		for j, header := range headers {
			if j < len(row) {
				data[header] = row[j]
			} else {
				data[header] = "" // Handle missing values
			}
		}

		// Marshal to JSON
		dataJSON, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal data for row %d: %w", i, err)
		}

		// Insert the row
		_, err = tx.Exec(query, datasetID, i, dataJSON, defaultUserID)
		if err != nil {
			return fmt.Errorf("failed to insert data for row %d: %w", i, err)
		}

		// Log progress for large datasets
		if (i+1)%1000 == 0 {
			log.Printf("Inserted %d rows", i+1)
		}
	}

	return tx.Commit()
}
