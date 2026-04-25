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

func TestInvoiceChangedFlowPublishesAndConsumesEventThroughDispatcher(t *testing.T) {
	payload := invoicingdomain.InvoiceChangedPayload{
		InvoiceID:              "invoice-123",
		CardDescription:        "Visa",
		UserID:                 "user-123",
		ReferenceDate:          time.Now(),
		Total:                  175.5,
		MarkImportTransactions: true,
	}

	syncPort := &fakeSyncPort{}
	dispatcher := platformevents.NewInProcessDispatcher()
	publisher := invoicingapp.NewInvoiceChangedEventPublisher(dispatcher)
	handler := invoicingapp.NewInvoiceChangedHandler(syncPort)
	dispatcher.AddListener("invoice_changed", invoicingapp.NewInvoiceChangedListenerAdapter(handler))

	err := publisher.PublishInvoiceChanged(t.Context(), payload)

	require.NoError(t, err)
	assert.True(t, syncPort.called)
}
