package vos

import (
	"fmt"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
)

// InstallmentNumber represents the sequential number of an installment (1..N).
type InstallmentNumber struct {
	value int
}

// NewInstallmentNumber constructs an InstallmentNumber.
// Returns ErrInstallmentCountOutOfRange when n < 1.
func NewInstallmentNumber(n int) (InstallmentNumber, error) {
	if n < 1 {
		return InstallmentNumber{}, domain.ErrInstallmentCountOutOfRange
	}
	return InstallmentNumber{value: n}, nil
}

func (n InstallmentNumber) Value() int     { return n.value }
func (n InstallmentNumber) String() string { return fmt.Sprintf("%d", n.value) }
