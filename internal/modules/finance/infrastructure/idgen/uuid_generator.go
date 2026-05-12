package idgen

import (
	"github.com/google/uuid"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/ports"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
)

var _ ports.IDGenerator = (*UUIDGenerator)(nil)

// UUIDGenerator is the production implementation of ports.IDGenerator using google/uuid.
type UUIDGenerator struct{}

func NewUUIDGenerator() *UUIDGenerator { return &UUIDGenerator{} }

func (UUIDGenerator) NewTransactionID() vos.TransactionID {
	return vos.TransactionID(uuid.NewString())
}

func (UUIDGenerator) NewInstallmentID() vos.InstallmentID {
	return vos.InstallmentID(uuid.NewString())
}

func (UUIDGenerator) NewInvoiceID() vos.InvoiceID {
	return vos.InvoiceID(uuid.NewString())
}
