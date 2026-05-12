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

func TestNewInvoiceFilter_ValidDateRange(t *testing.T) {
	t.Parallel()
	userID := identityvo.NewUserID()
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	state := vos.InvoiceStateOpen

	f, err := NewInvoiceFilter(userID, nil, &state, &from, &to, defaultPagination(t))

	require.NoError(t, err)
	assert.Equal(t, userID, f.UserID)
	assert.Equal(t, &state, f.State)
}

func TestNewInvoiceFilter_InvalidDateRange(t *testing.T) {
	t.Parallel()
	from := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	_, err := NewInvoiceFilter(identityvo.NewUserID(), nil, nil, &from, &to, defaultPagination(t))

	assert.ErrorIs(t, err, domain.ErrInvalidDateRange)
}
