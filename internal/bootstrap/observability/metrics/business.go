package metrics

import (
	"context"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/ports"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
)

// BusinessMetrics implements ports.FinancialMetricsRecorder using the devkit observability
// Metrics interface to register and record the three business instruments.
type BusinessMetrics struct {
	operationTotal  observability.Counter
	amountProcessed observability.Counter
	partnerLatency  observability.Histogram
}

// Compile-time assertion that BusinessMetrics satisfies the port.
var _ ports.FinancialMetricsRecorder = (*BusinessMetrics)(nil)

// NewBusinessMetrics registers the three business metric instruments via the
// provided observability provider. All three instruments are registered at startup
// even when not yet active (partner_integration_latency has no call sites this delivery).
func NewBusinessMetrics(obs observability.Observability) (*BusinessMetrics, error) {
	m := obs.Metrics()

	opTotal := m.Counter(
		string(MetricFinancialOperationTotal),
		"Total financial operations by operation type and status",
		"{operation}",
	)
	amtProcessed := m.Counter(
		string(MetricFinancialAmountProcessed),
		"Total amount processed in BRL centavos for successful operations",
		"centavos",
	)
	partnerLat := m.Histogram(
		string(MetricPartnerIntegrationLatency),
		"Partner integration latency in seconds; declared for contract; no active instrumentation this delivery",
		"s",
	)

	return &BusinessMetrics{
		operationTotal:  opTotal,
		amountProcessed: amtProcessed,
		partnerLatency:  partnerLat,
	}, nil
}

// RecordOperation increments financial_operation_total with operation and status labels.
func (b *BusinessMetrics) RecordOperation(ctx context.Context, op OperationLabel, status StatusBucket) {
	b.operationTotal.Increment(ctx,
		observability.String("operation", op.String()),
		observability.String("status", status.String()),
	)
}

// RecordAmountProcessed implements ports.FinancialMetricsRecorder.
// Increments financial_amount_processed (BRL centavos) for successful write operations.
func (b *BusinessMetrics) RecordAmountProcessed(ctx context.Context, money vos.Money, txType vos.TransactionType, method vos.PaymentMethod) {
	cents, err := FromMoney(money)
	if err != nil {
		return
	}
	b.amountProcessed.Add(ctx, int64(cents),
		observability.String("transaction_type", txType.String()),
		observability.String("payment_method", method.String()),
	)
}

// RecordPartnerLatency records a partner integration latency sample.
// Declared for contract completeness; no active call sites in this delivery.
func (b *BusinessMetrics) RecordPartnerLatency(ctx context.Context, partner PartnerLabel, d time.Duration) {
	b.partnerLatency.Record(ctx, d.Seconds(),
		observability.String("partner", partner.String()),
	)
}
