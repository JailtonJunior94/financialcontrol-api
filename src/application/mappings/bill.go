package mappings

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/responses"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/src/shared"
)

func ToBillEntity(r *requests.BillRequest) (e *entities.Bill) {
	entity := new(entities.Bill)
	entity.NewBill(r.Date)

	return entity
}

func ToBillResponse(e *entities.Bill) (r *responses.BillResponse) {
	return &responses.BillResponse{
		ID:           e.ID,
		Date:         shared.NewTime(shared.Time{Date: e.Date}).FormatDate(),
		Total:        e.Total,
		SixtyPercent: e.SixtyPercent,
		FortyPercent: e.FortyPercent,
		Active:       e.Active,
		BillItems:    ToManyBillItemResponse(e.BillItems),
	}
}

func ToManyBillResponse(entities []entities.Bill) (r []responses.BillResponse) {
	if len(entities) == 0 {
		return make([]responses.BillResponse, 0)
	}

	for _, e := range entities {
		bill := responses.BillResponse{
			ID:           e.ID,
			Date:         shared.NewTime(shared.Time{Date: e.Date}).FormatDate(),
			Total:        e.Total,
			SixtyPercent: e.SixtyPercent,
			FortyPercent: e.FortyPercent,
			Active:       e.Active,
		}
		r = append(r, bill)
	}

	return r
}

func ToBillItemEntity(r *requests.BillItemRequest, billId string) (e *entities.BillItem) {
	entity := new(entities.BillItem)
	entity.NewBillItem(billId, r.Title, r.Value)

	return entity
}

func ToBillItemResponse(e *entities.BillItem) (r *responses.BillItemResponse) {
	return &responses.BillItemResponse{
		ID:     e.ID,
		Title:  e.Title,
		Value:  e.Value,
		Active: e.Active,
	}
}

func ToManyBillItemResponse(entities []entities.BillItem) (r []responses.BillItemResponse) {
	for _, e := range entities {
		item := responses.BillItemResponse{
			ID:     e.ID,
			Title:  e.Title,
			Value:  e.Value,
			Active: e.Active,
		}
		r = append(r, item)
	}

	return r
}
