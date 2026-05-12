package vos_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
)

func TestMoneyHelpers(t *testing.T) {
	t.Parallel()

	t.Run("Equal", func(t *testing.T) {
		t.Parallel()
		a, _ := vos.NewMoney("10.00")
		b, _ := vos.NewMoney("10.00")
		c, _ := vos.NewMoney("20.00")
		assert.True(t, a.Equal(b))
		assert.False(t, a.Equal(c))
	})

	t.Run("Currency", func(t *testing.T) {
		t.Parallel()
		m, _ := vos.NewMoney("5.00")
		assert.Equal(t, "BRL", m.Currency())
	})

	t.Run("IsNegative", func(t *testing.T) {
		t.Parallel()
		neg, _ := vos.NewMoney("-1.00")
		pos, _ := vos.NewMoney("1.00")
		assert.True(t, neg.IsNegative())
		assert.False(t, pos.IsNegative())
	})

	t.Run("IsZero", func(t *testing.T) {
		t.Parallel()
		z, _ := vos.NewMoney("0")
		nz, _ := vos.NewMoney("0.01")
		assert.True(t, z.IsZero())
		assert.False(t, nz.IsZero())
	})

	t.Run("NewMoneyFromDecimal", func(t *testing.T) {
		t.Parallel()
		m, _ := vos.NewMoney("42.00")
		m2 := vos.NewMoneyFromDecimal(m.Amount())
		assert.True(t, m.Equal(m2))
	})
}

func TestAmountFromMoney(t *testing.T) {
	t.Parallel()

	t.Run("valid positive", func(t *testing.T) {
		t.Parallel()
		m, _ := vos.NewMoney("5.00")
		a, err := vos.NewAmountFromMoney(m)
		require.NoError(t, err)
		assert.Equal(t, "5.00", a.String())
	})

	t.Run("zero money rejected", func(t *testing.T) {
		t.Parallel()
		z := vos.ZeroMoney()
		_, err := vos.NewAmountFromMoney(z)
		require.Error(t, err)
	})
}

func TestInstallmentNumberString(t *testing.T) {
	t.Parallel()
	n, _ := vos.NewInstallmentNumber(3)
	assert.Equal(t, "3", n.String())
}

func TestIdempotencyKeyString(t *testing.T) {
	t.Parallel()
	k, _ := vos.NewIdempotencyKey("my-key")
	assert.Equal(t, "my-key", k.String())
}
