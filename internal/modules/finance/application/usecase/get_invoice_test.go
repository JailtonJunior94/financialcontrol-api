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
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/services"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

func newGetInvoiceUC(
	t *testing.T,
	mgr *mockManager,
	invRepo *portmocks.InvoiceRepository,
	instRepo *portmocks.InstallmentRepository,
	clock *portmocks.Clock,
) usecase.GetInvoice {
	t.Helper()
	closer := services.NewInvoiceCloser()
	return usecase.NewGetInvoice(mgr, invRepo, instRepo, closer, clock)
}

func TestGetInvoice_GoldenPath(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	cardID := vos.NewCardID()
	txID := vos.NewTransactionID()
	instID := vos.NewInstallmentID()
	invoiceID := vos.NewInvoiceID()

	mgr, _ := newMockMgr(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	instRepo := portmocks.NewInstallmentRepository(t)
	clock := portmocks.NewClock(t)

	inv := newOpenInvoice(t, userID, cardID)
	inst := newScheduledInstallment(instID, txID, invoiceID)

	clock.On("Now").Return(fixedNow)
	invRepo.On("GetByID", mock.Anything, userID, invoiceID).Return(inv, nil)
	instRepo.On("ListByInvoice", mock.Anything, invoiceID).Return([]entities.Installment{inst}, nil)

	uc := newGetInvoiceUC(t, mgr, invRepo, instRepo, clock)
	resp, err := uc.Execute(ctx, userID, invoiceID)

	require.NoError(t, err)
	assert.Equal(t, "open", resp.State)
	assert.Len(t, resp.Items, 1)
}

func TestGetInvoice_LazyClose_ClosesOverdueInvoice(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	cardID := vos.NewCardID()
	invoiceID := vos.NewInvoiceID()

	mgr, _ := newMockMgr(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	instRepo := portmocks.NewInstallmentRepository(t)
	clock := portmocks.NewClock(t)

	// Build invoice with closing date in the past so CloseIfDue triggers
	card := newActiveCardView(userID, cardID)
	pastStart := fixedNow.AddDate(0, -2, 0)
	pastEnd := fixedNow.AddDate(0, -1, 0)
	pastClosing := fixedNow.AddDate(0, -1, -5)
	pastDue := fixedNow.AddDate(0, -1, 10)
	overdueInv, err := entities.NewInvoice(card, pastStart, pastEnd, pastClosing, pastDue, pastStart)
	require.NoError(t, err)

	clock.On("Now").Return(fixedNow)
	invRepo.On("GetByID", mock.Anything, userID, invoiceID).Return(overdueInv, nil)
	invRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Invoice")).Return(nil)
	instRepo.On("ListByInvoice", mock.Anything, invoiceID).Return(nil, nil)

	uc := newGetInvoiceUC(t, mgr, invRepo, instRepo, clock)
	resp, err := uc.Execute(ctx, userID, invoiceID)

	require.NoError(t, err)
	assert.Equal(t, "closed", resp.State)
	invRepo.AssertCalled(t, "Update", mock.Anything, mock.AnythingOfType("*entities.Invoice"))
}

func TestGetInvoice_NotFound_ReturnsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	invoiceID := vos.NewInvoiceID()

	mgr, _ := newMockMgr(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	instRepo := portmocks.NewInstallmentRepository(t)
	clock := portmocks.NewClock(t)

	invRepo.On("GetByID", mock.Anything, userID, invoiceID).Return(nil, domain.ErrInvoiceNotFound)

	uc := newGetInvoiceUC(t, mgr, invRepo, instRepo, clock)
	_, err := uc.Execute(ctx, userID, invoiceID)

	assert.ErrorIs(t, err, domain.ErrInvoiceNotFound)
}
