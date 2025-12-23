package service_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/firmannf/recon/internal/models"
	"github.com/firmannf/recon/internal/service"
)

func TestReconciliation_FileUploadAndParsing(t *testing.T) {
	tests := []struct {
		name           string
		setupFiles     func(tmpDir string) (systemFile string, bankFiles []string)
		expectedError  bool
		expectedResult func(t *testing.T, result *models.ReconciliationResult)
	}{
		{
			name: "parse system transaction CSV successfully",
			setupFiles: func(tmpDir string) (string, []string) {
				systemCSV := filepath.Join(tmpDir, "transactions.csv")
				os.WriteFile(systemCSV, []byte(`trxID,amount,type,transactionTime
TRX001,1000.00,CREDIT,2024-01-15 10:30:00`), 0644)

				bankCSV := filepath.Join(tmpDir, "bank.csv")
				os.WriteFile(bankCSV, []byte(`unique_identifier,amount,date
BANK-001,1000.00,2024-01-15`), 0644)

				return systemCSV, []string{bankCSV}
			},
			expectedError: false,
			expectedResult: func(t *testing.T, result *models.ReconciliationResult) {
				if result.TotalMatchedTransactions != 1 {
					t.Errorf("Expected 1 matched transaction from single bank, got %d", result.TotalMatchedTransactions)
				}
			},
		},
		{
			name: "parse multiple bank statement files",
			setupFiles: func(tmpDir string) (string, []string) {
				systemCSV := filepath.Join(tmpDir, "transactions.csv")
				os.WriteFile(systemCSV, []byte(`trxID,amount,type,transactionTime
TRX001,1000.00,CREDIT,2024-01-15 10:30:00
TRX002,500.50,DEBIT,2024-01-16 14:22:00`), 0644)

				bcaCSV := filepath.Join(tmpDir, "bank_bca.csv")
				os.WriteFile(bcaCSV, []byte(`unique_identifier,amount,date
BCA-001,1000.00,2024-01-15`), 0644)

				mandiriCSV := filepath.Join(tmpDir, "bank_mandiri.csv")
				os.WriteFile(mandiriCSV, []byte(`unique_identifier,amount,date
MDR-001,-500.50,2024-01-16`), 0644)

				return systemCSV, []string{bcaCSV, mandiriCSV}
			},
			expectedError: false,
			expectedResult: func(t *testing.T, result *models.ReconciliationResult) {
				if result.TotalMatchedTransactions != 2 {
					t.Errorf("Expected 2 matched transactions from multiple banks, got %d", result.TotalMatchedTransactions)
				}
			},
		},
		{
			name: "parse dates in multiple formats",
			setupFiles: func(tmpDir string) (string, []string) {
				systemCSV := filepath.Join(tmpDir, "transactions.csv")
				os.WriteFile(systemCSV, []byte(`trxID,amount,type,transactionTime
TRX001,1000.00,CREDIT,2024-01-15 10:30:00
TRX002,500.50,DEBIT,2024-01-16 10:30:00`), 0644)

				bankCSV := filepath.Join(tmpDir, "bank.csv")
				os.WriteFile(bankCSV, []byte(`unique_identifier,amount,date
BANK-001,1000.00,15/01/2024
BANK-002,-500.50,16/01/2024`), 0644)

				return systemCSV, []string{bankCSV}
			},
			expectedError: false,
			expectedResult: func(t *testing.T, result *models.ReconciliationResult) {
				if result.TotalMatchedTransactions != 2 {
					t.Errorf("Expected 2 matched transactions with different date formats, got %d", result.TotalMatchedTransactions)
				}
			},
		},
		{
			name: "file not found",
			setupFiles: func(tmpDir string) (string, []string) {
				return "/nonexistent/path/transactions.csv", []string{"/nonexistent/path/bank.csv"}
			},
			expectedError: true,
		},
		{
			name: "non-CSV extension file",
			setupFiles: func(tmpDir string) (string, []string) {
				txtFile := filepath.Join(tmpDir, "transactions.txt")
				os.WriteFile(txtFile, []byte("some data"), 0644)
				return txtFile, []string{txtFile}
			},
			expectedError: true,
		},
		{
			name: "invalid amount in CSV",
			setupFiles: func(tmpDir string) (string, []string) {
				systemCSV := filepath.Join(tmpDir, "transactions.csv")
				os.WriteFile(systemCSV, []byte(`trxID,amount,type,transactionTime
TRX001,invalid-amount,CREDIT,2024-01-15`), 0644)

				bankCSV := filepath.Join(tmpDir, "bank.csv")
				os.WriteFile(bankCSV, []byte(`unique_identifier,amount,date
BANK-001,1000.00,2024-01-15`), 0644)

				return systemCSV, []string{bankCSV}
			},
			expectedError: true,
		},
		{
			name: "unsupported date format in CSV",
			setupFiles: func(tmpDir string) (string, []string) {
				systemCSV := filepath.Join(tmpDir, "transactions.csv")
				os.WriteFile(systemCSV, []byte(`trxID,amount,type,transactionTime
TRX001,1000.00,CREDIT,20240115`), 0644)

				bankCSV := filepath.Join(tmpDir, "bank.csv")
				os.WriteFile(bankCSV, []byte(`unique_identifier,amount,date
BANK-001,1000.00,20240115`), 0644)

				return systemCSV, []string{bankCSV}
			},
			expectedError: true,
		},
		{
			name: "unstandardized CSV format",
			setupFiles: func(tmpDir string) (string, []string) {
				systemCSV := filepath.Join(tmpDir, "transactions.csv")
				os.WriteFile(systemCSV, []byte(`trxID,amount,type,transactionTime,extraColumn
TRX001,1000.00,CREDIT,2024-01-15 10:30:00,extra`), 0644)

				bankCSV := filepath.Join(tmpDir, "bank.csv")
				os.WriteFile(bankCSV, []byte(`unique_identifier,amount,date,extraColumn
BANK-001,1000.00,2024-01-15,extra`), 0644)
				return systemCSV, []string{bankCSV}
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			systemFile, bankFiles := tt.setupFiles(tmpDir)

			reconService := service.NewReconciliationService()
			input := service.ReconciliationInput{
				SystemTransactionFile: systemFile,
				BankStatementFiles:    bankFiles,
				StartDate:             mustParseTime("2024-01-01 00:00:00"),
				EndDate:               mustParseTime("2024-12-31 23:59:59"),
				MatchStrategy:         service.NewExactMatchStrategy(),
			}

			result, err := reconService.Reconcile(input)

			if tt.expectedError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if tt.expectedResult != nil && result != nil {
				tt.expectedResult(t, result)
			}
		})
	}
}

func TestReconciliation_MatchingLogic(t *testing.T) {
	tests := []struct {
		name           string
		setupFiles     func(tmpDir string) (systemFile string, bankFiles []string)
		expectedResult func(t *testing.T, result *models.ReconciliationResult)
	}{
		{
			name: "match based on date and amount",
			setupFiles: func(tmpDir string) (string, []string) {
				systemCSV := filepath.Join(tmpDir, "transactions.csv")
				os.WriteFile(systemCSV, []byte(`trxID,amount,type,transactionTime
TRX001,1000.00,CREDIT,2024-01-15 10:30:00
TRX002,500.50,DEBIT,2024-01-15 10:31:00`), 0644)

				bankCSV := filepath.Join(tmpDir, "bank.csv")
				os.WriteFile(bankCSV, []byte(`unique_identifier,amount,date
BANK_BCA_001,1000.00,2024-01-15
BANK_BCA_002,-500.50,2024-01-15`), 0644)

				return systemCSV, []string{bankCSV}
			},
			expectedResult: func(t *testing.T, result *models.ReconciliationResult) {
				if result.TotalTransactionsProcessed != 4 {
					t.Errorf("Expected 4 transactions to be processed (2 system + 2 bank), got %d", result.TotalTransactionsProcessed)
				}
				if result.TotalMatchedTransactions != 2 {
					t.Errorf("Expected 2 transactions to match, got %d matches", result.TotalMatchedTransactions)
				}
				if result.TotalUnmatchedTransactions != 0 {
					t.Errorf("Expected 0 transaction to unmatched, got %d unmatched", result.TotalUnmatchedTransactions)
				}
				if len(result.UnmatchedSystemTransactions) != 0 {
					t.Errorf("Expected 0 transaction to unmatched, got %d unmatched", len(result.UnmatchedSystemTransactions))
				}
				if len(result.UnmatchedBankStatementLines) != 0 {
					t.Errorf("Expected 0 transaction to unmatched, got %d unmatched", len(result.UnmatchedBankStatementLines))
				}
				if result.TotalDiscrepancies.GreaterThan(decimal.Zero) {
					t.Errorf("Expected 0 discrepancies, got %s", result.TotalDiscrepancies.String())
				}
			},
		},
		{
			name: "exact same day matching",
			setupFiles: func(tmpDir string) (string, []string) {
				systemCSV := filepath.Join(tmpDir, "transactions.csv")
				os.WriteFile(systemCSV, []byte(`trxID,amount,type,transactionTime
TRX001,1000.00,CREDIT,2024-01-15 10:30:00
TRX002,500.50,DEBIT,2024-01-16 14:22:00`), 0644)

				bankCSV := filepath.Join(tmpDir, "bank.csv")
				os.WriteFile(bankCSV, []byte(`unique_identifier,amount,date
BANK_BCA_001,1000.00,2024-01-16
BANK_BCA_002,-500.50,2024-01-17`), 0644)

				return systemCSV, []string{bankCSV}
			},
			expectedResult: func(t *testing.T, result *models.ReconciliationResult) {
				if result.TotalTransactionsProcessed != 4 {
					t.Errorf("Expected 4 transactions to be processed (2 system + 2 bank), got %d", result.TotalTransactionsProcessed)
				}
				if result.TotalMatchedTransactions != 0 {
					t.Errorf("Expected no match with different dates (exact matching), got %d matches", result.TotalMatchedTransactions)
				}
				if result.TotalUnmatchedTransactions != 4 {
					t.Errorf("Expected 4 unmatched transactions (2 system + 2 bank), got %d", result.TotalUnmatchedTransactions)
				}
				if len(result.UnmatchedSystemTransactions) != 2 {
					t.Errorf("Expected 2 unmatched system transactions, got %d", len(result.UnmatchedSystemTransactions))
				}
				totalUnmatchedBank := 0
				for _, stmtLines := range result.UnmatchedBankStatementLines {
					totalUnmatchedBank += len(stmtLines)
				}
				if totalUnmatchedBank != 2 {
					t.Errorf("Expected 2 unmatched bank statements, got %d", totalUnmatchedBank)
				}
				if result.TotalDiscrepancies.GreaterThan(decimal.Zero) {
					t.Errorf("Expected 0 discrepancies, got %s", result.TotalDiscrepancies.String())
				}
			},
		},
		{
			name: "first-match-wins strategy",
			setupFiles: func(tmpDir string) (string, []string) {
				systemCSV := filepath.Join(tmpDir, "transactions.csv")
				os.WriteFile(systemCSV, []byte(`trxID,amount,type,transactionTime
TRX001,1000.00,CREDIT,2024-01-15 10:30:00`), 0644)

				bankCSV := filepath.Join(tmpDir, "bank.csv")
				os.WriteFile(bankCSV, []byte(`unique_identifier,amount,date
BANK_BCA_001,1000.00,2024-01-15
BANK_BCA_002,1000.00,2024-01-15`), 0644)

				return systemCSV, []string{bankCSV}
			},
			expectedResult: func(t *testing.T, result *models.ReconciliationResult) {
				if result.TotalTransactionsProcessed != 3 {
					t.Errorf("Expected 3 transactions to be processed (1 system + 2 bank), got %d", result.TotalTransactionsProcessed)
				}
				if result.TotalMatchedTransactions != 1 {
					t.Errorf("Expected only first match to succeed, got %d matches", result.TotalMatchedTransactions)
				}
				if result.TotalUnmatchedTransactions != 1 {
					t.Errorf("Expected 1 unmatched transaction, got %d", result.TotalUnmatchedTransactions)
				}
				if len(result.UnmatchedSystemTransactions) != 0 {
					t.Errorf("Expected 0 unmatched system transactions, got %d", len(result.UnmatchedSystemTransactions))
				}
				totalUnmatchedBank := 0
				for _, stmtLines := range result.UnmatchedBankStatementLines {
					totalUnmatchedBank += len(stmtLines)
				}
				if totalUnmatchedBank != 1 {
					t.Errorf("Expected second bank statement to remain unmatched, got %d unmatched", totalUnmatchedBank)
				}
				if result.TotalDiscrepancies.GreaterThan(decimal.Zero) {
					t.Errorf("Expected 0 discrepancies, got %s", result.TotalDiscrepancies.String())
				}
			},
		},
		{
			name: "mixed matched and unmatched transactions",
			setupFiles: func(tmpDir string) (string, []string) {
				systemCSV := filepath.Join(tmpDir, "transactions.csv")
				os.WriteFile(systemCSV, []byte(`trxID,amount,type,transactionTime
TRX001,1000.00,CREDIT,2024-01-15 10:30:00
TRX002,2000.00,CREDIT,2024-01-16 10:30:00
TRX003,3000.00,CREDIT,2024-01-17 10:30:00`), 0644)

				bankCSV := filepath.Join(tmpDir, "bank.csv")
				os.WriteFile(bankCSV, []byte(`unique_identifier,amount,date
BANK-001,1000.00,2024-01-15
BANK-002,4000.00,2024-01-18
BANK-003,5000.00,2024-01-19
BANK-004,6000.00,2024-01-20`), 0644)

				return systemCSV, []string{bankCSV}
			},
			expectedResult: func(t *testing.T, result *models.ReconciliationResult) {
				if result.TotalTransactionsProcessed != 7 {
					t.Errorf("Expected 7 transactions to be processed (3 system + 4 bank), got %d", result.TotalTransactionsProcessed)
				}
				if result.TotalMatchedTransactions != 1 {
					t.Errorf("Expected 1 matched transaction, got %d", result.TotalMatchedTransactions)
				}
				if result.TotalUnmatchedTransactions != 5 {
					t.Errorf("Expected 5 unmatched transactions (2 system + 3 bank), got %d", result.TotalUnmatchedTransactions)
				}
				if len(result.UnmatchedSystemTransactions) != 2 {
					t.Errorf("Expected 2 unmatched system transactions, got %d", len(result.UnmatchedSystemTransactions))
				}
				totalUnmatchedBank := 0
				for _, stmtLines := range result.UnmatchedBankStatementLines {
					totalUnmatchedBank += len(stmtLines)
				}
				if totalUnmatchedBank != 3 {
					t.Errorf("Expected 3 unmatched bank statements, got %d", totalUnmatchedBank)
				}
				if result.TotalDiscrepancies.GreaterThan(decimal.Zero) {
					t.Errorf("Expected 0 discrepancies, got %s", result.TotalDiscrepancies.String())
				}
			},
		},
		{
			name: "unmatched bank statements only",
			setupFiles: func(tmpDir string) (string, []string) {
				systemCSV := filepath.Join(tmpDir, "transactions.csv")
				os.WriteFile(systemCSV, []byte(`trxID,amount,type,transactionTime
TRX002,1000.00,CREDIT,2024-01-15 10:30:00
TRX003,500.50,DEBIT,2024-01-15 10:31:00`), 0644)

				bankCSV := filepath.Join(tmpDir, "bank.csv")
				os.WriteFile(bankCSV, []byte(`unique_identifier,amount,date
BANK_BCA_001,1500.00,2024-01-15
BANK_BCA_002,1000.00,2024-01-15
BANK_BCA_003,-500.50,2024-01-15`), 0644)

				return systemCSV, []string{bankCSV}
			},
			expectedResult: func(t *testing.T, result *models.ReconciliationResult) {
				if result.TotalTransactionsProcessed != 5 {
					t.Errorf("Expected 5 transactions to be processed (2 system + 3 bank), got %d", result.TotalTransactionsProcessed)
				}
				if result.TotalMatchedTransactions != 2 {
					t.Errorf("Expected 2 matched transaction, got %d", result.TotalMatchedTransactions)
				}
				if result.TotalUnmatchedTransactions != 1 {
					t.Errorf("Expected 1 unmatched transaction, got %d", result.TotalUnmatchedTransactions)
				}
				if len(result.UnmatchedSystemTransactions) != 0 {
					t.Errorf("Expected 0 unmatched system transactions, got %d", len(result.UnmatchedSystemTransactions))
				}
				totalUnmatchedBank := 0
				for _, stmtLines := range result.UnmatchedBankStatementLines {
					totalUnmatchedBank += len(stmtLines)
				}
				if totalUnmatchedBank != 1 {
					t.Errorf("Expected 1 unmatched bank statement, got %d", totalUnmatchedBank)
				}
				if result.TotalDiscrepancies.GreaterThan(decimal.Zero) {
					t.Errorf("Expected 0 discrepancies, got %s", result.TotalDiscrepancies.String())
				}
			},
		},
		{
			name: "unmatched system transactions only",
			setupFiles: func(tmpDir string) (string, []string) {
				systemCSV := filepath.Join(tmpDir, "transactions.csv")
				os.WriteFile(systemCSV, []byte(`trxID,amount,type,transactionTime
TRX001,3000.00,CREDIT,2024-01-15 10:30:00
TRX002,1000.00,CREDIT,2024-01-15 10:30:00
TRX003,500.50,DEBIT,2024-01-15 10:31:00`), 0644)

				bankCSV := filepath.Join(tmpDir, "bank.csv")
				os.WriteFile(bankCSV, []byte(`unique_identifier,amount,date
BANK_BCA_002,1000.00,2024-01-15
BANK_BCA_003,-500.50,2024-01-15`), 0644)

				return systemCSV, []string{bankCSV}
			},
			expectedResult: func(t *testing.T, result *models.ReconciliationResult) {
				if result.TotalTransactionsProcessed != 5 {
					t.Errorf("Expected 5 transactions to be processed (3 system + 2 bank), got %d", result.TotalTransactionsProcessed)
				}
				if result.TotalMatchedTransactions != 2 {
					t.Errorf("Expected 2 matched transaction, got %d", result.TotalMatchedTransactions)
				}
				if result.TotalUnmatchedTransactions != 1 {
					t.Errorf("Expected 1 unmatched transaction, got %d", result.TotalUnmatchedTransactions)
				}
				if len(result.UnmatchedSystemTransactions) != 1 {
					t.Errorf("Expected 1 unmatched system transactions, got %d", len(result.UnmatchedSystemTransactions))
				}
				totalUnmatchedBank := 0
				for _, stmtLines := range result.UnmatchedBankStatementLines {
					totalUnmatchedBank += len(stmtLines)
				}
				if totalUnmatchedBank != 0 {
					t.Errorf("Expected 0 unmatched bank statement, got %d", totalUnmatchedBank)
				}
				if result.TotalDiscrepancies.GreaterThan(decimal.Zero) {
					t.Errorf("Expected 0 discrepancies, got %s", result.TotalDiscrepancies.String())
				}
			},
		},
		{
			name: "unmatched transactions from multiple banks",
			setupFiles: func(tmpDir string) (string, []string) {
				systemCSV := filepath.Join(tmpDir, "transactions.csv")
				os.WriteFile(systemCSV, []byte(`trxID,amount,type,transactionTime
TRX001,1000.00,CREDIT,2024-01-15 10:30:00`), 0644)

				bcaCSV := filepath.Join(tmpDir, "bank_bca.csv")
				os.WriteFile(bcaCSV, []byte(`unique_identifier,amount,date
BCA-001,2000.00,2024-01-15`), 0644)

				mandiriCSV := filepath.Join(tmpDir, "bank_mandiri.csv")
				os.WriteFile(mandiriCSV, []byte(`unique_identifier,amount,date
MDR-001,3000.00,2024-01-15`), 0644)

				return systemCSV, []string{bcaCSV, mandiriCSV}
			},
			expectedResult: func(t *testing.T, result *models.ReconciliationResult) {
				if result.TotalTransactionsProcessed != 3 {
					t.Errorf("Expected 3 transactions to be processed (1 system + 2 bank), got %d", result.TotalTransactionsProcessed)
				}
				if result.TotalMatchedTransactions != 0 {
					t.Errorf("Expected 0 matched transactions, got %d", result.TotalMatchedTransactions)
				}
				if result.TotalUnmatchedTransactions != 3 {
					t.Errorf("Expected 3 unmatched transactions (1 system + 2 bank), got %d", result.TotalUnmatchedTransactions)
				}
				if len(result.UnmatchedSystemTransactions) != 1 {
					t.Errorf("Expected 1 unmatched system transaction, got %d", len(result.UnmatchedSystemTransactions))
				}
				if len(result.UnmatchedBankStatementLines) != 2 {
					t.Errorf("Expected unmatched statements from 2 banks, got %d", len(result.UnmatchedBankStatementLines))
				}
				if _, exists := result.UnmatchedBankStatementLines["bank_bca"]; !exists {
					t.Error("Expected unmatched statements from bank_bca")
				}
				if _, exists := result.UnmatchedBankStatementLines["bank_mandiri"]; !exists {
					t.Error("Expected unmatched statements from bank_mandiri")
				}
				totalUnmatchedBank := 0
				for _, stmtLines := range result.UnmatchedBankStatementLines {
					totalUnmatchedBank += len(stmtLines)
				}
				if totalUnmatchedBank != 2 {
					t.Errorf("Expected 2 unmatched bank statements, got %d", totalUnmatchedBank)
				}
				if result.TotalDiscrepancies.GreaterThan(decimal.Zero) {
					t.Errorf("Expected 0 discrepancies, got %s", result.TotalDiscrepancies.String())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			systemFile, bankFiles := tt.setupFiles(tmpDir)

			reconService := service.NewReconciliationService()
			input := service.ReconciliationInput{
				SystemTransactionFile: systemFile,
				BankStatementFiles:    bankFiles,
				StartDate:             mustParseTime("2024-01-01 00:00:00"),
				EndDate:               mustParseTime("2024-12-31 23:59:59"),
				MatchStrategy:         service.NewExactMatchStrategy(),
			}

			result, err := reconService.Reconcile(input)
			if err != nil {
				t.Fatalf("Reconciliation failed: %v", err)
			}

			if tt.expectedResult != nil {
				tt.expectedResult(t, result)
			}
		})
	}
}

func TestReconciliation_DateRangeFiltering(t *testing.T) {
	tests := []struct {
		name           string
		setupFiles     func(tmpDir string) (systemFile string, bankFiles []string)
		startDate      time.Time
		endDate        time.Time
		expectedError  bool
		expectedResult func(t *testing.T, result *models.ReconciliationResult)
	}{
		{
			name: "all transactions in specified year",
			setupFiles: func(tmpDir string) (string, []string) {
				systemCSV := filepath.Join(tmpDir, "transactions.csv")
				os.WriteFile(systemCSV, []byte(`trxID,amount,type,transactionTime
TRX001,1000.00,CREDIT,2024-01-15 10:30:00
TRX002,2000.00,CREDIT,2024-06-20 10:30:00
TRX003,3000.00,CREDIT,2024-12-25 10:30:00`), 0644)

				bankCSV := filepath.Join(tmpDir, "bank.csv")
				os.WriteFile(bankCSV, []byte(`unique_identifier,amount,date
BANK-001,1000.00,2024-01-15
BANK-002,2000.00,2024-06-20
BANK-003,3000.00,2024-12-25`), 0644)

				return systemCSV, []string{bankCSV}
			},
			startDate:     mustParseTime("2024-01-01 00:00:00"),
			endDate:       mustParseTime("2024-12-31 23:59:59"),
			expectedError: false,
			expectedResult: func(t *testing.T, result *models.ReconciliationResult) {
				if result.TotalTransactionsProcessed != 6 {
					t.Errorf("Expected 6 transactions processed, got %d", result.TotalTransactionsProcessed)
				}
			},
		},
		{
			name: "invalid date range - end before start",
			setupFiles: func(tmpDir string) (string, []string) {
				systemCSV := filepath.Join(tmpDir, "transactions.csv")
				os.WriteFile(systemCSV, []byte(`trxID,amount,type,transactionTime
TRX001,1000.00,CREDIT,2024-01-15 10:30:00`), 0644)

				bankCSV := filepath.Join(tmpDir, "bank.csv")
				os.WriteFile(bankCSV, []byte(`unique_identifier,amount,date
BANK-001,1000.00,2024-01-15`), 0644)

				return systemCSV, []string{bankCSV}
			},
			startDate:     mustParseTime("2024-12-31 00:00:00"),
			endDate:       mustParseTime("2024-01-01 23:59:59"),
			expectedError: true,
			expectedResult: func(t *testing.T, result *models.ReconciliationResult) {
				// Should not reach here - error expected
				t.Error("Expected error for invalid date range, but got result")
			},
		},
		{
			name: "same start and end date allowed",
			setupFiles: func(tmpDir string) (string, []string) {
				systemCSV := filepath.Join(tmpDir, "transactions.csv")
				os.WriteFile(systemCSV, []byte(`trxID,amount,type,transactionTime
TRX001,1000.00,CREDIT,2024-01-15 10:30:00`), 0644)

				bankCSV := filepath.Join(tmpDir, "bank.csv")
				os.WriteFile(bankCSV, []byte(`unique_identifier,amount,date
BANK-001,1000.00,2024-01-15`), 0644)

				return systemCSV, []string{bankCSV}
			},
			startDate:     mustParseTime("2024-01-15 00:00:00"),
			endDate:       mustParseTime("2024-01-15 23:59:59"),
			expectedError: false,
			expectedResult: func(t *testing.T, result *models.ReconciliationResult) {
				if result.TotalTransactionsProcessed != 2 {
					t.Errorf("Expected 2 transactions processed on same day, got %d", result.TotalTransactionsProcessed)
				}
			},
		},
		{
			name: "optional end date defaults to start date",
			setupFiles: func(tmpDir string) (string, []string) {
				systemCSV := filepath.Join(tmpDir, "transactions.csv")
				os.WriteFile(systemCSV, []byte(`trxID,amount,type,transactionTime
TRX001,1000.00,CREDIT,2024-01-15 10:30:00
TRX002,2000.00,CREDIT,2024-01-16 10:30:00`), 0644)

				bankCSV := filepath.Join(tmpDir, "bank.csv")
				os.WriteFile(bankCSV, []byte(`unique_identifier,amount,date
BANK-001,1000.00,2024-01-15
BANK-002,2000.00,2024-01-16`), 0644)

				return systemCSV, []string{bankCSV}
			},
			startDate:     mustParseTime("2024-01-15 00:00:00"),
			endDate:       time.Time{}, // Zero value - should default to start date
			expectedError: false,
			expectedResult: func(t *testing.T, result *models.ReconciliationResult) {
				// Should only process transactions on 2024-01-15, not 2024-01-16
				if result.TotalTransactionsProcessed != 2 {
					t.Errorf("Expected 2 transactions when end date defaults to start date, got %d", result.TotalTransactionsProcessed)
				}
			},
		},
		{
			name: "boundary dates included",
			setupFiles: func(tmpDir string) (string, []string) {
				systemCSV := filepath.Join(tmpDir, "transactions.csv")
				os.WriteFile(systemCSV, []byte(`trxID,amount,type,transactionTime
TRX001,1000.00,CREDIT,2024-01-01 00:00:00
TRX002,2000.00,CREDIT,2024-01-31 23:59:59`), 0644)

				bankCSV := filepath.Join(tmpDir, "bank.csv")
				os.WriteFile(bankCSV, []byte(`unique_identifier,amount,date
BANK-001,1000.00,2024-01-01
BANK-002,2000.00,2024-01-31`), 0644)

				return systemCSV, []string{bankCSV}
			},
			startDate:     mustParseTime("2024-01-01 00:00:00"),
			endDate:       mustParseTime("2024-01-31 23:59:59"),
			expectedError: false,
			expectedResult: func(t *testing.T, result *models.ReconciliationResult) {
				if result.TotalTransactionsProcessed != 4 {
					t.Errorf("Expected both boundary transactions to be included, got %d transactions processed", result.TotalTransactionsProcessed)
				}
			},
		},

		{
			name: "outside date range excluded",
			setupFiles: func(tmpDir string) (string, []string) {
				systemCSV := filepath.Join(tmpDir, "transactions.csv")
				os.WriteFile(systemCSV, []byte(`trxID,amount,type,transactionTime
TRX001,1000.00,CREDIT,2023-12-31 00:00:00
TRX002,2000.00,CREDIT,2024-02-01 23:59:59`), 0644)

				bankCSV := filepath.Join(tmpDir, "bank.csv")
				os.WriteFile(bankCSV, []byte(`unique_identifier,amount,date
BANK-001,1000.00,2023-12-31
BANK-002,2000.00,2024-02-01`), 0644)

				return systemCSV, []string{bankCSV}
			},
			startDate:     mustParseTime("2024-01-01 00:00:00"),
			endDate:       mustParseTime("2024-01-31 23:59:59"),
			expectedError: false,
			expectedResult: func(t *testing.T, result *models.ReconciliationResult) {
				if result.TotalTransactionsProcessed != 0 {
					t.Errorf("Expected both boundary transactions to be included, got %d transactions processed", result.TotalTransactionsProcessed)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			systemFile, bankFiles := tt.setupFiles(tmpDir)

			reconService := service.NewReconciliationService()
			input := service.ReconciliationInput{
				SystemTransactionFile: systemFile,
				BankStatementFiles:    bankFiles,
				StartDate:             tt.startDate,
				EndDate:               tt.endDate,
				MatchStrategy:         service.NewExactMatchStrategy(),
			}

			result, err := reconService.Reconcile(input)

			if tt.expectedError {
				if err == nil {
					t.Fatalf("Expected error but got success")
				}
				return // Test passed - error occurred as expected
			}

			if err != nil {
				t.Fatalf("Reconciliation failed: %v", err)
			}

			if tt.expectedResult != nil {
				tt.expectedResult(t, result)
			}
		})
	}
}

func mustParseTime(s string) time.Time {
	// Use UTC+7 timezone to match parser behavior
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.FixedZone("UTC+7", 7*60*60)
	}
	t, err := time.ParseInLocation("2006-01-02 15:04:05", s, loc)
	if err != nil {
		panic(err)
	}
	return t
}
