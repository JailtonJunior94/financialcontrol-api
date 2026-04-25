package infrastructure

import (
	gosync "sync"

	planningsync "github.com/jailtonjunior94/financialcontrol-api/internal/modules/planning/sync"
)

// SyncAdapter implements planning.SyncPort by orchestrating the legacy
// operational use cases (UpdateTransactionBill and UpdateTransactionUseCase)
// that synchronise transaction values from invoices and bills.
//
// The card IDs are configuration values that identify which cards to sync.
// They live here (infrastructure) rather than in business logic because they
// are deployment-specific identifiers, not domain rules.
type SyncAdapter struct {
	updateBill *planningsync.UpdateTransactionBill
	updateTx   *planningsync.UpdateTransactionUseCase
	cardIDs    []string
}

func NewSyncAdapter(
	updateBill *planningsync.UpdateTransactionBill,
	updateTx *planningsync.UpdateTransactionUseCase,
) *SyncAdapter {
	return &SyncAdapter{
		updateBill: updateBill,
		updateTx:   updateTx,
		cardIDs: []string{
			"B4351E7E-F9AC-4A84-A113-A0E159303281",
			"FF8C5393-2C43-4AE4-92F7-42AF4DD3AF08",
			"45DE5288-D5D0-471A-BF18-09FE1FD2FC86",
			"4FAE4733-FB19-4F0C-A678-3C6B7588F750",
		},
	}
}

func (a *SyncAdapter) Sync() error {
	var wg gosync.WaitGroup
	wg.Add(1 + len(a.cardIDs))

	go func() { _ = a.updateBill.Execute(&wg) }()
	for _, cardID := range a.cardIDs {
		id := cardID
		go func() { _ = a.updateTx.Execute(&wg, id) }()
	}

	wg.Wait()
	return nil
}
