package services

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/responses"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/mappings"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/customErrors"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/usecases"
	"github.com/jailtonjunior94/financialcontrol-api/src/shared"
)

type TransactionService struct {
	TransactionRepository interfaces.ITransactionRepository
}

func NewTransactionService(r interfaces.ITransactionRepository) usecases.ITransactionService {
	return &TransactionService{TransactionRepository: r}
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
		return responses.NotFound(customErrors.TransactionItemNotFound)
	}

	items, err := s.TransactionRepository.GetItemByTransactionId(id)
	if err != nil {
		return responses.ServerError()
	}
	transaction.AddItems(items)

	return responses.Ok(mappings.ToTransactionResponse(transaction))
}

func (s *TransactionService) CreateTransaction(request *requests.TransactionRequest, userId string) *responses.HttpResponse {
	time := shared.NewTime(shared.Time{Now: request.Date})

	isExist, err := s.TransactionRepository.GetTransactionByDate(time.StartDate(), time.EndDate(), userId)
	if err != nil {
		return responses.ServerError()
	}

	if isExist != nil {
		return responses.BadRequest(customErrors.TransactionExists)
	}

	newTransaction := mappings.ToTransactionEntity(request, userId)
	transaction, err := s.TransactionRepository.AddTransaction(newTransaction)
	if err != nil {
		return responses.ServerError()
	}

	return responses.Created(mappings.ToTransactionResponse(transaction))
}

func (s *TransactionService) CloneTransaction(id, userId string) *responses.HttpResponse {
	t, err := s.TransactionRepository.GetTransactionById(id, userId)
	if err != nil {
		return responses.ServerError()
	}

	items, err := s.TransactionRepository.GetItemByTransactionId(t.ID)
	if err != nil {
		return responses.ServerError()
	}

	newTransaction := entities.NewTransactionWithValues(t.Date.AddDate(0, 1, 0), userId, t.Total, t.Income, t.Outcome)
	transactionsItems := make([]entities.TransactionItem, len(items), cap(items))

	for index, item := range items {
		newItem := entities.NewTransactionItem(newTransaction.ID, item.Title, item.Type, item.Value)
		transactionsItems[index] = *newItem
	}

	newTransaction.AddItems(transactionsItems)
	if err := s.TransactionRepository.AddRangeTransactionItems(newTransaction, transactionsItems); err != nil {
		return responses.ServerError()
	}

	return responses.Ok(mappings.ToTransactionResponse(newTransaction))
}

func (s *TransactionService) TransactionItemById(transactionId, id string) *responses.HttpResponse {
	item, err := s.TransactionRepository.GetTransactionItemsById(transactionId, id)
	if err != nil {
		return responses.ServerError()
	}

	if item == nil {
		return responses.NotFound(customErrors.TransactionItemNotFound)
	}

	return responses.Ok(mappings.ToTransactionItemResponse(item))
}

func (s *TransactionService) CreateTransactionItem(request *requests.TransactionItemRequest, transactionId string, userId string) *responses.HttpResponse {
	newTransactionItem := mappings.ToTransactionItemEntity(request, transactionId)

	transactionItem, err := s.TransactionRepository.AddTransactionItem(newTransactionItem)
	if err != nil {
		return responses.ServerError()
	}

	if err := s.updatingTransactionValues(transactionId, userId); err != nil {
		return responses.ServerError()
	}

	return responses.Created(mappings.ToTransactionItemResponse(transactionItem))
}

func (s *TransactionService) UpdateTransactionItem(transactionId, id, userId string, request *requests.TransactionItemRequest) *responses.HttpResponse {
	item, err := s.TransactionRepository.GetTransactionItemsById(transactionId, id)
	if err != nil {
		return responses.ServerError()
	}

	if item == nil {
		return responses.NotFound(customErrors.TransactionItemNotFound)
	}

	item.UpdateTransactionItem(request.Title, request.Type, request.Value)
	item, err = s.TransactionRepository.UpdateTransactionItem(item)
	if err != nil {
		return responses.ServerError()
	}

	if err := s.updatingTransactionValues(item.TransactionId, userId); err != nil {
		return responses.ServerError()
	}
	return responses.Ok(mappings.ToTransactionItemResponse(item))
}

func (s *TransactionService) MarkAsPaidTransactionItem(transactionId, id, userId string, request *requests.TransactionMarkAsPaid) *responses.HttpResponse {
	item, err := s.TransactionRepository.GetTransactionItemsById(transactionId, id)
	if err != nil {
		return responses.ServerError()
	}

	if item == nil {
		return responses.NotFound(customErrors.TransactionItemNotFound)
	}

	item.MarkAsPaid(request.MarkAsPaid)
	_, err = s.TransactionRepository.UpdateTransactionItem(item)
	if err != nil {
		return responses.ServerError()
	}

	return responses.NoContent()
}

func (s *TransactionService) RemoveTransactionItem(transactionId, id, userId string) *responses.HttpResponse {
	item, err := s.TransactionRepository.GetTransactionItemsById(transactionId, id)
	if err != nil {
		return responses.ServerError()
	}

	if item == nil {
		return responses.NotFound(customErrors.TransactionItemNotFound)
	}

	item.UpdateStatus(false)
	_, err = s.TransactionRepository.UpdateTransactionItem(item)
	if err != nil {
		return responses.ServerError()
	}

	if err := s.updatingTransactionValues(item.TransactionId, userId); err != nil {
		return responses.ServerError()
	}
	return responses.NoContent()
}

func (s *TransactionService) updatingTransactionValues(transactionId, userId string) error {
	transaction, err := s.TransactionRepository.GetTransactionById(transactionId, userId)
	if err != nil {
		return err
	}

	items, err := s.TransactionRepository.GetItemByTransactionId(transactionId)
	if err != nil {
		return err
	}

	transaction.AddItems(items)
	transaction.UpdatingValues()

	if _, err := s.TransactionRepository.UpdateTransaction(transaction); err != nil {
		return err
	}
	return nil
}
