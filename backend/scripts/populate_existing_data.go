package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

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

	// Find datasets that have files but no data
	datasets, err := getDatasetsWithoutData(db)
	if err != nil {
		log.Fatal("Failed to get datasets:", err)
	}

	log.Printf("Found %d datasets without data", len(datasets))

	for _, dataset := range datasets {
		log.Printf("Processing dataset: %s (%s)", dataset.Name, dataset.ID)
		
		if err := processDatasetFile(db, dataset); err != nil {
			log.Printf("Error processing dataset %s: %v", dataset.ID, err)
		} else {
			log.Printf("Successfully processed dataset %s", dataset.ID)
		}
	}
}

type Dataset struct {
	ID       string `db:"id"`
	Name     string `db:"name"`
	FilePath string `db:"file_path"`
}

func getDatasetsWithoutData(db *sqlx.DB) ([]Dataset, error) {
	query := `
		SELECT d.id, d.name, d.file_path 
		FROM datasets d 
		WHERE d.file_path IS NOT NULL 
		AND NOT EXISTS (
			SELECT 1 FROM dataset_data dd WHERE dd.dataset_id = d.id
		)`
	
	var datasets []Dataset
	err := db.Select(&datasets, query)
	return datasets, err
}

func processDatasetFile(db *sqlx.DB, dataset Dataset) error {
	// Check if file exists
	if _, err := os.Stat(dataset.FilePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", dataset.FilePath)
	}

	// Only process CSV files for now
	if !strings.HasSuffix(strings.ToLower(dataset.FilePath), ".csv") {
		log.Printf("Skipping non-CSV file: %s", dataset.FilePath)
		return nil
	}

	// Read CSV file
	file, err := os.Open(dataset.FilePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		return fmt.Errorf("CSV file is empty")
	}

	headers := records[0]
	dataRows := records[1:]

	log.Printf("CSV has %d headers and %d data rows", len(headers), len(dataRows))

	// Insert data into database
	return bulkInsertData(db, dataset.ID, headers, dataRows)
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

	// Use a default user ID (you might want to get the actual user who uploaded)
	defaultUserID := "00000000-0000-0000-0000-000000000000"

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

	log.Printf("Inserted %d total rows", len(rows))
	return tx.Commit()
}
