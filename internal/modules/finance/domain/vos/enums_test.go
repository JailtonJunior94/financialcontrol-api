package vos_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
)

// ---- ParseInstallmentStatus ----

func TestParseInstallmentStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     string
		want    vos.InstallmentStatus
		wantErr error
	}{
		{"scheduled", "scheduled", vos.InstallmentStatusScheduled, nil},
		{"anticipated", "anticipated", vos.InstallmentStatusAnticipated, nil},
		{"paid_via_invoice", "paid_via_invoice", vos.InstallmentStatusPaidViaInvoice, nil},
		{"refunded", "refunded", vos.InstallmentStatusRefunded, nil},
		{"empty", "", vos.InstallmentStatus(""), domain.ErrInvalidTransactionType},
		{"unknown", "unknown_status", vos.InstallmentStatus(""), domain.ErrInvalidTransactionType},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := vos.ParseInstallmentStatus(tc.raw)
			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr))
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.raw, got.String())
		})
	}
}

func TestInstallmentStatusIsClosedOrPaid(t *testing.T) {
	t.Parallel()

	assert.True(t, vos.InstallmentStatusPaidViaInvoice.IsClosedOrPaid())
	assert.False(t, vos.InstallmentStatusScheduled.IsClosedOrPaid())
	assert.False(t, vos.InstallmentStatusAnticipated.IsClosedOrPaid())
	assert.False(t, vos.InstallmentStatusRefunded.IsClosedOrPaid())
}

// ---- ParseInvoiceState ----

func TestParseInvoiceState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     string
		want    vos.InvoiceState
		wantErr error
	}{
		{"open", "open", vos.InvoiceStateOpen, nil},
		{"closed", "closed", vos.InvoiceStateClosed, nil},
		{"paid", "paid", vos.InvoiceStatePaid, nil},
		{"empty", "", vos.InvoiceState(""), domain.ErrInvalidTransactionType},
		{"unknown", "draft", vos.InvoiceState(""), domain.ErrInvalidTransactionType},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := vos.ParseInvoiceState(tc.raw)
			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr))
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.raw, got.String())
		})
	}
}

// ---- ParseInvoiceStatusFilter ----

func TestParseInvoiceStatusFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		raw      string
		want     vos.InvoiceStatusFilter
		isActive bool
		wantErr  error
	}{
		{"none explicit", "none", vos.InvoiceStatusFilterNone, false, nil},
		{"empty treated as none", "", vos.InvoiceStatusFilterNone, false, nil},
		{"open", "open", vos.InvoiceStatusFilterOpen, true, nil},
		{"closed", "closed", vos.InvoiceStatusFilterClosed, true, nil},
		{"paid", "paid", vos.InvoiceStatusFilterPaid, true, nil},
		{"invalid", "draft", vos.InvoiceStatusFilter(""), false, domain.ErrInvalidTransactionType},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := vos.ParseInvoiceStatusFilter(tc.raw)
			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr))
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.isActive, got.IsActive())
		})
	}
}
