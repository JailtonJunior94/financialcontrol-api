package ports

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

// InstallmentRepository is the persistence port for Installment entities.
type InstallmentRepository interface {
	AddBatch(ctx context.Context, items []*entities.Installment) error
	UpdateBatch(ctx context.Context, items []*entities.Installment) error
	SoftDeleteByTransaction(ctx context.Context, transactionID vos.TransactionID, at time.Time) error
	ListByInvoice(ctx context.Context, invoiceID vos.InvoiceID) ([]entities.Installment, error)
	ListByTransaction(ctx context.Context, transactionID vos.TransactionID) ([]entities.Installment, error)
	HasClosedOrPaidForTransaction(ctx context.Context, transactionID vos.TransactionID) (bool, error)
	SumByMonthCompetence(ctx context.Context, userID identityvo.UserID, period vos.Period) (vos.Money, error)
}
