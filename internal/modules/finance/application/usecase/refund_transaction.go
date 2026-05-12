package usecase

import (
	"context"

	"github.com/JailtonJunior94/devkit-go/pkg/database/manager"

	database "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/database"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/ports"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/services"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

// RefundTransaction is the use case interface for creating a refund (RF-53/RF-13).
type RefundTransaction interface {
	Execute(ctx context.Context, userID identityvo.UserID, txID vos.TransactionID, req dtos.RefundTransactionRequest) (dtos.TransactionResponse, error)
}

type refundTransaction struct {
	mgr     manager.Manager
	txRepo  ports.TransactionRepository
	factory *services.RefundFactory
	clock   ports.Clock
	ids     ports.IDGenerator
}

// NewRefundTransaction constructs the RefundTransaction use case.
func NewRefundTransaction(
	mgr manager.Manager,
	txRepo ports.TransactionRepository,
	factory *services.RefundFactory,
	clock ports.Clock,
	ids ports.IDGenerator,
) RefundTransaction {
	return &refundTransaction{
		mgr:     mgr,
		txRepo:  txRepo,
		factory: factory,
		clock:   clock,
		ids:     ids,
	}
}

func (uc *refundTransaction) Execute(ctx context.Context, userID identityvo.UserID, txID vos.TransactionID, req dtos.RefundTransactionRequest) (dtos.TransactionResponse, error) {
	original, err := uc.txRepo.GetByID(ctx, userID, txID)
	if err != nil {
		return dtos.TransactionResponse{}, err
	}

	hasActiveRefund, err := uc.txRepo.HasActiveRefundFor(ctx, userID, txID)
	if err != nil {
		return dtos.TransactionResponse{}, err
	}

	override := services.RefundOverride{}
	if req.PaymentMethod != nil {
		pm, pmErr := vos.ParsePaymentMethod(*req.PaymentMethod)
		if pmErr != nil {
			return dtos.TransactionResponse{}, pmErr
		}
		override.PaymentMethod = &pm
	}
	override.Description = req.Description

	refund, err := uc.factory.Build(original, hasActiveRefund, override, uc.clock, uc.ids)
	if err != nil {
		return dtos.TransactionResponse{}, err
	}

	err = database.Do(ctx, uc.mgr, func(ctx context.Context) error {
		return uc.txRepo.Add(ctx, refund)
	})
	if err != nil {
		return dtos.TransactionResponse{}, err
	}

	return toTransactionResponse(refund), nil
}
