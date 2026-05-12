package filters

import (
	"time"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

// InvoiceFilter is the first-class collection for invoice listing criteria (RF-50).
type InvoiceFilter struct {
	UserID     identityvo.UserID
	CardID     *vos.CardID
	State      *vos.InvoiceState
	From       *time.Time
	To         *time.Time
	Pagination vos.Pagination
}

// NewInvoiceFilter constructs a validated InvoiceFilter.
// Returns ErrInvalidDateRange when from > to.
func NewInvoiceFilter(
	userID identityvo.UserID,
	cardID *vos.CardID,
	state *vos.InvoiceState,
	from *time.Time,
	to *time.Time,
	pagination vos.Pagination,
) (InvoiceFilter, error) {
	if from != nil && to != nil && from.After(*to) {
		return InvoiceFilter{}, domain.ErrInvalidDateRange
	}
	return InvoiceFilter{
		UserID:     userID,
		CardID:     cardID,
		State:      state,
		From:       from,
		To:         to,
		Pagination: pagination,
	}, nil
}
