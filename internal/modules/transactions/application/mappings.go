package application

import (
	transactionsdomain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/transactions/domain"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/shared"
)

func ToTransactionEntity(request *TransactionRequest, userID string) *transactionsdomain.Transaction {
	return transactionsdomain.NewTransaction(userID, request.Date)
}

func ToTransactionResponse(entity *transactionsdomain.Transaction) *TransactionResponse {
	return &TransactionResponse{
		ID:      entity.ID,
		Date:    shared.NewTime(shared.Time{Date: entity.Date}).FormatDate(),
		Total:   entity.Total,
		Income:  entity.Income,
		Outcome: entity.Outcome,
		Active:  entity.Active,
		Items:   ToManyTransactionItemResponse(entity.TransactionItems),
	}
}

func ToManyTransactionResponse(entities []transactionsdomain.Transaction) (response []TransactionResponse) {
	if len(entities) == 0 {
		return make([]TransactionResponse, 0)
	}

	for _, entity := range entities {
		response = append(response, TransactionResponse{
			ID:      entity.ID,
			Date:    shared.NewTime(shared.Time{Date: entity.Date}).FormatDate(),
			Total:   entity.Total,
			Income:  entity.Income,
			Outcome: entity.Outcome,
			Active:  entity.Active,
		})
	}

	return response
}

func ToTransactionItemEntity(request *TransactionItemRequest, transactionID string) *transactionsdomain.TransactionItem {
	return transactionsdomain.NewTransactionItem(transactionID, request.Title, request.Type, request.Value)
}

func ToTransactionItemResponse(entity *transactionsdomain.TransactionItem) *TransactionItemResponse {
	return &TransactionItemResponse{
		ID:     entity.ID,
		Title:  entity.Title,
		Value:  entity.Value,
		Type:   entity.Type,
		IsPaid: entity.IsPaid,
		Active: entity.Active,
	}
}

func ToManyTransactionItemResponse(entities []transactionsdomain.TransactionItem) (response []TransactionItemResponse) {
	if len(entities) == 0 {
		return make([]TransactionItemResponse, 0)
	}

	for _, entity := range entities {
		response = append(response, TransactionItemResponse{
			ID:     entity.ID,
			Title:  entity.Title,
			Value:  entity.Value,
			Type:   entity.Type,
			IsPaid: entity.IsPaid,
			Active: entity.Active,
		})
	}

	return response
}
