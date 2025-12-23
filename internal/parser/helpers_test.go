package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseDate_AllFormats(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Jakarta")

	tests := []struct {
		name         string
		dateStr      string
		shouldParse  bool
		expectedYear int
		expectedMon  int
		expectedDay  int
		expectedHour int
		expectedMin  int
		expectedSec  int
	}{
		// DateTime formats with time component
		{
			name:         "YYYY-MM-DD HH:MM:SS",
			dateStr:      "2024-01-15 10:30:45",
			shouldParse:  true,
			expectedYear: 2024,
			expectedMon:  1,
			expectedDay:  15,
			expectedHour: 10,
			expectedMin:  30,
			expectedSec:  45,
		},
		{
			name:         "YYYY-MM-DD HH:MM",
			dateStr:      "2024-01-15 10:30",
			shouldParse:  true,
			expectedYear: 2024,
			expectedMon:  1,
			expectedDay:  15,
			expectedHour: 10,
			expectedMin:  30,
			expectedSec:  0,
		},
		{
			name:         "DD/MM/YYYY HH:MM:SS",
			dateStr:      "15/01/2024 10:30:45",
			shouldParse:  true,
			expectedYear: 2024,
			expectedMon:  1,
			expectedDay:  15,
			expectedHour: 10,
			expectedMin:  30,
			expectedSec:  45,
		},
		{
			name:         "DD/MM/YYYY HH:MM",
			dateStr:      "15/01/2024 10:30",
			shouldParse:  true,
			expectedYear: 2024,
			expectedMon:  1,
			expectedDay:  15,
			expectedHour: 10,
			expectedMin:  30,
			expectedSec:  0,
		},
		{
			name:         "DD-MM-YYYY HH:MM:SS",
			dateStr:      "15-01-2024 10:30:45",
			shouldParse:  true,
			expectedYear: 2024,
			expectedMon:  1,
			expectedDay:  15,
			expectedHour: 10,
			expectedMin:  30,
			expectedSec:  45,
		},
		{
			name:         "DD-MM-YYYY HH:MM",
			dateStr:      "15-01-2024 10:30",
			shouldParse:  true,
			expectedYear: 2024,
			expectedMon:  1,
			expectedDay:  15,
			expectedHour: 10,
			expectedMin:  30,
			expectedSec:  0,
		},
		{
			name:         "YYYY-MM-DD",
			dateStr:      "2024-01-15",
			shouldParse:  true,
			expectedYear: 2024,
			expectedMon:  1,
			expectedDay:  15,
			expectedHour: 0,
			expectedMin:  0,
			expectedSec:  0,
		},
		{
			name:         "DD-MM-YYYY",
			dateStr:      "15-01-2024",
			shouldParse:  true,
			expectedYear: 2024,
			expectedMon:  1,
			expectedDay:  15,
			expectedHour: 0,
			expectedMin:  0,
			expectedSec:  0,
		},
		{
			name:         "DD/MM/YYYY",
			dateStr:      "15/01/2024",
			shouldParse:  true,
			expectedYear: 2024,
			expectedMon:  1,
			expectedDay:  15,
			expectedHour: 0,
			expectedMin:  0,
			expectedSec:  0,
		},
		// Invalid formats
		{
			name:        "Invalid format MM/DD/YYYY",
			dateStr:     "01/15/2024",
			shouldParse: false,
		},
		{
			name:        "Invalid format random string",
			dateStr:     "not-a-date",
			shouldParse: false,
		},
		{
			name:        "Empty string",
			dateStr:     "",
			shouldParse: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDate(tt.dateStr, loc)

			if tt.shouldParse {
				if err != nil {
					t.Errorf("Expected date '%s' to parse successfully, got error: %v", tt.dateStr, err)
					return
				}

				if result.Year() != tt.expectedYear {
					t.Errorf("Expected year %d, got %d", tt.expectedYear, result.Year())
				}

				if int(result.Month()) != tt.expectedMon {
					t.Errorf("Expected month %d, got %d", tt.expectedMon, result.Month())
				}

				if result.Day() != tt.expectedDay {
					t.Errorf("Expected day %d, got %d", tt.expectedDay, result.Day())
				}

				if result.Hour() != tt.expectedHour {
					t.Errorf("Expected hour %d, got %d", tt.expectedHour, result.Hour())
				}

				if result.Minute() != tt.expectedMin {
					t.Errorf("Expected minute %d, got %d", tt.expectedMin, result.Minute())
				}

				if result.Second() != tt.expectedSec {
					t.Errorf("Expected second %d, got %d", tt.expectedSec, result.Second())
				}
			} else {
				if err == nil {
					t.Errorf("Expected date '%s' to fail parsing, but it succeeded", tt.dateStr)
				}
			}
		})
	}
}

func TestParseDate_Timezone(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Jakarta")

	dateStr := "2024-01-15 10:30:00"
	result, err := parseDate(dateStr, loc)

	if err != nil {
		t.Fatalf("Expected successful parse, got error: %v", err)
	}

	// Verify timezone is applied
	zoneName, offset := result.Zone()
	if offset != 7*60*60 {
		t.Errorf("Expected UTC+7 timezone (offset %d), got %s with offset %d", 7*60*60, zoneName, offset)
	}
}

func TestExtractFileName(t *testing.T) {
	tests := []struct {
		name         string
		filePath     string
		expectedName string
	}{
		{
			name:         "Simple filename",
			filePath:     "bank_bca.csv",
			expectedName: "bank_bca",
		},
		{
			name:         "With path",
			filePath:     "/path/to/bank_mandiri.csv",
			expectedName: "bank_mandiri",
		},
		{
			name:         "Complex path",
			filePath:     "/Users/test/data/transactions/bank_bni.csv",
			expectedName: "bank_bni",
		},
		{
			name:         "No extension",
			filePath:     "bank_file",
			expectedName: "bank_file",
		},
		{
			name:         "Multiple dots",
			filePath:     "bank.statement.data.csv",
			expectedName: "bank.statement.data",
		},
		{
			name:         "Underscore in name",
			filePath:     "/data/bank_central_asia.csv",
			expectedName: "bank_central_asia",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractFileName(tt.filePath)

			if result != tt.expectedName {
				t.Errorf("Expected '%s', got '%s'", tt.expectedName, result)
			}
		})
	}
}

func TestValidateCSVExtension_Valid(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
	}{
		{"Lowercase .csv", "file.csv"},
		{"Uppercase .CSV", "file.CSV"},
		{"Mixed case .CsV", "file.CsV"},
		{"With path", "/path/to/file.csv"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCSVExtension(tt.filePath)

			if err != nil {
				t.Errorf("Expected '%s' to be valid CSV, got error: %v", tt.filePath, err)
			}
		})
	}
}

func TestValidateCSVExtension_Invalid(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
	}{
		{"Text file", "file.txt"},
		{"Excel file", "file.xlsx"},
		{"No extension", "file"},
		{"JSON file", "data.json"},
		{"XML file", "data.xml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCSVExtension(tt.filePath)

			if err == nil {
				t.Errorf("Expected '%s' to fail validation, but it passed", tt.filePath)
			}

			// Verify error message mentions CSV
			if !strings.Contains(strings.ToLower(err.Error()), "csv") {
				t.Errorf("Expected error message to mention CSV, got: %v", err)
			}
		})
	}
}

func TestReadCSVFile_Success(t *testing.T) {
	tests := []struct {
		name         string
		csvContent   string
		fileName     string
		expectedRows int
		verify       func(t *testing.T, records [][]string)
	}{
		{
			name: "basic CSV with header and data rows",
			csvContent: `header1,header2,header3
value1,value2,value3
value4,value5,value6`,
			fileName:     "test.csv",
			expectedRows: 3,
			verify: func(t *testing.T, records [][]string) {
				// Verify header
				if len(records[0]) != 3 || records[0][0] != "header1" {
					t.Errorf("Expected first row to be header, got %v", records[0])
				}
				// Verify data rows
				if len(records[1]) != 3 || records[1][0] != "value1" {
					t.Errorf("Expected second row to have correct values, got %v", records[1])
				}
			},
		},
		{
			name: "large file - 10000 rows",
			csvContent: func() string {
				var rows []string
				rows = append(rows, "col1,col2,col3")
				for i := 0; i < 10000; i++ {
					rows = append(rows, "val1,val2,val3")
				}
				return strings.Join(rows, "\n")
			}(),
			fileName:     "large.csv",
			expectedRows: 10001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			csvPath := filepath.Join(tmpDir, tt.fileName)

			if err := os.WriteFile(csvPath, []byte(tt.csvContent), 0644); err != nil {
				t.Fatalf("Failed to create test CSV: %v", err)
			}

			records, err := readCSVFile(csvPath)

			if err != nil {
				t.Fatalf("Expected successful read, got error: %v", err)
			}

			if len(records) != tt.expectedRows {
				t.Errorf("Expected %d records, got %d", tt.expectedRows, len(records))
			}

			if tt.verify != nil {
				tt.verify(t, records)
			}
		})
	}
}

func TestReadCSVFile_ErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		setupFile func(tmpDir string) string
	}{
		{
			name: "non-CSV extension",
			setupFile: func(tmpDir string) string {
				txtPath := filepath.Join(tmpDir, "test.txt")
				os.WriteFile(txtPath, []byte("some content"), 0644)
				return txtPath
			},
		},
		{
			name: "file not found",
			setupFile: func(tmpDir string) string {
				return "/nonexistent/path/file.csv"
			},
		},
		{
			name: "empty file",
			setupFile: func(tmpDir string) string {
				csvPath := filepath.Join(tmpDir, "empty.csv")
				os.WriteFile(csvPath, []byte(""), 0644)
				return csvPath
			},
		},
		{
			name: "only header - no data rows",
			setupFile: func(tmpDir string) string {
				csvPath := filepath.Join(tmpDir, "header_only.csv")
				os.WriteFile(csvPath, []byte("header1,header2,header3"), 0644)
				return csvPath
			},
		},
		{
			name: "malformed CSV - unclosed quote",
			setupFile: func(tmpDir string) string {
				csvPath := filepath.Join(tmpDir, "malformed.csv")
				content := `header1,header2,header3
value1,"unclosed quote,value3`
				os.WriteFile(csvPath, []byte(content), 0644)
				return csvPath
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			csvPath := tt.setupFile(tmpDir)

			_, err := readCSVFile(csvPath)

			if err == nil {
				t.Error("Expected error but got nil")
			}
		})
	}
}
