package usecase

import (
	"context"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/ports"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

// GetTransaction is the use case interface for retrieving a single transaction with its installments (RF-49).
type GetTransaction interface {
	Execute(ctx context.Context, userID identityvo.UserID, txID vos.TransactionID) (dtos.TransactionResponse, error)
}

type getTransaction struct {
	txRepo   ports.TransactionRepository
	instRepo ports.InstallmentRepository
}

// NewGetTransaction constructs the GetTransaction use case.
func NewGetTransaction(
	txRepo ports.TransactionRepository,
	instRepo ports.InstallmentRepository,
) GetTransaction {
	return &getTransaction{txRepo: txRepo, instRepo: instRepo}
}

func (uc *getTransaction) Execute(ctx context.Context, userID identityvo.UserID, txID vos.TransactionID) (dtos.TransactionResponse, error) {
	tx, err := uc.txRepo.GetByID(ctx, userID, txID)
	if err != nil {
		return dtos.TransactionResponse{}, err
	}

	insts, err := uc.instRepo.ListByTransaction(ctx, txID)
	if err != nil {
		return dtos.TransactionResponse{}, err
	}
	ptrs := make([]*entities.Installment, len(insts))
	for i := range insts {
		ptrs[i] = &insts[i]
	}
	tx.SetInstallments(ptrs)

	return toTransactionResponse(tx), nil
}
