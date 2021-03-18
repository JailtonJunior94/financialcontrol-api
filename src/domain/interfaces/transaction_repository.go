package interfaces

import "github.com/jailtonjunior94/financialcontrol-api/src/domain/entities"

type ITransactionRepository interface {
	AddTransaction(t *entities.Transaction) (transaction *entities.Transaction, err error)
	AddTransactionItem(t *entities.TransactionItem) (transactionItem *entities.TransactionItem, err error)
	GetTransactions(userId string) (transactions []entities.Transaction, err error)
	GetTransactionById(id string, userId string) (transaction *entities.Transaction, err error)
	GetItemByTransactionId(transactionId string) (items []entities.TransactionItem, err error)
}
