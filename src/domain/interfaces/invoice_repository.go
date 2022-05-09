package interfaces

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/src/domain/entities"
)

type IInvoiceRepository interface {
	DeleteInvoiceItem(invoiceControl int64) error
	GetInvoiceById(id string) (*entities.Invoice, error)
	UpdateManyInvoices(invoices []*entities.Invoice) error
	GetLastInvoiceControl() (invoiceControl int64, err error)
	AddManyInvoiceItems(invoiceItems []*entities.InvoiceItem) error
	GetInvoiceItemById(id string) (item *entities.InvoiceItem, err error)
	AddInvoice(p *entities.Invoice) (invoice *entities.Invoice, err error)
	UpdateInvoice(p *entities.Invoice) (invoice *entities.Invoice, err error)
	GetInvoiceByCardId(userId, cardId string) (invoices []entities.Invoice, err error)
	AddInvoiceItem(p *entities.InvoiceItem) (invoiceItem *entities.InvoiceItem, err error)
	GetInvoiceByDate(startDate, endDate time.Time, cardId string) (invoice *entities.Invoice, err error)
	GetInvoiceItemByInvoiceId(invoiceId, cardId, userId string) (items []entities.InvoiceItem, err error)
	GetInvoicesCategories(startDate, endDate time.Time, cardId string) (invoiceCategories []entities.InvoiceCategories, err error)
}
