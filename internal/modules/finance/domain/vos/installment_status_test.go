package vos_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
)

func TestInstallmentStatusCanTransitionTo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		from  vos.InstallmentStatus
		to    vos.InstallmentStatus
		canDo bool
	}{
		// scheduled transitions
		{"scheduled → anticipated", vos.InstallmentStatusScheduled, vos.InstallmentStatusAnticipated, true},
		{"scheduled → paid_via_invoice", vos.InstallmentStatusScheduled, vos.InstallmentStatusPaidViaInvoice, true},
		{"scheduled → refunded", vos.InstallmentStatusScheduled, vos.InstallmentStatusRefunded, true},
		{"scheduled → scheduled (no-op)", vos.InstallmentStatusScheduled, vos.InstallmentStatusScheduled, false},
		// anticipated transitions
		{"anticipated → paid_via_invoice", vos.InstallmentStatusAnticipated, vos.InstallmentStatusPaidViaInvoice, true},
		{"anticipated → refunded", vos.InstallmentStatusAnticipated, vos.InstallmentStatusRefunded, true},
		{"anticipated → scheduled (backward, blocked)", vos.InstallmentStatusAnticipated, vos.InstallmentStatusScheduled, false},
		{"anticipated → anticipated (no-op)", vos.InstallmentStatusAnticipated, vos.InstallmentStatusAnticipated, false},
		// paid_via_invoice — terminal
		{"paid_via_invoice → scheduled (blocked)", vos.InstallmentStatusPaidViaInvoice, vos.InstallmentStatusScheduled, false},
		{"paid_via_invoice → anticipated (blocked)", vos.InstallmentStatusPaidViaInvoice, vos.InstallmentStatusAnticipated, false},
		{"paid_via_invoice → refunded (blocked)", vos.InstallmentStatusPaidViaInvoice, vos.InstallmentStatusRefunded, false},
		// refunded — terminal
		{"refunded → scheduled (blocked)", vos.InstallmentStatusRefunded, vos.InstallmentStatusScheduled, false},
		{"refunded → paid_via_invoice (blocked)", vos.InstallmentStatusRefunded, vos.InstallmentStatusPaidViaInvoice, false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.canDo, tc.from.CanTransitionTo(tc.to))
		})
	}
}
