package usecase_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/application/usecase"
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	portmocks "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/ports/mocks"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/services"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

func newRefundTransactionUC(
	t *testing.T,
	mgr *mockManager,
	txRepo *portmocks.TransactionRepository,
	clock *portmocks.Clock,
	ids *portmocks.IDGenerator,
) usecase.RefundTransaction {
	t.Helper()
	factory := services.NewRefundFactory()
	return usecase.NewRefundTransaction(mgr, txRepo, factory, clock, ids)
}

func TestRefundTransaction_GoldenPath(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	txID := vos.NewTransactionID()
	refundID := vos.NewTransactionID()

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	clock := portmocks.NewClock(t)
	ids := portmocks.NewIDGenerator(t)

	original := newExpenseTransaction(t, userID)

	clock.On("Now").Return(fixedNow)
	ids.On("NewTransactionID").Return(refundID)

	txRepo.On("GetByID", mock.Anything, userID, txID).Return(original, nil)
	txRepo.On("HasActiveRefundFor", mock.Anything, userID, txID).Return(false, nil)
	txRepo.On("Add", mock.Anything, mock.AnythingOfType("*entities.Transaction")).Return(nil)

	uc := newRefundTransactionUC(t, mgr, txRepo, clock, ids)
	req := dtos.RefundTransactionRequest{}
	resp, err := uc.Execute(ctx, userID, txID, req)

	require.NoError(t, err)
	assert.Equal(t, refundID.String(), resp.ID)
	assert.Equal(t, "refund", resp.TransactionType)
}

func TestRefundTransaction_AlreadyHasRefund_ReturnsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	txID := vos.NewTransactionID()

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	clock := portmocks.NewClock(t)
	ids := portmocks.NewIDGenerator(t)

	original := newExpenseTransaction(t, userID)

	txRepo.On("GetByID", mock.Anything, userID, txID).Return(original, nil)
	txRepo.On("HasActiveRefundFor", mock.Anything, userID, txID).Return(true, nil)

	uc := newRefundTransactionUC(t, mgr, txRepo, clock, ids)
	req := dtos.RefundTransactionRequest{}
	_, err := uc.Execute(ctx, userID, txID, req)

	assert.ErrorIs(t, err, domain.ErrRefundAlreadyExists)
}

func TestRefundTransaction_NotFound_ReturnsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	txID := vos.NewTransactionID()

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	clock := portmocks.NewClock(t)
	ids := portmocks.NewIDGenerator(t)

	txRepo.On("GetByID", mock.Anything, userID, txID).Return(nil, domain.ErrTransactionNotFound)

	uc := newRefundTransactionUC(t, mgr, txRepo, clock, ids)
	req := dtos.RefundTransactionRequest{}
	_, err := uc.Execute(ctx, userID, txID, req)

	assert.ErrorIs(t, err, domain.ErrTransactionNotFound)
}
