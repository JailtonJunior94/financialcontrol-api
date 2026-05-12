package entities

import (
	"testing"
	"time"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockClock is a deterministic Clock for tests.
type mockClock struct{ now time.Time }

func (m *mockClock) Now() time.Time { return m.now }

// mockSplitter returns fixed amounts and IDs.
type mockSplitter struct {
	amounts []vos.Money
	ids     []vos.InstallmentID
}

func (m *mockSplitter) Split(_ vos.Money, _ vos.InstallmentCount) ([]vos.Money, []vos.InstallmentID) {
	return m.amounts, m.ids
}

// mockAssigner always returns a fixed invoice ID.
type mockAssigner struct{ invoiceID vos.InvoiceID }

func (m *mockAssigner) InvoiceForNumber(_ int) (vos.InvoiceID, error) { return m.invoiceID, nil }

func newTestTransaction(t *testing.T, now time.Time) *Transaction {
	t.Helper()
	clock := &mockClock{now: now}
	catID := vos.NewCategoryID()
	pm := vos.PaymentMethodPix
	tx, err := NewTransaction(
		vos.NewTransactionID(),
		identityvo.NewUserID(),
		"Test description",
		mustAmount(t, "100.00"),
		now,
		vos.TransactionTypeExpense,
		pm,
		nil,
		catID,
		nil,
		nil,
		clock,
	)
	require.NoError(t, err)
	return tx
}

func mustAmount(t *testing.T, s string) vos.Amount {
	t.Helper()
	m, err := vos.NewMoney(s)
	require.NoError(t, err)
	a, err := vos.NewAmountFromMoney(m)
	require.NoError(t, err)
	return a
}

func TestNewTransaction_ValidInvariants(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	tx := newTestTransaction(t, now)
	assert.NotEmpty(t, tx.ID())
	assert.Equal(t, "Test description", tx.Description())
	assert.False(t, tx.IsRefund())
	assert.False(t, tx.IsCardPurchase())
}

func TestNewTransaction_Validations(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	clock := &mockClock{now: now}
	catID := vos.NewCategoryID()
	amount := mustAmount(t, "100.00")

	tests := []struct {
		name        string
		description string
		pm          vos.PaymentMethod
		cardID      *vos.CardID
		occurredAt  time.Time
		wantErr     error
	}{
		{
			name:        "empty description",
			description: "   ",
			pm:          vos.PaymentMethodPix,
			occurredAt:  now,
			wantErr:     domain.ErrDescriptionRequired,
		},
		{
			name:        "description too long",
			description: string(make([]byte, 256)),
			pm:          vos.PaymentMethodPix,
			occurredAt:  now,
			wantErr:     domain.ErrDescriptionRequired,
		},
		{
			name:        "credit_card without card_id",
			description: "Compra",
			pm:          vos.PaymentMethodCreditCard,
			cardID:      nil,
			occurredAt:  now,
			wantErr:     domain.ErrCardRequiredForPaymentMethod,
		},
		{
			name:        "occurred_at too far in future",
			description: "Futura",
			pm:          vos.PaymentMethodPix,
			occurredAt:  now.AddDate(0, 0, 366),
			wantErr:     domain.ErrOccurredAtTooFarInFuture,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewTransaction(
				vos.NewTransactionID(),
				identityvo.NewUserID(),
				tc.description,
				amount,
				tc.occurredAt,
				vos.TransactionTypeExpense,
				tc.pm,
				tc.cardID,
				catID,
				nil,
				nil,
				clock,
			)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestTransaction_IsRefund_IsCardPurchase(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	clock := &mockClock{now: now}
	amount := mustAmount(t, "50.00")
	catID := vos.NewCategoryID()
	cardID := vos.NewCardID()

	tests := []struct {
		name         string
		txType       vos.TransactionType
		pm           vos.PaymentMethod
		cardID       *vos.CardID
		wantIsRefund bool
		wantIsCard   bool
	}{
		{
			name:         "refund",
			txType:       vos.TransactionTypeRefund,
			pm:           vos.PaymentMethodPix,
			wantIsRefund: true,
			wantIsCard:   false,
		},
		{
			name:       "credit_purchase",
			txType:     vos.TransactionTypeCreditPurchase,
			pm:         vos.PaymentMethodCreditCard,
			cardID:     &cardID,
			wantIsCard: true,
		},
		{
			name:   "income",
			txType: vos.TransactionTypeIncome,
			pm:     vos.PaymentMethodPix,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tx, err := NewTransaction(
				vos.NewTransactionID(),
				identityvo.NewUserID(),
				"desc",
				amount,
				now,
				tc.txType,
				tc.pm,
				tc.cardID,
				catID,
				nil,
				nil,
				clock,
			)
			require.NoError(t, err)
			assert.Equal(t, tc.wantIsRefund, tx.IsRefund())
			assert.Equal(t, tc.wantIsCard, tx.IsCardPurchase())
		})
	}
}

func TestTransaction_SoftDelete(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name                       string
		hasClosedOrPaidInstallment bool
		hasDependentRefund         bool
		wantErr                    error
	}{
		{
			name:                       "blocks when has closed/paid installment",
			hasClosedOrPaidInstallment: true,
			wantErr:                    domain.ErrInstallmentInClosedOrPaidInvoice,
		},
		{
			name:               "blocks when has dependent refund",
			hasDependentRefund: true,
			wantErr:            domain.ErrTransactionHasDependentRefund,
		},
		{
			name:    "soft deletes when no guards triggered",
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tx := newTestTransaction(t, now)
			err := tx.SoftDelete(now, tc.hasClosedOrPaidInstallment, tc.hasDependentRefund)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				assert.False(t, tx.IsDeleted())
				return
			}
			assert.NoError(t, err)
			assert.True(t, tx.IsDeleted())
		})
	}
}

func TestTransaction_Replace_RejectsClosedOrPaidInstallment(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	tx := newTestTransaction(t, now)

	// Attach a paid_via_invoice installment
	number, _ := vos.NewInstallmentNumber(1)
	total, _ := vos.NewInstallmentCount(1)
	amount, _ := vos.NewMoney("100.00")
	inst := RehydrateInstallment(
		vos.NewInstallmentID(), tx.ID(), vos.NewInvoiceID(),
		number, total, amount,
		vos.InstallmentStatusPaidViaInvoice, nil, now, now, nil,
	)
	tx.SetInstallments([]*Installment{inst})

	input := ReplaceInput{
		Description:     "New desc",
		Amount:          mustAmount(t, "100.00"),
		OccurredAt:      now,
		TransactionType: vos.TransactionTypeExpense,
		PaymentMethod:   vos.PaymentMethodPix,
		CategoryID:      vos.NewCategoryID(),
		Now:             now,
	}
	err := tx.Replace(input, nil, nil)
	assert.ErrorIs(t, err, domain.ErrInstallmentInClosedOrPaidInvoice)
}

func TestTransaction_Replace_RecreatestInstallmentsForInstallmentPurchase(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	clock := &mockClock{now: now}
	cardID := vos.NewCardID()
	catID := vos.NewCategoryID()
	amount := mustAmount(t, "90.00")

	tx, err := NewTransaction(
		vos.NewTransactionID(),
		identityvo.NewUserID(),
		"Purchase",
		amount,
		now,
		vos.TransactionTypeInstallmentPurchase,
		vos.PaymentMethodCreditCard,
		&cardID,
		catID,
		nil,
		nil,
		clock,
	)
	require.NoError(t, err)

	// Attach 1 old installment (scheduled)
	number, _ := vos.NewInstallmentNumber(1)
	total1, _ := vos.NewInstallmentCount(1)
	m, _ := vos.NewMoney("90.00")
	oldInst := newInstallment(vos.NewInstallmentID(), tx.ID(), vos.NewInvoiceID(), number, total1, m, now)
	tx.SetInstallments([]*Installment{oldInst})

	invID := vos.NewInvoiceID()
	splitter := &mockSplitter{
		amounts: []vos.Money{mustMoney(t, "30.00"), mustMoney(t, "30.00"), mustMoney(t, "30.00")},
		ids:     []vos.InstallmentID{vos.NewInstallmentID(), vos.NewInstallmentID(), vos.NewInstallmentID()},
	}
	assigner := &mockAssigner{invoiceID: invID}

	input := ReplaceInput{
		Description:      "Updated Purchase",
		Amount:           mustAmount(t, "90.00"),
		OccurredAt:       now,
		TransactionType:  vos.TransactionTypeInstallmentPurchase,
		PaymentMethod:    vos.PaymentMethodCreditCard,
		CardID:           &cardID,
		CategoryID:       catID,
		InstallmentCount: 3,
		Now:              now,
	}
	err = tx.Replace(input, splitter, assigner)
	require.NoError(t, err)

	allInst := tx.Installments()
	// 1 old (refunded/deleted) + 3 new
	assert.Len(t, allInst, 4)

	// old installment should be refunded and soft-deleted
	assert.Equal(t, vos.InstallmentStatusRefunded, allInst[0].Status())
	assert.NotNil(t, allInst[0].DeletedAt())

	// new installments should be scheduled
	for i := 1; i <= 3; i++ {
		assert.Equal(t, vos.InstallmentStatusScheduled, allInst[i].Status())
		assert.Equal(t, invID, allInst[i].InvoiceID())
	}
}

func TestTransaction_Anticipate(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	tx := newTestTransaction(t, now)

	card := makeTestCard()
	closingDate := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	targetInvoice, _ := NewInvoice(card, now, closingDate, closingDate,
		time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC), now)

	number, _ := vos.NewInstallmentNumber(1)
	total, _ := vos.NewInstallmentCount(1)
	m, _ := vos.NewMoney("100.00")
	instID := vos.NewInstallmentID()
	inst := newInstallment(instID, tx.ID(), vos.NewInvoiceID(), number, total, m, now)
	tx.SetInstallments([]*Installment{inst})

	t.Run("anticipate scheduled installment to open invoice", func(t *testing.T) {
		err := tx.Anticipate(instID, targetInvoice, now)
		assert.NoError(t, err)
		assert.Equal(t, vos.InstallmentStatusAnticipated, inst.Status())
		assert.Equal(t, targetInvoice.ID(), inst.InvoiceID())
	})

	t.Run("not found installment returns error", func(t *testing.T) {
		tx2 := newTestTransaction(t, now)
		err := tx2.Anticipate(vos.NewInstallmentID(), targetInvoice, now)
		assert.ErrorIs(t, err, domain.ErrInstallmentNotFound)
	})

	t.Run("already paid installment cannot be anticipated", func(t *testing.T) {
		tx3 := newTestTransaction(t, now)
		paidInst := RehydrateInstallment(
			vos.NewInstallmentID(), tx3.ID(), vos.NewInvoiceID(),
			number, total, m,
			vos.InstallmentStatusPaidViaInvoice, nil, now, now, nil,
		)
		tx3.SetInstallments([]*Installment{paidInst})
		err := tx3.Anticipate(paidInst.ID(), targetInvoice, now)
		assert.ErrorIs(t, err, domain.ErrInstallmentInClosedOrPaidInvoice)
	})
}

func mustMoney(t *testing.T, s string) vos.Money {
	t.Helper()
	m, err := vos.NewMoney(s)
	require.NoError(t, err)
	return m
}

// BUG-003 regression: Replace must reject when the caller signals that at least
// one installment is bound to a closed/paid invoice — even though every
// installment carries status=scheduled (RF-46 keeps status='scheduled' while the
// invoice is closed-but-unpaid, so the previous installment-status-only guard
// missed this case for RF-11/RF-54).
func TestTransaction_Replace_RejectsWhenHasClosedOrPaidInvoiceFlag(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	tx := newTestTransaction(t, now)

	number, _ := vos.NewInstallmentNumber(1)
	total, _ := vos.NewInstallmentCount(1)
	amount, _ := vos.NewMoney("100.00")
	scheduledInst := RehydrateInstallment(
		vos.NewInstallmentID(), tx.ID(), vos.NewInvoiceID(),
		number, total, amount,
		vos.InstallmentStatusScheduled, nil, now, now, nil,
	)
	tx.SetInstallments([]*Installment{scheduledInst})

	input := ReplaceInput{
		Description:            "New desc",
		Amount:                 mustAmount(t, "100.00"),
		OccurredAt:             now,
		TransactionType:        vos.TransactionTypeExpense,
		PaymentMethod:          vos.PaymentMethodPix,
		CategoryID:             vos.NewCategoryID(),
		Now:                    now,
		HasClosedOrPaidInvoice: true,
	}
	err := tx.Replace(input, nil, nil)
	assert.ErrorIs(t, err, domain.ErrInstallmentInClosedOrPaidInvoice)
}

// BUG-005 regression: Replace must soft-delete pre-existing installments when
// transitioning a transaction from installment_purchase to a non-card type so
// the update use case doesn't try to re-insert them (PK conflict in AddBatch).
func TestTransaction_Replace_SoftDeletesInstallmentsOnTypeTransitionAwayFromInstallmentPurchase(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	clock := &mockClock{now: now}
	cardID := vos.NewCardID()
	catID := vos.NewCategoryID()
	amount := mustAmount(t, "90.00")

	tx, err := NewTransaction(
		vos.NewTransactionID(),
		identityvo.NewUserID(),
		"Original installment",
		amount,
		now,
		vos.TransactionTypeInstallmentPurchase,
		vos.PaymentMethodCreditCard,
		&cardID,
		catID,
		nil,
		nil,
		clock,
	)
	require.NoError(t, err)

	number, _ := vos.NewInstallmentNumber(1)
	total, _ := vos.NewInstallmentCount(3)
	m, _ := vos.NewMoney("30.00")
	oldInst := newInstallment(vos.NewInstallmentID(), tx.ID(), vos.NewInvoiceID(), number, total, m, now)
	tx.SetInstallments([]*Installment{oldInst})

	input := ReplaceInput{
		Description:     "Converted to PIX",
		Amount:          mustAmount(t, "90.00"),
		OccurredAt:      now,
		TransactionType: vos.TransactionTypeExpense,
		PaymentMethod:   vos.PaymentMethodPix,
		CategoryID:      catID,
		Now:             now,
	}
	err = tx.Replace(input, nil, nil)
	require.NoError(t, err)

	allInst := tx.Installments()
	require.Len(t, allInst, 1)
	assert.Equal(t, vos.InstallmentStatusRefunded, allInst[0].Status())
	assert.NotNil(t, allInst[0].DeletedAt())
}

// BUG-008 regression: Replace must surface a contract error instead of silently
// no-oping when transitioning to installment_purchase without splitter/assigner.
func TestTransaction_Replace_RequiresSplitterAndAssignerForInstallmentPurchase(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	clock := &mockClock{now: now}
	cardID := vos.NewCardID()
	catID := vos.NewCategoryID()
	amount := mustAmount(t, "90.00")

	tx, err := NewTransaction(
		vos.NewTransactionID(),
		identityvo.NewUserID(),
		"Purchase",
		amount,
		now,
		vos.TransactionTypeExpense,
		vos.PaymentMethodPix,
		nil,
		catID,
		nil,
		nil,
		clock,
	)
	require.NoError(t, err)

	input := ReplaceInput{
		Description:      "Now installment_purchase",
		Amount:           amount,
		OccurredAt:       now,
		TransactionType:  vos.TransactionTypeInstallmentPurchase,
		PaymentMethod:    vos.PaymentMethodCreditCard,
		CardID:           &cardID,
		CategoryID:       catID,
		InstallmentCount: 3,
		Now:              now,
	}
	err = tx.Replace(input, nil, nil)
	assert.ErrorIs(t, err, domain.ErrSplitterRequired)
}
