package metrics_test

import (
	"context"
	"testing"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/observability/metrics"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
)

func newTestBusinessMetrics(t *testing.T) (*metrics.BusinessMetrics, *fake.FakeMetrics) {
	t.Helper()
	fp := fake.NewProvider()
	bm, err := metrics.NewBusinessMetrics(fp)
	require.NoError(t, err)
	fm := fp.Metrics().(*fake.FakeMetrics)
	return bm, fm
}

func TestRecordOperation(t *testing.T) {
	tests := []struct {
		name       string
		op         metrics.OperationLabel
		status     metrics.StatusBucket
		wantOp     string
		wantStatus string
	}{
		{
			name:       "create transaction success",
			op:         metrics.OperationCreateTransaction,
			status:     metrics.StatusSuccess,
			wantOp:     "create_transaction",
			wantStatus: "success",
		},
		{
			name:       "pay invoice declined",
			op:         metrics.OperationPayInvoice,
			status:     metrics.StatusDeclined,
			wantOp:     "pay_invoice",
			wantStatus: "declined",
		},
		{
			name:       "unknown operation technical error",
			op:         metrics.OperationUnknown,
			status:     metrics.StatusTechnicalError,
			wantOp:     "unknown",
			wantStatus: "technical_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bm, fm := newTestBusinessMetrics(t)
			ctx := context.Background()

			bm.RecordOperation(ctx, tt.op, tt.status)

			counter := fm.GetCounter("financial_operation_total")
			require.NotNil(t, counter)

			values := counter.GetValues()
			require.Len(t, values, 1)
			assert.Equal(t, int64(1), values[0].Value)
			assert.True(t, hasStringField(values[0].Fields, "operation", tt.wantOp))
			assert.True(t, hasStringField(values[0].Fields, "status", tt.wantStatus))
		})
	}
}

func TestRecordAmountProcessed(t *testing.T) {
	tests := []struct {
		name      string
		amount    string
		txType    vos.TransactionType
		method    vos.PaymentMethod
		wantCents int64
	}{
		{
			name:      "income via pix",
			amount:    "250.00",
			txType:    vos.TransactionTypeIncome,
			method:    vos.PaymentMethodPix,
			wantCents: 25000,
		},
		{
			name:      "refund via credit card",
			amount:    "100.50",
			txType:    vos.TransactionTypeRefund,
			method:    vos.PaymentMethodCreditCard,
			wantCents: 10050,
		},
		{
			name:      "expense via boleto",
			amount:    "10.00",
			txType:    vos.TransactionTypeExpense,
			method:    vos.PaymentMethodBoleto,
			wantCents: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bm, fm := newTestBusinessMetrics(t)
			ctx := context.Background()

			money, err := vos.NewMoney(tt.amount)
			require.NoError(t, err)

			bm.RecordAmountProcessed(ctx, money, tt.txType, tt.method)

			counter := fm.GetCounter("financial_amount_processed")
			require.NotNil(t, counter)

			values := counter.GetValues()
			require.Len(t, values, 1)
			assert.Equal(t, tt.wantCents, values[0].Value)
			assert.True(t, hasStringField(values[0].Fields, "transaction_type", tt.txType.String()))
			assert.True(t, hasStringField(values[0].Fields, "payment_method", tt.method.String()))
		})
	}
}

func TestRecordAmountProcessedOverflowIsAbsorbed(t *testing.T) {
	// Overflow amounts must be silently absorbed — must not panic or propagate errors.
	bm, fm := newTestBusinessMetrics(t)
	ctx := context.Background()

	// 1e17 BRL overflows int64 centavos.
	huge, err := vos.NewMoney("100000000000000000")
	require.NoError(t, err)

	bm.RecordAmountProcessed(ctx, huge, vos.TransactionTypeIncome, vos.PaymentMethodPix)

	counter := fm.GetCounter("financial_amount_processed")
	require.NotNil(t, counter)
	assert.Empty(t, counter.GetValues())
}

func TestRecordPartnerLatency(t *testing.T) {
	bm, fm := newTestBusinessMetrics(t)
	ctx := context.Background()

	bm.RecordPartnerLatency(ctx, metrics.PartnerUnknown, 150*time.Millisecond)

	hist := fm.GetHistogram("partner_integration_latency")
	require.NotNil(t, hist)

	values := hist.GetValues()
	require.Len(t, values, 1)
	assert.InDelta(t, 0.15, values[0].Value, 1e-9)
	assert.True(t, hasStringField(values[0].Fields, "partner", "unknown"))
}

func TestAllInstrumentsRegisteredAtStartup(t *testing.T) {
	// Instruments must be registered even if no values are recorded (RF-15.5).
	// Verifies NewBusinessMetrics does not return error and instruments exist.
	bm, fm := newTestBusinessMetrics(t)
	require.NotNil(t, bm)

	ctx := context.Background()
	// Touching each instrument to confirm it is usable (not nil/noop on error).
	bm.RecordOperation(ctx, metrics.OperationUnknown, metrics.StatusSuccess)
	assert.NotNil(t, fm.GetCounter("financial_operation_total"))

	bm.RecordAmountProcessed(ctx, vos.ZeroMoney(), vos.TransactionTypeExpense, vos.PaymentMethodCash)
	assert.NotNil(t, fm.GetCounter("financial_amount_processed"))

	// partner_integration_latency is declared but has no active call site; touch via RecordPartnerLatency.
	bm.RecordPartnerLatency(ctx, metrics.PartnerUnknown, 0)
	assert.NotNil(t, fm.GetHistogram("partner_integration_latency"))
}

// hasStringField reports whether fields contains a string field with the given key and value.
func hasStringField(fields []observability.Field, key, val string) bool {
	for _, f := range fields {
		if f.Key == key && f.Kind() == observability.FieldKindString && f.StringValue() == val {
			return true
		}
	}
	return false
}
