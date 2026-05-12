package vos_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
)

func TestNewIdempotencyKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr error
	}{
		{name: "minimum (1 char)", raw: "a", want: "a"},
		{name: "exact maximum (64 chars)", raw: strings.Repeat("x", 64), want: strings.Repeat("x", 64)},
		{name: "trimmed spaces preserved value", raw: "  abc  ", want: "abc"},
		{name: "only spaces", raw: "   ", wantErr: domain.ErrInvalidIdempotencyKeyFormat},
		{name: "empty string", raw: "", wantErr: domain.ErrInvalidIdempotencyKeyFormat},
		{name: "65 chars (one over)", raw: strings.Repeat("y", 65), wantErr: domain.ErrInvalidIdempotencyKeyFormat},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			k, err := vos.NewIdempotencyKey(tc.raw)
			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr))
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, k.Value())
		})
	}
}
