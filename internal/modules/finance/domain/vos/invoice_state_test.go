package vos_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
)

func TestInvoiceStateCanTransitionTo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		from  vos.InvoiceState
		to    vos.InvoiceState
		canDo bool
	}{
		// open transitions
		{"open → closed", vos.InvoiceStateOpen, vos.InvoiceStateClosed, true},
		{"open → paid (implicit close)", vos.InvoiceStateOpen, vos.InvoiceStatePaid, true},
		{"open → open (no-op, blocked)", vos.InvoiceStateOpen, vos.InvoiceStateOpen, false},
		// closed transitions
		{"closed → paid", vos.InvoiceStateClosed, vos.InvoiceStatePaid, true},
		{"closed → open (backward, blocked)", vos.InvoiceStateClosed, vos.InvoiceStateOpen, false},
		{"closed → closed (no-op, blocked)", vos.InvoiceStateClosed, vos.InvoiceStateClosed, false},
		// paid transitions — terminal state
		{"paid → open (blocked)", vos.InvoiceStatePaid, vos.InvoiceStateOpen, false},
		{"paid → closed (blocked)", vos.InvoiceStatePaid, vos.InvoiceStateClosed, false},
		{"paid → paid (blocked)", vos.InvoiceStatePaid, vos.InvoiceStatePaid, false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.canDo, tc.from.CanTransitionTo(tc.to))
		})
	}
}
