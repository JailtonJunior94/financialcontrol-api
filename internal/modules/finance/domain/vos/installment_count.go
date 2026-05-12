package vos

import domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"

const (
	minInstallmentCount = 1
	maxInstallmentCount = 24
)

// InstallmentCount represents the total number of installments in a purchase (1..24).
type InstallmentCount struct {
	value int
}

// NewInstallmentCount constructs an InstallmentCount.
// Returns ErrInstallmentCountOutOfRange when n is outside [1, 24].
func NewInstallmentCount(n int) (InstallmentCount, error) {
	if n < minInstallmentCount || n > maxInstallmentCount {
		return InstallmentCount{}, domain.ErrInstallmentCountOutOfRange
	}
	return InstallmentCount{value: n}, nil
}

func (c InstallmentCount) Value() int { return c.value }
