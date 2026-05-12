package services

import (
	"testing"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/projections"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeCard(closingDay int) projections.CardView {
	return projections.CardView{
		ID:           vos.NewCardID(),
		UserID:       identityvo.NewUserID(),
		ClosingDay:   closingDay,
		DueDay:       closingDay + 5,
		BillingCycle: 30,
		Active:       true,
	}
}

func makeOpenInvoice(card projections.CardView, closingDateSP time.Time) *entities.Invoice {
	saoPaulo, _ := time.LoadLocation("America/Sao_Paulo")
	// closing date stored as UTC midnight of that day in SP
	closingUTC := time.Date(closingDateSP.Year(), closingDateSP.Month(), closingDateSP.Day(), 3, 0, 0, 0, time.UTC)
	_ = saoPaulo
	inv, _ := entities.NewInvoice(
		card,
		closingUTC.AddDate(0, -1, 0),
		closingUTC,
		closingUTC,
		closingUTC.AddDate(0, 0, 5),
		time.Now(),
	)
	return inv
}

func TestInvoiceAssigner_AssignFor_Cutoff(t *testing.T) {
	t.Parallel()
	svc := NewInvoiceAssigner()
	saoPaulo, _ := time.LoadLocation("America/Sao_Paulo")

	// Card closing day = 10
	card := makeCard(10)

	// Prepare two open invoices: May 10 closing and June 10 closing (both in SP)
	mayClosingSP := time.Date(2026, 5, 10, 0, 0, 0, 0, saoPaulo)
	junClosingSP := time.Date(2026, 6, 10, 0, 0, 0, 0, saoPaulo)
	mayInv := makeOpenInvoice(card, mayClosingSP)
	junInv := makeOpenInvoice(card, junClosingSP)
	openInvoices := []*entities.Invoice{mayInv, junInv}

	tests := []struct {
		name        string
		occurredAt  time.Time // UTC
		wantNil     bool
		wantClosing time.Time // in SP
	}{
		{
			// 23:59:59 UTC → 20:59:59 SP on same day (May 9) → day=9 < closingDay=10 → May cycle
			name:        "23:59:59 UTC on May 9 → May 9 SP → current (May) cycle",
			occurredAt:  time.Date(2026, 5, 9, 23, 59, 59, 0, time.UTC),
			wantClosing: mayClosingSP,
		},
		{
			// 02:00:00 UTC on May 10 → 23:00:00 SP on May 9 → day=9 < closingDay=10 → May cycle
			name:        "02:00:00 UTC on May 10 → 23:00:00 SP May 9 → current (May) cycle",
			occurredAt:  time.Date(2026, 5, 10, 2, 0, 0, 0, time.UTC),
			wantClosing: mayClosingSP,
		},
		{
			// 02:00:00 UTC on May 11 → 23:00:00 SP on May 10 → day=10 >= closingDay=10 → June cycle
			name:        "02:00:00 UTC on May 11 → 23:00:00 SP May 10 → next (June) cycle",
			occurredAt:  time.Date(2026, 5, 11, 2, 0, 0, 0, time.UTC),
			wantClosing: junClosingSP,
		},
		{
			// Day >= closingDay in SP → next cycle (June)
			name:        "May 15 SP (day >= 10) → June cycle",
			occurredAt:  time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC),
			wantClosing: junClosingSP,
		},
		{
			// Day < closingDay → current (May) cycle
			name:        "May 5 SP (day < 10) → May cycle",
			occurredAt:  time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC),
			wantClosing: mayClosingSP,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			inv, err := svc.AssignFor(card, tc.occurredAt, openInvoices, nil)
			require.NoError(t, err)
			if tc.wantNil {
				assert.Nil(t, inv)
				return
			}
			require.NotNil(t, inv)
			closingSP := inv.ClosingDate().In(saoPaulo)
			assert.True(t, sameDate(closingSP, tc.wantClosing),
				"got %s want %s", closingSP.Format("2006-01-02"), tc.wantClosing.Format("2006-01-02"))
		})
	}
}

func TestInvoiceAssigner_AssignFor_NoMatchReturnsNil(t *testing.T) {
	t.Parallel()
	svc := NewInvoiceAssigner()
	card := makeCard(10)
	// No open invoices
	inv, err := svc.AssignFor(card, time.Now(), nil, nil)
	assert.NoError(t, err)
	assert.Nil(t, inv)
}
