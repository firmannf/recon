package service

import (
	"fmt"
	"time"

	"github.com/firmannf/recon/internal/models"
	"github.com/shopspring/decimal"
)

type MatchStrategy interface {
	// BuildKey creates an index key for lookup
	BuildKey(txType models.TransactionType, amount decimal.Decimal, date time.Time, id string) string

	// IsMatch validates if two transactions match (for additional validation after key lookup)
	IsMatch(sysTx models.Transaction, bankStmt models.BankStatementLine) bool
}

// ExactMatchStrategy matches by exact type, amount, and date
type ExactMatchStrategy struct{}

func NewExactMatchStrategy() *ExactMatchStrategy {
	return &ExactMatchStrategy{}
}

func (s *ExactMatchStrategy) BuildKey(txType models.TransactionType, amount decimal.Decimal, date time.Time, id string) string {
	return fmt.Sprintf("%s_%s_%s", txType, amount.String(), date.Format("2006-01-02"))
}

func (s *ExactMatchStrategy) IsMatch(sysTx models.Transaction, bankStmt models.BankStatementLine) bool {
	// Key already ensures type, amount, and date match
	// No additional validation needed for exact match
	return true
}
