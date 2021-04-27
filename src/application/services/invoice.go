package services

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/responses"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/mappings"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/usecases"
	"github.com/jailtonjunior94/financialcontrol-api/src/shared"
)

type InvoiceService struct {
	CardRepository    interfaces.ICardRepository
	InvoiceRepository interfaces.IInvoiceRepository
}

func NewInvoiceService(r interfaces.ICardRepository, i interfaces.IInvoiceRepository) usecases.IInvoiceService {
	return &InvoiceService{CardRepository: r, InvoiceRepository: i}
}

func (u *InvoiceService) Invoices(userId string, cardId string) *responses.HttpResponse {
	invoices, err := u.InvoiceRepository.GetInvoiceByCardId(userId, cardId)
	if err != nil {
		return responses.ServerError()
	}

	return responses.Ok(mappings.ToManyInvoiceResponse(invoices))
}

func (u *InvoiceService) CreateInvoice(userId string, request *requests.InvoiceRequest) *responses.HttpResponse {
	card, err := u.CardRepository.GetCardById(request.CardId, userId)
	if err != nil {
		return responses.ServerError()
	}

	startDate, endDate := u.getDates(request.PurchaseDate)
	for i := 0; i < request.QuantityInvoice; i++ {
		invoice, err := u.InvoiceRepository.GetInvoiceByDate(startDate.AddDate(0, i, 0), endDate.AddDate(0, i, 0), card.ID)
		if err != nil {
			return responses.ServerError()
		}

		if invoice == nil {
			newInvoice, err := u.InvoiceRepository.AddInvoice(mappings.ToInvoiceEntity(request, startDate.AddDate(0, i, 0), 0))
			if err != nil {
				return responses.ServerError()
			}

			if err := u.addInvoiceItemAndUpdateTotal(newInvoice, request, i); err != nil {
				return responses.ServerError()
			}
			continue
		}

		if err := u.addInvoiceItemAndUpdateTotal(invoice, request, i); err != nil {
			return responses.ServerError()
		}
	}

	return responses.Created(map[string]string{"message": "Cadastrado com sucesso"})
}

func (u *InvoiceService) getDates(purchaseDate time.Time) (startDate, endDate time.Time) {
	time := shared.NewTime(shared.Time{Now: purchaseDate})
	closing := time.EndDate().AddDate(0, 0, -6)

	if purchaseDate.Day() > closing.Day() {
		startDate = time.StartDate().AddDate(0, 2, 0)
		endDate = time.EndDate().AddDate(0, 2, 0)
	} else {
		startDate = time.StartDate().AddDate(0, 1, 0)
		endDate = time.EndDate().AddDate(0, 1, 0)
	}

	return startDate, endDate
}

func (u *InvoiceService) addInvoiceItemAndUpdateTotal(invoice *entities.Invoice, request *requests.InvoiceRequest, installment int) error {
	_, err := u.InvoiceRepository.AddInvoiceItem(mappings.ToInvoiceItemEntity(request, invoice.ID, installment+1))
	if err != nil {
		return err
	}

	items, err := u.InvoiceRepository.GetInvoiceItemByInvoiceId(invoice.ID)
	if err != nil {
		return err
	}

	invoice.AddInvoiceItems(items)
	invoice.UpdatingValues()

	_, err = u.InvoiceRepository.UpdateInvoice(invoice)
	if err != nil {
		return err
	}

	return nil
}
