package invoicingapp

import (
	"context"
	"mime/multipart"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/platform/persistence"
	"github.com/jailtonjunior94/financialcontrol-api/internal/platform/web"
)

type HttpResponse = web.HttpResponse

type CardRepository interface {
	GetCardById(id, userID string) (*entities.Card, error)
}

type InvoiceRepository interface {
	DeleteInvoiceItem(invoiceControl int64) error
	GetInvoiceById(id string) (*entities.Invoice, error)
	UpdateManyInvoices(invoices []*entities.Invoice) error
	GetLastInvoiceControl() (int64, error)
	AddManyInvoiceItems(invoiceItems []*entities.InvoiceItem) error
	GetInvoiceItemById(id string) (*entities.InvoiceItem, error)
	AddInvoice(invoice *entities.Invoice) (*entities.Invoice, error)
	UpdateInvoice(invoice *entities.Invoice) (*entities.Invoice, error)
	GetInvoiceByCardId(userID, cardID string) ([]entities.Invoice, error)
	AddInvoiceItem(item *entities.InvoiceItem) (*entities.InvoiceItem, error)
	GetInvoiceItemByInvoiceControl(invoiceControl int64) ([]*entities.InvoiceItem, error)
	GetInvoiceByDate(startDate, endDate time.Time, cardID string) (*entities.Invoice, error)
	GetInvoiceItemByInvoiceId(invoiceID, cardID, userID string) ([]entities.InvoiceItem, error)
	GetInvoicesCategories(startDate, endDate time.Time, cardID string) ([]entities.InvoiceCategories, error)
	FetchInvoiceByCard(cardID string) ([]persistence.InvoiceQuery, error)
	GetInvoices(date time.Time) (*persistence.InvoiceRead, error)
}

type InvoiceChangedPublisher interface {
	PublishInvoiceChanged(ctx context.Context, invoiceID string) error
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

type ClaimsResolver interface {
	UserID(authorizationHeader string) (string, error)
}
