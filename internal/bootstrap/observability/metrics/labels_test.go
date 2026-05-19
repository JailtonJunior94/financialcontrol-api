package metrics_test

import (
	"fmt"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/observability/metrics"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
)

func TestStatusFromHTTP(t *testing.T) {
	tests := []struct {
		code int
		want metrics.StatusBucket
	}{
		{200, metrics.StatusSuccess},
		{201, metrics.StatusSuccess},
		{204, metrics.StatusSuccess},
		{301, metrics.StatusSuccess},
		{302, metrics.StatusSuccess},
		{400, metrics.StatusDeclined},
		{404, metrics.StatusDeclined},
		{422, metrics.StatusDeclined},
		{500, metrics.StatusTechnicalError},
		{503, metrics.StatusTechnicalError},
		{0, metrics.StatusTechnicalError},
		{199, metrics.StatusTechnicalError},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("HTTP_%d", tt.code), func(t *testing.T) {
			assert.Equal(t, tt.want, metrics.StatusFromHTTP(tt.code))
		})
	}
}

func TestParseOperationLabel(t *testing.T) {
	tests := []struct {
		input string
		want  metrics.OperationLabel
	}{
		{"create_transaction", metrics.OperationCreateTransaction},
		{"update_transaction", metrics.OperationUpdateTransaction},
		{"delete_transaction", metrics.OperationDeleteTransaction},
		{"pay_invoice", metrics.OperationPayInvoice},
		{"refund_transaction", metrics.OperationRefundTransaction},
		{"anticipate_installment", metrics.OperationAnticipateInstallment},
		{"unknown", metrics.OperationUnknown},
		{"foo", metrics.OperationUnknown},
		{"", metrics.OperationUnknown},
		{"CREATE_TRANSACTION", metrics.OperationUnknown}, // case-sensitive closed list
	}
	for _, tt := range tests {
		name := tt.input
		if name == "" {
			name = "<empty>"
		}
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, metrics.ParseOperationLabel(tt.input))
		})
	}
}

func TestAmountCentsFromMoney(t *testing.T) {
	brl := func(s string) vos.Money {
		m, err := vos.NewMoney(s)
		require.NoError(t, err)
		return m
	}

	tests := []struct {
		name    string
		money   vos.Money
		want    metrics.AmountCents
		wantErr bool
	}{
		{
			name:  "whole BRL",
			money: brl("100"),
			want:  10000,
		},
		{
			name:  "fractional BRL",
			money: brl("1.50"),
			want:  150,
		},
		{
			name:  "zero",
			money: vos.ZeroMoney(),
			want:  0,
		},
		{
			name:  "negative amount",
			money: brl("-5.00"),
			want:  -500,
		},
		{
			name:  "large valid amount",
			money: brl("100000000"),
			want:  10000000000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := metrics.FromMoney(tt.money)
			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, metrics.ErrAmountOverflow)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAmountCentsOverflow(t *testing.T) {
	// math.MaxInt64 ≈ 9.22e18; divided by 100 ≈ 9.22e16 BRL.
	// A value of 1e17 BRL exceeds that threshold when converted to centavos.
	hugeDecimal := decimal.NewFromFloat(1e17)
	m := vos.NewMoneyFromDecimal(hugeDecimal)
	_, err := metrics.FromMoney(m)
	assert.ErrorIs(t, err, metrics.ErrAmountOverflow)
}

func TestEnvironmentTag(t *testing.T) {
	tests := []struct {
		env     string
		appName string
	}{
		{"development", "financialcontrol-api-development"},
		{"staging", "financialcontrol-api-staging"},
		{"production", "financialcontrol-api-production"},
		{"", "financialcontrol-api-unknown"},
	}
	for _, tt := range tests {
		name := tt.env
		if name == "" {
			name = "<empty>"
		}
		t.Run(name, func(t *testing.T) {
			tag := metrics.NewEnvironmentTag(tt.env)
			assert.Equal(t, tt.appName, tag.AppName())
		})
	}
}

func TestPartnerLabelAlwaysUnknown(t *testing.T) {
	tests := []string{"stripe", "paypal", "", "any_partner"}
	for _, s := range tests {
		name := s
		if name == "" {
			name = "<empty>"
		}
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, metrics.PartnerUnknown, metrics.ParsePartnerLabel(s))
		})
	}
}
