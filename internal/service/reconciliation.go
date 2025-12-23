package service

import (
	"time"

	"github.com/firmannf/recon/internal/models"
	"github.com/firmannf/recon/internal/parser"
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

func (s *ReconciliationService) Reconcile(input ReconciliationInput) (*models.ReconciliationResult, error) {
	return nil, nil
}
