package usecase

import (
	"context"

	"github.com/JailtonJunior94/devkit-go/pkg/database/manager"

	database "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/database"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/ports"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

type payInvoice struct {
	mgr     manager.Manager
	invRepo ports.InvoiceRepository
	clock   ports.Clock
	metrics ports.FinancialMetricsRecorder
}

// NewPayInvoice constructs the PayInvoice use case.
func NewPayInvoice(
	mgr manager.Manager,
	invRepo ports.InvoiceRepository,
	clock ports.Clock,
	metrics ports.FinancialMetricsRecorder,
) PayInvoice {
	return &payInvoice{
		mgr:     mgr,
		invRepo: invRepo,
		clock:   clock,
		metrics: metrics,
	}
}

func (uc *payInvoice) Execute(ctx context.Context, userID identityvo.UserID, invoiceID vos.InvoiceID, req dtos.PayInvoiceRequest) (dtos.InvoiceResponse, error) {
	inv, err := uc.invRepo.GetByID(ctx, userID, invoiceID)
	if err != nil {
		return dtos.InvoiceResponse{}, err
	}

	now := uc.clock.Now()
	if err := inv.MarkPaid(now); err != nil {
		return dtos.InvoiceResponse{}, err
	}

	err = database.Do(ctx, uc.mgr, func(ctx context.Context) error {
		return uc.invRepo.Update(ctx, inv)
	})
	if err != nil {
		return dtos.InvoiceResponse{}, err
	}

	pm := vos.PaymentMethodCreditCard
	if req.PaymentMethod != nil {
		if parsed, parseErr := vos.ParsePaymentMethod(*req.PaymentMethod); parseErr == nil {
			pm = parsed
		}
	}
	safeRecord(ctx, func() {
		uc.metrics.RecordAmountProcessed(ctx, inv.Total(), vos.TransactionTypeCreditPurchase, pm)
	})

	return toInvoiceResponse(inv), nil
}
