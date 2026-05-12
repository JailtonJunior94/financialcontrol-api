package vos_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
)

func TestNewMoney(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     string
		wantErr error
	}{
		{name: "valid positive", raw: "100.00"},
		{name: "valid with 4 decimals", raw: "99.9999"},
		{name: "zero is valid", raw: "0"},
		{name: "negative is valid (sign managed by TransactionType)", raw: "-50.00"},
		{name: "invalid string", raw: "abc", wantErr: domain.ErrInvalidMoneyFormat},
		{name: "empty string", raw: "", wantErr: domain.ErrInvalidMoneyFormat},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m, err := vos.NewMoney(tc.raw)
			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr), "expected %v got %v", tc.wantErr, err)
				return
			}
			require.NoError(t, err)
			assert.False(t, m.Amount().IsZero() && tc.raw != "0", "unexpected zero for %q", tc.raw)
		})
	}
}

func TestMoneyAdd(t *testing.T) {
	t.Parallel()

	a, _ := vos.NewMoney("100.00")
	b, _ := vos.NewMoney("23.50")
	result := a.Add(b)
	assert.Equal(t, "123.50", result.String())
}

func TestMoneySub(t *testing.T) {
	t.Parallel()

	a, _ := vos.NewMoney("100.00")
	b, _ := vos.NewMoney("30.00")
	result := a.Sub(b)
	assert.Equal(t, "70.00", result.String())
}

func TestMoneyMul(t *testing.T) {
	t.Parallel()

	m, _ := vos.NewMoney("25.00")
	result := m.Mul(4)
	assert.Equal(t, "100.00", result.String())
}

func TestMoneyDivideEvenly(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		total     string
		n         int
		wantParts []string // optional: only checked when non-nil
		wantSum   string   // empty when n=0 (nil result)
		nilResult bool
	}{
		{
			// DivRound(3, 4) → 33.3333; first = 100.00 - 33.3333*2 = 33.3334
			name:      "R$100 / 3 — resíduo na 1ª",
			total:     "100.00",
			n:         3,
			wantParts: []string{"33.3334", "33.3333", "33.3333"},
			wantSum:   "100.00",
		},
		{
			name:    "R$10 / 7",
			total:   "10.00",
			n:       7,
			wantSum: "10.00",
		},
		{
			name:    "R$0.01 / 24",
			total:   "0.01",
			n:       24,
			wantSum: "0.01",
		},
		{
			name:      "exact division",
			total:     "90.00",
			n:         3,
			wantParts: []string{"30.0000", "30.0000", "30.0000"},
			wantSum:   "90.00",
		},
		{
			name:      "count = 1",
			total:     "55.00",
			n:         1,
			wantParts: []string{"55.0000"},
			wantSum:   "55.00",
		},
		{
			name:      "count = 0 returns nil",
			total:     "100.00",
			n:         0,
			nilResult: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			total, err := vos.NewMoney(tc.total)
			require.NoError(t, err)

			parts := total.DivideEvenly(tc.n)

			if tc.nilResult {
				assert.Nil(t, parts)
				return
			}

			require.Len(t, parts, tc.n)

			// Verify specific parts when provided.
			for i, want := range tc.wantParts {
				d, _ := decimal.NewFromString(want)
				assert.True(t, parts[i].Amount().Equal(d), "part[%d]: got %s want %s", i, parts[i].Amount(), want)
			}

			// Sum invariant: must equal original total exactly.
			sum := vos.ZeroMoney()
			for _, p := range parts {
				sum = sum.Add(p)
			}
			expected, _ := vos.NewMoney(tc.wantSum)
			assert.True(t, sum.Amount().Equal(expected.Amount()),
				"sum=%s expected=%s", sum.Amount(), expected.Amount())
		})
	}
}

func TestMoneyMarshalJSON(t *testing.T) {
	t.Parallel()

	m, _ := vos.NewMoney("123.45")
	data, err := json.Marshal(m)
	require.NoError(t, err)
	assert.Equal(t, `"123.45"`, string(data))
}

func TestMoneyUnmarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{name: "string decimal", input: `"123.45"`},
		{name: "string zero", input: `"0.00"`},
		{name: "raw number", input: `123.45`, wantErr: domain.ErrInvalidMoneyFormat},
		{name: "raw integer", input: `100`, wantErr: domain.ErrInvalidMoneyFormat},
		{name: "null", input: `null`, wantErr: domain.ErrInvalidMoneyFormat},
		{name: "string non-numeric", input: `"abc"`, wantErr: domain.ErrInvalidMoneyFormat},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var m vos.Money
			err := json.Unmarshal([]byte(tc.input), &m)
			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr))
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestMoneyJSONRoundTrip(t *testing.T) {
	t.Parallel()

	original, _ := vos.NewMoney("99.99")
	data, err := json.Marshal(original)
	require.NoError(t, err)

	var recovered vos.Money
	err = json.Unmarshal(data, &recovered)
	require.NoError(t, err)

	assert.Equal(t, original.String(), recovered.String())
}
