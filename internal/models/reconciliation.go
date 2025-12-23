package models

import (
	"fmt"
	"strings"

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

// Print outputs a formatted reconciliation summary
func (r *ReconciliationResult) Print() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("TRANSACTION RECONCILIATION SUMMARY")
	fmt.Println(strings.Repeat("=", 80))

	fmt.Printf("\nTotal Transactions Processed: %d\n", r.TotalTransactionsProcessed)
	fmt.Printf("Total Matched Transactions: %d\n", r.TotalMatchedTransactions)
	fmt.Printf("Total Unmatched Transactions: %d\n", r.TotalUnmatchedTransactions)
	fmt.Printf("Total Discrepancies (Amount): Rp. %s\n", r.TotalDiscrepancies)

	// Print unmatched system transactions
	if len(r.UnmatchedSystemTransactions) > 0 {
		fmt.Println("\n" + strings.Repeat("-", 80))
		fmt.Printf("UNMATCHED SYSTEM TRANSACTIONS: %d\n", len(r.UnmatchedSystemTransactions))
		fmt.Println(strings.Repeat("-", 80))
		fmt.Printf("%-20s %-10s %-25s %20s \n", "TrxID", "Type", "Transaction Time", "Amount")
		for _, trx := range r.UnmatchedSystemTransactions {
			fmt.Printf("%-20s %-10s %-25s %20s\n", trx.TrxID, trx.Type, trx.TransactionTime.Format("2006-01-02 15:04:05"), fmt.Sprintf("Rp. %v", trx.Amount.StringFixed(2)))
		}
	}

	// Print unmatched bank statements grouped by bank
	if len(r.UnmatchedBankStatementLines) > 0 {
		totalUnmatchedBank := 0
		for _, statements := range r.UnmatchedBankStatementLines {
			totalUnmatchedBank += len(statements)
		}

		fmt.Println("\n" + strings.Repeat("-", 80))
		fmt.Printf("UNMATCHED BANK STATEMENTS: %d\n", totalUnmatchedBank)
		fmt.Println(strings.Repeat("-", 80))

		for bankName, statements := range r.UnmatchedBankStatementLines {
			fmt.Printf("\nBank: %s (%d transactions)\n", bankName, len(statements))
			fmt.Printf("%-20s %-10s %20s\n", "Unique Identifier", "Date", "Amount")
			for _, stmt := range statements {
				fmt.Printf("%-20s %-10s %20s\n", stmt.UniqueIdentifier, stmt.Date.Format("2006-01-02"), fmt.Sprintf("Rp. %v", stmt.Amount.StringFixed(2)))
			}
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
}
