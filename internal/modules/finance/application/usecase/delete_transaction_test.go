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

	clock.EXPECT().Now().Return(fixedNow)
	txRepo.EXPECT().GetByID(mock.Anything, userID, txID).Return(tx, nil)
	txRepo.EXPECT().HasActiveRefundFor(mock.Anything, userID, txID).Return(false, nil)
	instRepo.EXPECT().HasClosedOrPaidForTransaction(mock.Anything, txID).Return(false, nil)
	txRepo.EXPECT().SoftDelete(mock.Anything, mock.Anything, mock.AnythingOfType("time.Time")).Return(nil)
	instRepo.EXPECT().SoftDeleteByTransaction(mock.Anything, userID, txID, mock.AnythingOfType("time.Time")).Return(nil)

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

	clock.EXPECT().Now().Return(fixedNow)
	txRepo.EXPECT().GetByID(mock.Anything, userID, txID).Return(tx, nil)
	txRepo.EXPECT().HasActiveRefundFor(mock.Anything, userID, txID).Return(true, nil)
	instRepo.EXPECT().HasClosedOrPaidForTransaction(mock.Anything, txID).Return(false, nil)

	uc := newDeleteTransactionUC(t, mgr, txRepo, instRepo, clock)
	err := uc.Execute(ctx, userID, txID)

	assert.ErrorIs(t, err, domain.ErrTransactionHasDependentRefund)
}

// BUG-003 regression: DeleteTransaction must surface
// ErrInstallmentInClosedOrPaidInvoice when any installment is bound to a closed
// or paid invoice. After fixing the SQL to JOIN dbo.FinanceInvoices, the repository
// returns true even when installment.status='scheduled' (RF-46 keeps it that
// way while the invoice is closed-but-unpaid).
func TestDeleteTransaction_BlocksWhenInstallmentInClosedOrPaidInvoice(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	txID := vos.NewTransactionID()

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	instRepo := portmocks.NewInstallmentRepository(t)
	clock := portmocks.NewClock(t)

	tx := newExpenseTransaction(t, userID)

	clock.EXPECT().Now().Return(fixedNow)
	txRepo.EXPECT().GetByID(mock.Anything, userID, txID).Return(tx, nil)
	txRepo.EXPECT().HasActiveRefundFor(mock.Anything, userID, txID).Return(false, nil)
	// repository now reflects the invoice-state check — returns true for closed-unpaid
	instRepo.EXPECT().HasClosedOrPaidForTransaction(mock.Anything, txID).Return(true, nil)

	uc := newDeleteTransactionUC(t, mgr, txRepo, instRepo, clock)
	err := uc.Execute(ctx, userID, txID)

	assert.ErrorIs(t, err, domain.ErrInstallmentInClosedOrPaidInvoice)
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

	txRepo.EXPECT().GetByID(mock.Anything, userID, txID).Return(nil, domain.ErrTransactionNotFound)

	uc := newDeleteTransactionUC(t, mgr, txRepo, instRepo, clock)
	err := uc.Execute(ctx, userID, txID)

	assert.ErrorIs(t, err, domain.ErrTransactionNotFound)
}
