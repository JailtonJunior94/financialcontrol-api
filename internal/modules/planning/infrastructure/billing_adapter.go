package infrastructure

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/planning"
	"github.com/jailtonjunior94/financialcontrol-api/internal/platform/persistence"
)

// billDataSource is the minimal interface the billing adapter requires from the
// billing repository. Defined here (consumer side) so planning never imports
// the concrete billing repository or the billing application package.
type billDataSource interface {
	Get(date time.Time) (*persistence.BillQuery, error)
}

// BillingReadAdapter implements planning.BillingReadPort by querying the
// billing repository and projecting the result into the planning read model.
type BillingReadAdapter struct {
	repo billDataSource
}

func NewBillingReadAdapter(repo billDataSource) *BillingReadAdapter {
	return &BillingReadAdapter{repo: repo}
}

func (a *BillingReadAdapter) GetMonthlyBills(_ context.Context, referenceDate time.Time) (*planning.MonthlyBillsReadModel, error) {
	query, err := a.repo.Get(referenceDate)
	if err != nil {
		return nil, err
	}

	model := &planning.MonthlyBillsReadModel{ReferenceDate: referenceDate}
	if query == nil {
		return model, nil
	}

	model.Items = make([]planning.BillReadItem, 0, len(query.Items))
	for _, i := range query.Items {
		model.Items = append(model.Items, planning.BillReadItem{
			ID:          i.ID,
			Description: i.Description,
			Total:       i.Total,
		})
	}

	return model, nil
}
