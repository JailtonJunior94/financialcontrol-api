package interfaces

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/src/domain/entities"
)

type ITransactionRepository interface {
	GetTransactions(userId string) (transactions []entities.Transaction, err error)
	GetTransactionById(id string, userId string) (transaction *entities.Transaction, err error)
	GetTransactionByDate(startDate, endDate time.Time, userId string) (transaction *entities.Transaction, err error)
	AddTransaction(t *entities.Transaction) (transaction *entities.Transaction, err error)
	UpdateTransaction(t *entities.Transaction) (transaction *entities.Transaction, err error)

	GetItemByTransactionId(transactionId string) (items []entities.TransactionItem, err error)
	GetTransactionItemsById(transactionId, id string) (transactionItem *entities.TransactionItem, err error)
	AddTransactionItem(t *entities.TransactionItem) (transactionItem *entities.TransactionItem, err error)
	AddRangeTransactionItems(t *entities.Transaction, ti []entities.TransactionItem) error
	UpdateTransactionItem(t *entities.TransactionItem) (transactionItem *entities.TransactionItem, err error)
}
