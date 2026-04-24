package requests

import transactionsapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/transactions/application"

type TransactionItemRequest = transactionsapp.TransactionItemRequest

func NewTransactionItemRequest(title, transactionType string, value float64) *TransactionItemRequest {
	return transactionsapp.NewTransactionItemRequest(title, transactionType, value)
}
