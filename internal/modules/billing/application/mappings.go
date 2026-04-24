package application

import (
	appresponses "github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos/responses"
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/shared"
)

func ToBillEntity(request *BillRequest) *entities.Bill {
	entity := new(entities.Bill)
	entity.NewBill(request.Date)
	return entity
}

func ToBillResponse(entity *entities.Bill) *appresponses.BillResponse {
	return &appresponses.BillResponse{
		ID:           entity.ID,
		Date:         shared.NewTime(shared.Time{Date: entity.Date}).FormatDate(),
		Total:        entity.Total,
		SixtyPercent: entity.SixtyPercent,
		FortyPercent: entity.FortyPercent,
		Active:       entity.Active,
		BillItems:    ToManyBillItemResponse(entity.BillItems),
	}
}

func ToManyBillResponse(entities []entities.Bill) []appresponses.BillResponse {
	if len(entities) == 0 {
		return make([]appresponses.BillResponse, 0)
	}

	response := make([]appresponses.BillResponse, 0, len(entities))
	for _, entity := range entities {
		response = append(response, appresponses.BillResponse{
			ID:           entity.ID,
			Date:         shared.NewTime(shared.Time{Date: entity.Date}).FormatDate(),
			Total:        entity.Total,
			SixtyPercent: entity.SixtyPercent,
			FortyPercent: entity.FortyPercent,
			Active:       entity.Active,
		})
	}

	return response
}

func ToBillItemEntity(request *BillItemRequest, billID string) *entities.BillItem {
	entity := new(entities.BillItem)
	entity.NewBillItem(billID, request.Title, request.Value)
	return entity
}

func ToBillItemResponse(entity *entities.BillItem) *appresponses.BillItemResponse {
	return &appresponses.BillItemResponse{
		ID:     entity.ID,
		Title:  entity.Title,
		Value:  entity.Value,
		Active: entity.Active,
	}
}

func ToManyBillItemResponse(entities []entities.BillItem) []appresponses.BillItemResponse {
	response := make([]appresponses.BillItemResponse, 0, len(entities))
	for _, entity := range entities {
		response = append(response, appresponses.BillItemResponse{
			ID:     entity.ID,
			Title:  entity.Title,
			Value:  entity.Value,
			Active: entity.Active,
		})
	}

	return response
}
