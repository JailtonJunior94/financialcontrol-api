package application

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/shared"
)

// InvoiceSyncAdapter implements the TransactionSyncPort required by the
// invoicing module's InvoiceChangedHandler. It lives in the transactions
// module so that the invoicing layer never imports a concrete transactions
// repository directly.
type InvoiceSyncAdapter struct {
	transactionRepo    TransactionRepository
	transactionService TransactionAppService
}

// NewInvoiceSyncAdapter returns an adapter ready to synchronise transaction
// values when an invoice total changes.
func NewInvoiceSyncAdapter(
	repo TransactionRepository,
	svc TransactionAppService,
) *InvoiceSyncAdapter {
	return &InvoiceSyncAdapter{
		transactionRepo:    repo,
		transactionService: svc,
	}
}

// SyncTransactionWithInvoice finds the transaction for the given reference
// month and user, locates the item whose title contains cardDescription, and
// updates its value to match the invoice total.
func (a *InvoiceSyncAdapter) SyncTransactionWithInvoice(
	_ context.Context,
	invoiceID string,
	cardDescription string,
	userID string,
	referenceDate time.Time,
	total float64,
) error {
	t := shared.NewTime(shared.Time{Now: referenceDate})

	transaction, err := a.transactionRepo.GetTransactionByDate(t.StartDate(), t.EndDate(), userID)
	if err != nil {
		return fmt.Errorf("sync_invoice: fetching transaction for invoice %s: %w", invoiceID, err)
	}

	if transaction == nil {
		log.Printf("[transactions] sync_invoice: no transaction found for invoice %s on %s", invoiceID, referenceDate.Format("2006-01"))
		return nil
	}

	items, err := a.transactionRepo.GetItemByTransactionId(transaction.ID)
	if err != nil {
		return fmt.Errorf("sync_invoice: fetching items for transaction %s: %w", transaction.ID, err)
	}

	for _, item := range items {
		if strings.Contains(item.Title, cardDescription) {
			req := NewTransactionItemRequest(cardDescription, "OUTCOME", total)
			resp := a.transactionService.UpdateTransactionItem(transaction.ID, item.ID, transaction.UserId, req)
			if resp == nil || resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
				statusCode := 0
				if resp != nil {
					statusCode = resp.StatusCode
				}
				return fmt.Errorf("sync_invoice: update item %s for invoice %s failed with status %d", item.ID, invoiceID, statusCode)
			}
			log.Printf("[transactions] sync_invoice: updated item %s for invoice %s", item.ID, invoiceID)
			return nil
		}
	}

	log.Printf("[transactions] sync_invoice: no matching item found for card %q in transaction %s", cardDescription, transaction.ID)
	return nil
}
