package usecase_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/application/usecase"
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/entities"
	portmocks "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/ports/mocks"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

func newAnticipateInstallmentUC(
	t *testing.T,
	mgr *mockManager,
	txRepo *portmocks.TransactionRepository,
	invRepo *portmocks.InvoiceRepository,
	instRepo *portmocks.InstallmentRepository,
	clock *portmocks.Clock,
) usecase.AnticipateInstallment {
	t.Helper()
	return usecase.NewAnticipateInstallment(mgr, txRepo, invRepo, instRepo, clock)
}

func TestAnticipateInstallment_GoldenPath(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	cardID := vos.NewCardID()
	txID := vos.NewTransactionID()
	instID := vos.NewInstallmentID()
	invoiceID := vos.NewInvoiceID()

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	instRepo := portmocks.NewInstallmentRepository(t)
	clock := portmocks.NewClock(t)

	tx := newCreditPurchaseTransaction(t, userID, cardID)
	inst := newScheduledInstallment(instID, tx.ID(), invoiceID)
	targetInv := newOpenInvoice(t, userID, cardID)

	clock.On("Now").Return(fixedNow)
	txRepo.On("GetByID", mock.Anything, userID, txID).Return(tx, nil)
	instRepo.On("ListByTransaction", mock.Anything, txID).Return([]entities.Installment{inst}, nil)
	invRepo.On("NextOpenFor", mock.Anything, userID, cardID, mock.Anything).Return(targetInv, nil)
	txRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Transaction")).Return(nil)
	instRepo.On("UpdateBatch", mock.Anything, mock.Anything).Return(nil)

	uc := newAnticipateInstallmentUC(t, mgr, txRepo, invRepo, instRepo, clock)
	resp, err := uc.Execute(ctx, userID, txID, instID)

	require.NoError(t, err)
	assert.Equal(t, tx.ID().String(), resp.ID)
}

func TestAnticipateInstallment_NoCardID_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	txID := vos.NewTransactionID()
	instID := vos.NewInstallmentID()

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	instRepo := portmocks.NewInstallmentRepository(t)
	clock := portmocks.NewClock(t)

	// expense transaction has no cardID
	tx := newExpenseTransaction(t, userID)

	txRepo.On("GetByID", mock.Anything, userID, txID).Return(tx, nil)
	instRepo.On("ListByTransaction", mock.Anything, txID).Return([]entities.Installment{}, nil)

	uc := newAnticipateInstallmentUC(t, mgr, txRepo, invRepo, instRepo, clock)
	resp, err := uc.Execute(ctx, userID, txID, instID)

	require.NoError(t, err)
	// No cardID means empty response (no-op)
	assert.Empty(t, resp.ID)
}

func TestAnticipateInstallment_TransactionNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	txID := vos.NewTransactionID()
	instID := vos.NewInstallmentID()

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	instRepo := portmocks.NewInstallmentRepository(t)
	clock := portmocks.NewClock(t)

	txRepo.On("GetByID", mock.Anything, userID, txID).Return(nil, domain.ErrTransactionNotFound)

	uc := newAnticipateInstallmentUC(t, mgr, txRepo, invRepo, instRepo, clock)
	_, err := uc.Execute(ctx, userID, txID, instID)

	assert.ErrorIs(t, err, domain.ErrTransactionNotFound)
}
