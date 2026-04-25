package invoicingapp_test

import (
	"testing"
	"time"

	invoicingdomain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/invoicing/domain"
	platformevents "github.com/jailtonjunior94/financialcontrol-api/pkg/events"
	invoicingapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/invoicing/application"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeDispatcher struct {
	dispatched []platformevents.Event
}

func (f *fakeDispatcher) Dispatch(event platformevents.Event) error {
	f.dispatched = append(f.dispatched, event)
	return nil
}

func (f *fakeDispatcher) AddListener(_ string, _ platformevents.Listener) {}

func TestPublishInvoiceChangedReturnsErrorWhenInvoiceIDIsEmpty(t *testing.T) {
	pub := invoicingapp.NewInvoiceChangedEventPublisher(&fakeDispatcher{})
	err := pub.PublishInvoiceChanged(t.Context(), invoicingdomain.InvoiceChangedPayload{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "InvoiceID must not be empty")
}

func TestPublishInvoiceChangedDispatchesEventWithCorrectKeyAndPayload(t *testing.T) {
	dispatcher := &fakeDispatcher{}
	pub := invoicingapp.NewInvoiceChangedEventPublisher(dispatcher)

	payload := invoicingdomain.InvoiceChangedPayload{
		InvoiceID:       "inv-99",
		CardDescription: "Visa",
		UserID:          "user-1",
		ReferenceDate:   time.Now(),
		Total:           200.0,
	}

	err := pub.PublishInvoiceChanged(t.Context(), payload)

	require.NoError(t, err)
	require.Len(t, dispatcher.dispatched, 1)
	assert.Equal(t, "invoice_changed", dispatcher.dispatched[0].GetKey())
	got, ok := dispatcher.dispatched[0].GetData().(invoicingdomain.InvoiceChangedPayload)
	require.True(t, ok, "event data should be InvoiceChangedPayload")
	assert.Equal(t, "inv-99", got.InvoiceID)
	assert.Equal(t, "Visa", got.CardDescription)
}
