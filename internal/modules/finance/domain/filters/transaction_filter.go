package filters

import (
	"time"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

// TransactionFilter is the first-class collection that carries all criteria for
// listing transactions (RF-19, Object Calisthenics #4).
type TransactionFilter struct {
	UserID              identityvo.UserID
	TransactionType     *vos.TransactionType
	PaymentMethod       *vos.PaymentMethod
	CardID              *vos.CardID
	CategoryID          *vos.CategoryID
	InvoiceStatus       vos.InvoiceStatusFilter
	DescriptionContains string
	From                *time.Time
	To                  *time.Time
	Pagination          vos.Pagination
}

// NewTransactionFilter constructs a validated TransactionFilter.
// Returns ErrInvalidDateRange when from > to (defensive; handlers also validate).
func NewTransactionFilter(
	userID identityvo.UserID,
	txType *vos.TransactionType,
	paymentMethod *vos.PaymentMethod,
	cardID *vos.CardID,
	categoryID *vos.CategoryID,
	invoiceStatus vos.InvoiceStatusFilter,
	descriptionContains string,
	from *time.Time,
	to *time.Time,
	pagination vos.Pagination,
) (TransactionFilter, error) {
	if from != nil && to != nil && from.After(*to) {
		return TransactionFilter{}, domain.ErrInvalidDateRange
	}
	return TransactionFilter{
		UserID:              userID,
		TransactionType:     txType,
		PaymentMethod:       paymentMethod,
		CardID:              cardID,
		CategoryID:          categoryID,
		InvoiceStatus:       invoiceStatus,
		DescriptionContains: descriptionContains,
		From:                from,
		To:                  to,
		Pagination:          pagination,
	}, nil
}

// RequiresCardPurchaseScope returns true when the invoice_status filter is active,
// meaning the query must be restricted to card-purchase transactions that have
// installments assigned to invoices (RF-19 semantics).
func (f TransactionFilter) RequiresCardPurchaseScope() bool {
	return f.InvoiceStatus.IsActive()
}
