package usecases

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/responses"
)

type ITransactionService interface {
	CreateTransaction(request *requests.TransactionRequest, userId string) *responses.HttpResponse
	CreateTransactionItem(request *requests.TransactionItemRequest, transactionId string, userId string) *responses.HttpResponse
	Transactions(userId string) *responses.HttpResponse
	TransactionById(id string, userId string) *responses.HttpResponse
	TransactionItemById(id string) *responses.HttpResponse
	UpdateTransactionItem(id, userId string, request *requests.TransactionItemRequest) *responses.HttpResponse
	RemoveTransactionItem(id, userId string) *responses.HttpResponse
}
