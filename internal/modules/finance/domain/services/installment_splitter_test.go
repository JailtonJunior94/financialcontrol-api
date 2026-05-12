package services

import (
	"math/rand"
	"testing"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// deterministicIDs returns predictable IDs for tests.
type deterministicIDs struct{ seq int }

func (d *deterministicIDs) NewInstallmentID() vos.InstallmentID {
	d.seq++
	return vos.InstallmentID("inst-" + string(rune('0'+d.seq)))
}
func (d *deterministicIDs) NewTransactionID() vos.TransactionID { return vos.NewTransactionID() }
func (d *deterministicIDs) NewInvoiceID() vos.InvoiceID         { return vos.NewInvoiceID() }

func TestInstallmentSplitter_Split_SumEqualsTotal(t *testing.T) {
	t.Parallel()
	svc := NewInstallmentSplitter()

	tests := []struct {
		name  string
		total string
		count int
	}{
		{"100 / 3", "100.00", 3},
		{"10 / 7", "10.00", 7},
		{"0.01 / 24", "0.01", 24},
		{"100 / 1", "100.00", 1},
		{"1234.56 / 12", "1234.56", 12},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m, err := vos.NewMoney(tc.total)
			require.NoError(t, err)
			count, err := vos.NewInstallmentCount(tc.count)
			require.NoError(t, err)

			parts, ids := svc.Split(m, count, &deterministicIDs{})

			require.Len(t, parts, tc.count)
			require.Len(t, ids, tc.count)

			sum := vos.ZeroMoney()
			for _, p := range parts {
				sum = sum.Add(p)
			}
			assert.True(t, sum.Equal(m), "sum=%s total=%s", sum.String(), m.String())
		})
	}
}

func TestInstallmentSplitter_Split_ResidueInFirstPart(t *testing.T) {
	t.Parallel()
	svc := NewInstallmentSplitter()
	m, _ := vos.NewMoney("100.00")
	count, _ := vos.NewInstallmentCount(3)

	parts, _ := svc.Split(m, count, &deterministicIDs{})

	// 100 / 3 = 33.3333... → base = 33.3333, parts[1]=parts[2]=33.3333
	// parts[0] = 100 - 2*33.3333 = 33.3334 (residue in first)
	require.Len(t, parts, 3)
	assert.True(t, parts[0].Amount().GreaterThanOrEqual(parts[1].Amount()),
		"first part should be >= others (holds residue)")
}

func TestInstallmentSplitter_Split_PropertyInvariant(t *testing.T) {
	t.Parallel()
	// 10 random seeds, fixed for determinism
	svc := NewInstallmentSplitter()
	rng := rand.New(rand.NewSource(42)) //nolint:gosec

	for i := 0; i < 10; i++ {
		cents := rng.Intn(100000) + 1
		totalDecimal := decimal.NewFromInt(int64(cents)).Div(decimal.NewFromInt(100))
		rawTotal := totalDecimal.StringFixed(2)
		n := rng.Intn(24) + 1

		m, err := vos.NewMoney(rawTotal)
		require.NoError(t, err)
		count, err := vos.NewInstallmentCount(n)
		require.NoError(t, err)

		parts, _ := svc.Split(m, count, &deterministicIDs{})
		sum := vos.ZeroMoney()
		for _, p := range parts {
			sum = sum.Add(p)
		}
		assert.True(t, sum.Equal(m), "seed42 iter=%d total=%s n=%d sum=%s", i, rawTotal, n, sum.String())
	}
}
