package entities

import (
	"testing"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInstallment_Scheduled(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	id := vos.NewInstallmentID()
	txID := vos.NewTransactionID()
	invID := vos.NewInvoiceID()
	number, _ := vos.NewInstallmentNumber(1)
	total, _ := vos.NewInstallmentCount(3)
	amount, _ := vos.NewMoney("100.00")

	inst := newInstallment(id, txID, invID, number, total, amount, now)

	require.NotNil(t, inst)
	assert.Equal(t, id, inst.ID())
	assert.Equal(t, txID, inst.TransactionID())
	assert.Equal(t, invID, inst.InvoiceID())
	assert.Equal(t, vos.InstallmentStatusScheduled, inst.Status())
	assert.Equal(t, now.UTC(), inst.CreatedAt())
	assert.False(t, inst.IsDeleted())
	assert.False(t, inst.IsClosedOrPaid())
}

func TestRehydrateInstallment(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	id := vos.NewInstallmentID()
	txID := vos.NewTransactionID()
	invID := vos.NewInvoiceID()
	number, _ := vos.NewInstallmentNumber(2)
	total, _ := vos.NewInstallmentCount(3)
	amount, _ := vos.NewMoney("50.00")
	origin := "Transaction:abc"

	inst := RehydrateInstallment(id, txID, invID, number, total, amount,
		vos.InstallmentStatusPaidViaInvoice, &origin, now, now, nil)

	assert.Equal(t, vos.InstallmentStatusPaidViaInvoice, inst.Status())
	assert.True(t, inst.IsClosedOrPaid())
	assert.Equal(t, &origin, inst.LegacyOrigin())
	assert.False(t, inst.IsDeleted())
}

func TestInstallment_ChangeInvoice_PackageInternal(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	id := vos.NewInstallmentID()
	txID := vos.NewTransactionID()
	invID := vos.NewInvoiceID()
	number, _ := vos.NewInstallmentNumber(1)
	total, _ := vos.NewInstallmentCount(1)
	amount, _ := vos.NewMoney("200.00")

	inst := newInstallment(id, txID, invID, number, total, amount, now)

	newInvID := vos.NewInvoiceID()
	later := now.Add(time.Hour)
	inst.changeInvoice(newInvID, vos.InstallmentStatusAnticipated, later)

	assert.Equal(t, newInvID, inst.InvoiceID())
	assert.Equal(t, vos.InstallmentStatusAnticipated, inst.Status())
	assert.Equal(t, later.UTC(), inst.UpdatedAt())
}
