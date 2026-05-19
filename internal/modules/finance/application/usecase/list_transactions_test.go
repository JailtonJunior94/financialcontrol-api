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
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/services"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

func newListTransactionsUC(
	t *testing.T,
	mgr *mockManager,
	txRepo *portmocks.TransactionRepository,
	invRepo *portmocks.InvoiceRepository,
	clock *portmocks.Clock,
) usecase.ListTransactions {
	t.Helper()
	closer := services.NewInvoiceCloser()
	return usecase.NewListTransactions(mgr, txRepo, invRepo, closer, clock)
}

func mustPagination(t *testing.T) vos.Pagination {
	t.Helper()
	pag, err := vos.NewPagination(1, 10)
	if err != nil {
		t.Fatalf("mustPagination: %v", err)
	}
	return pag
}

func TestListTransactions_GoldenPath_NoLazyClose(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	clock := portmocks.NewClock(t)

	clock.EXPECT().Now().Return(fixedNow)

	// open invoices query returns empty — no lazy close
	invRepo.EXPECT().List(mock.Anything, userID, mock.AnythingOfType("filters.InvoiceFilter")).Return([]entities.Invoice{}, int64(0), nil)

	tx := newExpenseTransaction(t, userID)
	txs := []entities.Transaction{*tx}
	txRepo.EXPECT().List(mock.Anything, userID, mock.AnythingOfType("filters.TransactionFilter")).Return(txs, int64(1), nil)

	pag := mustPagination(t)
	f, err := filters.NewTransactionFilter(userID, nil, nil, nil, nil, vos.InvoiceStatusFilterNone, "", nil, nil, pag)
	require.NoError(t, err)

	uc := newListTransactionsUC(t, mgr, txRepo, invRepo, clock)
	resp, err := uc.Execute(ctx, userID, f)

	require.NoError(t, err)
	assert.Len(t, resp.Items, 1)
	assert.Equal(t, int64(1), resp.Total)
}

func TestListTransactions_LazyClose_UpdatesChangedInvoices(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	cardID := vos.NewCardID()

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	clock := portmocks.NewClock(t)

	// Build an open invoice whose closing date is BEFORE fixedNow so CloseIfDue triggers.
	card := newActiveCardView(userID, cardID)
	pastClosing := fixedNow.AddDate(0, -1, -5)
	pastEnd := fixedNow.AddDate(0, -1, 0)
	pastStart := fixedNow.AddDate(0, -2, 0)
	dueDate := fixedNow.AddDate(0, -1, 10)
	openInv, err := entities.NewInvoice(card, pastStart, pastEnd, pastClosing, dueDate, fixedNow.AddDate(0, -2, 0))
	require.NoError(t, err)

	clock.EXPECT().Now().Return(fixedNow)
	invRepo.EXPECT().List(mock.Anything, userID, mock.AnythingOfType("filters.InvoiceFilter")).Return([]entities.Invoice{*openInv}, int64(1), nil)
	invRepo.EXPECT().Update(mock.Anything, mock.AnythingOfType("*entities.Invoice")).Return(nil)

	txRepo.EXPECT().List(mock.Anything, userID, mock.AnythingOfType("filters.TransactionFilter")).Return([]entities.Transaction{}, int64(0), nil)

	pag := mustPagination(t)
	f, ferr := filters.NewTransactionFilter(userID, nil, nil, nil, nil, vos.InvoiceStatusFilterNone, "", nil, nil, pag)
	require.NoError(t, ferr)

	uc := newListTransactionsUC(t, mgr, txRepo, invRepo, clock)
	resp, execErr := uc.Execute(ctx, userID, f)

	require.NoError(t, execErr)
	assert.Empty(t, resp.Items)
	invRepo.AssertCalled(t, "Update", mock.Anything, mock.AnythingOfType("*entities.Invoice"))
}
