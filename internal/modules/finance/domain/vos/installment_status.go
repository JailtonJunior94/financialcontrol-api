package vos

import domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"

// InstallmentStatus enumerates the lifecycle states of an installment.
type InstallmentStatus string

const (
	InstallmentStatusScheduled      InstallmentStatus = "scheduled"
	InstallmentStatusAnticipated    InstallmentStatus = "anticipated"
	InstallmentStatusPaidViaInvoice InstallmentStatus = "paid_via_invoice"
	InstallmentStatusRefunded       InstallmentStatus = "refunded"
)

var validInstallmentStatuses = map[InstallmentStatus]struct{}{
	InstallmentStatusScheduled:      {},
	InstallmentStatusAnticipated:    {},
	InstallmentStatusPaidViaInvoice: {},
	InstallmentStatusRefunded:       {},
}

// ParseInstallmentStatus parses a raw string into an InstallmentStatus.
func ParseInstallmentStatus(s string) (InstallmentStatus, error) {
	st := InstallmentStatus(s)
	if _, ok := validInstallmentStatuses[st]; !ok {
		return "", domain.ErrInvalidTransactionType
	}
	return st, nil
}

func (s InstallmentStatus) String() string { return string(s) }

// CanTransitionTo reports whether a transition from s to target is allowed.
// Allowed transitions:
//   - scheduled   → anticipated
//   - scheduled   → paid_via_invoice
//   - scheduled   → refunded
//   - anticipated → paid_via_invoice
//   - anticipated → refunded
func (s InstallmentStatus) CanTransitionTo(target InstallmentStatus) bool {
	switch s {
	case InstallmentStatusScheduled:
		return target == InstallmentStatusAnticipated ||
			target == InstallmentStatusPaidViaInvoice ||
			target == InstallmentStatusRefunded
	case InstallmentStatusAnticipated:
		return target == InstallmentStatusPaidViaInvoice ||
			target == InstallmentStatusRefunded
	default:
		return false
	}
}

// IsClosedOrPaid reports whether the installment is in a terminal state that
// blocks edits (paid_via_invoice is treated as "closed/paid" for edit guards).
func (s InstallmentStatus) IsClosedOrPaid() bool {
	return s == InstallmentStatusPaidViaInvoice
}
