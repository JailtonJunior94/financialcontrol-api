package metrics

import (
	"errors"
	"math"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/shopspring/decimal"
)

// ErrAmountOverflow is returned when a Money value cannot be represented as int64 centavos.
var ErrAmountOverflow = errors.New("amount exceeds int64 range in BRL centavos")

// OperationLabel is a closed-set value object for the `operation` metric label.
// Closed list: create_transaction, update_transaction, delete_transaction,
// pay_invoice, refund_transaction, anticipate_installment, unknown.
// Values outside this list resolve to unknown (RF-04.1).
type OperationLabel string

const (
	OperationCreateTransaction     OperationLabel = "create_transaction"
	OperationUpdateTransaction     OperationLabel = "update_transaction"
	OperationDeleteTransaction     OperationLabel = "delete_transaction"
	OperationPayInvoice            OperationLabel = "pay_invoice"
	OperationRefundTransaction     OperationLabel = "refund_transaction"
	OperationAnticipateInstallment OperationLabel = "anticipate_installment"
	OperationUnknown               OperationLabel = "unknown"
)

var validOperationLabels = map[OperationLabel]struct{}{
	OperationCreateTransaction:     {},
	OperationUpdateTransaction:     {},
	OperationDeleteTransaction:     {},
	OperationPayInvoice:            {},
	OperationRefundTransaction:     {},
	OperationAnticipateInstallment: {},
	OperationUnknown:               {},
}

// ParseOperationLabel returns the OperationLabel for s.
// Returns OperationUnknown for any value not in the closed list.
func ParseOperationLabel(s string) OperationLabel {
	op := OperationLabel(s)
	if _, ok := validOperationLabels[op]; ok {
		return op
	}
	return OperationUnknown
}

func (o OperationLabel) String() string { return string(o) }

// StatusBucket is the metric label derived from an HTTP status code.
// Mapping per RF-01.1: 2xx/3xx → success, 4xx → declined, 5xx → technical_error.
type StatusBucket string

const (
	StatusSuccess        StatusBucket = "success"
	StatusDeclined       StatusBucket = "declined"
	StatusTechnicalError StatusBucket = "technical_error"
)

// StatusFromHTTP maps an HTTP status code to a StatusBucket.
func StatusFromHTTP(code int) StatusBucket {
	switch {
	case code >= 200 && code < 400:
		return StatusSuccess
	case code >= 400 && code < 500:
		return StatusDeclined
	default:
		return StatusTechnicalError
	}
}

func (s StatusBucket) String() string { return string(s) }

// PartnerLabel is a closed-set value object for the `partner` metric label.
// The list is empty in this delivery (no active partner integrations).
// Any value resolves to unknown (RF-04.1).
type PartnerLabel string

const PartnerUnknown PartnerLabel = "unknown"

// ParsePartnerLabel returns PartnerUnknown since no partners are active in this delivery.
func ParsePartnerLabel(_ string) PartnerLabel {
	return PartnerUnknown
}

func (p PartnerLabel) String() string { return string(p) }

// AmountCents represents a monetary amount in BRL centavos as a non-negative int64.
// Constructed from vos.Money via FromMoney.
type AmountCents int64

// FromMoney converts m to BRL centavos (int64).
// Returns ErrAmountOverflow when the value exceeds int64 range.
func FromMoney(m vos.Money) (AmountCents, error) {
	cents := m.Amount().Mul(decimal.NewFromInt(100))

	maxInt64 := decimal.NewFromInt(math.MaxInt64)
	minInt64 := decimal.NewFromInt(math.MinInt64)
	if cents.GreaterThan(maxInt64) || cents.LessThan(minInt64) {
		return 0, ErrAmountOverflow
	}

	return AmountCents(cents.IntPart()), nil
}

// EnvironmentTag identifies the service environment in MSSQL application_name.
// Produces "financialcontrol-api-{env}"; defaults to "unknown" for empty env.
type EnvironmentTag string

// NewEnvironmentTag constructs an EnvironmentTag from an environment string.
func NewEnvironmentTag(env string) EnvironmentTag {
	if env == "" {
		return EnvironmentTag("unknown")
	}
	return EnvironmentTag(env)
}

// AppName returns the MSSQL application_name value: "financialcontrol-api-{env}".
func (e EnvironmentTag) AppName() string {
	return "financialcontrol-api-" + string(e)
}

func (e EnvironmentTag) String() string { return string(e) }
