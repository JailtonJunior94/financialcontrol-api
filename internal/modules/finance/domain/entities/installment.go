package entities

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
)

// Installment is a child entity of Transaction, projected via FK in Invoice.
// Mutation is package-internal: only Transaction may call changeInvoice.
type Installment struct {
	id            vos.InstallmentID
	transactionID vos.TransactionID
	invoiceID     vos.InvoiceID
	number        vos.InstallmentNumber
	total         vos.InstallmentCount
	amount        vos.Money
	status        vos.InstallmentStatus
	legacyOrigin  *string
	createdAt     time.Time
	updatedAt     time.Time
	deletedAt     *time.Time
}

// newInstallment is the package-internal factory. Only Transaction may call it.
func newInstallment(
	id vos.InstallmentID,
	transactionID vos.TransactionID,
	invoiceID vos.InvoiceID,
	number vos.InstallmentNumber,
	total vos.InstallmentCount,
	amount vos.Money,
	now time.Time,
) *Installment {
	return &Installment{
		id:            id,
		transactionID: transactionID,
		invoiceID:     invoiceID,
		number:        number,
		total:         total,
		amount:        amount,
		status:        vos.InstallmentStatusScheduled,
		createdAt:     now.UTC(),
		updatedAt:     now.UTC(),
	}
}

// RehydrateInstallment reconstructs an Installment from persisted data.
func RehydrateInstallment(
	id vos.InstallmentID,
	transactionID vos.TransactionID,
	invoiceID vos.InvoiceID,
	number vos.InstallmentNumber,
	total vos.InstallmentCount,
	amount vos.Money,
	status vos.InstallmentStatus,
	legacyOrigin *string,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) *Installment {
	return &Installment{
		id:            id,
		transactionID: transactionID,
		invoiceID:     invoiceID,
		number:        number,
		total:         total,
		amount:        amount,
		status:        status,
		legacyOrigin:  legacyOrigin,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
		deletedAt:     deletedAt,
	}
}

// changeInvoice transitions the installment to a new invoice and status.
// Package-internal: only Transaction.Anticipate may call this.
func (i *Installment) changeInvoice(newInvoiceID vos.InvoiceID, newStatus vos.InstallmentStatus, now time.Time) {
	i.invoiceID = newInvoiceID
	i.status = newStatus
	i.updatedAt = now.UTC()
}

func (i *Installment) ID() vos.InstallmentID            { return i.id }
func (i *Installment) TransactionID() vos.TransactionID { return i.transactionID }
func (i *Installment) InvoiceID() vos.InvoiceID         { return i.invoiceID }
func (i *Installment) Number() vos.InstallmentNumber    { return i.number }
func (i *Installment) Total() vos.InstallmentCount      { return i.total }
func (i *Installment) Amount() vos.Money                { return i.amount }
func (i *Installment) Status() vos.InstallmentStatus    { return i.status }
func (i *Installment) LegacyOrigin() *string            { return i.legacyOrigin }
func (i *Installment) CreatedAt() time.Time             { return i.createdAt }
func (i *Installment) UpdatedAt() time.Time             { return i.updatedAt }
func (i *Installment) DeletedAt() *time.Time            { return i.deletedAt }
func (i *Installment) IsDeleted() bool                  { return i.deletedAt != nil }
func (i *Installment) IsClosedOrPaid() bool             { return i.status.IsClosedOrPaid() }
