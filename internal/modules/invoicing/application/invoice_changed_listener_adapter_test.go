package invoicingapp_test

import (
	"errors"
	"testing"
	"time"

	invoicingdomain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/invoicing/domain"
	invoicingapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/invoicing/application"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListenerAdapterSetDataAndHandleCallsHandler(t *testing.T) {
	sync := &fakeSyncPort{}
	handler := invoicingapp.NewInvoiceChangedHandler(sync)
	adapter := invoicingapp.NewInvoiceChangedListenerAdapter(handler)

	payload := invoicingdomain.InvoiceChangedPayload{
		InvoiceID:       "inv-adapter-1",
		CardDescription: "Visa",
		UserID:          "user-1",
		ReferenceDate:   time.Now(),
		Total:           50.0,
	}

	adapter.SetData(payload)
	err := adapter.Handle()

	require.NoError(t, err)
	assert.True(t, sync.called)
}

func TestListenerAdapterHandleIgnoresUnexpectedDataType(t *testing.T) {
	handler := invoicingapp.NewInvoiceChangedHandler(&fakeSyncPort{})
	adapter := invoicingapp.NewInvoiceChangedListenerAdapter(handler)

	adapter.SetData(12345) // wrong type — must be InvoiceChangedPayload
	err := adapter.Handle()

	// Should return nil (logged and skipped) rather than panicking.
	require.NoError(t, err)
}

func TestListenerAdapterPropagatesHandlerError(t *testing.T) {
	sync := &fakeSyncPort{err: errors.New("sync fail")}
	handler := invoicingapp.NewInvoiceChangedHandler(sync)
	adapter := invoicingapp.NewInvoiceChangedListenerAdapter(handler)

	payload := invoicingdomain.InvoiceChangedPayload{
		InvoiceID: "inv-err",
	}

	adapter.SetData(payload)
	err := adapter.Handle()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "syncing transaction")
}
