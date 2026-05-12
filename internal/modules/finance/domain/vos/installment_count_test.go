package vos_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
)

func TestNewInstallmentCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		n       int
		wantErr error
	}{
		{name: "minimum", n: 1},
		{name: "mid-range", n: 12},
		{name: "maximum", n: 24},
		{name: "zero", n: 0, wantErr: domain.ErrInstallmentCountOutOfRange},
		{name: "negative", n: -1, wantErr: domain.ErrInstallmentCountOutOfRange},
		{name: "over maximum", n: 25, wantErr: domain.ErrInstallmentCountOutOfRange},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c, err := vos.NewInstallmentCount(tc.n)
			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr))
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.n, c.Value())
		})
	}
}

func TestNewInstallmentNumber(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		n       int
		wantErr error
	}{
		{name: "minimum", n: 1},
		{name: "large number", n: 100},
		{name: "zero is invalid", n: 0, wantErr: domain.ErrInstallmentCountOutOfRange},
		{name: "negative", n: -1, wantErr: domain.ErrInstallmentCountOutOfRange},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			n, err := vos.NewInstallmentNumber(tc.n)
			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr))
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.n, n.Value())
		})
	}
}
