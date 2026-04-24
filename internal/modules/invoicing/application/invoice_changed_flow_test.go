package invoicingapp_test

import (
	"context"
	"testing"

	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/entities"
	platformevents "github.com/jailtonjunior94/financialcontrol-api/internal/platform/events"
	invoicingapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/invoicing/application"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInvoiceChangedFlowPublishesAndConsumesEventThroughDispatcher(t *testing.T) {
	invoice := &entities.Invoice{
		MarkImportTransactions: true,
		Total:                  175.5,
	}
	invoice.ID = "invoice-123"
	invoice.Card.Description = "Visa"
	invoice.Card.UserId = "user-123"

	repository := &fakeInvoiceRepo{invoice: invoice}
	syncPort := &fakeSyncPort{}
	dispatcher := platformevents.NewDispatcher()
	publisher := invoicingapp.NewInvoiceChangedEventPublisher(dispatcher)
	handler := invoicingapp.NewInvoiceChangedHandler(repository, syncPort)
	dispatcher.AddListener("invoice_changed", invoicingapp.NewInvoiceChangedListenerAdapter(handler))

	err := publisher.PublishInvoiceChanged(context.Background(), invoice.ID)

	require.NoError(t, err)
	assert.True(t, repository.updated)
	assert.True(t, syncPort.called)
}
