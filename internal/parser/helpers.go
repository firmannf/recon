package parser

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// parseDate tries to parse date/datetime in multiple formats
func parseDate(dateStr string) (time.Time, error) {
	formats := []string{
		// DateTime formats with time component
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"02/01/2006 15:04:05",
		"02/01/2006 15:04",
		"02-01-2006 15:04:05",
		"02-01-2006 15:04",
		// Date-only formats
		"2006-01-02",
		"02-01-2006",
		"02/01/2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// extractFileName extracts a file name without extension from the file path
func extractFileName(filePath string) string {
	fileName := filepath.Base(filePath)
	name := fileName[:len(fileName)-len(filepath.Ext(fileName))]
	return name
}

// validateCSVExtension checks if the file has a .csv extension
func validateCSVExtension(filePath string) error {
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".csv" {
		return fmt.Errorf("file must be a CSV file (got %s): %s", ext, filePath)
	}
	return nil
}

// readCSVFile reads and validates a CSV file, returning all records
func readCSVFile(filePath string) ([][]string, error) {
	// Validate extension
	if err := validateCSVExtension(filePath); err != nil {
		return nil, err
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read CSV
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	// Validate not empty
	if len(records) <= headerRowCount {
		return nil, fmt.Errorf("CSV file is empty or has no data rows")
	}

	return records, nil
}
