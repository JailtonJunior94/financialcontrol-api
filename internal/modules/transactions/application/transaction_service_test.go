package application_test

import (
	"errors"
	"testing"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos"
	appresponses "github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos/responses"
	transactionsapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/transactions/application"
	transactionsdomain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/transactions/domain"

	"github.com/stretchr/testify/require"
)

type transactionRepositoryStub struct {
	transaction           *transactionsdomain.Transaction
	transactionByDate     *transactionsdomain.Transaction
	items                 []transactionsdomain.TransactionItem
	item                  *transactionsdomain.TransactionItem
	addRangeTransaction   *transactionsdomain.Transaction
	addRangeItems         []transactionsdomain.TransactionItem
	updatedTransaction    *transactionsdomain.Transaction
	updatedItem           *transactionsdomain.TransactionItem
	getTransactionByIDErr error
	addRangeErr           error
	updateItemErr         error
	updateTransactionErr  error
}

func (s *transactionRepositoryStub) GetTransactions(string) ([]transactionsdomain.Transaction, error) {
	return nil, nil
}

func (s *transactionRepositoryStub) GetTransactionById(string, string) (*transactionsdomain.Transaction, error) {
	return s.transaction, s.getTransactionByIDErr
}

func (s *transactionRepositoryStub) GetTransactionByDate(time.Time, time.Time, string) (*transactionsdomain.Transaction, error) {
	return s.transactionByDate, nil
}

func (s *transactionRepositoryStub) AddTransaction(*transactionsdomain.Transaction) (*transactionsdomain.Transaction, error) {
	return nil, nil
}

func (s *transactionRepositoryStub) UpdateTransaction(transaction *transactionsdomain.Transaction) (*transactionsdomain.Transaction, error) {
	s.updatedTransaction = transaction
	return transaction, s.updateTransactionErr
}

func (s *transactionRepositoryStub) GetItemByTransactionId(string) ([]transactionsdomain.TransactionItem, error) {
	return s.items, nil
}

func (s *transactionRepositoryStub) GetTransactionItemsById(string, string) (*transactionsdomain.TransactionItem, error) {
	return s.item, nil
}

func (s *transactionRepositoryStub) AddTransactionItem(item *transactionsdomain.TransactionItem) (*transactionsdomain.TransactionItem, error) {
	return item, nil
}

func (s *transactionRepositoryStub) AddRangeTransactionItems(transaction *transactionsdomain.Transaction, items []transactionsdomain.TransactionItem) error {
	s.addRangeTransaction = transaction
	s.addRangeItems = items
	return s.addRangeErr
}

func (s *transactionRepositoryStub) UpdateTransactionItem(item *transactionsdomain.TransactionItem) (*transactionsdomain.TransactionItem, error) {
	s.updatedItem = item
	return item, s.updateItemErr
}

func (s *transactionRepositoryStub) FetchTransactionByDate(time.Time, string) (*dtos.TransactionQuery, error) {
	return nil, nil
}

func TestCloneTransactionCopiesItemsToNextMonth(t *testing.T) {
	now := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
	transaction := transactionsdomain.NewTransactionWithValues(now, "user-1", 350, 500, 150)
	transaction.ID = "tx-1"

	repository := &transactionRepositoryStub{
		transaction: transaction,
		items: []transactionsdomain.TransactionItem{
			{Title: "Salary", Type: "INCOME", Value: 500},
			{Title: "Rent", Type: "OUTCOME", Value: 150},
		},
	}

	service := transactionsapp.NewTransactionService(repository)
	response := service.CloneTransaction("tx-1", "user-1")

	require.Equal(t, appresponses.Ok(nil).StatusCode, response.StatusCode)
	require.NotNil(t, repository.addRangeTransaction)
	require.Len(t, repository.addRangeItems, 2)
	require.Equal(t, now.AddDate(0, 1, 0).Month(), repository.addRangeTransaction.Date.Month())
	require.Equal(t, repository.addRangeTransaction.ID, repository.addRangeItems[0].TransactionId)
	require.Equal(t, "Salary", repository.addRangeItems[0].Title)
	require.Equal(t, "Rent", repository.addRangeItems[1].Title)
}

func TestCloneTransactionReturnsServerErrorWhenPersistingCloneFails(t *testing.T) {
	transaction := transactionsdomain.NewTransactionWithValues(time.Now(), "user-1", 10, 20, 10)
	transaction.ID = "tx-2"

	repository := &transactionRepositoryStub{
		transaction: transaction,
		items: []transactionsdomain.TransactionItem{
			{Title: "Rent", Type: "OUTCOME", Value: 10},
		},
		addRangeErr: errors.New("write failed"),
	}

	service := transactionsapp.NewTransactionService(repository)
	response := service.CloneTransaction("tx-2", "user-1")

	require.Equal(t, appresponses.ServerError().StatusCode, response.StatusCode)
}

func TestMarkAsPaidTransactionItemRecalculatesTransactionTotals(t *testing.T) {
	transaction := transactionsdomain.NewTransactionWithValues(time.Now(), "user-1", 200, 500, 300)
	transaction.ID = "tx-3"
	item := transactionsdomain.NewTransactionItem("tx-3", "Rent", "OUTCOME", 300)
	item.ID = "item-1"

	repository := &transactionRepositoryStub{
		transaction: transaction,
		item:        item,
		items: []transactionsdomain.TransactionItem{
			*transactionsdomain.NewTransactionItem("tx-3", "Salary", "INCOME", 500),
			*item,
		},
	}

	service := transactionsapp.NewTransactionService(repository)
	response := service.MarkAsPaidTransactionItem("tx-3", "item-1", "user-1", &transactionsapp.TransactionMarkAsPaid{MarkAsPaid: true})

	require.Equal(t, appresponses.NoContent().StatusCode, response.StatusCode)
	require.NotNil(t, repository.updatedItem)
	require.True(t, repository.updatedItem.IsPaid)
	require.NotNil(t, repository.updatedTransaction)
	require.Equal(t, 200.0, repository.updatedTransaction.Total)
	require.Equal(t, 500.0, repository.updatedTransaction.Income)
	require.Equal(t, 300.0, repository.updatedTransaction.Outcome)
}

func TestTransactionDomainUpdatesTotalsFromItems(t *testing.T) {
	transaction := transactionsdomain.NewTransaction("user-1", time.Now())
	transaction.AddItems([]transactionsdomain.TransactionItem{
		*transactionsdomain.NewTransactionItem(transaction.ID, "Salary", "INCOME", 1000),
		*transactionsdomain.NewTransactionItem(transaction.ID, "Rent", "OUTCOME", 400),
		*transactionsdomain.NewTransactionItem(transaction.ID, "Groceries", "OUTCOME", 150),
	})

	transaction.UpdatingValues()

	require.Equal(t, 1000.0, transaction.Income)
	require.Equal(t, 550.0, transaction.Outcome)
	require.Equal(t, 450.0, transaction.Total)
}
