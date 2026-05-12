package usecase

import (
	"context"

	"github.com/JailtonJunior94/devkit-go/pkg/database/manager"

	database "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/database"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/ports"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

// AnticipateInstallment is the use case interface for moving an installment to the next open invoice (RF-52/RF-14).
type AnticipateInstallment interface {
	Execute(ctx context.Context, userID identityvo.UserID, txID vos.TransactionID, installmentID vos.InstallmentID) (dtos.TransactionResponse, error)
}

type anticipateInstallment struct {
	mgr      manager.Manager
	txRepo   ports.TransactionRepository
	invRepo  ports.InvoiceRepository
	instRepo ports.InstallmentRepository
	clock    ports.Clock
}

// NewAnticipateInstallment constructs the AnticipateInstallment use case.
func NewAnticipateInstallment(
	mgr manager.Manager,
	txRepo ports.TransactionRepository,
	invRepo ports.InvoiceRepository,
	instRepo ports.InstallmentRepository,
	clock ports.Clock,
) AnticipateInstallment {
	return &anticipateInstallment{
		mgr:      mgr,
		txRepo:   txRepo,
		invRepo:  invRepo,
		instRepo: instRepo,
		clock:    clock,
	}
}

func (uc *anticipateInstallment) Execute(ctx context.Context, userID identityvo.UserID, txID vos.TransactionID, installmentID vos.InstallmentID) (dtos.TransactionResponse, error) {
	tx, err := uc.txRepo.GetByID(ctx, userID, txID)
	if err != nil {
		return dtos.TransactionResponse{}, err
	}

	existingInsts, err := uc.instRepo.ListByTransaction(ctx, txID)
	if err != nil {
		return dtos.TransactionResponse{}, err
	}
	ptrs := make([]*entities.Installment, len(existingInsts))
	for i := range existingInsts {
		ptrs[i] = &existingInsts[i]
	}
	tx.SetInstallments(ptrs)

	if tx.CardID() == nil {
		return dtos.TransactionResponse{}, nil
	}

	targetInvoice, err := uc.invRepo.NextOpenFor(ctx, userID, *tx.CardID(), uc.clock)
	if err != nil {
		return dtos.TransactionResponse{}, err
	}

	now := uc.clock.Now()
	if err := tx.Anticipate(installmentID, targetInvoice, now); err != nil {
		return dtos.TransactionResponse{}, err
	}

	err = database.Do(ctx, uc.mgr, func(ctx context.Context) error {
		if updErr := uc.txRepo.Update(ctx, tx); updErr != nil {
			return updErr
		}
		return uc.instRepo.UpdateBatch(ctx, tx.Installments())
	})
	if err != nil {
		return dtos.TransactionResponse{}, err
	}

	return toTransactionResponse(tx), nil
}
