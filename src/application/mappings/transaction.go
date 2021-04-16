package mappings

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/responses"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/src/shared"
)

func ToTransactionEntity(r *requests.TransactionRequest, userId string) (e *entities.Transaction) {
	entity := new(entities.Transaction)
	entity.NewTransaction(userId, r.Date)

	return entity
}

func ToTransactionResponse(e *entities.Transaction) (r *responses.TransactionResponse) {
	return &responses.TransactionResponse{
		ID:      e.ID,
		Date:    shared.NewTime(shared.Time{Date: e.Date}).FormatDate(),
		Total:   e.Total,
		Income:  e.Income,
		Outcome: e.Outcome,
		Active:  e.Active,
		Items:   ToManyTransactionItemResponse(e.TransactionItems),
	}
}

func ToManyTransactionResponse(entities []entities.Transaction) (r []responses.TransactionResponse) {
	if len(entities) == 0 {
		return make([]responses.TransactionResponse, 0)
	}

	for _, e := range entities {
		transaction := responses.TransactionResponse{
			ID:      e.ID,
			Date:    shared.NewTime(shared.Time{Date: e.Date}).FormatDate(),
			Total:   e.Total,
			Income:  e.Income,
			Outcome: e.Outcome,
			Active:  e.Active,
		}
		r = append(r, transaction)
	}

	return r
}

func ToTransactionItemEntity(r *requests.TransactionItemRequest, transactionId string) (e *entities.TransactionItem) {
	entity := new(entities.TransactionItem)
	entity.NewTransactionItem(transactionId, r.Title, r.Type, r.Value)

	return entity
}

func ToTransactionItemResponse(e *entities.TransactionItem) (r *responses.TransactionItemResponse) {
	return &responses.TransactionItemResponse{
		ID:     e.ID,
		Title:  e.Title,
		Value:  e.Value,
		Type:   e.Type,
		Active: e.Active,
	}
}

func ToManyTransactionItemResponse(entities []entities.TransactionItem) (r []responses.TransactionItemResponse) {
	for _, e := range entities {
		item := responses.TransactionItemResponse{
			ID:     e.ID,
			Title:  e.Title,
			Value:  e.Value,
			Type:   e.Type,
			Active: e.Active,
		}
		r = append(r, item)
	}

	return r
}
