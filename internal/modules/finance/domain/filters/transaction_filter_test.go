package filters

import (
	"testing"
	"time"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func defaultPagination(t *testing.T) vos.Pagination {
	t.Helper()
	p, err := vos.NewPagination(1, 20)
	require.NoError(t, err)
	return p
}

func TestNewTransactionFilter_ValidDateRange(t *testing.T) {
	t.Parallel()
	userID := identityvo.NewUserID()
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)

	f, err := NewTransactionFilter(userID, nil, nil, nil, nil,
		vos.InvoiceStatusFilterNone, "", &from, &to, defaultPagination(t))

	require.NoError(t, err)
	assert.Equal(t, userID, f.UserID)
	assert.Equal(t, &from, f.From)
	assert.Equal(t, &to, f.To)
}

func TestNewTransactionFilter_InvalidDateRange(t *testing.T) {
	t.Parallel()
	from := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	_, err := NewTransactionFilter(identityvo.NewUserID(), nil, nil, nil, nil,
		vos.InvoiceStatusFilterNone, "", &from, &to, defaultPagination(t))

	assert.ErrorIs(t, err, domain.ErrInvalidDateRange)
}

func TestTransactionFilter_RequiresCardPurchaseScope(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		invoiceStatus vos.InvoiceStatusFilter
		want          bool
	}{
		{name: "none filter → false", invoiceStatus: vos.InvoiceStatusFilterNone, want: false},
		{name: "open filter → true", invoiceStatus: vos.InvoiceStatusFilterOpen, want: true},
		{name: "closed filter → true", invoiceStatus: vos.InvoiceStatusFilterClosed, want: true},
		{name: "paid filter → true", invoiceStatus: vos.InvoiceStatusFilterPaid, want: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			f, err := NewTransactionFilter(identityvo.NewUserID(), nil, nil, nil, nil,
				tc.invoiceStatus, "", nil, nil, defaultPagination(t))
			require.NoError(t, err)
			assert.Equal(t, tc.want, f.RequiresCardPurchaseScope())
		})
	}
}
