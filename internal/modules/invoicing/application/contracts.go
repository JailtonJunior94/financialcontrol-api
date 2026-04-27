package invoicingapp

import (
	"context"
	"mime/multipart"
	"time"

	invoicingdomain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/invoicing/domain"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/persistence"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/web"
)

type HttpResponse = web.HttpResponse

// CardView carries the card fields needed by the invoicing module.
// It is a local type to avoid cross-module domain imports.
type CardView struct {
	ID         string
	ClosingDay int
}

type CardRepository interface {
	GetCardById(id, userID string) (*CardView, error)
}

type InvoiceRepository interface {
	DeleteInvoiceItem(invoiceControl int64) error
	GetInvoiceById(id string) (*invoicingdomain.Invoice, error)
	UpdateManyInvoices(invoices []*invoicingdomain.Invoice) error
	GetLastInvoiceControl() (int64, error)
	AddManyInvoiceItems(invoiceItems []*invoicingdomain.InvoiceItem) error
	GetInvoiceItemById(id string) (*invoicingdomain.InvoiceItem, error)
	AddInvoice(invoice *invoicingdomain.Invoice) (*invoicingdomain.Invoice, error)
	UpdateInvoice(invoice *invoicingdomain.Invoice) (*invoicingdomain.Invoice, error)
	GetInvoiceByCardId(userID, cardID string) ([]invoicingdomain.Invoice, error)
	AddInvoiceItem(item *invoicingdomain.InvoiceItem) (*invoicingdomain.InvoiceItem, error)
	GetInvoiceItemByInvoiceControl(invoiceControl int64) ([]*invoicingdomain.InvoiceItem, error)
	GetInvoiceByDate(startDate, endDate time.Time, cardID string) (*invoicingdomain.Invoice, error)
	GetInvoiceItemByInvoiceId(invoiceID, cardID, userID string) ([]invoicingdomain.InvoiceItem, error)
	GetInvoicesCategories(startDate, endDate time.Time, cardID string) ([]invoicingdomain.InvoiceCategories, error)
	FetchInvoiceByCard(cardID string) ([]persistence.InvoiceQuery, error)
	GetInvoices(date time.Time) (*persistence.InvoiceRead, error)
}

type InvoiceChangedPublisher interface {
	PublishInvoiceChanged(ctx context.Context, payload invoicingdomain.InvoiceChangedPayload) error
}

type InvoiceService interface {
	DeleteInvoiceItem(id string) *HttpResponse
	InvoiceById(userID, id string) *HttpResponse
	Invoices(userID, cardID string) *HttpResponse
	ImportInvoices(userID string, request *multipart.FileHeader) *HttpResponse
	CreateInvoice(userID string, request *InvoiceRequest) *HttpResponse
	InvoiceCategories(startDate, endDate time.Time, cardID string) *HttpResponse
	UpdateInvoice(id, userID string, request *InvoiceRequest) *HttpResponse
}
