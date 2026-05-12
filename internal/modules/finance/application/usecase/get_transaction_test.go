package usecase_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/application/usecase"
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	portmocks "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/ports/mocks"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

func newGetTransactionUC(
	t *testing.T,
	txRepo *portmocks.TransactionRepository,
	instRepo *portmocks.InstallmentRepository,
) usecase.GetTransaction {
	t.Helper()
	return usecase.NewGetTransaction(txRepo, instRepo)
}

func TestGetTransaction_GoldenPath(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	txID := vos.NewTransactionID()

	txRepo := portmocks.NewTransactionRepository(t)
	instRepo := portmocks.NewInstallmentRepository(t)

	tx := newExpenseTransaction(t, userID)

	txRepo.On("GetByID", mock.Anything, userID, txID).Return(tx, nil)
	instRepo.On("ListByTransaction", mock.Anything, txID).Return(nil, nil)

	uc := newGetTransactionUC(t, txRepo, instRepo)
	resp, err := uc.Execute(ctx, userID, txID)

	require.NoError(t, err)
	assert.Equal(t, tx.ID().String(), resp.ID)
	assert.Equal(t, "expense", resp.TransactionType)
}

func TestGetTransaction_NotFound_ReturnsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	txID := vos.NewTransactionID()

	txRepo := portmocks.NewTransactionRepository(t)
	instRepo := portmocks.NewInstallmentRepository(t)

	txRepo.On("GetByID", mock.Anything, userID, txID).Return(nil, domain.ErrTransactionNotFound)

	uc := newGetTransactionUC(t, txRepo, instRepo)
	_, err := uc.Execute(ctx, userID, txID)

	assert.ErrorIs(t, err, domain.ErrTransactionNotFound)
}
