package ports

import (
	"context"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
)

// FinancialMetricsRecorder is the out-port consumed by use cases that need to
// emit business metrics on successful writes. The implementation lives in
// internal/bootstrap/observability/metrics and is injected via finance.Deps.
type FinancialMetricsRecorder interface {
	RecordAmountProcessed(ctx context.Context, amount vos.Money, txType vos.TransactionType, method vos.PaymentMethod)
}

// NoopRecorder is the default for tests and for code paths that opt out of metrics.
type NoopRecorder struct{}

func (NoopRecorder) RecordAmountProcessed(context.Context, vos.Money, vos.TransactionType, vos.PaymentMethod) {
}
