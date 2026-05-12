package vos_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
)

func TestNewAmount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     string
		wantErr error
	}{
		{name: "positive amount", raw: "10.00"},
		{name: "minimal positive", raw: "0.0001"},
		{name: "zero is invalid", raw: "0", wantErr: domain.ErrInvalidAmount},
		{name: "zero string variants", raw: "0.00", wantErr: domain.ErrInvalidAmount},
		{name: "negative is invalid", raw: "-1.00", wantErr: domain.ErrInvalidAmount},
		{name: "invalid format", raw: "abc", wantErr: domain.ErrInvalidMoneyFormat},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a, err := vos.NewAmount(tc.raw)
			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr), "expected %v got %v", tc.wantErr, err)
				return
			}
			require.NoError(t, err)
			assert.True(t, a.Money().IsPositive())
		})
	}
}
