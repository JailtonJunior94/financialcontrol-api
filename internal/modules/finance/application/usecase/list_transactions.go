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
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

type listTransactions struct {
	mgr     manager.Manager
	txRepo  ports.TransactionRepository
	invRepo ports.InvoiceRepository
	closer  *services.InvoiceCloser
	clock   ports.Clock
}

// NewListTransactions constructs the ListTransactions use case.
func NewListTransactions(
	mgr manager.Manager,
	txRepo ports.TransactionRepository,
	invRepo ports.InvoiceRepository,
	closer *services.InvoiceCloser,
	clock ports.Clock,
) ListTransactions {
	return &listTransactions{
		mgr:     mgr,
		txRepo:  txRepo,
		invRepo: invRepo,
		closer:  closer,
		clock:   clock,
	}
}

func (uc *listTransactions) Execute(ctx context.Context, userID identityvo.UserID, f filters.TransactionFilter) (dtos.PaginatedResponse[dtos.TransactionResponse], error) {
	now := uc.clock.Now()

	maxPag, pagErr := vos.NewPagination(1, 100)
	if pagErr != nil {
		return dtos.PaginatedResponse[dtos.TransactionResponse]{}, pagErr
	}
	openFilter, err := filters.NewInvoiceFilter(userID, nil, statePtr(vos.InvoiceStateOpen), nil, nil, maxPag)
	if err != nil {
		return dtos.PaginatedResponse[dtos.TransactionResponse]{}, err
	}
	openInvoices, _, err := uc.invRepo.List(ctx, userID, openFilter)
	if err != nil {
		return dtos.PaginatedResponse[dtos.TransactionResponse]{}, err
	}
	openPtrs := make([]*entities.Invoice, len(openInvoices))
	for i := range openInvoices {
		openPtrs[i] = &openInvoices[i]
	}
	changed := uc.closer.CloseIfDue(openPtrs, now)

	if len(changed) > 0 {
		if err := database.Do(ctx, uc.mgr, func(ctx context.Context) error {
			for _, inv := range changed {
				if updErr := uc.invRepo.Update(ctx, inv); updErr != nil {
					return updErr
				}
			}
			return nil
		}); err != nil {
			return dtos.PaginatedResponse[dtos.TransactionResponse]{}, err
		}
	}

	txs, total, err := uc.txRepo.List(ctx, userID, f)
	if err != nil {
		return dtos.PaginatedResponse[dtos.TransactionResponse]{}, err
	}

	items := make([]dtos.TransactionResponse, len(txs))
	for i := range txs {
		items[i] = toTransactionResponse(&txs[i])
	}
	return dtos.PaginatedResponse[dtos.TransactionResponse]{
		Items:    items,
		Total:    total,
		Page:     f.Pagination.Page(),
		PageSize: f.Pagination.PageSize(),
	}, nil
}

func statePtr(s vos.InvoiceState) *vos.InvoiceState { return &s }
