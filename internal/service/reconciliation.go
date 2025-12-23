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
	MatchStrategy         MatchStrategy
}

// Reconcile performs the reconciliation process
func (s *ReconciliationService) Reconcile(input ReconciliationInput) (*models.ReconciliationResult, error) {
	// If end date is not provided (zero value), set it to end of start date
	if input.EndDate.IsZero() {
		input.EndDate = input.StartDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	}

	// Validate date range
	if input.StartDate.After(input.EndDate) {
		return nil, fmt.Errorf("start date must not be after end date")
	}

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
	result := s.performReconciliation(systemTransactions, bankStatements, input.MatchStrategy)

	return result, nil
}

func (s *ReconciliationService) performReconciliation(
	systemTxs []models.Transaction,
	bankStmts []models.BankStatementLine,
	matchStrategy MatchStrategy,
) *models.ReconciliationResult {
	result := &models.ReconciliationResult{
		TotalSystemTransactions:     len(systemTxs),
		TotalBankStatementLines:     len(bankStmts),
		UnmatchedBankStatementLines: make(map[string][]models.BankStatementLine),
		TotalDiscrepancies:          decimal.Zero,
	}

	// Build index of bank statements by matching key for O(1) lookup
	// Key format depends on strategy (e.g., "TYPE_AMOUNT_DATE", "TYPE_DATE", "ID", etc.)
	bankStmtIndex := make(map[string][]int)
	for bankIdx, bankStmt := range bankStmts {
		key := matchStrategy.BuildKey(bankStmt.Type, bankStmt.GetAbsoluteAmount(), bankStmt.Date, bankStmt.UniqueIdentifier)
		bankStmtIndex[key] = append(bankStmtIndex[key], bankIdx)
	}

	// Track which statements have been matched
	matchedSystemTxs := make(map[int]bool)
	matchedBankStmts := make(map[int]bool)

	// Try to match each system transaction with bank statements
	for sysIdx, sysTrx := range systemTxs {
		matched := false

		// Look up potential matches using index - O(1) instead of O(m)
		key := matchStrategy.BuildKey(sysTrx.Type, sysTrx.Amount, sysTrx.TransactionTime, sysTrx.TrxID)
		if candidates, exists := bankStmtIndex[key]; exists {
			for _, bankIdx := range candidates {
				// Skip already matched bank statements
				if matchedBankStmts[bankIdx] {
					continue
				}

				// Validate match using strategy (for tolerance checking, etc.)
				if !matchStrategy.IsMatch(sysTrx, bankStmts[bankIdx]) {
					continue
				}

				// Found a match (first available candidate)
				matched = true
				matchedSystemTxs[sysIdx] = true
				matchedBankStmts[bankIdx] = true
				result.TotalMatchedTransactions++

				// Check for amount discrepancies
				bankAbsAmount := bankStmts[bankIdx].GetAbsoluteAmount()
				diff := sysTrx.Amount.Sub(bankAbsAmount).Abs()

				// This always zero since isMatch checks for exact amount match
				// However, keeping this for future enhancements (e.g., tolerance amount)
				if !diff.IsZero() {
					result.TotalDiscrepancies = result.TotalDiscrepancies.Add(diff)
				}

				break // Move to next system transaction to avoid multiple matches
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
	result.TotalTransactionsProcessed = len(systemTxs) + len(bankStmts)
	result.TotalUnmatchedTransactions = len(result.UnmatchedSystemTransactions)
	for _, stmts := range result.UnmatchedBankStatementLines {
		result.TotalUnmatchedTransactions += len(stmts)
	}

	return result
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
