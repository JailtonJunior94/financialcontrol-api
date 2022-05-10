package usecases

import (
	"mime/multipart"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/responses"
)

type IInvoiceService interface {
	DeleteInvoiceItem(id string) *responses.HttpResponse
	InvoiceById(userId, id string) *responses.HttpResponse
	Invoices(userId, cardId string) *responses.HttpResponse
	ImportInvoices(userId string, request *multipart.FileHeader) *responses.HttpResponse
	CreateInvoice(userId string, request *requests.InvoiceRequest) *responses.HttpResponse
	InvoiceCategories(startDate, endDate time.Time, cardId string) *responses.HttpResponse
	UpdateInvoice(id, userId string, request *requests.InvoiceRequest) *responses.HttpResponse
}
