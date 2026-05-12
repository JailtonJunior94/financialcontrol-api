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

func newDeleteTransactionUC(
	t *testing.T,
	mgr *mockManager,
	txRepo *portmocks.TransactionRepository,
	instRepo *portmocks.InstallmentRepository,
	clock *portmocks.Clock,
) usecase.DeleteTransaction {
	t.Helper()
	return usecase.NewDeleteTransaction(mgr, txRepo, instRepo, clock)
}

func TestDeleteTransaction_GoldenPath(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	txID := vos.NewTransactionID()

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	instRepo := portmocks.NewInstallmentRepository(t)
	clock := portmocks.NewClock(t)

	tx := newExpenseTransaction(t, userID)

	clock.On("Now").Return(fixedNow)
	txRepo.On("GetByID", mock.Anything, userID, txID).Return(tx, nil)
	txRepo.On("HasActiveRefundFor", mock.Anything, userID, txID).Return(false, nil)
	instRepo.On("HasClosedOrPaidForTransaction", mock.Anything, txID).Return(false, nil)
	txRepo.On("SoftDelete", mock.Anything, mock.Anything, mock.AnythingOfType("time.Time")).Return(nil)
	instRepo.On("SoftDeleteByTransaction", mock.Anything, txID, mock.AnythingOfType("time.Time")).Return(nil)

	uc := newDeleteTransactionUC(t, mgr, txRepo, instRepo, clock)
	err := uc.Execute(ctx, userID, txID)

	require.NoError(t, err)
}

func TestDeleteTransaction_HasActiveRefund_ReturnsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	txID := vos.NewTransactionID()

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	instRepo := portmocks.NewInstallmentRepository(t)
	clock := portmocks.NewClock(t)

	tx := newExpenseTransaction(t, userID)

	clock.On("Now").Return(fixedNow)
	txRepo.On("GetByID", mock.Anything, userID, txID).Return(tx, nil)
	txRepo.On("HasActiveRefundFor", mock.Anything, userID, txID).Return(true, nil)
	instRepo.On("HasClosedOrPaidForTransaction", mock.Anything, txID).Return(false, nil)

	uc := newDeleteTransactionUC(t, mgr, txRepo, instRepo, clock)
	err := uc.Execute(ctx, userID, txID)

	assert.ErrorIs(t, err, domain.ErrTransactionHasDependentRefund)
}

func TestDeleteTransaction_NotFound_ReturnsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	txID := vos.NewTransactionID()

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	instRepo := portmocks.NewInstallmentRepository(t)
	clock := portmocks.NewClock(t)

	txRepo.On("GetByID", mock.Anything, userID, txID).Return(nil, domain.ErrTransactionNotFound)

	uc := newDeleteTransactionUC(t, mgr, txRepo, instRepo, clock)
	err := uc.Execute(ctx, userID, txID)

	assert.ErrorIs(t, err, domain.ErrTransactionNotFound)
}
