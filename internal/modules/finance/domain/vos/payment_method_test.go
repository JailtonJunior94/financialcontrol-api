package vos_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
)

func TestParsePaymentMethod(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     string
		want    vos.PaymentMethod
		wantErr error
	}{
		{name: "pix", raw: "pix", want: vos.PaymentMethodPix},
		{name: "boleto", raw: "boleto", want: vos.PaymentMethodBoleto},
		{name: "ted", raw: "ted", want: vos.PaymentMethodTED},
		{name: "debit_card", raw: "debit_card", want: vos.PaymentMethodDebitCard},
		{name: "credit_card", raw: "credit_card", want: vos.PaymentMethodCreditCard},
		{name: "cash", raw: "cash", want: vos.PaymentMethodCash},
		{name: "empty", raw: "", wantErr: domain.ErrInvalidPaymentMethod},
		{name: "unknown", raw: "paypal", wantErr: domain.ErrInvalidPaymentMethod},
		{name: "uppercase", raw: "PIX", wantErr: domain.ErrInvalidPaymentMethod},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := vos.ParsePaymentMethod(tc.raw)
			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr))
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestPaymentMethodRequiresCard(t *testing.T) {
	t.Parallel()

	tests := []struct {
		pm           vos.PaymentMethod
		requiresCard bool
	}{
		{vos.PaymentMethodDebitCard, true},
		{vos.PaymentMethodCreditCard, true},
		{vos.PaymentMethodPix, false},
		{vos.PaymentMethodBoleto, false},
		{vos.PaymentMethodTED, false},
		{vos.PaymentMethodCash, false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.pm.String(), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.requiresCard, tc.pm.RequiresCard())
		})
	}
}
