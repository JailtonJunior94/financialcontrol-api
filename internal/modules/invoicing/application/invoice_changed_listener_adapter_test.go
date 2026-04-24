package invoicingapp_test

import (
	"errors"
	"testing"

	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/entities"
	invoicingapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/invoicing/application"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListenerAdapterSetDataAndHandleCallsHandler(t *testing.T) {
	inv := newFlaggedInvoice()
	repo := &fakeInvoiceRepo{invoice: inv}
	sync := &fakeSyncPort{}
	handler := invoicingapp.NewInvoiceChangedHandler(repo, sync)
	adapter := invoicingapp.NewInvoiceChangedListenerAdapter(handler)

	adapter.SetData("inv-adapter-1")
	err := adapter.Handle()

	require.NoError(t, err)
	assert.True(t, repo.updated)
}

func TestListenerAdapterHandleIgnoresNonStringData(t *testing.T) {
	handler := invoicingapp.NewInvoiceChangedHandler(&fakeInvoiceRepo{}, &fakeSyncPort{})
	adapter := invoicingapp.NewInvoiceChangedListenerAdapter(handler)

	adapter.SetData(12345) // wrong type
	err := adapter.Handle()

	// Should return nil (logged and skipped) rather than panicking.
	require.NoError(t, err)
}

func TestListenerAdapterPropagatesHandlerError(t *testing.T) {
	repo := &fakeInvoiceRepo{getErr: errors.New("db fail")}
	handler := invoicingapp.NewInvoiceChangedHandler(repo, &fakeSyncPort{})
	adapter := invoicingapp.NewInvoiceChangedListenerAdapter(handler)

	adapter.SetData("inv-err")
	err := adapter.Handle()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "fetching invoice")
}

// helper to produce an invoice that skips the sync path
func newFlaggedInvoice() *entities.Invoice {
	inv := &entities.Invoice{MarkImportTransactions: false}
	return inv
}
