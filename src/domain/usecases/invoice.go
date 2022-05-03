package usecases

import (
	"mime/multipart"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/responses"
)

type IInvoiceService interface {
	Invoices(userId, cardId string) *responses.HttpResponse
	InvoiceById(userId, cardId, id string) *responses.HttpResponse
	InvoiceCategories(startDate, endDate time.Time, cardId string) *responses.HttpResponse
	CreateInvoice(userId string, request *requests.InvoiceRequest) *responses.HttpResponse
	UpdateInvoice(id, userId string, request *requests.InvoiceRequest) *responses.HttpResponse
	ImportInvoices(userId string, request *multipart.FileHeader) *responses.HttpResponse
}
