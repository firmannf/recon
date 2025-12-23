package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/firmannf/recon/internal/models"
)

func TestBankStatementLineParser_ParseCSV_Success(t *testing.T) {
	tests := []struct {
		name             string
		csvContent       string
		fileName         string
		expectedCount    int
		expectedBankName string
		verify           func(t *testing.T, statements []models.BankStatementLine)
	}{
		{
			name: "multiple statements",
			csvContent: `unique_identifier,amount,date
BANK-001,1000.50,2024-01-15
BANK-002,-250,2024-01-16`,
			fileName:         "bank_bca.csv",
			expectedCount:    2,
			expectedBankName: "bank_bca",
			verify: func(t *testing.T, statements []models.BankStatementLine) {
				if statements[0].UniqueIdentifier != "BANK-001" {
					t.Errorf("Expected ID 'BANK-001', got '%s'", statements[0].UniqueIdentifier)
				}
				if !statements[0].Amount.Equal(decimal.NewFromFloat(1000.50)) {
					t.Errorf("Expected amount 1000.50, got %s", statements[0].Amount)
				}
				if !statements[1].Amount.Equal(decimal.NewFromFloat(-250)) {
					t.Errorf("Expected amount 250, got %s", statements[1].Amount)
				}
				if statements[0].Type != models.TransactionTypeCredit {
					t.Errorf("Expected CREDIT, got %s", statements[0].Type)
				}
				if statements[1].Type != models.TransactionTypeDebit {
					t.Errorf("Expected DEBIT, got %s", statements[1].Type)
				}
			},
		},
		{
			name: "single statement",
			csvContent: `unique_identifier,amount,date
BANK-001,1000.00,2024-01-15`,
			fileName:         "bank_mandiri.csv",
			expectedCount:    1,
			expectedBankName: "bank_mandiri",
		},
		{
			name: "format date YYYY-MM-DD",
			csvContent: `unique_identifier,amount,date
BANK-001,1000.00,2024-01-15`,
			fileName:         "bank.csv",
			expectedCount:    1,
			expectedBankName: "bank",
			verify: func(t *testing.T, statements []models.BankStatementLine) {
				if statements[0].Date.Year() != 2024 || statements[0].Date.Day() != 15 {
					t.Errorf("Expected 2024-01-15, got %v", statements[0].Date)
				}
			},
		},
		{
			name: "format date DD/MM/YYYY",
			csvContent: `unique_identifier,amount,date
BANK-001,1000.00,15/01/2024`,
			fileName:         "bank.csv",
			expectedCount:    1,
			expectedBankName: "bank",
			verify: func(t *testing.T, statements []models.BankStatementLine) {
				if statements[0].Date.Year() != 2024 || statements[0].Date.Day() != 15 {
					t.Errorf("Expected 2024-01-15, got %v", statements[0].Date)
				}
			},
		},
		{
			name: "format date DD-MM-YYYY",
			csvContent: `unique_identifier,amount,date
BANK-001,1000.00,15-01-2024`,
			fileName:         "bank.csv",
			expectedCount:    1,
			expectedBankName: "bank",
			verify: func(t *testing.T, statements []models.BankStatementLine) {
				if statements[0].Date.Year() != 2024 || statements[0].Date.Day() != 15 {
					t.Errorf("Expected 2024-01-15, got %v", statements[0].Date)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			csvPath := filepath.Join(tmpDir, tt.fileName)

			if err := os.WriteFile(csvPath, []byte(tt.csvContent), 0644); err != nil {
				t.Fatalf("Failed to create test CSV: %v", err)
			}

			parser := NewBankStatementLineParser()
			statements, err := parser.ParseCSV(csvPath)

			if err != nil {
				t.Fatalf("Expected successful parse, got error: %v", err)
			}

			if len(statements) != tt.expectedCount {
				t.Fatalf("Expected %d statements, got %d", tt.expectedCount, len(statements))
			}

			if statements[0].BankName != tt.expectedBankName {
				t.Errorf("Expected bank name '%s', got '%s'", tt.expectedBankName, statements[0].BankName)
			}

			if tt.verify != nil {
				tt.verify(t, statements)
			}
		})
	}
}

func TestBankStatementLineParser_ParseCSV_ErrorCases(t *testing.T) {
	tests := []struct {
		name       string
		setupFile  func(tmpDir string) string
		shouldFail bool
	}{
		{
			name: "file not found",
			setupFile: func(tmpDir string) string {
				return "/nonexistent/path/bank.csv"
			},
			shouldFail: true,
		},
		{
			name: "non-CSV extension",
			setupFile: func(tmpDir string) string {
				txtPath := filepath.Join(tmpDir, "bank.txt")
				os.WriteFile(txtPath, []byte("some content"), 0644)
				return txtPath
			},
			shouldFail: true,
		},
		{
			name: "empty file - only header",
			setupFile: func(tmpDir string) string {
				csvPath := filepath.Join(tmpDir, "bank.csv")
				os.WriteFile(csvPath, []byte("unique_identifier,amount,date\n"), 0644)
				return csvPath
			},
			shouldFail: true,
		},
		{
			name: "invalid amount",
			setupFile: func(tmpDir string) string {
				csvPath := filepath.Join(tmpDir, "bank.csv")
				content := `unique_identifier,amount,date
BANK-001,not-a-number,2024-01-15`
				os.WriteFile(csvPath, []byte(content), 0644)
				return csvPath
			},
			shouldFail: true,
		},
		{
			name: "invalid date",
			setupFile: func(tmpDir string) string {
				csvPath := filepath.Join(tmpDir, "bank.csv")
				content := `unique_identifier,amount,date
BANK-001,1000.00,invalid-date`
				os.WriteFile(csvPath, []byte(content), 0644)
				return csvPath
			},
			shouldFail: true,
		},
		{
			name: "missing columns",
			setupFile: func(tmpDir string) string {
				csvPath := filepath.Join(tmpDir, "bank.csv")
				content := `unique_identifier,amount,date
BANK-001,1000.00`
				os.WriteFile(csvPath, []byte(content), 0644)
				return csvPath
			},
			shouldFail: true,
		},
		{
			name: "row count is not bank format standard",
			setupFile: func(tmpDir string) string {
				csvPath := filepath.Join(tmpDir, "bank.csv")
				content := `unique_identifier,amount,date1,date2
BANK-001,1000.00,2024-01-15,2024-01-15
BANK-002,500.00,2024-01-15,2024-01-15
BANK-003,250.00,2024-01-17,2024-01-17`
				os.WriteFile(csvPath, []byte(content), 0644)
				return csvPath
			},
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			csvPath := tt.setupFile(tmpDir)

			parser := NewBankStatementLineParser()
			_, err := parser.ParseCSV(csvPath)

			if tt.shouldFail && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.shouldFail && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestBankStatementLineParser_ParseMultipleCSVs(t *testing.T) {
	tests := []struct {
		name               string
		setupFiles         func(tmpDir string) []string
		expectedCount      int
		expectedBankCounts map[string]int
		shouldFail         bool
	}{
		{
			name: "multiple bank files - success",
			setupFiles: func(tmpDir string) []string {
				bca := filepath.Join(tmpDir, "bank_bca.csv")
				bcaContent := `unique_identifier,amount,date
BCA-001,1000.00,2024-01-15
BCA-002,2000.00,2024-01-16`
				os.WriteFile(bca, []byte(bcaContent), 0644)

				mandiri := filepath.Join(tmpDir, "bank_mandiri.csv")
				mandiriContent := `unique_identifier,amount,date
MDR-001,-500.00,2024-01-15
MDR-002,-750.00,2024-01-16`
				os.WriteFile(mandiri, []byte(mandiriContent), 0644)

				return []string{bca, mandiri}
			},
			expectedCount: 4,
			expectedBankCounts: map[string]int{
				"bank_bca":     2,
				"bank_mandiri": 2,
			},
			shouldFail: false,
		},
		{
			name: "one file fails - non-existent",
			setupFiles: func(tmpDir string) []string {
				bca := filepath.Join(tmpDir, "bank_bca.csv")
				bcaContent := `unique_identifier,amount,date
BCA-001,1000.00,2024-01-15`
				os.WriteFile(bca, []byte(bcaContent), 0644)

				nonExistent := filepath.Join(tmpDir, "bank_nonexistent.csv")
				return []string{bca, nonExistent}
			},
			shouldFail: true,
		},
		{
			name: "empty file list",
			setupFiles: func(tmpDir string) []string {
				return []string{}
			},
			expectedCount: 0,
			shouldFail:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			files := tt.setupFiles(tmpDir)

			parser := NewBankStatementLineParser()
			statements, err := parser.ParseMultipleCSVs(files)

			if tt.shouldFail {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if len(statements) != tt.expectedCount {
				t.Errorf("Expected %d statements, got %d", tt.expectedCount, len(statements))
			}

			// Verify bank counts if specified
			if len(tt.expectedBankCounts) > 0 {
				bankCounts := make(map[string]int)
				for _, stmt := range statements {
					bankCounts[stmt.BankName]++
				}

				for bankName, expectedCount := range tt.expectedBankCounts {
					if bankCounts[bankName] != expectedCount {
						t.Errorf("Expected %d statements from %s, got %d", expectedCount, bankName, bankCounts[bankName])
					}
				}
			}
		})
	}
}
