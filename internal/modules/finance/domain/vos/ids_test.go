package vos_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
)

// ---- TransactionID ----

func TestTransactionID(t *testing.T) {
	t.Parallel()

	t.Run("NewTransactionID generates valid UUID", func(t *testing.T) {
		t.Parallel()
		id := vos.NewTransactionID()
		assert.NotEmpty(t, id.String())
		// Round-trip via Parse.
		parsed, err := vos.ParseTransactionID(id.String())
		require.NoError(t, err)
		assert.Equal(t, id, parsed)
	})

	tests := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{"valid uuid", "550e8400-e29b-41d4-a716-446655440000", false},
		{"empty string", "", true},
		{"not a uuid", "not-a-uuid", true},
		{"wrong format", "550e8400-e29b-41d4-a716-44665544000z", true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run("ParseTransactionID/"+tc.name, func(t *testing.T) {
			t.Parallel()
			id, err := vos.ParseTransactionID(tc.raw)
			if tc.wantErr {
				require.Error(t, err)
				assert.True(t, errors.Is(err, vos.ErrInvalidTransactionID))
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.raw, id.String())
		})
	}
}

// ---- InstallmentID ----

func TestInstallmentID(t *testing.T) {
	t.Parallel()

	t.Run("NewInstallmentID generates valid UUID", func(t *testing.T) {
		t.Parallel()
		id := vos.NewInstallmentID()
		assert.NotEmpty(t, id.String())
		parsed, err := vos.ParseInstallmentID(id.String())
		require.NoError(t, err)
		assert.Equal(t, id, parsed)
	})

	tests := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{"valid uuid", "550e8400-e29b-41d4-a716-446655440000", false},
		{"empty", "", true},
		{"invalid", "bad-id", true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run("ParseInstallmentID/"+tc.name, func(t *testing.T) {
			t.Parallel()
			id, err := vos.ParseInstallmentID(tc.raw)
			if tc.wantErr {
				require.Error(t, err)
				assert.True(t, errors.Is(err, vos.ErrInvalidInstallmentID))
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.raw, id.String())
		})
	}
}

// ---- InvoiceID ----

func TestInvoiceID(t *testing.T) {
	t.Parallel()

	t.Run("NewInvoiceID generates valid UUID", func(t *testing.T) {
		t.Parallel()
		id := vos.NewInvoiceID()
		assert.NotEmpty(t, id.String())
		parsed, err := vos.ParseInvoiceID(id.String())
		require.NoError(t, err)
		assert.Equal(t, id, parsed)
	})

	tests := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{"valid uuid", "550e8400-e29b-41d4-a716-446655440000", false},
		{"empty", "", true},
		{"invalid", "bad", true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run("ParseInvoiceID/"+tc.name, func(t *testing.T) {
			t.Parallel()
			id, err := vos.ParseInvoiceID(tc.raw)
			if tc.wantErr {
				require.Error(t, err)
				assert.True(t, errors.Is(err, vos.ErrInvalidInvoiceID))
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.raw, id.String())
		})
	}
}

// ---- CardID ----

func TestCardID(t *testing.T) {
	t.Parallel()

	t.Run("NewCardID generates valid UUID", func(t *testing.T) {
		t.Parallel()
		id := vos.NewCardID()
		assert.NotEmpty(t, id.String())
	})

	tests := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{"valid uuid", "550e8400-e29b-41d4-a716-446655440000", false},
		{"empty", "", true},
		{"invalid", "x", true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run("ParseCardID/"+tc.name, func(t *testing.T) {
			t.Parallel()
			id, err := vos.ParseCardID(tc.raw)
			if tc.wantErr {
				require.Error(t, err)
				assert.True(t, errors.Is(err, vos.ErrInvalidCardID))
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.raw, id.String())
		})
	}
}

// ---- CategoryID ----

func TestCategoryID(t *testing.T) {
	t.Parallel()

	t.Run("NewCategoryID generates valid UUID", func(t *testing.T) {
		t.Parallel()
		id := vos.NewCategoryID()
		assert.NotEmpty(t, id.String())
	})

	tests := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{"valid uuid", "550e8400-e29b-41d4-a716-446655440000", false},
		{"empty", "", true},
		{"invalid", "not-uuid", true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run("ParseCategoryID/"+tc.name, func(t *testing.T) {
			t.Parallel()
			id, err := vos.ParseCategoryID(tc.raw)
			if tc.wantErr {
				require.Error(t, err)
				assert.True(t, errors.Is(err, vos.ErrInvalidCategoryID))
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.raw, id.String())
		})
	}
}
