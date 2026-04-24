package invoicingapp_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos"
	invoicingapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/invoicing/application"

	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/entities"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- fakes ---

type fakeInvoiceRepo struct {
	invoice *entities.Invoice
	getErr  error
	updErr  error
	updated bool
}

func (f *fakeInvoiceRepo) GetInvoiceById(_ string) (*entities.Invoice, error) {
	return f.invoice, f.getErr
}

func (f *fakeInvoiceRepo) UpdateInvoice(inv *entities.Invoice) (*entities.Invoice, error) {
	f.updated = true
	return inv, f.updErr
}

// Satisfy the full IInvoiceRepository interface with no-op stubs.
func (f *fakeInvoiceRepo) DeleteInvoiceItem(_ int64) error                     { return nil }
func (f *fakeInvoiceRepo) UpdateManyInvoices(_ []*entities.Invoice) error      { return nil }
func (f *fakeInvoiceRepo) GetLastInvoiceControl() (int64, error)               { return 0, nil }
func (f *fakeInvoiceRepo) AddManyInvoiceItems(_ []*entities.InvoiceItem) error { return nil }
func (f *fakeInvoiceRepo) GetInvoiceItemById(_ string) (*entities.InvoiceItem, error) {
	return nil, nil
}
func (f *fakeInvoiceRepo) AddInvoice(_ *entities.Invoice) (*entities.Invoice, error) { return nil, nil }
func (f *fakeInvoiceRepo) GetInvoiceByCardId(_, _ string) ([]entities.Invoice, error) {
	return nil, nil
}
func (f *fakeInvoiceRepo) AddInvoiceItem(_ *entities.InvoiceItem) (*entities.InvoiceItem, error) {
	return nil, nil
}
func (f *fakeInvoiceRepo) GetInvoiceItemByInvoiceControl(_ int64) ([]*entities.InvoiceItem, error) {
	return nil, nil
}
func (f *fakeInvoiceRepo) GetInvoiceByDate(_, _ time.Time, _ string) (*entities.Invoice, error) {
	return nil, nil
}
func (f *fakeInvoiceRepo) GetInvoiceItemByInvoiceId(_, _, _ string) ([]entities.InvoiceItem, error) {
	return nil, nil
}
func (f *fakeInvoiceRepo) GetInvoicesCategories(_, _ time.Time, _ string) ([]entities.InvoiceCategories, error) {
	return nil, nil
}
func (f *fakeInvoiceRepo) FetchInvoiceByCard(_ string) ([]dtos.InvoiceQuery, error) {
	return nil, nil
}
func (f *fakeInvoiceRepo) GetInvoices(_ time.Time) (*dtos.InvoiceRead, error) { return nil, nil }

type fakeSyncPort struct {
	called bool
	err    error
}

func (f *fakeSyncPort) SyncTransactionWithInvoice(_ context.Context, _, _, _ string, _ time.Time, _ float64) error {
	f.called = true
	return f.err
}

// --- tests ---

func TestHandleReturnsErrorWhenInvoiceIDIsEmpty(t *testing.T) {
	handler := invoicingapp.NewInvoiceChangedHandler(&fakeInvoiceRepo{}, &fakeSyncPort{})
	err := handler.Handle(context.Background(), "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invoiceID must not be empty")
}

func TestHandleReturnsErrorWhenGetInvoiceFails(t *testing.T) {
	repo := &fakeInvoiceRepo{getErr: errors.New("db error")}
	handler := invoicingapp.NewInvoiceChangedHandler(repo, &fakeSyncPort{})
	err := handler.Handle(context.Background(), "inv-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "fetching invoice")
}

func TestHandleReturnsErrorWhenUpdateInvoiceFails(t *testing.T) {
	repo := &fakeInvoiceRepo{
		invoice: &entities.Invoice{},
		updErr:  errors.New("update error"),
	}
	handler := invoicingapp.NewInvoiceChangedHandler(repo, &fakeSyncPort{})
	err := handler.Handle(context.Background(), "inv-2")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "updating invoice")
}

func TestHandleDoesNotCallSyncWhenNotFlaggedForImport(t *testing.T) {
	inv := &entities.Invoice{MarkImportTransactions: false}
	repo := &fakeInvoiceRepo{invoice: inv}
	sync := &fakeSyncPort{}
	handler := invoicingapp.NewInvoiceChangedHandler(repo, sync)

	err := handler.Handle(context.Background(), "inv-3")

	require.NoError(t, err)
	assert.True(t, repo.updated)
	assert.False(t, sync.called, "sync should not be called when MarkImportTransactions is false")
}

func TestHandleCallsSyncWhenFlaggedForImport(t *testing.T) {
	inv := &entities.Invoice{MarkImportTransactions: true}
	inv.Card.Description = "Card X"
	inv.Card.UserId = "user-1"
	repo := &fakeInvoiceRepo{invoice: inv}
	sync := &fakeSyncPort{}
	handler := invoicingapp.NewInvoiceChangedHandler(repo, sync)

	err := handler.Handle(context.Background(), "inv-4")

	require.NoError(t, err)
	assert.True(t, repo.updated)
	assert.True(t, sync.called, "sync should be called when MarkImportTransactions is true")
}

func TestHandleReturnsErrorWhenSyncFails(t *testing.T) {
	inv := &entities.Invoice{MarkImportTransactions: true}
	repo := &fakeInvoiceRepo{invoice: inv}
	sync := &fakeSyncPort{err: errors.New("sync error")}
	handler := invoicingapp.NewInvoiceChangedHandler(repo, sync)

	err := handler.Handle(context.Background(), "inv-5")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "syncing transaction")
}
