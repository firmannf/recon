package parser

const (
	// CSV column counts
	transactionColumnCount   = 4 // trxID, amount, type, transactionTime
	bankStatementColumnCount = 3 // unique_identifier, amount, date
	headerRowCount           = 1

	// Transaction CSV column indices
	transactionColTrxID           = 0
	transactionColAmount          = 1
	transactionColType            = 2
	transactionColTransactionTime = 3

	// Bank statement CSV column indices
	bankStatementColUniqueIdentifier = 0
	bankStatementColAmount           = 1
	bankStatementColDate             = 2
)
