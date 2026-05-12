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
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

func newPayInvoiceUC(
	t *testing.T,
	mgr *mockManager,
	invRepo *portmocks.InvoiceRepository,
	clock *portmocks.Clock,
) usecase.PayInvoice {
	t.Helper()
	return usecase.NewPayInvoice(mgr, invRepo, clock)
}

func TestPayInvoice_GoldenPath(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	cardID := vos.NewCardID()
	invoiceID := vos.NewInvoiceID()

	mgr, _ := newMockMgr(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	clock := portmocks.NewClock(t)

	inv := newOpenInvoice(t, userID, cardID)

	clock.On("Now").Return(fixedNow)
	invRepo.On("GetByID", mock.Anything, userID, invoiceID).Return(inv, nil)
	invRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Invoice")).Return(nil)

	uc := newPayInvoiceUC(t, mgr, invRepo, clock)
	resp, err := uc.Execute(ctx, userID, invoiceID, dtos.PayInvoiceRequest{})

	require.NoError(t, err)
	assert.Equal(t, "paid", resp.State)
}

func TestPayInvoice_AlreadyPaid_ReturnsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	cardID := vos.NewCardID()
	invoiceID := vos.NewInvoiceID()

	mgr, _ := newMockMgr(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	clock := portmocks.NewClock(t)

	inv := newOpenInvoice(t, userID, cardID)
	// pay once to make it paid
	_ = inv.MarkPaid(fixedNow)

	clock.On("Now").Return(fixedNow)
	invRepo.On("GetByID", mock.Anything, userID, invoiceID).Return(inv, nil)

	uc := newPayInvoiceUC(t, mgr, invRepo, clock)
	_, err := uc.Execute(ctx, userID, invoiceID, dtos.PayInvoiceRequest{})

	assert.ErrorIs(t, err, domain.ErrInvoiceAlreadyPaid)
}

func TestPayInvoice_NotFound_ReturnsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	invoiceID := vos.NewInvoiceID()

	mgr, _ := newMockMgr(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	clock := portmocks.NewClock(t)

	invRepo.On("GetByID", mock.Anything, userID, invoiceID).Return(nil, domain.ErrInvoiceNotFound)

	uc := newPayInvoiceUC(t, mgr, invRepo, clock)
	_, err := uc.Execute(ctx, userID, invoiceID, dtos.PayInvoiceRequest{})

	assert.ErrorIs(t, err, domain.ErrInvoiceNotFound)
}
