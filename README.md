# Overview

A simple Go application for reconciling internal system transactions with bank statements. 

## Features

- **Multi-Bank Support**: Reconcile transactions across multiple bank statement files (with same format)
- **Date Range Filtering**: Process transactions within specific time periods
- **Automatic Matching**: Matching transactions based on amount
- **Saving Result**: Saving result to a file

## Getting Started

### Prerequisites

- Go 1.23 or higher

### Installation

1. Clone the repository:
```bash
git clone https://github.com/firmannf/recon.git
```

2. Build the application:
```bash
make deps
make build
```

### Usage

Run the reconciliation service with the following command:

```bash
# Using binary
./bin/recon -system=<path-to-system-csv> \
            -banks=<path-to-bank-csv-1>,<path-to-bank-csv-2> \
            -start=<YYYY-MM-DD> \
            -end=<YYYY-MM-DD> \
            -output=output.txt

# Or use go run 
go run cmd/recon/main.go -system=<path-to-system-csv> \
            -banks=<path-to-bank-csv-1>,<path-to-bank-csv-2> \
            -start=<YYYY-MM-DD> \
            -end=<YYYY-MM-DD> \
            -output=output.txt  
```

#### CLI Options

- `-system`: Path to system transactions CSV file (required)
- `-banks`: Comma-separated paths to bank statement CSV files (required)
- `-start`: Start date for reconciliation in YYYY-MM-DD format (required)
- `-end`: End date for reconciliation (YYYY-MM-DD) (optional, defaults to start date)
- `-otuput`: Path to output file, only support txt at the moment. (optional)

## CSV File Formats

### System Transactions CSV

Format: `trxID,amount,type,transactionTime`

```csv
trxID,amount,type,transactionTime
TRX001,1000.00,CREDIT,2024-01-15 10:30:00
TRX002,500.50,DEBIT,2024-01-16 14:22:00
```

Fields:
- `trxID`: Unique transaction identifier
- `amount`: Transaction amount (positive number)
- `type`: Either `DEBIT` or `CREDIT`
- `transactionTime`: Date and time (supports multiple formats)

### Bank Statement CSV

Format: `unique_identifier,amount,date`

```csv
unique_identifier,amount,date
BCA-20240115-001,1000.00,2024-01-15
BCA-20240116-002,-500.50,2024-01-16
```

Fields:
- `unique_identifier`: Bank's unique transaction identifier
- `amount`: Transaction amount (negative for debits, positive for credits)
- `date`: Transaction date (supports multiple formats)

## Output

The service generates a comprehensive reconciliation report:

```
================================================================================
TRANSACTION RECONCILIATION SUMMARY
================================================================================

Reconciliation Parameters:
  System Transaction File: testdata/scenario1_all_matched_system.csv
  Bank Statement Files: testdata/scenario1_all_matched_bank_bca.csv, testdata/scenario1_all_matched_bank_bri.csv
  Date Range: 2024-01-15 to 2024-01-30

Reconciliation Results:
  Total Transactions Processed: 20 (System: 10 | Bank: 10)
  Total Matched Transactions: 8 pairs
  Total Unmatched Transactions: 4
  Total Discrepancies (Amount): Rp. 0.00

--------------------------------------------------------------------------------
UNMATCHED SYSTEM TRANSACTIONS: 1
--------------------------------------------------------------------------------
TrxID                Type       Transaction Time          Amount
TRX005               CREDIT     2024-01-20 11:00:00       Rp. 3000000.00
TRX006               CREDIT     2024-01-20 11:00:00       Rp. 4000000.00

--------------------------------------------------------------------------------
UNMATCHED BANK STATEMENTS: 2
--------------------------------------------------------------------------------

Bank: bank_mandiri (1 transactions)
Unique Identifier    Date                   Amount
MDR-05022024-999     2024-02-05        Rp. 1500000.00

Bank: bank_bri (1 transactions)
Unique Identifier    Date                   Amount
BRI-20240117-001     2024-01-17        Rp. 2499000.00
================================================================================
```
