package invoicingapp_test

import (
	"context"
	"errors"
	"testing"
	"time"

	invoicingdomain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/invoicing/domain"
	invoicingapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/invoicing/application"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- fakes ---

type fakeSyncPort struct {
	called bool
	err    error
}

func (f *fakeSyncPort) SyncTransactionWithInvoice(_ context.Context, _, _, _ string, _ time.Time, _ float64) error {
	f.called = true
	return f.err
}

// --- tests ---

func TestHandleReturnsErrorWhenInvoiceIDIsEmpty(t *testing.T) {
	handler := invoicingapp.NewInvoiceChangedHandler(&fakeSyncPort{})
	err := handler.Handle(t.Context(), invoicingdomain.InvoiceChangedPayload{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "InvoiceID must not be empty")
}

func TestHandleCallsSync(t *testing.T) {
	sync := &fakeSyncPort{}
	handler := invoicingapp.NewInvoiceChangedHandler(sync)

	payload := invoicingdomain.InvoiceChangedPayload{
		InvoiceID:       "inv-1",
		CardDescription: "Card X",
		UserID:          "user-1",
		ReferenceDate:   time.Now(),
		Total:           100.0,
	}

	err := handler.Handle(t.Context(), payload)

	require.NoError(t, err)
	assert.True(t, sync.called, "sync should be called")
}

func TestHandleReturnsErrorWhenSyncFails(t *testing.T) {
	sync := &fakeSyncPort{err: errors.New("sync error")}
	handler := invoicingapp.NewInvoiceChangedHandler(sync)

	payload := invoicingdomain.InvoiceChangedPayload{
		InvoiceID:       "inv-2",
		CardDescription: "Card X",
		UserID:          "user-1",
		ReferenceDate:   time.Now(),
		Total:           50.0,
	}

	err := handler.Handle(t.Context(), payload)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "syncing transaction")
}

func TestHandleTableDriven(t *testing.T) {
	tests := []struct {
		name     string
		payload  invoicingdomain.InvoiceChangedPayload
		syncErr  error
		wantErr  bool
		errContains string
	}{
		{
			name:        "empty InvoiceID returns error",
			payload:     invoicingdomain.InvoiceChangedPayload{},
			wantErr:     true,
			errContains: "InvoiceID must not be empty",
		},
		{
			name: "valid payload calls sync",
			payload: invoicingdomain.InvoiceChangedPayload{
				InvoiceID:       "inv-ok",
				CardDescription: "Visa",
				UserID:          "user-42",
				Total:           99.9,
			},
			wantErr: false,
		},
		{
			name: "sync failure propagates error",
			payload: invoicingdomain.InvoiceChangedPayload{
				InvoiceID: "inv-fail",
			},
			syncErr:     errors.New("db down"),
			wantErr:     true,
			errContains: "syncing transaction",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sync := &fakeSyncPort{err: tt.syncErr}
			handler := invoicingapp.NewInvoiceChangedHandler(sync)

			err := handler.Handle(t.Context(), tt.payload)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
