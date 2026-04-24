package invoicingapp_test

import (
	"context"
	"testing"

	platformevents "github.com/jailtonjunior94/financialcontrol-api/internal/platform/events"
	invoicingapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/invoicing/application"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeDispatcher struct {
	dispatched []platformevents.Event
}

func (f *fakeDispatcher) Dispatch(event platformevents.Event) {
	f.dispatched = append(f.dispatched, event)
}

func TestPublishInvoiceChangedReturnsErrorWhenInvoiceIDIsEmpty(t *testing.T) {
	pub := invoicingapp.NewInvoiceChangedEventPublisher(&fakeDispatcher{})
	err := pub.PublishInvoiceChanged(context.Background(), "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invoiceID must not be empty")
}

func TestPublishInvoiceChangedDispatchesEventWithCorrectKey(t *testing.T) {
	dispatcher := &fakeDispatcher{}
	pub := invoicingapp.NewInvoiceChangedEventPublisher(dispatcher)

	err := pub.PublishInvoiceChanged(context.Background(), "inv-99")

	require.NoError(t, err)
	require.Len(t, dispatcher.dispatched, 1)
	assert.Equal(t, "invoice_changed", dispatcher.dispatched[0].GetKey())
	assert.Equal(t, "inv-99", dispatcher.dispatched[0].GetData())
}
