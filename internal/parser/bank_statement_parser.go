package parser

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"github.com/firmannf/recon/internal/models"
)

// BankStatementLineParser handles parsing of bank statement CSV files
type BankStatementLineParser struct {
	timezone *time.Location
}

// NewBankStatementLineParser creates a new BankStatementLineParser with UTC+7 timezone
func NewBankStatementLineParser() *BankStatementLineParser {
	// Load Asia/Jakarta timezone (UTC+7) by default
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		// Fallback to fixed offset if timezone database unavailable
		loc = time.FixedZone("UTC+7", 7*60*60)
	}
	return &BankStatementLineParser{
		timezone: loc,
	}
}

// ParseCSV reads and parses a bank statement CSV file
// Expected CSV format: unique_identifier,amount,date
func (p *BankStatementLineParser) ParseCSV(filePath string) ([]models.BankStatementLine, error) {
	records, err := readCSVFile(filePath)
	if err != nil {
		return nil, err
	}

	// Extract bank name for grouping from the file path
	bankName := extractFileName(filePath)

	var statementLines []models.BankStatementLine

	// Skip header row
	for i, record := range records[1:] {
		if len(record) < bankStatementColumnCount {
			return nil, fmt.Errorf("invalid record at row %d: expected %d columns, got %d", i+2, bankStatementColumnCount, len(record))
		}

		amount, err := decimal.NewFromString(record[bankStatementColAmount])
		if err != nil {
			return nil, fmt.Errorf("invalid amount at row %d: %w", i+2, err)
		}

		date, err := parseDate(record[bankStatementColDate], p.timezone)
		if err != nil {
			return nil, fmt.Errorf("invalid date at row %d: %w", i+2, err)
		}

		// Derive transaction type from amount sign
		trxType := models.TransactionTypeCredit
		if amount.IsNegative() {
			trxType = models.TransactionTypeDebit
		}

		statementLines = append(statementLines, models.BankStatementLine{
			UniqueIdentifier: record[bankStatementColUniqueIdentifier],
			Amount:           amount,
			Type:             trxType,
			Date:             date,
			BankName:         bankName,
		})
	}

	return statementLines, nil
}

// ParseMultipleCSVs reads and parses multiple bank statement CSV files
func (p *BankStatementLineParser) ParseMultipleCSVs(filePaths []string) ([]models.BankStatementLine, error) {
	var allStatements []models.BankStatementLine

	for _, filePath := range filePaths {
		statements, err := p.ParseCSV(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", filePath, err)
		}
		allStatements = append(allStatements, statements...)
	}

	return allStatements, nil
}
