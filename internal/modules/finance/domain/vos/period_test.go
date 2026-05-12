package vos_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
)

func TestNewPeriod(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		year    int
		month   int
		wantErr error
	}{
		{name: "valid may 2026", year: 2026, month: 5},
		{name: "valid january", year: 2024, month: 1},
		{name: "valid december", year: 2024, month: 12},
		{name: "month 0 invalid", year: 2024, month: 0, wantErr: domain.ErrInvalidYearMonth},
		{name: "month 13 invalid", year: 2024, month: 13, wantErr: domain.ErrInvalidYearMonth},
		{name: "year 0 invalid", year: 0, month: 5, wantErr: domain.ErrInvalidYearMonth},
	}

	loc, _ := time.LoadLocation("America/Sao_Paulo")
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			p, err := vos.NewPeriod(tc.year, tc.month, loc)
			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr))
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.year, p.Year())
			assert.Equal(t, time.Month(tc.month), p.Month())
		})
	}
}

func TestPeriodBounds(t *testing.T) {
	t.Parallel()

	loc, err := time.LoadLocation("America/Sao_Paulo")
	require.NoError(t, err)

	tests := []struct {
		name      string
		year      int
		month     int
		wantStart time.Time
		wantEnd   time.Time
	}{
		{
			name:  "May 2026 — America/Sao_Paulo UTC-3",
			year:  2026,
			month: 5,
			// UTC-3 → midnight SP = 03:00 UTC
			wantStart: time.Date(2026, 5, 1, 3, 0, 0, 0, time.UTC),
			wantEnd:   time.Date(2026, 6, 1, 3, 0, 0, 0, time.UTC),
		},
		{
			name:      "January 2026",
			year:      2026,
			month:     1,
			wantStart: time.Date(2026, 1, 1, 3, 0, 0, 0, time.UTC),
			wantEnd:   time.Date(2026, 2, 1, 3, 0, 0, 0, time.UTC),
		},
		{
			name:      "December 2025 — year boundary",
			year:      2025,
			month:     12,
			wantStart: time.Date(2025, 12, 1, 3, 0, 0, 0, time.UTC),
			wantEnd:   time.Date(2026, 1, 1, 3, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			p, err := vos.NewPeriod(tc.year, tc.month, loc)
			require.NoError(t, err)

			start, end := p.Bounds()
			assert.True(t, start.Equal(tc.wantStart),
				"start: got %s want %s", start, tc.wantStart)
			assert.True(t, end.Equal(tc.wantEnd),
				"end: got %s want %s", end, tc.wantEnd)
		})
	}
}
