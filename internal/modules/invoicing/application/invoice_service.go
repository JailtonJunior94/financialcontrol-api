package invoicingapp

import (
	"bufio"
	"context"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	invoicingdomain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/invoicing/domain"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/shared"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/web"
)

type DefaultInvoiceService struct {
	cardRepository    CardRepository
	invoiceRepository InvoiceRepository
	eventPublisher    InvoiceChangedPublisher
}

func NewInvoiceService(
	cardRepository CardRepository,
	invoiceRepository InvoiceRepository,
	eventPublisher InvoiceChangedPublisher,
) InvoiceService {
	return &DefaultInvoiceService{
		cardRepository:    cardRepository,
		invoiceRepository: invoiceRepository,
		eventPublisher:    eventPublisher,
	}
}

func (s *DefaultInvoiceService) Invoices(userID, cardID string) *HttpResponse {
	invoices, err := s.invoiceRepository.GetInvoiceByCardId(userID, cardID)
	if err != nil {
		return web.ServerError()
	}

	return web.Ok(ToManyInvoiceResponse(invoices))
}

func (s *DefaultInvoiceService) InvoiceById(userID, id string) *HttpResponse {
	invoice, err := s.invoiceRepository.GetInvoiceById(id)
	if err != nil {
		return web.BadRequest(err)
	}

	return web.Ok(ToInvoiceResponse(invoice))
}

func (s *DefaultInvoiceService) InvoiceCategories(startDate, endDate time.Time, cardID string) *HttpResponse {
	start := shared.NewTime(shared.Time{Now: startDate})
	end := shared.NewTime(shared.Time{Now: endDate})

	categories, err := s.invoiceRepository.GetInvoicesCategories(start.StartDate(), end.EndDate(), cardID)
	if err != nil {
		return web.ServerError()
	}

	return web.Ok(categories)
}

func (s *DefaultInvoiceService) CreateInvoice(userID string, request *InvoiceRequest) *HttpResponse {
	return s.create(userID, request)
}

func (s *DefaultInvoiceService) UpdateInvoice(id, userID string, request *InvoiceRequest) *HttpResponse {
	item, err := s.invoiceRepository.GetInvoiceItemById(id)
	if err != nil {
		return web.ServerError()
	}

	if err := s.invoiceRepository.DeleteInvoiceItem(item.InvoiceControl); err != nil {
		return web.ServerError()
	}

	return s.create(userID, request)
}

func (s *DefaultInvoiceService) DeleteInvoiceItem(id string) *HttpResponse {
	item, err := s.invoiceRepository.GetInvoiceItemById(id)
	if err != nil {
		return web.ServerError()
	}

	items, err := s.invoiceRepository.GetInvoiceItemByInvoiceControl(item.InvoiceControl)
	if err != nil {
		return web.ServerError()
	}

	if err := s.invoiceRepository.DeleteInvoiceItem(item.InvoiceControl); err != nil {
		return web.ServerError()
	}

	for _, invoiceItem := range items {
		if err := s.publishInvoiceChanged(context.Background(), invoiceItem.InvoiceId); err != nil {
			return web.ServerError()
		}
	}

	return web.NoContent()
}

func (s *DefaultInvoiceService) ImportInvoices(userID string, request *multipart.FileHeader) *HttpResponse {
	body, err := request.Open()
	if err != nil {
		return web.ServerError()
	}
	defer func() {
		_ = body.Close()
	}()

	scanner := bufio.NewScanner(body)
	scanner.Scan()

	var invoices []*InvoiceRequest
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ";")
		invoices = append(invoices, NewInvoiceRequest(fields[0], fields[1], fields[2], fields[3], fields[4], fields[5], fields[6]))
	}

	for _, invoice := range invoices {
		s.CreateInvoice(userID, invoice)
	}

	return web.Created(map[string]string{"message": "Cadastrado com sucesso"})
}

func (s *DefaultInvoiceService) create(userID string, request *InvoiceRequest) *HttpResponse {
	card, err := s.cardRepository.GetCardById(request.CardId, userID)
	if err != nil {
		return web.BadRequest(err)
	}

	invoiceControl, err := s.invoiceRepository.GetLastInvoiceControl()
	if err != nil {
		return web.BadRequest(err)
	}

	startDate, endDate := s.getDates(request.PurchaseDate, card.ClosingDay)
	items := make([]*invoicingdomain.InvoiceItem, request.QuantityInvoice)
	for i := 0; i < request.QuantityInvoice; i++ {
		invoice, err := s.invoiceRepository.GetInvoiceByDate(startDate.AddDate(0, i, 0), endDate.AddDate(0, i, 0), card.ID)
		if err != nil {
			return web.ServerError()
		}

		if invoice == nil {
			invoice, err = s.invoiceRepository.AddInvoice(ToInvoiceEntity(request, startDate.AddDate(0, i, 0), 0))
			if err != nil {
				return web.ServerError()
			}

			items[i] = ToInvoiceItemEntity(request, invoice.ID, i+1, invoiceControl+1)
			continue
		}

		items[i] = ToInvoiceItemEntity(request, invoice.ID, i+1, invoiceControl+1)
	}

	if err := s.invoiceRepository.AddManyInvoiceItems(items); err != nil {
		return web.ServerError()
	}

	for _, item := range items {
		if err := s.publishInvoiceChanged(context.Background(), item.InvoiceId); err != nil {
			return web.ServerError()
		}
	}

	return web.Created(map[string]string{"message": "Cadastrado com sucesso"})
}

func (s *DefaultInvoiceService) getDates(purchaseDate time.Time, closingDay int) (time.Time, time.Time) {
	timer := shared.NewTime(shared.Time{Now: purchaseDate})
	closing := timer.EndDate().AddDate(0, 0, closingDay).AddDate(0, 0, -7)

	if purchaseDate.Day() >= closing.Day() {
		return timer.StartDate().AddDate(0, 2, 0), timer.EndDate().AddDate(0, 2, 0)
	}

	return timer.StartDate().AddDate(0, 1, 0), timer.EndDate().AddDate(0, 1, 0)
}

// publishInvoiceChanged fetches the invoice, recalculates its total, persists
// the updated value and then publishes the invoice_changed event with an
// auto-contained payload. Publishing is skipped when MarkImportTransactions
// is false because no downstream sync is needed.
func (s *DefaultInvoiceService) publishInvoiceChanged(ctx context.Context, invoiceID string) error {
	invoice, err := s.invoiceRepository.GetInvoiceById(invoiceID)
	if err != nil {
		return fmt.Errorf("publishInvoiceChanged: fetching invoice %s: %w", invoiceID, err)
	}

	invoice.UpdatingValues()

	if _, err = s.invoiceRepository.UpdateInvoice(invoice); err != nil {
		return fmt.Errorf("publishInvoiceChanged: updating invoice %s: %w", invoiceID, err)
	}

	if !invoice.MarkImportTransactions {
		return nil
	}

	payload := invoicingdomain.InvoiceChangedPayload{
		InvoiceID:              invoiceID,
		CardDescription:        invoice.Card.Description,
		UserID:                 invoice.Card.UserId,
		ReferenceDate:          invoice.Date,
		Total:                  invoice.Total,
		MarkImportTransactions: invoice.MarkImportTransactions,
	}

	return s.eventPublisher.PublishInvoiceChanged(ctx, payload)
}
