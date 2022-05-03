package services

import (
	"bufio"
	"mime/multipart"
	"strings"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/responses"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/mappings"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/events"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/usecases"
	"github.com/jailtonjunior94/financialcontrol-api/src/shared"
)

type InvoiceService struct {
	CardRepository    interfaces.ICardRepository
	InvoiceRepository interfaces.IInvoiceRepository
	Dispatcher        *events.EventDispatcher
}

func NewInvoiceService(r interfaces.ICardRepository, i interfaces.IInvoiceRepository, d *events.EventDispatcher) usecases.IInvoiceService {
	return &InvoiceService{CardRepository: r, InvoiceRepository: i, Dispatcher: d}
}

func (u *InvoiceService) Invoices(userId string, cardId string) *responses.HttpResponse {
	invoices, err := u.InvoiceRepository.GetInvoiceByCardId(userId, cardId)
	if err != nil {
		return responses.ServerError()
	}

	return responses.Ok(mappings.ToManyInvoiceResponse(invoices))
}

func (u *InvoiceService) InvoiceById(userId, cardId, id string) *responses.HttpResponse {
	invoiceItems, err := u.InvoiceRepository.GetInvoiceItemByInvoiceId(id, cardId, userId)
	if err != nil {
		return responses.ServerError()
	}

	return responses.Ok(mappings.ToManyInvoiceItemResponse(invoiceItems))
}

func (u *InvoiceService) InvoiceCategories(startDate, endDate time.Time, cardId string) *responses.HttpResponse {
	start := shared.NewTime(shared.Time{Now: startDate})
	end := shared.NewTime(shared.Time{Now: endDate})

	invoiceCategories, err := u.InvoiceRepository.GetInvoicesCategories(start.StartDate(), end.EndDate(), cardId)
	if err != nil {
		return responses.ServerError()
	}

	return responses.Ok(invoiceCategories)
}

func (u *InvoiceService) CreateInvoice(userId string, request *requests.InvoiceRequest) *responses.HttpResponse {
	return u.create(userId, request)
}

func (u *InvoiceService) UpdateInvoice(id, userId string, request *requests.InvoiceRequest) *responses.HttpResponse {
	item, err := u.InvoiceRepository.GetInvoiceItemById(id)
	if err != nil {
		return responses.ServerError()
	}

	if err := u.InvoiceRepository.DeleteInvoiceItem(item.InvoiceControl); err != nil {
		return responses.ServerError()
	}

	return u.create(userId, request)
}

func (u *InvoiceService) ImportInvoices(userId string, request *multipart.FileHeader) *responses.HttpResponse {
	body, err := request.Open()
	if err != nil {
		return responses.ServerError()
	}
	defer body.Close()

	scanner := bufio.NewScanner(body)
	scanner.Scan()

	var newInvoices []*requests.InvoiceRequest

	for scanner.Scan() {
		line := scanner.Text()
		split := strings.Split(line, ";")

		invoice := requests.NewInvoiceRequest(split[0], split[1], split[2], split[3], split[4], split[5], split[6])
		newInvoices = append(newInvoices, invoice)
	}

	for _, invoiceRequest := range newInvoices {
		u.CreateInvoice(userId, invoiceRequest)
	}

	return responses.Created(map[string]string{"message": "Cadastrado com sucesso"})
}

func (u *InvoiceService) create(userId string, request *requests.InvoiceRequest) *responses.HttpResponse {
	card, err := u.CardRepository.GetCardById(request.CardId, userId)
	if err != nil {
		return responses.ServerError()
	}

	invoiceControl, err := u.InvoiceRepository.GetLastInvoiceControl()
	if err != nil {
		return responses.ServerError()
	}

	startDate, endDate := u.getDates(request.PurchaseDate, card.ClosingDay)
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

			if err := u.addInvoiceItemAndUpdateTotal(newInvoice, request, i, invoiceControl, userId); err != nil {
				return responses.ServerError()
			}
			continue
		}

		if err := u.addInvoiceItemAndUpdateTotal(invoice, request, i, invoiceControl, userId); err != nil {
			return responses.ServerError()
		}
		u.Dispatcher.Dispatch(events.NewInvoiceChangedEvent(invoice.ID))
	}

	return responses.Created(map[string]string{"message": "Cadastrado com sucesso"})
}

func (u *InvoiceService) getDates(purchaseDate time.Time, closingDay int) (startDate, endDate time.Time) {
	time := shared.NewTime(shared.Time{Now: purchaseDate})
	closing := time.EndDate().AddDate(0, 0, closingDay).AddDate(0, 0, -7)

	if purchaseDate.Day() >= closing.Day() {
		startDate = time.StartDate().AddDate(0, 2, 0)
		endDate = time.EndDate().AddDate(0, 2, 0)
	} else {
		startDate = time.StartDate().AddDate(0, 1, 0)
		endDate = time.EndDate().AddDate(0, 1, 0)
	}

	return startDate, endDate
}

func (u *InvoiceService) addInvoiceItemAndUpdateTotal(invoice *entities.Invoice, request *requests.InvoiceRequest, installment int, invoiceControl int64, userId string) error {
	_, err := u.InvoiceRepository.AddInvoiceItem(mappings.ToInvoiceItemEntity(request, invoice.ID, installment+1, invoiceControl+1))
	if err != nil {
		return err
	}

	items, err := u.InvoiceRepository.GetInvoiceItemByInvoiceId(invoice.ID, invoice.CardId, userId)
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
