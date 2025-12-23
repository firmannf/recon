package service

import (
	"fmt"
	"time"

	"github.com/firmannf/recon/internal/models"
	"github.com/firmannf/recon/internal/parser"
	"github.com/shopspring/decimal"
)

type ReconciliationService struct {
	transactionParser   *parser.TransactionParser
	bankStatementParser *parser.BankStatementLineParser
}

func NewReconciliationService() *ReconciliationService {
	return &ReconciliationService{
		transactionParser:   parser.NewTransactionParser(),
		bankStatementParser: parser.NewBankStatementLineParser(),
	}
}

type ReconciliationInput struct {
	SystemTransactionFile string
	BankStatementFiles    []string
	StartDate             time.Time
	EndDate               time.Time
	OutputFile            string
}

// Reconcile performs the reconciliation process
func (s *ReconciliationService) Reconcile(input ReconciliationInput) (*models.ReconciliationResult, error) {
	// Parse system transactions
	systemTransactions, err := s.transactionParser.ParseCSV(input.SystemTransactionFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse system transactions: %w", err)
	}

	// Filter system transactions by date range
	systemTransactions = s.filterTransactionsByDateRange(
		systemTransactions,
		input.StartDate,
		input.EndDate,
	)

	// Parse bank statements from multiple files
	bankStatements, err := s.bankStatementParser.ParseMultipleCSVs(input.BankStatementFiles)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bank statements: %w", err)
	}

	// Filter bank statements by date range
	bankStatements = s.filterBankStatementsByDateRange(
		bankStatements,
		input.StartDate,
		input.EndDate,
	)

	// Perform reconciliation
	result := s.performReconciliation(systemTransactions, bankStatements)

	return result, nil
}

func (s *ReconciliationService) performReconciliation(
	systemTxs []models.Transaction,
	bankStmts []models.BankStatementLine,
) *models.ReconciliationResult {
	result := &models.ReconciliationResult{
		UnmatchedBankStatementLines: make(map[string][]models.BankStatementLine),
		TotalDiscrepancies:          decimal.Zero,
	}

	// Track which transactions and statements have been matched
	matchedSystemTxs := make(map[int]bool)
	matchedBankStmts := make(map[int]bool)

	// Try to match each system transaction with bank statements
	for sysIdx, sysTrx := range systemTxs {
		matched := false

		for bankIdx, bankStmt := range bankStmts {
			// Skip already matched bank statements
			if matchedBankStmts[bankIdx] {
				continue
			}

			if s.isMatch(sysTrx, bankStmt) {
				matched = true
				matchedSystemTxs[sysIdx] = true
				matchedBankStmts[bankIdx] = true
				result.TotalMatchedTransactions++

				// Check for amount discrepancies
				bankAbsAmount := bankStmt.GetAbsoluteAmount()
				diff := sysTrx.Amount.Sub(bankAbsAmount).Abs()

				// This always zero since isMatch checks for exact amount match
				// However, keeping this for future enhancements (e.g., tolerance amount)
				if !diff.IsZero() {
					result.TotalDiscrepancies = result.TotalDiscrepancies.Add(diff)
				}

				break // Move to next system transaction
			}
		}

		if !matched {
			result.UnmatchedSystemTransactions = append(result.UnmatchedSystemTransactions, sysTrx)
		}
	}

	// Collect unmatched bank statements grouped by bank
	for bankIdx, bankStmt := range bankStmts {
		if !matchedBankStmts[bankIdx] {
			if result.UnmatchedBankStatementLines[bankStmt.BankName] == nil {
				result.UnmatchedBankStatementLines[bankStmt.BankName] = []models.BankStatementLine{}
			}
			result.UnmatchedBankStatementLines[bankStmt.BankName] = append(
				result.UnmatchedBankStatementLines[bankStmt.BankName],
				bankStmt,
			)
		}
	}

	// Calculate totals
	result.TotalTransactionsProcessed = max(len(systemTxs), len(bankStmts))
	result.TotalUnmatchedTransactions = len(result.UnmatchedSystemTransactions)
	for _, stmts := range result.UnmatchedBankStatementLines {
		result.TotalUnmatchedTransactions += len(stmts)
	}

	return result
}

// isMatch determines if a system transaction matches a bank statement
func (s *ReconciliationService) isMatch(sysTx models.Transaction, bankStmt models.BankStatementLine) bool {
	// Check transaction type
	if sysTx.Type != bankStmt.Type {
		return false
	}

	// Check amount (exact match)
	bankAbsAmount := bankStmt.GetAbsoluteAmount()
	if !sysTx.Amount.Equal(bankAbsAmount) {
		return false
	}

	// Check date (exact match - same day)
	sysTxDate := time.Date(sysTx.TransactionTime.Year(), sysTx.TransactionTime.Month(), sysTx.TransactionTime.Day(), 0, 0, 0, 0, sysTx.TransactionTime.Location())
	bankStmtDate := time.Date(bankStmt.Date.Year(), bankStmt.Date.Month(), bankStmt.Date.Day(), 0, 0, 0, 0, bankStmt.Date.Location())
	if !sysTxDate.Equal(bankStmtDate) {
		return false
	}

	return true
}

func (s *ReconciliationService) filterTransactionsByDateRange(transactions []models.Transaction, startDate, endDate time.Time) []models.Transaction {
	var filtered []models.Transaction
	for _, trx := range transactions {
		if !trx.TransactionTime.Before(startDate) && !trx.TransactionTime.After(endDate) {
			filtered = append(filtered, trx)
		}
	}
	return filtered
}

func (s *ReconciliationService) filterBankStatementsByDateRange(statements []models.BankStatementLine, startDate, endDate time.Time) []models.BankStatementLine {
	var filtered []models.BankStatementLine
	for _, stmt := range statements {
		if !stmt.Date.Before(startDate) && !stmt.Date.After(endDate) {
			filtered = append(filtered, stmt)
		}
	}
	return filtered
}
