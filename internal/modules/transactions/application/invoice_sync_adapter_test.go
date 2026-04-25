package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/pkg/persistence"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/web"
	transactionsapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/transactions/application"
	transactionsdomain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/transactions/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- fakes ---

type fakeTransactionRepo struct {
	transaction  *transactionsdomain.Transaction
	getByDateErr error
	items        []transactionsdomain.TransactionItem
	getItemsErr  error
}

func (f *fakeTransactionRepo) GetTransactionByDate(_, _ time.Time, _ string) (*transactionsdomain.Transaction, error) {
	return f.transaction, f.getByDateErr
}

func (f *fakeTransactionRepo) GetItemByTransactionId(_ string) ([]transactionsdomain.TransactionItem, error) {
	return f.items, f.getItemsErr
}

// no-op stubs
func (f *fakeTransactionRepo) GetTransactions(_ string) ([]transactionsdomain.Transaction, error) {
	return nil, nil
}
func (f *fakeTransactionRepo) GetTransactionById(_, _ string) (*transactionsdomain.Transaction, error) {
	return nil, nil
}
func (f *fakeTransactionRepo) AddTransaction(_ *transactionsdomain.Transaction) (*transactionsdomain.Transaction, error) {
	return nil, nil
}
func (f *fakeTransactionRepo) UpdateTransaction(_ *transactionsdomain.Transaction) (*transactionsdomain.Transaction, error) {
	return nil, nil
}
func (f *fakeTransactionRepo) GetTransactionItemsById(_, _ string) (*transactionsdomain.TransactionItem, error) {
	return nil, nil
}
func (f *fakeTransactionRepo) AddTransactionItem(_ *transactionsdomain.TransactionItem) (*transactionsdomain.TransactionItem, error) {
	return nil, nil
}
func (f *fakeTransactionRepo) AddRangeTransactionItems(_ *transactionsdomain.Transaction, _ []transactionsdomain.TransactionItem) error {
	return nil
}
func (f *fakeTransactionRepo) UpdateTransactionItem(_ *transactionsdomain.TransactionItem) (*transactionsdomain.TransactionItem, error) {
	return nil, nil
}
func (f *fakeTransactionRepo) FetchTransactionByDate(_ time.Time, _ string) (*persistence.TransactionQuery, error) {
	return nil, nil
}

type fakeTransactionService struct {
	updateCalled   bool
	updateResponse *web.HttpResponse
}

func (f *fakeTransactionService) Transactions(_ string) *web.HttpResponse { return nil }
func (f *fakeTransactionService) TransactionById(_, _ string) *web.HttpResponse {
	return web.Ok(nil)
}
func (f *fakeTransactionService) CreateTransaction(_ *transactionsapp.TransactionRequest, _ string) *web.HttpResponse {
	return nil
}
func (f *fakeTransactionService) CloneTransaction(_, _ string) *web.HttpResponse { return nil }
func (f *fakeTransactionService) TransactionItemById(_, _ string) *web.HttpResponse {
	return nil
}
func (f *fakeTransactionService) CreateTransactionItem(_ *transactionsapp.TransactionItemRequest, _, _ string) *web.HttpResponse {
	return nil
}
func (f *fakeTransactionService) UpdateTransactionItem(_, _, _ string, _ *transactionsapp.TransactionItemRequest) *web.HttpResponse {
	f.updateCalled = true
	if f.updateResponse != nil {
		return f.updateResponse
	}
	return web.Ok(nil)
}
func (f *fakeTransactionService) MarkAsPaidTransactionItem(_, _, _ string, _ *transactionsapp.TransactionMarkAsPaid) *web.HttpResponse {
	return nil
}
func (f *fakeTransactionService) RemoveTransactionItem(_, _, _ string) *web.HttpResponse {
	return nil
}

// --- tests ---

func TestSyncReturnsErrorWhenGetTransactionByDateFails(t *testing.T) {
	repo := &fakeTransactionRepo{getByDateErr: errors.New("db error")}
	adapter := transactionsapp.NewInvoiceSyncAdapter(repo, &fakeTransactionService{})
	err := adapter.SyncTransactionWithInvoice(context.Background(), "inv-1", "Card", "user-1", time.Now(), 100)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "fetching transaction")
}

func TestSyncReturnsNilWhenNoTransactionFound(t *testing.T) {
	repo := &fakeTransactionRepo{transaction: nil}
	adapter := transactionsapp.NewInvoiceSyncAdapter(repo, &fakeTransactionService{})
	err := adapter.SyncTransactionWithInvoice(context.Background(), "inv-1", "Card", "user-1", time.Now(), 100)
	require.NoError(t, err)
}

func TestSyncUpdatesMatchingItemByCardDescription(t *testing.T) {
	tx := &transactionsdomain.Transaction{}
	tx.ID = "tx-1"
	tx.UserId = "user-1"

	items := []transactionsdomain.TransactionItem{
		{Title: "Card Visa", Type: "OUTCOME", Value: 500},
	}
	items[0].ID = "item-1"
	items[0].TransactionId = "tx-1"

	repo := &fakeTransactionRepo{transaction: tx, items: items}
	svc := &fakeTransactionService{}
	adapter := transactionsapp.NewInvoiceSyncAdapter(repo, svc)

	err := adapter.SyncTransactionWithInvoice(context.Background(), "inv-1", "Visa", "user-1", time.Now(), 700)

	require.NoError(t, err)
	assert.True(t, svc.updateCalled, "UpdateTransactionItem should be called for matching card description")
}

func TestSyncDoesNotUpdateWhenNoItemMatchesCardDescription(t *testing.T) {
	tx := &transactionsdomain.Transaction{}
	tx.ID = "tx-2"

	items := []transactionsdomain.TransactionItem{
		{Title: "Grocery", Type: "OUTCOME", Value: 200},
	}

	repo := &fakeTransactionRepo{transaction: tx, items: items}
	svc := &fakeTransactionService{}
	adapter := transactionsapp.NewInvoiceSyncAdapter(repo, svc)

	err := adapter.SyncTransactionWithInvoice(context.Background(), "inv-2", "Visa", "user-1", time.Now(), 700)

	require.NoError(t, err)
	assert.False(t, svc.updateCalled, "UpdateTransactionItem should not be called when no item matches")
}

func TestSyncReturnsErrorWhenUpdateTransactionItemReturnsNonSuccessStatus(t *testing.T) {
	tx := &transactionsdomain.Transaction{}
	tx.ID = "tx-3"
	tx.UserId = "user-1"

	items := []transactionsdomain.TransactionItem{
		{Title: "Card Visa", Type: "OUTCOME", Value: 500},
	}
	items[0].ID = "item-3"
	items[0].TransactionId = "tx-3"

	repo := &fakeTransactionRepo{transaction: tx, items: items}
	svc := &fakeTransactionService{updateResponse: web.ServerError()}
	adapter := transactionsapp.NewInvoiceSyncAdapter(repo, svc)

	err := adapter.SyncTransactionWithInvoice(context.Background(), "inv-3", "Visa", "user-1", time.Now(), 900)

	require.Error(t, err, "SyncTransactionWithInvoice must return an error when UpdateTransactionItem fails")
	assert.Contains(t, err.Error(), "update item")
	assert.True(t, svc.updateCalled, "UpdateTransactionItem should have been called")
}

func TestSyncReturnsErrorWhenUpdateTransactionItemReturnsNilResponse(t *testing.T) {
	tx := &transactionsdomain.Transaction{}
	tx.ID = "tx-4"
	tx.UserId = "user-1"

	items := []transactionsdomain.TransactionItem{
		{Title: "Card Master", Type: "OUTCOME", Value: 300},
	}
	items[0].ID = "item-4"
	items[0].TransactionId = "tx-4"

	repo := &fakeTransactionRepo{transaction: tx, items: items}
	svc := &fakeTransactionService{updateResponse: &web.HttpResponse{StatusCode: 0, Data: nil}}
	adapter := transactionsapp.NewInvoiceSyncAdapter(repo, svc)

	err := adapter.SyncTransactionWithInvoice(context.Background(), "inv-4", "Master", "user-1", time.Now(), 400)

	require.Error(t, err, "SyncTransactionWithInvoice must return an error when UpdateTransactionItem returns status 0")
	assert.Contains(t, err.Error(), "update item")
}
