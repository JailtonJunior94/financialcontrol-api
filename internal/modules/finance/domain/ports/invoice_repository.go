package ports

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/filters"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/projections"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

// InvoiceRepository is the persistence port for Invoice aggregates.
type InvoiceRepository interface {
	AssignOrCreateOpen(ctx context.Context, userID identityvo.UserID, cardID vos.CardID, occurredAtLocal time.Time, card projections.CardView, clock Clock) (*entities.Invoice, error)
	GetByID(ctx context.Context, userID identityvo.UserID, id vos.InvoiceID) (*entities.Invoice, error)
	List(ctx context.Context, userID identityvo.UserID, f filters.InvoiceFilter) ([]entities.Invoice, int64, error)
	Update(ctx context.Context, inv *entities.Invoice) error
	NextOpenFor(ctx context.Context, userID identityvo.UserID, cardID vos.CardID, clock Clock) (*entities.Invoice, error)
	SumPaidInPeriod(ctx context.Context, userID identityvo.UserID, period vos.Period) (vos.Money, error)
	SumOpenForUser(ctx context.Context, userID identityvo.UserID, period vos.Period) (vos.Money, error)
}
