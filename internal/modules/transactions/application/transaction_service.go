package application

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/pkg/customerrors"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/web"
	transactionsdomain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/transactions/domain"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/shared"
)

type DefaultTransactionService struct {
	repository TransactionRepository
}

func NewTransactionService(repository TransactionRepository) TransactionAppService {
	return &DefaultTransactionService{repository: repository}
}

func (s *DefaultTransactionService) Transactions(userID string) *HttpResponse {
	transactions, err := s.repository.GetTransactions(userID)
	if err != nil {
		return web.ServerError()
	}

	return web.Ok(ToManyTransactionResponse(transactions))
}

func (s *DefaultTransactionService) TransactionById(id, userID string) *HttpResponse {
	transaction, err := s.repository.GetTransactionById(id, userID)
	if err != nil {
		return web.ServerError()
	}

	if transaction == nil {
		return web.NotFound(transactionsdomain.TransactionItemNotFound)
	}

	items, err := s.repository.GetItemByTransactionId(id)
	if err != nil {
		return web.ServerError()
	}

	transaction.AddItems(items)
	return web.Ok(ToTransactionResponse(transaction))
}

func (s *DefaultTransactionService) CreateTransaction(request *TransactionRequest, userID string) *HttpResponse {
	exists, err := s.isTransactionExist(request.Date, userID)
	if err != nil {
		return web.ServerError()
	}

	if exists {
		return web.BadRequest(transactionsdomain.TransactionExists)
	}

	transaction, err := s.repository.AddTransaction(ToTransactionEntity(request, userID))
	if err != nil {
		return web.ServerError()
	}

	return web.Created(ToTransactionResponse(transaction))
}

func (s *DefaultTransactionService) CloneTransaction(id, userID string) *HttpResponse {
	transaction, err := s.repository.GetTransactionById(id, userID)
	if err != nil {
		return web.ServerError()
	}

	if transaction == nil {
		return web.NotFound(transactionsdomain.TransactionItemNotFound)
	}

	items, err := s.repository.GetItemByTransactionId(transaction.ID)
	if err != nil {
		return web.ServerError()
	}

	exists, err := s.isTransactionExist(transaction.Date.AddDate(0, 1, 0), userID)
	if err != nil {
		return web.ServerError()
	}

	if exists {
		return web.BadRequest(transactionsdomain.TransactionExists)
	}

	clonedTransaction := transactionsdomain.NewTransactionWithValues(
		transaction.Date.AddDate(0, 1, 0),
		userID,
		transaction.Total,
		transaction.Income,
		transaction.Outcome,
	)
	clonedItems := make([]transactionsdomain.TransactionItem, len(items))
	for index, item := range items {
		clonedItems[index] = *transactionsdomain.NewTransactionItem(clonedTransaction.ID, item.Title, item.Type, item.Value)
	}

	clonedTransaction.AddItems(clonedItems)
	if err := s.repository.AddRangeTransactionItems(clonedTransaction, clonedItems); err != nil {
		return web.ServerError()
	}

	return web.Ok(ToTransactionResponse(clonedTransaction))
}

func (s *DefaultTransactionService) TransactionItemById(transactionID, id string) *HttpResponse {
	item, err := s.repository.GetTransactionItemsById(transactionID, id)
	if err != nil {
		return web.ServerError()
	}

	if item == nil {
		return web.NotFound(transactionsdomain.TransactionItemNotFound)
	}

	return web.Ok(ToTransactionItemResponse(item))
}

func (s *DefaultTransactionService) CreateTransactionItem(request *TransactionItemRequest, transactionID, userID string) *HttpResponse {
	item, err := s.repository.AddTransactionItem(ToTransactionItemEntity(request, transactionID))
	if err != nil {
		return web.ServerError()
	}

	if err := s.updatingTransactionValues(transactionID, userID); err != nil {
		return web.ServerError()
	}

	return web.Created(ToTransactionItemResponse(item))
}

func (s *DefaultTransactionService) UpdateTransactionItem(transactionID, id, userID string, request *TransactionItemRequest) *HttpResponse {
	item, err := s.repository.GetTransactionItemsById(transactionID, id)
	if err != nil {
		return web.ServerError()
	}

	if item == nil {
		return web.NotFound(transactionsdomain.TransactionItemNotFound)
	}

	item.UpdateTransactionItem(request.Title, request.Type, request.Value)
	item, err = s.repository.UpdateTransactionItem(item)
	if err != nil {
		return web.ServerError()
	}

	if err := s.updatingTransactionValues(item.TransactionId, userID); err != nil {
		return web.ServerError()
	}

	return web.Ok(ToTransactionItemResponse(item))
}

func (s *DefaultTransactionService) MarkAsPaidTransactionItem(transactionID, id, userID string, request *TransactionMarkAsPaid) *HttpResponse {
	item, err := s.repository.GetTransactionItemsById(transactionID, id)
	if err != nil {
		return web.ServerError()
	}

	if item == nil {
		return web.NotFound(transactionsdomain.TransactionItemNotFound)
	}

	item.MarkAsPaid(request.MarkAsPaid)
	if _, err := s.repository.UpdateTransactionItem(item); err != nil {
		return web.ServerError()
	}

	if err := s.updatingTransactionValues(item.TransactionId, userID); err != nil {
		return web.ServerError()
	}

	return web.NoContent()
}

func (s *DefaultTransactionService) RemoveTransactionItem(transactionID, id, userID string) *HttpResponse {
	item, err := s.repository.GetTransactionItemsById(transactionID, id)
	if err != nil {
		return web.ServerError()
	}

	if item == nil {
		return web.NotFound(transactionsdomain.TransactionItemNotFound)
	}

	item.UpdateStatus(false)
	if _, err := s.repository.UpdateTransactionItem(item); err != nil {
		return web.ServerError()
	}

	if err := s.updatingTransactionValues(item.TransactionId, userID); err != nil {
		return web.ServerError()
	}

	return web.NoContent()
}

func (s *DefaultTransactionService) updatingTransactionValues(transactionID, userID string) error {
	transaction, err := s.repository.GetTransactionById(transactionID, userID)
	if err != nil {
		return err
	}

	items, err := s.repository.GetItemByTransactionId(transactionID)
	if err != nil {
		return err
	}

	transaction.AddItems(items)
	transaction.UpdatingValues()
	_, err = s.repository.UpdateTransaction(transaction)
	return err
}

func (s *DefaultTransactionService) isTransactionExist(date time.Time, userID string) (bool, error) {
	timer := shared.NewTime(shared.Time{Now: date})

	transaction, err := s.repository.GetTransactionByDate(timer.StartDate(), timer.EndDate(), userID)
	if err != nil {
		return false, customerrors.InternalServerError
	}

	return transaction != nil, nil
}
