package parser_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/firmannf/recon/internal/models"
	"github.com/firmannf/recon/internal/parser"
)

func TestTransactionParser_ParseCSV_Success(t *testing.T) {
	tests := []struct {
		name          string
		csvContent    string
		expectedCount int
		verify        func(t *testing.T, transactions []models.Transaction)
	}{
		{
			name: "multiple transactions",
			csvContent: `trxID,amount,type,transactionTime
TRX001,1000.50,CREDIT,2024-01-15 10:30:00
TRX002,250,DEBIT,2024-01-16 14:22:30`,
			expectedCount: 2,
			verify: func(t *testing.T, transactions []models.Transaction) {
				if transactions[0].TrxID != "TRX001" {
					t.Errorf("Expected TrxID 'TRX001', got '%s'", transactions[0].TrxID)
				}
				if !transactions[0].Amount.Equal(decimal.NewFromFloat(1000.50)) {
					t.Errorf("Expected amount 1000.50, got %s", transactions[0].Amount)
				}
				if !transactions[1].Amount.Equal(decimal.NewFromFloat(250)) {
					t.Errorf("Expected amount 250, got %s", transactions[1].Amount)
				}
				if transactions[0].Type != models.TransactionTypeCredit {
					t.Errorf("Expected type CREDIT, got %s", transactions[0].Type)
				}
				if transactions[1].Type != models.TransactionTypeDebit {
					t.Errorf("Expected type DEBIT, got %s", transactions[1].Type)
				}
			},
		},
		{
			name: "format date YYYY-MM-DD HH:MM:SS",
			csvContent: `trxID,amount,type,transactionTime
TRX001,1000.00,CREDIT,2024-01-15 10:30:00`,
			expectedCount: 1,
			verify: func(t *testing.T, transactions []models.Transaction) {
				if transactions[0].TransactionTime.Year() != 2024 || transactions[0].TransactionTime.Day() != 15 {
					t.Errorf("Expected 2024-01-15, got %v", transactions[0].TransactionTime)
				}
			},
		},
		{
			name: "format date YYYY-MM-DD HH:MM",
			csvContent: `trxID,amount,type,transactionTime
TRX001,1000.00,CREDIT,2024-01-15 10:30`,
			expectedCount: 1,
			verify: func(t *testing.T, transactions []models.Transaction) {
				if transactions[0].TransactionTime.Year() != 2024 || transactions[0].TransactionTime.Day() != 15 {
					t.Errorf("Expected 2024-01-15, got %v", transactions[0].TransactionTime)
				}
			},
		},
		{
			name: "format date DD/MM/YYYY HH:MM:SS",
			csvContent: `trxID,amount,type,transactionTime
TRX001,1000.00,CREDIT,15/01/2024 10:30:00`,
			expectedCount: 1,
			verify: func(t *testing.T, transactions []models.Transaction) {
				if transactions[0].TransactionTime.Year() != 2024 || transactions[0].TransactionTime.Day() != 15 {
					t.Errorf("Expected 2024-01-15, got %v", transactions[0].TransactionTime)
				}
			},
		},
		{
			name: "format date DD/MM/YYYY HH:MM",
			csvContent: `trxID,amount,type,transactionTime
TRX001,1000.00,CREDIT,15/01/2024 10:30`,
			expectedCount: 1,
			verify: func(t *testing.T, transactions []models.Transaction) {
				if transactions[0].TransactionTime.Year() != 2024 || transactions[0].TransactionTime.Day() != 15 {
					t.Errorf("Expected 2024-01-15, got %v", transactions[0].TransactionTime)
				}
			},
		},

		{
			name: "format date DD-MM-YYYY HH:MM:SS",
			csvContent: `trxID,amount,type,transactionTime
TRX001,1000.00,CREDIT,15-01-2024 10:30:00`,
			expectedCount: 1,
			verify: func(t *testing.T, transactions []models.Transaction) {
				if transactions[0].TransactionTime.Year() != 2024 || transactions[0].TransactionTime.Day() != 15 {
					t.Errorf("Expected 2024-01-15, got %v", transactions[0].TransactionTime)
				}
			},
		},
		{
			name: "format date DD-MM-YYYY HH:MM",
			csvContent: `trxID,amount,type,transactionTime
TRX001,1000.00,CREDIT,15-01-2024 10:30`,
			expectedCount: 1,
			verify: func(t *testing.T, transactions []models.Transaction) {
				if transactions[0].TransactionTime.Year() != 2024 || transactions[0].TransactionTime.Day() != 15 {
					t.Errorf("Expected 2024-01-15, got %v", transactions[0].TransactionTime)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			csvPath := filepath.Join(tmpDir, "transactions.csv")

			if err := os.WriteFile(csvPath, []byte(tt.csvContent), 0644); err != nil {
				t.Fatalf("Failed to create test CSV: %v", err)
			}

			parser := parser.NewTransactionParser()
			transactions, err := parser.ParseCSV(csvPath)

			if err != nil {
				t.Fatalf("Expected successful parse, got error: %v", err)
			}

			if len(transactions) != tt.expectedCount {
				t.Fatalf("Expected %d transactions, got %d", tt.expectedCount, len(transactions))
			}

			if tt.verify != nil {
				tt.verify(t, transactions)
			}
		})
	}
}

func TestTransactionParser_ParseCSV_ErrorCases(t *testing.T) {
	tests := []struct {
		name       string
		setupFile  func(tmpDir string) string
		shouldFail bool
	}{
		{
			name: "file not found",
			setupFile: func(tmpDir string) string {
				return "/nonexistent/path/transactions.csv"
			},
			shouldFail: true,
		},
		{
			name: "non-CSV extension",
			setupFile: func(tmpDir string) string {
				txtPath := filepath.Join(tmpDir, "transactions.txt")
				os.WriteFile(txtPath, []byte("some content"), 0644)
				return txtPath
			},
			shouldFail: true,
		},
		{
			name: "empty file - only header",
			setupFile: func(tmpDir string) string {
				csvPath := filepath.Join(tmpDir, "transactions.csv")
				os.WriteFile(csvPath, []byte("trxID,amount,type,transactionTime\n"), 0644)
				return csvPath
			},
			shouldFail: true,
		},
		{
			name: "invalid amount",
			setupFile: func(tmpDir string) string {
				csvPath := filepath.Join(tmpDir, "transactions.csv")
				content := `trxID,amount,type,transactionTime
TRX001,invalid-amount,CREDIT,2024-01-15 10:30:00`
				os.WriteFile(csvPath, []byte(content), 0644)
				return csvPath
			},
			shouldFail: true,
		},
		{
			name: "invalid date",
			setupFile: func(tmpDir string) string {
				csvPath := filepath.Join(tmpDir, "transactions.csv")
				content := `trxID,amount,type,transactionTime
TRX001,1000.00,CREDIT,invalid-date`
				os.WriteFile(csvPath, []byte(content), 0644)
				return csvPath
			},
			shouldFail: true,
		},
		{
			name: "missing columns",
			setupFile: func(tmpDir string) string {
				csvPath := filepath.Join(tmpDir, "transactions.csv")
				content := `trxID,amount,type,transactionTime
TRX001,1000.00,CREDIT`
				os.WriteFile(csvPath, []byte(content), 0644)
				return csvPath
			},
			shouldFail: true,
		},
		{
			name: "invalid type empty",
			setupFile: func(tmpDir string) string {
				csvPath := filepath.Join(tmpDir, "transactions.csv")
				content := `trxID,amount,type,transactionTime
TRX001,1000.00,,2024-01-15 10:30:00`
				os.WriteFile(csvPath, []byte(content), 0644)
				return csvPath
			},
			shouldFail: true,
		},
		{
			name: "invalid type random",
			setupFile: func(tmpDir string) string {
				csvPath := filepath.Join(tmpDir, "transactions.csv")
				content := `trxID,amount,type,transactionTime
TRX001,1000.00,PAYMENT,2024-01-15 10:30:00`
				os.WriteFile(csvPath, []byte(content), 0644)
				return csvPath
			},
			shouldFail: true,
		},

		{
			name: "row count is not system transaction standard format",
			setupFile: func(tmpDir string) string {
				csvPath := filepath.Join(tmpDir, "bank.csv")
				content := `trxID,amount,type,transactionTime,extraColumn
TRX001,1000.00,CREDIT,2024-01-15 10:30:00,2024-01-15 10:30:00`
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

			parser := parser.NewTransactionParser()
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

// Helper function
func mustDecimal(s string) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		panic(err)
	}
	return d
}
