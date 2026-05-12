package vos

import domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"

// InvoiceState enumerates the lifecycle states of an invoice.
type InvoiceState string

const (
	InvoiceStateOpen   InvoiceState = "open"
	InvoiceStateClosed InvoiceState = "closed"
	InvoiceStatePaid   InvoiceState = "paid"
)

var validInvoiceStates = map[InvoiceState]struct{}{
	InvoiceStateOpen:   {},
	InvoiceStateClosed: {},
	InvoiceStatePaid:   {},
}

// ParseInvoiceState parses a raw string into an InvoiceState.
func ParseInvoiceState(s string) (InvoiceState, error) {
	st := InvoiceState(s)
	if _, ok := validInvoiceStates[st]; !ok {
		return "", domain.ErrInvalidTransactionType
	}
	return st, nil
}

func (s InvoiceState) String() string { return string(s) }

// CanTransitionTo reports whether a transition from s to target is allowed.
// Allowed transitions:
//   - open   → closed
//   - open   → paid  (implicit close then pay)
//   - closed → paid
//   - paid   → (none)
func (s InvoiceState) CanTransitionTo(target InvoiceState) bool {
	switch s {
	case InvoiceStateOpen:
		return target == InvoiceStateClosed || target == InvoiceStatePaid
	case InvoiceStateClosed:
		return target == InvoiceStatePaid
	default:
		return false
	}
}
