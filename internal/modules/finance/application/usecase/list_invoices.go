package usecase

import (
	"context"

	"github.com/JailtonJunior94/devkit-go/pkg/database/manager"

	database "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/database"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/filters"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/ports"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/services"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

type listInvoices struct {
	mgr     manager.Manager
	invRepo ports.InvoiceRepository
	closer  *services.InvoiceCloser
	clock   ports.Clock
}

// NewListInvoices constructs the ListInvoices use case.
func NewListInvoices(
	mgr manager.Manager,
	invRepo ports.InvoiceRepository,
	closer *services.InvoiceCloser,
	clock ports.Clock,
) ListInvoices {
	return &listInvoices{
		mgr:     mgr,
		invRepo: invRepo,
		closer:  closer,
		clock:   clock,
	}
}

func (uc *listInvoices) Execute(ctx context.Context, userID identityvo.UserID, f filters.InvoiceFilter) (dtos.PaginatedResponse[dtos.InvoiceResponse], error) {
	invs, total, err := uc.invRepo.List(ctx, userID, f)
	if err != nil {
		return dtos.PaginatedResponse[dtos.InvoiceResponse]{}, err
	}

	ptrs := make([]*entities.Invoice, len(invs))
	for i := range invs {
		ptrs[i] = &invs[i]
	}

	now := uc.clock.Now()
	changed := uc.closer.CloseIfDue(ptrs, now)
	if len(changed) > 0 {
		if err := database.Do(ctx, uc.mgr, func(ctx context.Context) error {
			for _, inv := range changed {
				if updErr := uc.invRepo.Update(ctx, inv); updErr != nil {
					return updErr
				}
			}
			return nil
		}); err != nil {
			return dtos.PaginatedResponse[dtos.InvoiceResponse]{}, err
		}
	}

	items := make([]dtos.InvoiceResponse, len(invs))
	for i := range invs {
		items[i] = toInvoiceResponse(ptrs[i])
	}
	return dtos.PaginatedResponse[dtos.InvoiceResponse]{
		Items:    items,
		Total:    total,
		Page:     f.Pagination.Page(),
		PageSize: f.Pagination.PageSize(),
	}, nil
}
