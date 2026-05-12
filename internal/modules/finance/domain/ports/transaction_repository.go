package ports

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/filters"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

// SummaryAggregates holds aggregate totals for the monthly summary use case (RF-20).
type SummaryAggregates struct {
	TotalIncome     vos.Money
	TotalExpense    vos.Money
	TotalRefundsIn  vos.Money // refunds whose original transaction type is income
	TotalRefundsOut vos.Money // refunds whose original transaction type is non-income
}

// TransactionRepository is the persistence port for Transaction aggregates.
type TransactionRepository interface {
	Add(ctx context.Context, t *entities.Transaction) error
	Update(ctx context.Context, t *entities.Transaction) error
	GetByID(ctx context.Context, userID identityvo.UserID, id vos.TransactionID) (*entities.Transaction, error)
	List(ctx context.Context, userID identityvo.UserID, f filters.TransactionFilter) ([]entities.Transaction, int64, error)
	SoftDelete(ctx context.Context, t *entities.Transaction, at time.Time) error
	HasActiveRefundFor(ctx context.Context, userID identityvo.UserID, original vos.TransactionID) (bool, error)
	SumForSummary(ctx context.Context, userID identityvo.UserID, period vos.Period) (SummaryAggregates, error)
}
