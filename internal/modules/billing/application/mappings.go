package application

import (
	appresponses "github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos/responses"
	billingdomain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/billing/domain"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/shared"
)

func ToBillEntity(request *BillRequest) *billingdomain.Bill {
	entity := new(billingdomain.Bill)
	entity.NewBill(request.Date)
	return entity
}

func ToBillResponse(entity *billingdomain.Bill) *appresponses.BillResponse {
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

func ToManyBillResponse(entities []billingdomain.Bill) []appresponses.BillResponse {
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

func ToBillItemEntity(request *BillItemRequest, billID string) *billingdomain.BillItem {
	entity := new(billingdomain.BillItem)
	entity.NewBillItem(billID, request.Title, request.Value)
	return entity
}

func ToBillItemResponse(entity *billingdomain.BillItem) *appresponses.BillItemResponse {
	return &appresponses.BillItemResponse{
		ID:     entity.ID,
		Title:  entity.Title,
		Value:  entity.Value,
		Active: entity.Active,
	}
}

func ToManyBillItemResponse(entities []billingdomain.BillItem) []appresponses.BillItemResponse {
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
