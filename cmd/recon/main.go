package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/shopspring/decimal"

	"github.com/firmannf/recon/internal/models"
	"github.com/firmannf/recon/internal/service"
)

const (
	DEFAULT_DATE_FORMAT = "2006-01-02"
)

type ReconciliationParams struct {
	SystemFile string
	BankFiles  string
	StartDate  string
	EndDate    string
	OutputFile string
}

func main() {
	// Define CLI flags
	var (
		fSystemFile = flag.String("system", "", "Path to system transactions CSV file (required)")
		fBankFiles  = flag.String("banks", "", "Comma-separated paths to bank statement CSV files (required)")
		fStartDate  = flag.String("start", "", "Start date for reconciliation (YYYY-MM-DD)in UTC+7 (required)")
		fEndDate    = flag.String("end", "", "End date for reconciliation (YYYY-MM-DD) in UTC+7 (optional, defaults to start date)")
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

	params := ReconciliationParams{
		SystemFile: *fSystemFile,
		BankFiles:  *fBankFiles,
		StartDate:  *fStartDate,
		EndDate:    *fEndDate,
		OutputFile: *fOutputFile,
	}
	// Validate required flags
	if params.SystemFile == "" || params.BankFiles == "" || params.StartDate == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Load timezone for parsing (use UTC+7 to match parser behavior)
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.FixedZone("UTC+7", 7*60*60)
	}

	// Parse dates
	start, err := time.ParseInLocation(DEFAULT_DATE_FORMAT, params.StartDate, loc)
	if err != nil {
		log.Fatalf("Invalid start date format: %v. Expected format: YYYY-MM-DD", err)
	}

	var end time.Time
	if params.EndDate != "" {
		end, err = time.ParseInLocation(DEFAULT_DATE_FORMAT, params.EndDate, loc)
		if err != nil {
			log.Fatalf("Invalid end date format: %v. Expected format: YYYY-MM-DD", err)
		}
		// Set end date to end of day
		end = end.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

		// Validate date range
		if start.After(end) {
			log.Fatalf("Start date must not be after end date")
		}
	} else {
		// If no end date provided, set to end of start day
		end = start.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	}

	// Split bank files
	bankFileList := strings.Split(params.BankFiles, ",")
	for i, v := range bankFileList {
		bankFileList[i] = strings.TrimSpace(v)
	}

	// Validate files exist
	if err := validateFileExists(params.SystemFile); err != nil {
		log.Fatalf("System transaction file error: %v", err)
	}
	for _, bankFile := range bankFileList {
		if err := validateFileExists(bankFile); err != nil {
			log.Fatalf("Bank statement file error: %v", err)
		}
	}

	// Run reconciliation
	fmt.Println("Starting reconciliation process...")
	fmt.Printf("System Transactions: %s\n", params.SystemFile)
	fmt.Printf("Bank Statements: %s\n", strings.Join(bankFileList, ", "))
	fmt.Printf("Date Range: %s to %s\n", start.Format(DEFAULT_DATE_FORMAT), end.Format(DEFAULT_DATE_FORMAT))
	fmt.Printf("Output file: %s\n", params.OutputFile)
	reconService := service.NewReconciliationService()

	input := service.ReconciliationInput{
		SystemTransactionFile: params.SystemFile,
		BankStatementFiles:    bankFileList,
		StartDate:             start,
		EndDate:               end,
		OutputFile:            params.OutputFile,
		MatchStrategy:         service.NewExactMatchStrategy(),
	}

	result, err := reconService.Reconcile(input)
	if err != nil {
		log.Fatalf("Reconciliation failed: %v", err)
	}

	// Print results
	printResult(result, params)

	// Save to output file if specified
	if params.OutputFile != "" {
		if err := writeResultToFile(result, params.OutputFile, params); err != nil {
			log.Fatalf("Failed to write output file: %v", err)
		}
		fmt.Printf("\nResults saved to: %s\n", params.OutputFile)
	}

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

func printResult(result *models.ReconciliationResult, params ReconciliationParams) {
	formatResult(os.Stdout, result, params)
}

func writeResultToFile(result *models.ReconciliationResult, filepath string, params ReconciliationParams) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	formatResult(file, result, params)
	return nil
}

func formatResult(w io.Writer, result *models.ReconciliationResult, params ReconciliationParams) {
	fmt.Fprintln(w, "\n"+strings.Repeat("=", 80))
	fmt.Fprintln(w, "TRANSACTION RECONCILIATION SUMMARY")
	fmt.Fprintln(w, strings.Repeat("=", 80))

	fmt.Fprintln(w, "\nReconciliation Parameters:")
	fmt.Fprintf(w, "  System Transaction File: %s\n", params.SystemFile)
	fmt.Fprintf(w, "  Bank Statement Files: %s\n", params.BankFiles)
	fmt.Fprintf(w, "  Date Range: %s to %s\n", params.StartDate, params.EndDate)

	fmt.Fprintln(w, "\nReconciliation Results:")
	fmt.Fprintf(w, "  Total Transactions Processed: %d\n", result.TotalTransactionsProcessed)
	fmt.Fprintf(w, "  Total Matched Transactions: %d\n", result.TotalMatchedTransactions)
	fmt.Fprintf(w, "  Total Unmatched Transactions: %d\n", result.TotalUnmatchedTransactions)
	fmt.Fprintf(w, "  Total Discrepancies (Amount): Rp. %s\n", result.TotalDiscrepancies)

	// Write unmatched system transactions
	if len(result.UnmatchedSystemTransactions) > 0 {
		fmt.Fprintln(w, "\n"+strings.Repeat("-", 80))
		fmt.Fprintf(w, "UNMATCHED SYSTEM TRANSACTIONS: %d\n", len(result.UnmatchedSystemTransactions))
		fmt.Fprintln(w, strings.Repeat("-", 80))
		fmt.Fprintf(w, "%-20s %-10s %-25s %20s \n", "TrxID", "Type", "Transaction Time", "Amount")
		for _, trx := range result.UnmatchedSystemTransactions {
			fmt.Fprintf(w, "%-20s %-10s %-25s %20s\n", trx.TrxID, trx.Type, trx.TransactionTime.Format("2006-01-02 15:04:05"), fmt.Sprintf("Rp. %v", trx.Amount.StringFixed(2)))
		}
	}

	// Write unmatched bank statements grouped by bank
	if len(result.UnmatchedBankStatementLines) > 0 {
		totalUnmatchedBank := 0
		for _, statements := range result.UnmatchedBankStatementLines {
			totalUnmatchedBank += len(statements)
		}

		fmt.Fprintln(w, "\n"+strings.Repeat("-", 80))
		fmt.Fprintf(w, "UNMATCHED BANK STATEMENTS: %d\n", totalUnmatchedBank)
		fmt.Fprintln(w, strings.Repeat("-", 80))

		for bankName, statements := range result.UnmatchedBankStatementLines {
			fmt.Fprintf(w, "\nBank: %s (%d transactions)\n", bankName, len(statements))
			fmt.Fprintf(w, "%-20s %-10s %20s\n", "Unique Identifier", "Date", "Amount")
			for _, stmt := range statements {
				fmt.Fprintf(w, "%-20s %-10s %20s\n", stmt.UniqueIdentifier, stmt.Date.Format("2006-01-02"), fmt.Sprintf("Rp. %v", stmt.Amount.StringFixed(2)))
			}
		}
	}

	fmt.Fprintln(w, "\n"+strings.Repeat("=", 80))
}
