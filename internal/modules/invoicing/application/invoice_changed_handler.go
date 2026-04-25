package invoicingapp

import (
	"context"
	"fmt"
	"log"
	"time"

	invoicingdomain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/invoicing/domain"
)

// TransactionSyncPort is the port this handler calls to update a transaction
// when an invoice total changes. Defined on the consumer side so invoicing
// never imports a concrete transactions repository.
type TransactionSyncPort interface {
	SyncTransactionWithInvoice(ctx context.Context, invoiceID string, cardDescription string, userID string, referenceDate time.Time, total float64) error
}

// InvoiceChangedHandler reacts to the invoice_changed event. Because the
// payload is auto-contained (ADR-003), no database fetch is needed — all data
// required for the sync is carried in the payload itself.
type InvoiceChangedHandler struct {
	transactionSync TransactionSyncPort
}

// NewInvoiceChangedHandler constructs the handler with its required port.
func NewInvoiceChangedHandler(transactionSync TransactionSyncPort) *InvoiceChangedHandler {
	return &InvoiceChangedHandler{transactionSync: transactionSync}
}

// Handle processes the invoice_changed event payload. It calls the transaction
// sync port to keep values consistent with the updated invoice total.
func (h *InvoiceChangedHandler) Handle(ctx context.Context, payload invoicingdomain.InvoiceChangedPayload) error {
	if payload.InvoiceID == "" {
		return fmt.Errorf("invoice_changed: InvoiceID must not be empty")
	}

	if err := h.transactionSync.SyncTransactionWithInvoice(
		ctx,
		payload.InvoiceID,
		payload.CardDescription,
		payload.UserID,
		payload.ReferenceDate,
		payload.Total,
	); err != nil {
		return fmt.Errorf("invoice_changed: syncing transaction for invoice %s: %w", payload.InvoiceID, err)
	}

	log.Printf("[invoicing] invoice_changed: processed successfully for invoice %s", payload.InvoiceID)
	return nil
}
