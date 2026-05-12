package usecase_test

import (
	"context"
	"testing"
	"time"

	devkitdb "github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/stretchr/testify/mock"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/projections"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

// --- time helpers ---

var fixedNow = time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)

// --- mock manager / tx ---

type mockTx struct{ mock.Mock }

func (m *mockTx) ExecContext(_ context.Context, _ string, _ ...any) (devkitdb.Result, error) {
	return nil, nil
}
func (m *mockTx) QueryContext(_ context.Context, _ string, _ ...any) (devkitdb.Rows, error) {
	return nil, nil
}
func (m *mockTx) QueryRowContext(_ context.Context, _ string, _ ...any) devkitdb.Row { return nil }
func (m *mockTx) Commit(ctx context.Context) error                                   { return m.Called(ctx).Error(0) }
func (m *mockTx) Rollback(ctx context.Context) error                                 { return m.Called(ctx).Error(0) }

type mockManager struct{ mock.Mock }

func (m *mockManager) Driver() devkitdb.Driver { return m.Called().Get(0).(devkitdb.Driver) }
func (m *mockManager) DBTX(ctx context.Context) devkitdb.DBTX {
	ret := m.Called(ctx)
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(devkitdb.DBTX)
}
func (m *mockManager) BeginTx(ctx context.Context, opts devkitdb.TxOptions) (devkitdb.Tx, error) {
	ret := m.Called(ctx, opts)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(devkitdb.Tx), ret.Error(1)
}
func (m *mockManager) Ping(ctx context.Context) error     { return m.Called(ctx).Error(0) }
func (m *mockManager) Shutdown(ctx context.Context) error { return m.Called(ctx).Error(0) }

// newMockMgr creates a manager that successfully opens and commits a tx.
func newMockMgr(t *testing.T) (*mockManager, *mockTx) {
	t.Helper()
	tx := &mockTx{}
	mgr := &mockManager{}
	mgr.On("BeginTx", mock.Anything, devkitdb.TxOptions{}).Return(tx, nil)
	tx.On("Commit", mock.Anything).Return(nil)
	return mgr, tx
}

// --- domain helpers ---

func mustAmount(t *testing.T, s string) vos.Amount {
	t.Helper()
	m, err := vos.NewMoney(s)
	if err != nil {
		t.Fatalf("mustAmount: %v", err)
	}
	a, err := vos.NewAmountFromMoney(m)
	if err != nil {
		t.Fatalf("mustAmount: %v", err)
	}
	return a
}

// newExpenseTransaction builds a minimal non-card expense transaction for tests.
func newExpenseTransaction(t *testing.T, userID identityvo.UserID) *entities.Transaction {
	t.Helper()
	catID := vos.NewCategoryID()
	a := mustAmount(t, "100.00")
	tx, err := entities.NewTransaction(
		vos.NewTransactionID(),
		userID,
		"Test expense",
		a,
		fixedNow,
		vos.TransactionTypeExpense,
		vos.PaymentMethodPix,
		nil,
		catID,
		nil,
		nil,
		&fixedClockImpl{fixedNow},
	)
	if err != nil {
		t.Fatalf("newExpenseTransaction: %v", err)
	}
	return tx
}

type fixedClockImpl struct{ t time.Time }

func (c *fixedClockImpl) Now() time.Time { return c.t }

// newCreditPurchaseTransaction builds a minimal credit_purchase transaction with a card for tests.
func newCreditPurchaseTransaction(t *testing.T, userID identityvo.UserID, cardID vos.CardID) *entities.Transaction {
	t.Helper()
	catID := vos.NewCategoryID()
	a := mustAmount(t, "100.00")
	cid := cardID
	tx, err := entities.NewTransaction(
		vos.NewTransactionID(),
		userID,
		"Test credit purchase",
		a,
		fixedNow,
		vos.TransactionTypeCreditPurchase,
		vos.PaymentMethodCreditCard,
		&cid,
		catID,
		nil,
		nil,
		&fixedClockImpl{fixedNow},
	)
	if err != nil {
		t.Fatalf("newCreditPurchaseTransaction: %v", err)
	}
	return tx
}

// newScheduledInstallment builds a SCHEDULED installment for a given transaction and invoice.
func newScheduledInstallment(instID vos.InstallmentID, txID vos.TransactionID, invoiceID vos.InvoiceID) entities.Installment {
	num, _ := vos.NewInstallmentNumber(1)
	count, _ := vos.NewInstallmentCount(1)
	amt, _ := vos.NewMoney("100.00")
	inst := entities.RehydrateInstallment(instID, txID, invoiceID, num, count, amt, vos.InstallmentStatusScheduled, nil, fixedNow, fixedNow, nil)
	return *inst
}

// newActiveCardView builds a valid active CardView.
func newActiveCardView(userID identityvo.UserID, cardID vos.CardID) projections.CardView {
	return projections.CardView{
		ID:           cardID,
		UserID:       userID,
		FlagName:     "Visa",
		ClosingDay:   10,
		DueDay:       20,
		BillingCycle: 30,
		Active:       true,
	}
}

// newActiveCategoryView builds a valid active CategoryView (root category).
func newActiveCategoryView(userID identityvo.UserID, catID vos.CategoryID) projections.CategoryView {
	return projections.CategoryView{
		ID:       catID,
		UserID:   userID,
		Name:     "Food",
		ParentID: nil,
		Active:   true,
	}
}

// newOpenInvoice builds a simple open invoice for tests.
func newOpenInvoice(t *testing.T, userID identityvo.UserID, cardID vos.CardID) *entities.Invoice {
	t.Helper()
	card := newActiveCardView(userID, cardID)
	cycleStart := fixedNow.AddDate(0, 0, -5)
	cycleEnd := fixedNow.AddDate(0, 1, 0)
	closingDate := fixedNow.AddDate(0, 1, 10)
	dueDate := fixedNow.AddDate(0, 1, 20)
	inv, err := entities.NewInvoice(card, cycleStart, cycleEnd, closingDate, dueDate, fixedNow)
	if err != nil {
		t.Fatalf("newOpenInvoice: %v", err)
	}
	return inv
}

// newClosedInvoice builds an invoice already past its closing date (state=closed).
func newClosedInvoice(t *testing.T, userID identityvo.UserID, cardID vos.CardID) *entities.Invoice {
	t.Helper()
	inv := newOpenInvoice(t, userID, cardID)
	// Force state=closed by calling CloseIfDue with a time past closing.
	if !inv.CloseIfDue(inv.ClosingDate().Add(time.Hour)) {
		t.Fatalf("newClosedInvoice: CloseIfDue did not transition open invoice")
	}
	return inv
}

// newPaidInvoice builds an invoice in state=paid for tests.
func newPaidInvoice(t *testing.T, userID identityvo.UserID, cardID vos.CardID) *entities.Invoice {
	t.Helper()
	inv := newClosedInvoice(t, userID, cardID)
	if err := inv.MarkPaid(inv.ClosingDate().Add(2 * time.Hour)); err != nil {
		t.Fatalf("newPaidInvoice: MarkPaid failed: %v", err)
	}
	return inv
}
