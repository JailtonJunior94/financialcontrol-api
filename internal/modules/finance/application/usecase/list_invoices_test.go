package usecase_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/application/usecase"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/filters"
	portmocks "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/ports/mocks"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/projections"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/services"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

func newListInvoicesUC(
	t *testing.T,
	mgr *mockManager,
	invRepo *portmocks.InvoiceRepository,
	clock *portmocks.Clock,
) usecase.ListInvoices {
	t.Helper()
	closer := services.NewInvoiceCloser()
	return usecase.NewListInvoices(mgr, invRepo, closer, clock)
}

func TestListInvoices_GoldenPath(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	cardID := vos.NewCardID()

	mgr, _ := newMockMgr(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	clock := portmocks.NewClock(t)

	inv := newOpenInvoice(t, userID, cardID)
	invs := []entities.Invoice{*inv}

	clock.EXPECT().Now().Return(fixedNow)
	invRepo.EXPECT().List(mock.Anything, userID, mock.AnythingOfType("filters.InvoiceFilter")).Return(invs, int64(1), nil)

	pag, err := vos.NewPagination(1, 10)
	require.NoError(t, err)
	f, err := filters.NewInvoiceFilter(userID, nil, nil, nil, nil, pag)
	require.NoError(t, err)

	uc := newListInvoicesUC(t, mgr, invRepo, clock)
	resp, err := uc.Execute(ctx, userID, f)

	require.NoError(t, err)
	assert.Len(t, resp.Items, 1)
	assert.Equal(t, int64(1), resp.Total)
	assert.Equal(t, "open", resp.Items[0].State)
}

func TestListInvoices_LazyClose_ClosesOverdueInvoices(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	cardID := vos.NewCardID()

	mgr, _ := newMockMgr(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	clock := portmocks.NewClock(t)

	card := projections.CardView{
		ID:           cardID,
		UserID:       userID,
		FlagName:     "Visa",
		ClosingDay:   10,
		DueDay:       20,
		BillingCycle: 30,
		Active:       true,
	}
	pastStart := fixedNow.AddDate(0, -2, 0)
	pastEnd := fixedNow.AddDate(0, -1, 0)
	pastClosing := fixedNow.AddDate(0, -1, -5)
	pastDue := fixedNow.AddDate(0, -1, 10)
	overdueInv, err := entities.NewInvoice(card, pastStart, pastEnd, pastClosing, pastDue, pastStart)
	require.NoError(t, err)

	clock.EXPECT().Now().Return(fixedNow)
	invRepo.EXPECT().List(mock.Anything, userID, mock.AnythingOfType("filters.InvoiceFilter")).Return([]entities.Invoice{*overdueInv}, int64(1), nil)
	invRepo.EXPECT().Update(mock.Anything, mock.AnythingOfType("*entities.Invoice")).Return(nil)

	pag, err := vos.NewPagination(1, 10)
	require.NoError(t, err)
	f, err := filters.NewInvoiceFilter(userID, nil, nil, nil, nil, pag)
	require.NoError(t, err)

	uc := newListInvoicesUC(t, mgr, invRepo, clock)
	resp, err := uc.Execute(ctx, userID, f)

	require.NoError(t, err)
	assert.Len(t, resp.Items, 1)
	assert.Equal(t, "closed", resp.Items[0].State)
	invRepo.AssertCalled(t, "Update", mock.Anything, mock.AnythingOfType("*entities.Invoice"))
}

func TestListInvoices_Empty_ReturnsEmptyPage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()

	mgr, _ := newMockMgr(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	clock := portmocks.NewClock(t)

	clock.EXPECT().Now().Return(fixedNow)
	invRepo.EXPECT().List(mock.Anything, userID, mock.AnythingOfType("filters.InvoiceFilter")).Return([]entities.Invoice{}, int64(0), nil)

	pag, err := vos.NewPagination(1, 10)
	require.NoError(t, err)
	f, err := filters.NewInvoiceFilter(userID, nil, nil, nil, nil, pag)
	require.NoError(t, err)

	uc := newListInvoicesUC(t, mgr, invRepo, clock)
	resp, err := uc.Execute(ctx, userID, f)

	require.NoError(t, err)
	assert.Empty(t, resp.Items)
	assert.Equal(t, int64(0), resp.Total)
}
