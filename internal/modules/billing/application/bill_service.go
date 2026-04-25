package application

import (
	billingdomain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/billing/domain"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/shared"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/web"
)

type DefaultBillService struct {
	repository BillRepository
}

func NewBillService(repository BillRepository) BillService {
	return &DefaultBillService{repository: repository}
}

func (s *DefaultBillService) Bills() *HttpResponse {
	bills, err := s.repository.GetBills()
	if err != nil {
		return web.ServerError()
	}

	return web.Ok(ToManyBillResponse(bills))
}

func (s *DefaultBillService) BillById(id string) *HttpResponse {
	bill, err := s.repository.GetBillById(id)
	if err != nil {
		return web.ServerError()
	}

	if bill == nil {
		return web.NotFound(billingdomain.BillNotFound)
	}

	items, err := s.repository.GetBillItemByBillId(id)
	if err != nil {
		return web.ServerError()
	}

	bill.AddBillItems(items)
	return web.Ok(ToBillResponse(bill))
}

func (s *DefaultBillService) CreateBill(request *BillRequest) *HttpResponse {
	timer := shared.NewTime(shared.Time{Now: request.Date})

	exists, err := s.repository.GetBillByDate(timer.StartDate(), timer.EndDate())
	if err != nil {
		return web.ServerError()
	}

	if exists != nil {
		return web.BadRequest(billingdomain.BillExists)
	}

	bill, err := s.repository.AddBill(ToBillEntity(request))
	if err != nil {
		return web.ServerError()
	}

	return web.Created(ToBillResponse(bill))
}

func (s *DefaultBillService) BillItemById(id, billID string) *HttpResponse {
	item, err := s.repository.GetBillItemById(id, billID)
	if err != nil {
		return web.ServerError()
	}

	if item == nil {
		return web.NotFound(billingdomain.BillItemNotFound)
	}

	return web.Ok(ToBillItemResponse(item))
}

func (s *DefaultBillService) CreateBillItem(request *BillItemRequest, billID string) *HttpResponse {
	item, err := s.repository.AddBillItem(ToBillItemEntity(request, billID))
	if err != nil {
		return web.ServerError()
	}

	if err := s.updateBillValues(billID); err != nil {
		return web.ServerError()
	}

	return web.Created(ToBillItemResponse(item))
}

func (s *DefaultBillService) UpdateBillItem(billID, id string, request *BillItemRequest) *HttpResponse {
	item, err := s.repository.GetBillItemById(id, billID)
	if err != nil {
		return web.ServerError()
	}

	if item == nil {
		return web.NotFound(billingdomain.BillItemNotFound)
	}

	item.Update(request.Title, request.Value)
	item, err = s.repository.UpdateBillItem(item)
	if err != nil {
		return web.ServerError()
	}

	if err := s.updateBillValues(billID); err != nil {
		return web.ServerError()
	}

	return web.Ok(ToBillItemResponse(item))
}

func (s *DefaultBillService) RemoveBillItem(billID, id string) *HttpResponse {
	item, err := s.repository.GetBillItemById(id, billID)
	if err != nil {
		return web.ServerError()
	}

	if item == nil {
		return web.NotFound(billingdomain.BillItemNotFound)
	}

	item.UpdateStatus(false)
	if _, err := s.repository.UpdateBillItem(item); err != nil {
		return web.ServerError()
	}

	if err := s.updateBillValues(item.BillId); err != nil {
		return web.ServerError()
	}

	return web.NoContent()
}

func (s *DefaultBillService) updateBillValues(id string) error {
	bill, err := s.repository.GetBillById(id)
	if err != nil {
		return err
	}

	items, err := s.repository.GetBillItemByBillId(id)
	if err != nil {
		return err
	}

	bill.AddBillItems(items)
	bill.UpdatingValues()
	_, err = s.repository.UpdateBill(bill)
	return err
}
