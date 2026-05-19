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
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/ports"
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
	return usecase.NewRefundTransaction(mgr, txRepo, factory, clock, ids, ports.NoopRecorder{})
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

	clock.EXPECT().Now().Return(fixedNow)
	ids.On("NewTransactionID").Return(refundID)

	txRepo.EXPECT().GetByID(mock.Anything, userID, txID).Return(original, nil)
	txRepo.EXPECT().HasActiveRefundFor(mock.Anything, userID, txID).Return(false, nil)
	txRepo.EXPECT().Add(mock.Anything, mock.AnythingOfType("*entities.Transaction")).Return(nil)

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

	txRepo.EXPECT().GetByID(mock.Anything, userID, txID).Return(original, nil)
	txRepo.EXPECT().HasActiveRefundFor(mock.Anything, userID, txID).Return(true, nil)

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

	txRepo.EXPECT().GetByID(mock.Anything, userID, txID).Return(nil, domain.ErrTransactionNotFound)

	uc := newRefundTransactionUC(t, mgr, txRepo, clock, ids)
	req := dtos.RefundTransactionRequest{}
	_, err := uc.Execute(ctx, userID, txID, req)

	assert.ErrorIs(t, err, domain.ErrTransactionNotFound)
}

func TestRefundTransaction_RecordsMetrics_WithRefundType(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	txID := vos.NewTransactionID()
	refundID := vos.NewTransactionID()

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	clock := portmocks.NewClock(t)
	ids := portmocks.NewIDGenerator(t)
	spy := &spyRecorder{}

	original := newExpenseTransaction(t, userID)
	clock.EXPECT().Now().Return(fixedNow)
	ids.On("NewTransactionID").Return(refundID)
	txRepo.EXPECT().GetByID(mock.Anything, userID, txID).Return(original, nil)
	txRepo.EXPECT().HasActiveRefundFor(mock.Anything, userID, txID).Return(false, nil)
	txRepo.EXPECT().Add(mock.Anything, mock.AnythingOfType("*entities.Transaction")).Return(nil)

	factory := services.NewRefundFactory()
	uc := usecase.NewRefundTransaction(mgr, txRepo, factory, clock, ids, spy)
	resp, err := uc.Execute(ctx, userID, txID, dtos.RefundTransactionRequest{})

	require.NoError(t, err)
	assert.Equal(t, "refund", resp.TransactionType)
	require.Len(t, spy.calls, 1)
	assert.Equal(t, vos.TransactionTypeRefund, spy.calls[0].txType, "refund must record transaction_type=refund")
}
