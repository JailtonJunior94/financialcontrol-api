package metrics

// MetricName is a typed constant for OTel instrument names.
// Values are frozen by RF-15.5 — rename requires a new PRD + deprecation period.
type MetricName string

const (
	MetricFinancialOperationTotal   MetricName = "financial_operation_total"
	MetricFinancialAmountProcessed  MetricName = "financial_amount_processed"
	MetricPartnerIntegrationLatency MetricName = "partner_integration_latency"
)
