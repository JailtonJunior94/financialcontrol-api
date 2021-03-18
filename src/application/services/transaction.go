package services

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/responses"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/mappings"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/usecases"
)

type TransactionService struct {
	TransactionRepository interfaces.ITransactionRepository
}

func NewTransactionService(r interfaces.ITransactionRepository) usecases.ITransactionService {
	return &TransactionService{TransactionRepository: r}
}

func (s *TransactionService) CreateTransaction(request *requests.TransactionRequest, userId string) *responses.HttpResponse {
	newTransaction := mappings.ToTransactionEntity(request, userId)

	transaction, err := s.TransactionRepository.AddTransaction(newTransaction)
	if err != nil {
		return responses.ServerError()
	}

	return responses.Created(mappings.ToTransactionResponse(transaction))
}

func (s *TransactionService) CreateTransactionItem(request *requests.TransactionItemRequest, transactionId string, userId string) *responses.HttpResponse {
	newTransactionItem := mappings.ToTransactionItemEntity(request, transactionId)

	transactionItem, err := s.TransactionRepository.AddTransactionItem(newTransactionItem)
	if err != nil {
		return responses.ServerError()
	}

	transaction, err := s.TransactionRepository.GetTransactionById(transactionId, userId)
	if err != nil {
		return responses.ServerError()
	}

	items, err := s.TransactionRepository.GetItemByTransactionId(transactionId)
	if err != nil {
		return responses.ServerError()
	}

	transaction.AddItems(items)
	transaction.UpdatingValues()

	if _, err := s.TransactionRepository.UpdateTransaction(transaction); err != nil {
		return responses.ServerError()
	}
	return responses.Created(mappings.ToTransactionItemResponse(transactionItem))
}

func (s *TransactionService) Transactions(userId string) *responses.HttpResponse {
	transactions, err := s.TransactionRepository.GetTransactions(userId)
	if err != nil {
		return responses.ServerError()
	}

	return responses.Ok(mappings.ToManyTransactionResponse(transactions))
}

func (s *TransactionService) TransactionById(id string, userId string) *responses.HttpResponse {
	transaction, err := s.TransactionRepository.GetTransactionById(id, userId)
	if err != nil {
		return responses.ServerError()
	}

	if transaction == nil {
		return responses.NotFound("Não foi encontrado nenhuma Transação")
	}

	items, err := s.TransactionRepository.GetItemByTransactionId(id)
	if err != nil {
		return responses.ServerError()
	}
	transaction.AddItems(items)

	return responses.Ok(mappings.ToTransactionResponse(transaction))
}
