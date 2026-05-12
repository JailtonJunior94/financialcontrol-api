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

func makeInvoice(closingDate time.Time) *entities.Invoice {
	card := projections.CardView{
		ID:     vos.NewCardID(),
		UserID: identityvo.NewUserID(),
	}
	inv, _ := entities.NewInvoice(card,
		closingDate.AddDate(0, -1, 0),
		closingDate,
		closingDate,
		closingDate.AddDate(0, 0, 5),
		time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
	)
	return inv
}

func TestInvoiceCloser_CloseIfDue(t *testing.T) {
	t.Parallel()
	svc := NewInvoiceCloser()
	closing := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		now         time.Time
		wantChanged int
	}{
		{
			name:        "before closing date: no invoices closed",
			now:         time.Date(2026, 5, 9, 23, 59, 59, 0, time.UTC),
			wantChanged: 0,
		},
		{
			name:        "at closing date: invoice closed",
			now:         closing,
			wantChanged: 1,
		},
		{
			name:        "after closing date: invoice closed",
			now:         time.Date(2026, 5, 11, 0, 0, 0, 0, time.UTC),
			wantChanged: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			inv := makeInvoice(closing)
			changed := svc.CloseIfDue([]*entities.Invoice{inv}, tc.now)
			assert.Len(t, changed, tc.wantChanged)
		})
	}
}

func TestInvoiceCloser_CloseIfDue_Idempotent(t *testing.T) {
	t.Parallel()
	svc := NewInvoiceCloser()
	closing := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)
	inv := makeInvoice(closing)
	now := time.Date(2026, 5, 11, 0, 0, 0, 0, time.UTC)

	first := svc.CloseIfDue([]*entities.Invoice{inv}, now)
	require.Len(t, first, 1)

	// Second call with already-closed invoice → returns empty slice
	second := svc.CloseIfDue([]*entities.Invoice{inv}, now)
	assert.Len(t, second, 0)
}
