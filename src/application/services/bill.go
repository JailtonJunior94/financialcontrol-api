package services

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/responses"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/mappings"
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
		return responses.NotFound("Não foi encontrado conta do mês")
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
		return responses.BadRequest("Já existe mês cadastrado para despesas")
	}

	newBill := mappings.ToBillEntity(request)
	bill, err := s.BillRepository.AddBill(newBill)
	if err != nil {
		return responses.ServerError()
	}

	return responses.Created(mappings.ToBillResponse(bill))
}

func (s *BillService) CreateBillItem(request *requests.BillItemRequest, billId string) *responses.HttpResponse {
	bill, err := s.BillRepository.GetBillById(billId)
	if err != nil {
		return responses.ServerError()
	}

	newBillItem := mappings.ToBillItemEntity(request, billId)

	billItem, err := s.BillRepository.AddBillItem(newBillItem)
	if err != nil {
		return responses.ServerError()
	}

	items, err := s.BillRepository.GetBillItemByBillId(bill.ID)
	if err != nil {
		return responses.ServerError()
	}

	bill.AddBillItems(items)
	bill.UpdatingValues()

	_, err = s.BillRepository.UpdateBill(bill)
	if err != nil {
		return responses.ServerError()
	}

	return responses.Created(mappings.ToBillItemResponse(billItem))
}
