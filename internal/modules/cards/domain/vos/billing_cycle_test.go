package vos_test

import (
	"testing"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/vos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustDate(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestNewBillingCycle(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		closing int
		due     int
		wantErr bool
	}{
		{"closing != due válido", 25, 1, false},
		{"closing == due inválido", 15, 15, true},
		{"closing=1 due=31 válido", 1, 31, false},
		{"closing=31 due=1 válido", 31, 1, false},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := vos.NewBillingCycle(vos.MustClosingDay(tc.closing), vos.MustDueDay(tc.due))
			if tc.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, vos.ErrInvalidBillingCycle)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.closing, got.Closing().Int())
			assert.Equal(t, tc.due, got.Due().Int())
		})
	}
}

func TestBillingCycle_DueDateFor(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		closing  int
		due      int
		purchase time.Time
		want     time.Time
	}{
		{
			name:    "compra antes do fechamento",
			closing: 25, due: 1,
			purchase: mustDate("2026-04-24"),
			want:     mustDate("2026-05-01"),
		},
		{
			name:    "compra no dia do fechamento",
			closing: 25, due: 1,
			purchase: mustDate("2026-04-25"),
			want:     mustDate("2026-06-01"),
		},
		{
			name:    "compra após o fechamento",
			closing: 25, due: 1,
			purchase: mustDate("2026-04-26"),
			want:     mustDate("2026-06-01"),
		},
		{
			name:    "due 31 em fevereiro clamp (ano bissexto)",
			closing: 5, due: 31,
			purchase: mustDate("2024-02-04"),
			want:     mustDate("2024-02-29"),
		},
		{
			name:    "virada de ano",
			closing: 25, due: 10,
			purchase: mustDate("2026-12-26"),
			want:     mustDate("2027-02-10"),
		},
		{
			name:    "mês de 28 dias (fevereiro não bissexto)",
			closing: 5, due: 31,
			purchase: mustDate("2025-02-04"),
			want:     mustDate("2025-02-28"),
		},
		{
			name:    "mês de 30 dias clamp",
			closing: 5, due: 31,
			purchase: mustDate("2026-04-04"),
			want:     mustDate("2026-04-30"),
		},
		{
			name:    "compra em dezembro dentro do ciclo",
			closing: 25, due: 1,
			purchase: mustDate("2026-12-10"),
			want:     mustDate("2027-01-01"),
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c, err := vos.NewBillingCycle(vos.MustClosingDay(tc.closing), vos.MustDueDay(tc.due))
			require.NoError(t, err)
			got := c.DueDateFor(tc.purchase)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestBillingCycle_BestPurchaseDay(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		closing int
		due     int
		want    int
	}{
		{"closing 25 → best 24", 25, 1, 24},
		{"closing 1 → best 1 (clamp)", 1, 31, 1},
		{"closing 2 → best 1", 2, 31, 1},
		{"closing 31 → best 30", 31, 1, 30},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c, err := vos.NewBillingCycle(vos.MustClosingDay(tc.closing), vos.MustDueDay(tc.due))
			require.NoError(t, err)
			assert.Equal(t, tc.want, c.BestPurchaseDay())
		})
	}
}
