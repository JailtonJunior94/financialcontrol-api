package application

import (
	billingdomain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/billing/domain"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/shared"
)

func ToBillEntity(request *BillRequest) *billingdomain.Bill {
	entity := new(billingdomain.Bill)
	entity.NewBill(request.Date)
	return entity
}

func ToBillResponse(entity *billingdomain.Bill) *BillResponse {
	return &BillResponse{
		ID:           entity.ID,
		Date:         shared.NewTime(shared.Time{Date: entity.Date}).FormatDate(),
		Total:        entity.Total,
		SixtyPercent: entity.SixtyPercent,
		FortyPercent: entity.FortyPercent,
		Active:       entity.Active,
		BillItems:    ToManyBillItemResponse(entity.BillItems),
	}
}

func ToManyBillResponse(entities []billingdomain.Bill) []BillResponse {
	if len(entities) == 0 {
		return make([]BillResponse, 0)
	}

	response := make([]BillResponse, 0, len(entities))
	for _, entity := range entities {
		response = append(response, BillResponse{
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

func ToBillItemResponse(entity *billingdomain.BillItem) *BillItemResponse {
	return &BillItemResponse{
		ID:     entity.ID,
		Title:  entity.Title,
		Value:  entity.Value,
		Active: entity.Active,
	}
}

func ToManyBillItemResponse(entities []billingdomain.BillItem) []BillItemResponse {
	response := make([]BillItemResponse, 0, len(entities))
	for _, entity := range entities {
		response = append(response, BillItemResponse{
			ID:     entity.ID,
			Title:  entity.Title,
			Value:  entity.Value,
			Active: entity.Active,
		})
	}

	return response
}
