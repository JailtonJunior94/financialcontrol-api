package vos_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
)

func TestParseTransactionType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     string
		want    vos.TransactionType
		wantErr error
	}{
		{name: "income", raw: "income", want: vos.TransactionTypeIncome},
		{name: "expense", raw: "expense", want: vos.TransactionTypeExpense},
		{name: "credit_purchase", raw: "credit_purchase", want: vos.TransactionTypeCreditPurchase},
		{name: "installment_purchase", raw: "installment_purchase", want: vos.TransactionTypeInstallmentPurchase},
		{name: "refund", raw: "refund", want: vos.TransactionTypeRefund},
		{name: "empty string", raw: "", wantErr: domain.ErrInvalidTransactionType},
		{name: "unknown value", raw: "transfer", wantErr: domain.ErrInvalidTransactionType},
		{name: "uppercase", raw: "INCOME", wantErr: domain.ErrInvalidTransactionType},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := vos.ParseTransactionType(tc.raw)
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

func TestTransactionTypeHelpers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		tt          vos.TransactionType
		isCardBased bool
		isRefund    bool
	}{
		{"credit_purchase is card-based", vos.TransactionTypeCreditPurchase, true, false},
		{"installment_purchase is card-based", vos.TransactionTypeInstallmentPurchase, true, false},
		{"income is not card-based", vos.TransactionTypeIncome, false, false},
		{"expense is not card-based", vos.TransactionTypeExpense, false, false},
		{"refund is not card-based", vos.TransactionTypeRefund, false, true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.isCardBased, tc.tt.IsCardBased())
			assert.Equal(t, tc.isRefund, tc.tt.IsRefund())
		})
	}
}
