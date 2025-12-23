package models

import (
	"github.com/shopspring/decimal"
)

// ReconciliationResult represents the result of a reconciliation process
type ReconciliationResult struct {
	TotalTransactionsProcessed  int
	TotalMatchedTransactions    int
	TotalUnmatchedTransactions  int
	UnmatchedSystemTransactions []Transaction
	UnmatchedBankStatementLines map[string][]BankStatementLine // Grouped by bank
	TotalDiscrepancies          decimal.Decimal
}
