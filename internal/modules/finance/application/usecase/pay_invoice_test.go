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
	return usecase.NewPayInvoice(mgr, invRepo, clock, ports.NoopRecorder{})
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

	clock.EXPECT().Now().Return(fixedNow)
	invRepo.EXPECT().GetByID(mock.Anything, userID, invoiceID).Return(inv, nil)
	invRepo.EXPECT().Update(mock.Anything, mock.AnythingOfType("*entities.Invoice")).Return(nil)

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

	clock.EXPECT().Now().Return(fixedNow)
	invRepo.EXPECT().GetByID(mock.Anything, userID, invoiceID).Return(inv, nil)

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

	invRepo.EXPECT().GetByID(mock.Anything, userID, invoiceID).Return(nil, domain.ErrInvoiceNotFound)

	uc := newPayInvoiceUC(t, mgr, invRepo, clock)
	_, err := uc.Execute(ctx, userID, invoiceID, dtos.PayInvoiceRequest{})

	assert.ErrorIs(t, err, domain.ErrInvoiceNotFound)
}

func TestPayInvoice_RecordsMetrics_WithPaymentMethodFromRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	cardID := vos.NewCardID()
	invoiceID := vos.NewInvoiceID()

	mgr, _ := newMockMgr(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	clock := portmocks.NewClock(t)
	spy := &spyRecorder{}

	inv := newOpenInvoice(t, userID, cardID)
	clock.EXPECT().Now().Return(fixedNow)
	invRepo.EXPECT().GetByID(mock.Anything, userID, invoiceID).Return(inv, nil)
	invRepo.EXPECT().Update(mock.Anything, mock.AnythingOfType("*entities.Invoice")).Return(nil)

	pm := "pix"
	req := dtos.PayInvoiceRequest{PaymentMethod: &pm}
	uc := usecase.NewPayInvoice(mgr, invRepo, clock, spy)
	_, err := uc.Execute(ctx, userID, invoiceID, req)

	require.NoError(t, err)
	require.Len(t, spy.calls, 1)
	assert.Equal(t, vos.PaymentMethodPix, spy.calls[0].method, "payment method from request must be used")
	assert.Equal(t, vos.TransactionTypeCreditPurchase, spy.calls[0].txType)
}

func TestPayInvoice_RecordsMetrics_DefaultsCreditCard_WhenNoPaymentMethod(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	cardID := vos.NewCardID()
	invoiceID := vos.NewInvoiceID()

	mgr, _ := newMockMgr(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	clock := portmocks.NewClock(t)
	spy := &spyRecorder{}

	inv := newOpenInvoice(t, userID, cardID)
	clock.EXPECT().Now().Return(fixedNow)
	invRepo.EXPECT().GetByID(mock.Anything, userID, invoiceID).Return(inv, nil)
	invRepo.EXPECT().Update(mock.Anything, mock.AnythingOfType("*entities.Invoice")).Return(nil)

	uc := usecase.NewPayInvoice(mgr, invRepo, clock, spy)
	_, err := uc.Execute(ctx, userID, invoiceID, dtos.PayInvoiceRequest{})

	require.NoError(t, err)
	require.Len(t, spy.calls, 1)
	assert.Equal(t, vos.PaymentMethodCreditCard, spy.calls[0].method, "credit_card default must be used when no payment method in request")
}
