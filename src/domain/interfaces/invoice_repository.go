package interfaces

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/src/domain/entities"
)

type IInvoiceRepository interface {
	GetInvoiceByCardId(userId, cardId string) (invoices []entities.Invoice, err error)
	GetInvoiceByDate(startDate, endDate time.Time, cardId string) (invoice *entities.Invoice, err error)
	AddInvoice(p *entities.Invoice) (invoice *entities.Invoice, err error)
	UpdateInvoice(p *entities.Invoice) (invoice *entities.Invoice, err error)
	GetInvoiceItemByInvoiceId(invoiceId, cardId, userId string) (items []entities.InvoiceItem, err error)
	AddInvoiceItem(p *entities.InvoiceItem) (invoiceItem *entities.InvoiceItem, err error)
	GetLastInvoiceControl() (invoiceControl int64, err error)
	GetInvoicesCategories(startDate, endDate time.Time, cardId string) (invoiceCategories []entities.InvoiceCategories, err error)
}
