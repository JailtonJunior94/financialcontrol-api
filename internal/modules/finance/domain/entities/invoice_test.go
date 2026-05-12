package entities

import (
	"testing"
	"time"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/projections"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeTestCard() projections.CardView {
	return projections.CardView{
		ID:           vos.NewCardID(),
		UserID:       identityvo.NewUserID(),
		FlagName:     "Visa",
		ClosingDay:   10,
		DueDay:       15,
		BillingCycle: 30,
		Active:       true,
	}
}

func TestNewInvoice_Open(t *testing.T) {
	t.Parallel()
	card := makeTestCard()
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	cycleStart := time.Date(2026, 4, 11, 0, 0, 0, 0, time.UTC)
	cycleEnd := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)
	closingDate := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)
	dueDate := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)

	inv, err := NewInvoice(card, cycleStart, cycleEnd, closingDate, dueDate, now)

	require.NoError(t, err)
	require.NotNil(t, inv)
	assert.True(t, inv.IsOpen())
	assert.False(t, inv.IsClosed())
	assert.False(t, inv.IsPaid())
	assert.Equal(t, card.ID, inv.CardID())
	assert.Equal(t, card.UserID, inv.UserID())
	assert.Nil(t, inv.PaidAt())
}

func TestInvoice_CloseIfDue(t *testing.T) {
	t.Parallel()
	card := makeTestCard()
	closingDate := time.Date(2026, 5, 10, 3, 0, 0, 0, time.UTC) // 00:00 SP (UTC-3)

	tests := []struct {
		name        string
		now         time.Time
		wantChanged bool
		wantClosed  bool
	}{
		{
			name:        "before closing date stays open",
			now:         time.Date(2026, 5, 9, 23, 59, 59, 0, time.UTC),
			wantChanged: false,
			wantClosed:  false,
		},
		{
			name:        "at closing date closes",
			now:         closingDate,
			wantChanged: true,
			wantClosed:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// use a fresh invoice for each sub-test
			fresh, _ := NewInvoice(card,
				time.Date(2026, 4, 11, 0, 0, 0, 0, time.UTC),
				closingDate,
				closingDate,
				time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC),
				time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
			)
			changed := fresh.CloseIfDue(tc.now)
			assert.Equal(t, tc.wantChanged, changed)
			assert.Equal(t, tc.wantClosed, fresh.IsClosed())
		})
	}
}

func TestInvoice_CloseIfDue_Idempotent(t *testing.T) {
	t.Parallel()
	card := makeTestCard()
	closingDate := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)
	inv, _ := NewInvoice(card,
		time.Date(2026, 4, 11, 0, 0, 0, 0, time.UTC),
		closingDate, closingDate,
		time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
	)
	now := time.Date(2026, 5, 11, 0, 0, 0, 0, time.UTC)
	inv.CloseIfDue(now)

	// Second call: already closed → returns false
	changed := inv.CloseIfDue(now)
	assert.False(t, changed)
	assert.True(t, inv.IsClosed())
}

func TestInvoice_MarkPaid(t *testing.T) {
	t.Parallel()
	card := makeTestCard()
	closingDate := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)
	baseNow := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	at := time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC)

	newInv := func() *Invoice {
		inv, _ := NewInvoice(card,
			time.Date(2026, 4, 11, 0, 0, 0, 0, time.UTC),
			closingDate, closingDate,
			time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC),
			baseNow,
		)
		return inv
	}

	t.Run("closed_to_paid", func(t *testing.T) {
		inv := newInv()
		inv.CloseIfDue(closingDate)
		require.True(t, inv.IsClosed())
		err := inv.MarkPaid(at)
		assert.NoError(t, err)
		assert.True(t, inv.IsPaid())
		assert.NotNil(t, inv.PaidAt())
	})

	t.Run("open_to_paid_with_implicit_close", func(t *testing.T) {
		inv := newInv()
		require.True(t, inv.IsOpen())
		err := inv.MarkPaid(at)
		assert.NoError(t, err)
		assert.True(t, inv.IsPaid())
	})

	t.Run("paid_to_paid_returns_409_sentinel", func(t *testing.T) {
		inv := newInv()
		inv.CloseIfDue(closingDate)
		_ = inv.MarkPaid(at)
		err := inv.MarkPaid(at)
		assert.ErrorIs(t, err, domain.ErrInvoiceAlreadyPaid)
	})
}

func TestInvoice_RecalculateTotal(t *testing.T) {
	t.Parallel()
	card := makeTestCard()
	now := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	inv, _ := NewInvoice(card, now, now, now, now, now)

	number, _ := vos.NewInstallmentNumber(1)
	total, _ := vos.NewInstallmentCount(2)
	amount1, _ := vos.NewMoney("50.00")
	amount2, _ := vos.NewMoney("33.33")

	items := []*Installment{
		newInstallment(vos.NewInstallmentID(), vos.NewTransactionID(), inv.ID(), number, total, amount1, now),
		newInstallment(vos.NewInstallmentID(), vos.NewTransactionID(), inv.ID(), number, total, amount2, now),
	}
	inv.RecalculateTotal(items)

	expected, _ := vos.NewMoney("83.33")
	assert.True(t, inv.Total().Equal(expected))
}
