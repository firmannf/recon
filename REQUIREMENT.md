# Reconciliation Service - Requirements

## Assumptions

### Given Assumption
- Both system transactions and bank statements are provided as separate CSV files
- Discrepancies only occur in amount

### Additional Assumption
- Each bank provides statements in standardized CSV format and the data already sorted
- System transaction data is in standardized CSV format and the data already sorted
- All date are using UTC+7
- All amount are using Indonesian Rupiah and only 2 max decimal points
- One CSV can contains multiple date transactions
- Transaction only has one-to-one relationship
- Transaction IDs between system and bank cannot be relied upon for matching
- Transaction matching is based on amount since
- Transactions are reconciled within the same calendar date unless later introduced function for handling cutoff period time
- Transaction type (CREDIT, DEBIT) in system transaction data are the same with bank's point of view (Credit +, Debit -)
- A discrepancy is defined as an amount difference between matched transactions, however since amount is used for matching, this value should be zero unless a tolerance-based strategy is later introduced
- System handles multiple banks in a single reconciliation run
  
---

## User Stories
### As a finance officer, I want to upload internal data and bank statement data, So that I can proceed the reconciliation process between those uploaded data  
```
GIVEN system transaction CSV file as internal data with format "trxID,amount,type,transactionTime"
WHEN I provide the correct file path to the system
THEN the system should accept the file
    AND parse the file successfully
```

```
GIVEN multiple bank statement CSV files as external data with format "unique_identifier,amount,date"
WHEN I provide comma-separated file paths to the system
THEN the system should accept all bank statement files
    AND the system should parse each file successfully
```

```
GIVEN a CSV file with dates in supported format YYYY-MM-DD, DD/MM/YYYY, YYYY-MM-DD HH:MM:SS, DD/MM/YYYY HH:MM:SS
WHEN the system parses the transaction time field
THEN the system should correctly interpret the date
```

```
GIVEN a file path that does not exist
WHEN the system attempts to open the file
THEN the system should detect the file access error
    AND the system should display a clear error message
```

```
GIVEN a file path with non CSV extension
WHEN the system attempts to open the file
THEN the system should not proceed with reconciliation
    AND the system should display a clear error message
```

```
GIVEN a CSV file with invalid format (different column, unsupported date format, non-numbers amount)
WHEN the system attempts to parse the file
THEN the system should not proceed with reconciliation
    AND the system should display a clear error message
```

### As a finance officer, I want to reconcile internal data with bank statement data, So that I can check if there is any amount discrepancy  
```
GIVEN a system transaction data with type "CREDIT", date "2024-01-15", amount "1000.00"
    AND a bank statement data with amount "1000.00", date "2024-01-15"
WHEN the system performs matching
THEN the transactions should be matched
    AND the match should be based on date and amount
```

```
GIVEN a bank statement data with amount "-500.50"
WHEN the system interprets the transaction type
THEN the system should classify it as "DEBIT"
    AND the absolute amount should be "500.50"
```

```
GIVEN a bank statement data with amount "1000.00"
WHEN the system interprets the transaction type
THEN the system should classify it as "CREDIT"
    AND the amount should remain "1000.00"
```

```
GIVEN a system transaction data TRX001 with amount "1000.00" on "2024-01-15"
    AND bank statements BANK_BCA_001 with amount "1000.00" on "2024-01-16" (one day difference)
WHEN the system performs matching
THEN TRX001 should unmatch with BANK_BCA_001
```

```
GIVEN a system transaction data TRX001 with amount "1000.00" on "2024-01-15"
    AND two bank statements BANK_BCA_001 and BANK_BCA_002 with amount "1000.00" on "2024-01-15"
WHEN the system performs matching
THEN TRX001 should match with BANK_BCA_001 (first encountered)
    AND BANK_BCA_002 should remain unmatched
```

```
GIVEN 10 transaction in system transaction data and 9 transactions in bank statement in the date range
WHEN the reconciliation completes
THEN the report should display "Total Transactions Processed: 10"
```

```
GIVEN 10 transactions were matched successfully
WHEN the report is generated
THEN the report should display "Total Matched Transactions: 10"
```

```
GIVEN 2 unmatched system transactions
    AND 3 unmatched bank statements
WHEN the report is generated
THEN the report should display "Total Unmatched Transactions: 5"
```

```
GIVEN a system transaction data TRX001 with amount "3000.00"
    AND no matching bank statement exists
WHEN the reconciliation completes
THEN TRX001 should be identified as unmatched
    AND it should appear in the "Unmatched System Transactions" section
```

```
GIVEN a bank statement BANK_BCA_001 with amount "1500.00"
    AND no matching system transaction exists
WHEN the reconciliation completes
THEN BANK_BCA_001 should be identified as unmatched
    AND it should appear in the "Unmatched Bank Statements" section
```

```
GIVEN unmatched transactions from both sources
WHEN the system generates the report
THEN unmatched system transactions should be in a separate section
    AND unmatched bank statements should be in a separate section
    AND each section should be clearly labeled with its source
```

```
GIVEN an unmatched system transaction:
  | TrxID  | Amount  | Type   | Time                |
  | TRX001 | 1000.00 | CREDIT | 2024-01-15 10:30:00 |
WHEN the report is generated
THEN the report should include a detailed table
    AND the table should show TrxID, Amount, Type, and Transaction Time
```

```
GIVEN unmatched bank statements from "bank_bca" and "bank_mandiri"
WHEN the report is generated
THEN statements should be grouped under bank headers
    AND each group should show "Bank: <bank_name> (N transactions)"
    AND the table should show Unique Identifier, Amount, and Date
```

```
GIVEN all matched transactions
WHEN the report is generated
THEN the report should display "Total Discrepancies: 0"
```

```
GIVEN an unmatched system transaction
WHEN the report is generated
THEN the report should display "Total Discrepancies: 0"
```

```
GIVEN an unmatched bank statements
WHEN the report is generated
THEN the report should display "Total Discrepancies: 0"
```

### As a finance officer, I want to reconcile transactions with specific date range only, So that I can check discrepancy more specific
```
GIVEN I want to reconcile transactions from a specific date
WHEN I provide start date and end date in YYYY-MM-DD format
THEN the system should accept the start date parameter
    AND the system should parse it correctly
```

```
GIVEN start date is "2024-12-31"
    AND end date is "2024-01-01"
WHEN the system validates the date range
THEN the system should reject the date range
    AND the system should display a clear error message
```

```
GIVEN start date is "2024-12-31"
    AND end date is empty
WHEN the system validates the date range
THEN the end date should be same with start date
```

```
GIVEN start date is "2024-01-01"
    AND end date is "2024-01-31"
    AND a transaction exists on "2024-01-01" at "10:30:00"
    AND a transaction exists on "2024-01-31" at "23:59:59"
WHEN the system filters transactions
THEN both boundary transactions should be included
```

```
GIVEN start date is "2024-01-01"
    AND end date is "2024-01-31"
    AND a transaction exists on "2023-12-31"
    AND a transaction exists on "2024-02-01"
WHEN the system filters transactions
    THEN both transactions should be excluded from reconciliation
```

### As a finance officer, I want to save reconciliation result, So that I can keep the reconciliation result as a proof
```
GIVEN I want to save reconciliation results
WHEN I provide "-output=results.txt" parameter
THEN the system should accept the output path
    AND the system should write it to the results.txt file in the current directory
```

```
GIVEN I do the reconciliation
WHEN I don't provide "-output=results.txt" parameter
THEN the system only log the results in the console
```

---

## Non-Functional Requirements
- System can be run in console application and input files provided via arguments
- Runs on macOS, Linux, and Windows
- Requires Go 1.23 or higher
- No external system dependencies
- Single binary deployment

---

## Future Improvements
  