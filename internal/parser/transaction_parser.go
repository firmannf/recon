package parser

import (
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/firmannf/recon/internal/models"
)

// TransactionParser handles parsing of system transaction CSV files
type TransactionParser struct {
	// Future: timezone, date format preferences, etc.
}

// NewTransactionParser creates a new TransactionParser
func NewTransactionParser() *TransactionParser {
	return &TransactionParser{}
}

// ParseCSV reads and parses a transaction CSV file
// Expected CSV format: trxID,amount,type,transactionTime
func (p *TransactionParser) ParseCSV(filePath string) ([]models.Transaction, error) {
	records, err := readCSVFile(filePath)
	if err != nil {
		return nil, err
	}

	var transactions []models.Transaction

	// Skip header row
	for i, record := range records[1:] {
		if len(record) < transactionColumnCount {
			return nil, fmt.Errorf("invalid record at row %d: expected %d columns, got %d", i+2, transactionColumnCount, len(record))
		}

		amount, err := decimal.NewFromString(record[transactionColAmount])
		if err != nil {
			return nil, fmt.Errorf("invalid amount at row %d: %w", i+2, err)
		}

		txType := models.TransactionType(record[transactionColType])
		if txType != models.TransactionTypeDebit && txType != models.TransactionTypeCredit {
			return nil, fmt.Errorf("invalid transaction type at row %d: %s", i+2, record[transactionColType])
		}

		// Try multiple date formats
		transactionTime, err := parseDate(record[transactionColTransactionTime])
		if err != nil {
			return nil, fmt.Errorf("invalid transaction time at row %d: %w", i+2, err)
		}

		transactions = append(transactions, models.Transaction{
			TrxID:           record[transactionColTrxID],
			Amount:          amount,
			Type:            txType,
			TransactionTime: transactionTime,
		})
	}

	return transactions, nil
}
