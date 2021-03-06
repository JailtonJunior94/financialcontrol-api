package services

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/responses"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/mappings"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/customErrors"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/usecases"
	"github.com/jailtonjunior94/financialcontrol-api/src/shared"
)

type BillService struct {
	BillRepository interfaces.IBillRepository
}

func NewBillService(r interfaces.IBillRepository) usecases.IBillService {
	return &BillService{BillRepository: r}
}

func (s *BillService) Bills() *responses.HttpResponse {
	bills, err := s.BillRepository.GetBills()
	if err != nil {
		return responses.ServerError()
	}

	return responses.Ok(mappings.ToManyBillResponse(bills))
}

func (s *BillService) BillById(id string) *responses.HttpResponse {
	bill, err := s.BillRepository.GetBillById(id)
	if err != nil {
		return responses.ServerError()
	}

	if bill == nil {
		return responses.NotFound(customErrors.BillNotFound)
	}

	items, err := s.BillRepository.GetBillItemByBillId(id)
	if err != nil {
		return responses.ServerError()
	}
	bill.AddBillItems(items)

	return responses.Ok(mappings.ToBillResponse(bill))
}

func (s *BillService) CreateBill(request *requests.BillRequest) *responses.HttpResponse {
	time := shared.NewTime(shared.Time{Now: request.Date})

	isExist, err := s.BillRepository.GetBillByDate(time.StartDate(), time.EndDate())
	if err != nil {
		return responses.ServerError()
	}

	if isExist != nil {
		return responses.BadRequest(customErrors.BillExists)
	}

	newBill := mappings.ToBillEntity(request)
	bill, err := s.BillRepository.AddBill(newBill)
	if err != nil {
		return responses.ServerError()
	}

	return responses.Created(mappings.ToBillResponse(bill))
}

func (s *BillService) BillItemById(id, billId string) *responses.HttpResponse {
	billItem, err := s.BillRepository.GetBillItemById(id, billId)
	if err != nil {
		return responses.ServerError()
	}

	if billItem == nil {
		return responses.NotFound(customErrors.BillItemNotFound)
	}

	return responses.Ok(mappings.ToBillItemResponse(billItem))
}

func (s *BillService) CreateBillItem(request *requests.BillItemRequest, billId string) *responses.HttpResponse {
	newBillItem := mappings.ToBillItemEntity(request, billId)

	billItem, err := s.BillRepository.AddBillItem(newBillItem)
	if err != nil {
		return responses.ServerError()
	}

	if err := s.updatingBillValues(billId); err != nil {
		return responses.ServerError()
	}

	return responses.Created(mappings.ToBillItemResponse(billItem))
}

func (s *BillService) UpdateBillItem(billId, id string, request *requests.BillItemRequest) *responses.HttpResponse {
	item, err := s.BillRepository.GetBillItemById(id, billId)
	if err != nil {
		return responses.ServerError()
	}

	if item == nil {
		return responses.NotFound(customErrors.BillItemNotFound)
	}

	item.Update(request.Title, request.Value)
	item, err = s.BillRepository.UpdateBillItem(item)
	if err != nil {
		return responses.ServerError()
	}

	if err := s.updatingBillValues(billId); err != nil {
		return responses.ServerError()
	}
	return responses.Ok(mappings.ToBillItemResponse(item))
}

func (s *BillService) RemoveBillItem(billId, id string) *responses.HttpResponse {
	item, err := s.BillRepository.GetBillItemById(id, billId)
	if err != nil {
		return responses.ServerError()
	}

	if item == nil {
		return responses.NotFound(customErrors.BillItemNotFound)
	}

	item.UpdateStatus(false)
	_, err = s.BillRepository.UpdateBillItem(item)
	if err != nil {
		return responses.ServerError()
	}

	if err := s.updatingBillValues(item.BillId); err != nil {
		return responses.ServerError()
	}
	return responses.NoContent()
}

func (s *BillService) updatingBillValues(id string) error {
	bill, err := s.BillRepository.GetBillById(id)
	if err != nil {
		return err
	}

	items, err := s.BillRepository.GetBillItemByBillId(id)
	if err != nil {
		return err
	}

	bill.AddBillItems(items)
	bill.UpdatingValues()

	if _, err := s.BillRepository.UpdateBill(bill); err != nil {
		return err
	}
	return nil
}
