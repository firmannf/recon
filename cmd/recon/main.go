package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/firmannf/recon/internal/service"
	"github.com/shopspring/decimal"
)

const (
	DEFAULT_DATE_FORMAT = "2006-01-02"
)

func main() {
	// Define CLI flags
	var (
		fSystemFile = flag.String("system", "", "Path to system transactions CSV file (required)")
		fBankFiles  = flag.String("banks", "", "Comma-separated paths to bank statement CSV files (required)")
		fStartDate  = flag.String("start", "", "Start date for reconciliation (YYYY-MM-DD) (required)")
		fEndDate    = flag.String("end", "", "End date for reconciliation (YYYY-MM-DD) (required)")
		fOutputFile = flag.String("output", "", "Path to output file, only support txt at the moment. (optional)")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Reconciliation Service\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s -system=transactions.csv -banks=bank_bca.csv/bank_mandiri.csv -start=2024-01-01 -end=2024-12-31 -output=output.txt\n", os.Args[0])
	}

	flag.Parse()

	systemFile := *fSystemFile
	bankFiles := *fBankFiles
	startDate := *fStartDate
	endDate := *fEndDate
	outputFile := *fOutputFile

	// Validate required flags
	if systemFile == "" || bankFiles == "" || startDate == "" || endDate == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Parse dates
	start, err := time.Parse(DEFAULT_DATE_FORMAT, startDate)
	if err != nil {
		log.Fatalf("Invalid start date format: %v. Expected format: YYYY-MM-DD", err)
	}
	end, err := time.Parse(DEFAULT_DATE_FORMAT, endDate)
	if err != nil {
		log.Fatalf("Invalid end date format: %v. Expected format: YYYY-MM-DD", err)
	}

	// Set end date to end of day
	end = end.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	// Validate date range
	if end.Before(start) {
		log.Fatalf("End date must be after start date")
	}

	// Split bank files
	bankFileList := strings.Split(bankFiles, ",")
	for i, v := range bankFileList {
		bankFileList[i] = strings.TrimSpace(v)
	}

	// Validate files exist
	if err := validateFileExists(systemFile); err != nil {
		log.Fatalf("System transaction file error: %v", err)
	}
	for _, bankFile := range bankFileList {
		if err := validateFileExists(bankFile); err != nil {
			log.Fatalf("Bank statement file error: %v", err)
		}
	}

	// Run reconciliation
	fmt.Println("Starting reconciliation process...")
	fmt.Printf("System Transactions: %s\n", systemFile)
	fmt.Printf("Bank Statements: %s\n", strings.Join(bankFileList, ", "))
	fmt.Printf("Date Range: %s to %s\n", start.Format(DEFAULT_DATE_FORMAT), end.Format(DEFAULT_DATE_FORMAT))
	fmt.Printf("Output file: %s\n", outputFile)
	reconService := service.NewReconciliationService()

	input := service.ReconciliationInput{
		SystemTransactionFile: systemFile,
		BankStatementFiles:    bankFileList,
		StartDate:             start,
		EndDate:               end,
		OutputFile:            outputFile,
	}

	result, err := reconService.Reconcile(input)
	if err != nil {
		log.Fatalf("Reconciliation failed: %v", err)
	}

	// Print results
	result.Print()

	// Exit with additional info
	if result.TotalUnmatchedTransactions > 0 || result.TotalDiscrepancies.GreaterThan(decimal.Zero) {
		fmt.Println("\nReconciliation completed successfully - There are UNMATCHED transactions or discrepancies.")
	} else {
		fmt.Println("\nReconciliation completed successfully - All transactions MATCHED!")
	}
	os.Exit(0)
}

func validateFileExists(filePath string) error {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}
	if err != nil {
		return fmt.Errorf("error accessing file: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", filePath)
	}
	return nil
}
