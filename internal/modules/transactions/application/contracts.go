package application

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/pkg/persistence"
	transactionsdomain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/transactions/domain"
)

type TransactionRepository interface {
	GetTransactions(userID string) ([]transactionsdomain.Transaction, error)
	GetTransactionById(id, userID string) (*transactionsdomain.Transaction, error)
	GetTransactionByDate(startDate, endDate time.Time, userID string) (*transactionsdomain.Transaction, error)
	AddTransaction(transaction *transactionsdomain.Transaction) (*transactionsdomain.Transaction, error)
	UpdateTransaction(transaction *transactionsdomain.Transaction) (*transactionsdomain.Transaction, error)

	GetItemByTransactionId(transactionID string) ([]transactionsdomain.TransactionItem, error)
	GetTransactionItemsById(transactionID, id string) (*transactionsdomain.TransactionItem, error)
	AddTransactionItem(item *transactionsdomain.TransactionItem) (*transactionsdomain.TransactionItem, error)
	AddRangeTransactionItems(transaction *transactionsdomain.Transaction, items []transactionsdomain.TransactionItem) error
	UpdateTransactionItem(item *transactionsdomain.TransactionItem) (*transactionsdomain.TransactionItem, error)

	FetchTransactionByDate(date time.Time, cardDescription string) (*persistence.TransactionQuery, error)
}

type TransactionAppService interface {
	Transactions(userID string) *HttpResponse
	TransactionById(id, userID string) *HttpResponse
	CreateTransaction(request *TransactionRequest, userID string) *HttpResponse
	CloneTransaction(id, userID string) *HttpResponse

	TransactionItemById(transactionID, id string) *HttpResponse
	CreateTransactionItem(request *TransactionItemRequest, transactionID, userID string) *HttpResponse
	UpdateTransactionItem(transactionID, id, userID string, request *TransactionItemRequest) *HttpResponse
	MarkAsPaidTransactionItem(transactionID, id, userID string, request *TransactionMarkAsPaid) *HttpResponse
	RemoveTransactionItem(transactionID, id, userID string) *HttpResponse
}
