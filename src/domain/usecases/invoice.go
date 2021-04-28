package usecases

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/responses"
)

type IInvoiceService interface {
	Invoices(userId, cardId string) *responses.HttpResponse
	InvoiceById(userId, cardId, id string) *responses.HttpResponse
	CreateInvoice(userId string, request *requests.InvoiceRequest) *responses.HttpResponse
}
