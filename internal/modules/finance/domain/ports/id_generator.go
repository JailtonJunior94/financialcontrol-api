package ports

import "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"

// IDGenerator abstracts UUID generation for deterministic testing in domain
// services that create multiple IDs in sequence (InstallmentSplitter, RefundFactory — B3.c).
type IDGenerator interface {
	NewTransactionID() vos.TransactionID
	NewInstallmentID() vos.InstallmentID
	NewInvoiceID() vos.InvoiceID
}
