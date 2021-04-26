package usecases

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/responses"
)

type IInvoiceService interface {
	CreateInvoice(userId string, request *requests.InvoiceRequest) *responses.HttpResponse
}
