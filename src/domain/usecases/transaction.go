package usecases

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/responses"
)

type ITransactionService interface {
	Transactions(userId string) *responses.HttpResponse
	TransactionById(id string, userId string) *responses.HttpResponse
	CreateTransaction(request *requests.TransactionRequest, userId string) *responses.HttpResponse
	CloneTransaction(id, userId string) *responses.HttpResponse

	TransactionItemById(transactionId, id string) *responses.HttpResponse
	CreateTransactionItem(request *requests.TransactionItemRequest, transactionId string, userId string) *responses.HttpResponse
	UpdateTransactionItem(transactionId, id, userId string, request *requests.TransactionItemRequest) *responses.HttpResponse
	MarkAsPaidTransactionItem(transactionId, id, userId string, request *requests.TransactionMarkAsPaid) *responses.HttpResponse
	RemoveTransactionItem(transactionId, id, userId string) *responses.HttpResponse
}
