package usecase

import (
	"context"

	"github.com/JailtonJunior94/devkit-go/pkg/database/manager"

	database "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/database"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/ports"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

type deleteTransaction struct {
	mgr      manager.Manager
	txRepo   ports.TransactionRepository
	instRepo ports.InstallmentRepository
	clock    ports.Clock
}

// NewDeleteTransaction constructs the DeleteTransaction use case.
func NewDeleteTransaction(
	mgr manager.Manager,
	txRepo ports.TransactionRepository,
	instRepo ports.InstallmentRepository,
	clock ports.Clock,
) DeleteTransaction {
	return &deleteTransaction{
		mgr:      mgr,
		txRepo:   txRepo,
		instRepo: instRepo,
		clock:    clock,
	}
}

func (uc *deleteTransaction) Execute(ctx context.Context, userID identityvo.UserID, txID vos.TransactionID) error {
	tx, err := uc.txRepo.GetByID(ctx, userID, txID)
	if err != nil {
		return err
	}

	hasRefund, err := uc.txRepo.HasActiveRefundFor(ctx, userID, txID)
	if err != nil {
		return err
	}

	hasClosedOrPaid, err := uc.instRepo.HasClosedOrPaidForTransaction(ctx, txID)
	if err != nil {
		return err
	}

	now := uc.clock.Now()
	if err := tx.SoftDelete(now, hasClosedOrPaid, hasRefund); err != nil {
		return err
	}

	return database.Do(ctx, uc.mgr, func(ctx context.Context) error {
		if delErr := uc.txRepo.SoftDelete(ctx, tx, now); delErr != nil {
			return delErr
		}
		return uc.instRepo.SoftDeleteByTransaction(ctx, userID, txID, now)
	})
}
