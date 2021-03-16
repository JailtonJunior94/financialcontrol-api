package usecases

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/responses"
)

type ITransactionService interface {
	CreateTransaction(request *requests.TransactionRequest, userId string) *responses.HttpResponse
	CreateTransactionItem(request *requests.TransactionItemRequest, transactionId string) *responses.HttpResponse
	TransactionById(id string) *responses.HttpResponse
}
