package usecase

import (
	"context"

	"github.com/JailtonJunior94/devkit-go/pkg/database/manager"

	database "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/database"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/ports"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/services"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

// GetInvoice is the use case interface for retrieving an invoice detail with items (RF-21).
type GetInvoice interface {
	Execute(ctx context.Context, userID identityvo.UserID, invoiceID vos.InvoiceID) (dtos.InvoiceDetailResponse, error)
}

type getInvoice struct {
	mgr      manager.Manager
	invRepo  ports.InvoiceRepository
	instRepo ports.InstallmentRepository
	closer   *services.InvoiceCloser
	clock    ports.Clock
}

// NewGetInvoice constructs the GetInvoice use case.
func NewGetInvoice(
	mgr manager.Manager,
	invRepo ports.InvoiceRepository,
	instRepo ports.InstallmentRepository,
	closer *services.InvoiceCloser,
	clock ports.Clock,
) GetInvoice {
	return &getInvoice{
		mgr:      mgr,
		invRepo:  invRepo,
		instRepo: instRepo,
		closer:   closer,
		clock:    clock,
	}
}

func (uc *getInvoice) Execute(ctx context.Context, userID identityvo.UserID, invoiceID vos.InvoiceID) (dtos.InvoiceDetailResponse, error) {
	inv, err := uc.invRepo.GetByID(ctx, userID, invoiceID)
	if err != nil {
		return dtos.InvoiceDetailResponse{}, err
	}

	now := uc.clock.Now()
	changed := uc.closer.CloseIfDue([]*entities.Invoice{inv}, now)
	if len(changed) > 0 {
		if err := database.Do(ctx, uc.mgr, func(ctx context.Context) error {
			return uc.invRepo.Update(ctx, inv)
		}); err != nil {
			return dtos.InvoiceDetailResponse{}, err
		}
	}

	insts, err := uc.instRepo.ListByInvoice(ctx, invoiceID)
	if err != nil {
		return dtos.InvoiceDetailResponse{}, err
	}

	items := make([]dtos.InvoiceItemResponse, len(insts))
	for i := range insts {
		items[i] = toInvoiceItemResponse(&insts[i])
	}

	return dtos.InvoiceDetailResponse{
		InvoiceResponse: toInvoiceResponse(inv),
		Items:           items,
	}, nil
}
