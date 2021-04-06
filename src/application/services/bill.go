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
