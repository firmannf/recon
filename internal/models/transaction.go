package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// TransactionType represents the type of transaction
type TransactionType string

const (
	TransactionTypeDebit  TransactionType = "DEBIT"
	TransactionTypeCredit TransactionType = "CREDIT"
)

// Transaction represents a system transaction entry
type Transaction struct {
	TrxID           string
	Amount          decimal.Decimal
	Type            TransactionType
	TransactionTime time.Time
}

// BankStatement represents a bank statement line
type BankStatementLine struct {
	UniqueIdentifier string
	Amount           decimal.Decimal // Can be negative for debit
	Type             TransactionType // Derived from amount sign
	Date             time.Time
	BankName         string
}

// GetAbsoluteAmount returns the absolute value of the amount
func (bs *BankStatementLine) GetAbsoluteAmount() decimal.Decimal {
	return bs.Amount.Abs()
}
