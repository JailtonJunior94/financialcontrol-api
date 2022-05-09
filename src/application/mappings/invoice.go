package mappings

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/responses"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/src/shared"
)

func ToInvoiceEntity(r *requests.InvoiceRequest, date time.Time, total float64) (e *entities.Invoice) {
	invoice := entities.NewInvoice(r.CardId, date, total)
	return invoice
}

func ToInvoiceItemEntity(r *requests.InvoiceRequest, invoiceId string, installment int, invoiceControl int64) (e *entities.InvoiceItem) {
	entity := entities.NewInvoiceItem(invoiceId,
		r.CategoryId,
		r.Description,
		r.Tags,
		r.PurchaseDate,
		r.TotalAmount,
	)

	entity.AddInstallment(installment, r.TotalAmount/float64(r.QuantityInvoice), invoiceControl)
	return entity
}

func ToInvoiceResponse(entity *entities.Invoice) *responses.InvoiceResponse {
	if entity == nil {
		return nil
	}

	invoice := &responses.InvoiceResponse{
		ID:           entity.ID,
		CardId:       entity.CardId,
		Date:         shared.NewTime(shared.Time{Date: entity.Date}).FormatDate(),
		Total:        entity.Total,
		Card:         ToCardMinimalResponse(&entity.Card),
		InvoiceItems: ToManyInvoiceItemResponse(entity.InvoiceItems),
	}

	return invoice
}

func ToManyInvoiceResponse(entities []entities.Invoice) []*responses.InvoiceResponse {
	invoices := make([]*responses.InvoiceResponse, len(entities), len(entities))
	if len(invoices) == 0 {
		return invoices
	}

	for index, entity := range entities {
		invoice := &responses.InvoiceResponse{
			ID:     entity.ID,
			CardId: entity.CardId,
			Date:   shared.NewTime(shared.Time{Date: entity.Date}).FormatDate(),
			Total:  entity.Total,
		}

		invoices[index] = invoice
	}

	return invoices
}

func ToManyInvoiceItemResponse(entities []entities.InvoiceItem) []*responses.InvoiceItemResponse {
	items := make([]*responses.InvoiceItemResponse, len(entities), len(entities))
	if len(items) == 0 {
		return items
	}

	for index, entity := range entities {
		item := &responses.InvoiceItemResponse{
			ID:               entity.ID,
			InvoiceControl:   entity.InvoiceControl,
			PurchaseDate:     shared.NewTime(shared.Time{Date: entity.PurchaseDate}).FormatDate(),
			Description:      entity.Description,
			TotalAmount:      entity.TotalAmount,
			Installment:      entity.Installment,
			InstallmentValue: entity.InstallmentValue,
			Tags:             entity.Tags,
			Category: responses.CategoryResponse{
				ID:     entity.Category.ID,
				Name:   entity.Category.Name,
				Active: entity.Category.Active,
			},
		}
		items[index] = item
	}

	return items
}
