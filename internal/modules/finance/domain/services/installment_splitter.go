package services

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/ports"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
)

// InstallmentSplitter splits a total Money amount into N equal parts using half-even
// rounding. The first part absorbs any rounding residue, guaranteeing sum(parts)==total (RF-10).
type InstallmentSplitter struct{}

func NewInstallmentSplitter() *InstallmentSplitter { return &InstallmentSplitter{} }

// Split divides total into count parts, generating installment IDs via ids.
// Returns (amounts, installmentIDs) in the same order (index 0 = installment #1).
func (s *InstallmentSplitter) Split(
	total vos.Money,
	count vos.InstallmentCount,
	ids ports.IDGenerator,
) ([]vos.Money, []vos.InstallmentID) {
	n := count.Value()
	parts := total.DivideEvenly(n)
	installmentIDs := make([]vos.InstallmentID, n)
	for i := range installmentIDs {
		installmentIDs[i] = ids.NewInstallmentID()
	}
	return parts, installmentIDs
}
