package vos_test

import (
	"testing"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/vos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCategoryID(t *testing.T) {
	t.Parallel()
	id := vos.NewCategoryID()
	assert.NotEmpty(t, id.String())
	parsed, err := vos.ParseCategoryID(id.String())
	require.NoError(t, err)
	assert.Equal(t, id, parsed)
}

func TestParseCategoryID(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"UUID válido", "550e8400-e29b-41d4-a716-446655440000", false},
		{"string vazia", "", true},
		{"UUID inválido", "not-a-uuid", true},
		{"UUID com formato errado", "550e8400-e29b-41d4-a716", true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := vos.ParseCategoryID(tc.input)
			if tc.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, domain.ErrInvalidCategoryID)
				assert.Empty(t, got.String())
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.input, got.String())
		})
	}
}
