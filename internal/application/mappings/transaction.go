package mappings

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos/responses"
	transactionsdomain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/transactions/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/shared"
)

func ToTransactionEntity(request *requests.TransactionRequest, userID string) *transactionsdomain.Transaction {
	return transactionsdomain.NewTransaction(userID, request.Date)
}

func ToTransactionResponse(entity *transactionsdomain.Transaction) *responses.TransactionResponse {
	return &responses.TransactionResponse{
		ID:      entity.ID,
		Date:    shared.NewTime(shared.Time{Date: entity.Date}).FormatDate(),
		Total:   entity.Total,
		Income:  entity.Income,
		Outcome: entity.Outcome,
		Active:  entity.Active,
		Items:   ToManyTransactionItemResponse(entity.TransactionItems),
	}
}

func ToManyTransactionResponse(entities []transactionsdomain.Transaction) []responses.TransactionResponse {
	if len(entities) == 0 {
		return make([]responses.TransactionResponse, 0)
	}

	response := make([]responses.TransactionResponse, 0, len(entities))
	for _, entity := range entities {
		response = append(response, responses.TransactionResponse{
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

func ToTransactionItemEntity(request *requests.TransactionItemRequest, transactionID string) *transactionsdomain.TransactionItem {
	return transactionsdomain.NewTransactionItem(transactionID, request.Title, request.Type, request.Value)
}

func ToTransactionItemResponse(entity *transactionsdomain.TransactionItem) *responses.TransactionItemResponse {
	return &responses.TransactionItemResponse{
		ID:     entity.ID,
		Title:  entity.Title,
		Value:  entity.Value,
		Type:   entity.Type,
		IsPaid: entity.IsPaid,
		Active: entity.Active,
	}
}

func ToManyTransactionItemResponse(entities []transactionsdomain.TransactionItem) []responses.TransactionItemResponse {
	if len(entities) == 0 {
		return make([]responses.TransactionItemResponse, 0)
	}

	response := make([]responses.TransactionItemResponse, 0, len(entities))
	for _, entity := range entities {
		response = append(response, responses.TransactionItemResponse{
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
