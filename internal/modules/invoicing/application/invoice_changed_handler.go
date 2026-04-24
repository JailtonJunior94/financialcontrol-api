package invoicingapp

import (
	"context"
	"fmt"
	"log"
	"time"
)

// TransactionSyncPort is the port this handler calls to update a transaction
// when an invoice total changes. Defined on the consumer side so invoicing
// never imports a concrete transactions repository.
type TransactionSyncPort interface {
	SyncTransactionWithInvoice(ctx context.Context, invoiceID string, cardDescription string, userID string, referenceDate time.Time, total float64) error
}

// InvoiceChangedHandler reacts to the invoice_changed event. It recalculates
// the invoice totals and, when the invoice is flagged for transaction import,
// calls the transaction sync port to keep values consistent.
//
// This handler replaces the legacy invoiceChangedListener in
// internal/application/handlers and removes direct use of repositories
// from another module.
type InvoiceChangedHandler struct {
	invoiceRepo     InvoiceRepository
	transactionSync TransactionSyncPort
}

// NewInvoiceChangedHandler constructs the handler with its required ports.
func NewInvoiceChangedHandler(
	invoiceRepo InvoiceRepository,
	transactionSync TransactionSyncPort,
) *InvoiceChangedHandler {
	return &InvoiceChangedHandler{
		invoiceRepo:     invoiceRepo,
		transactionSync: transactionSync,
	}
}

// Handle processes the invoice_changed event for the given invoiceID.
// It recalculates totals and optionally propagates the new total to
// the linked transaction via the sync port.
func (h *InvoiceChangedHandler) Handle(ctx context.Context, invoiceID string) error {
	if invoiceID == "" {
		return fmt.Errorf("invoice_changed: invoiceID must not be empty")
	}

	invoice, err := h.invoiceRepo.GetInvoiceById(invoiceID)
	if err != nil {
		return fmt.Errorf("invoice_changed: fetching invoice %s: %w", invoiceID, err)
	}

	invoice.UpdatingValues()

	if _, err = h.invoiceRepo.UpdateInvoice(invoice); err != nil {
		return fmt.Errorf("invoice_changed: updating invoice %s: %w", invoiceID, err)
	}

	if !invoice.MarkImportTransactions {
		log.Printf("[invoicing] invoice_changed: invoice %s not flagged for transaction import", invoiceID)
		return nil
	}

	if err = h.transactionSync.SyncTransactionWithInvoice(
		ctx,
		invoiceID,
		invoice.Card.Description,
		invoice.Card.UserId,
		invoice.Date,
		invoice.Total,
	); err != nil {
		return fmt.Errorf("invoice_changed: syncing transaction for invoice %s: %w", invoiceID, err)
	}

	log.Printf("[invoicing] invoice_changed: processed successfully for invoice %s", invoiceID)
	return nil
}
