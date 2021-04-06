package usecases

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/responses"
)

type IBillService interface {
	Bills() *responses.HttpResponse
	CreateBill(request *requests.BillRequest) *responses.HttpResponse
}
