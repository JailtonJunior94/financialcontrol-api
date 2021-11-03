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
	entity := new(entities.InvoiceItem)
	entity.NewInvoiceItem(invoiceId,
		r.CategoryId,
		r.Description,
		r.Tags,
		r.PurchaseDate,
		r.TotalAmount,
	)
	entity.AddInstallment(installment, r.TotalAmount/float64(r.QuantityInvoice), invoiceControl)

	return entity
}

func ToManyInvoiceResponse(entities []entities.Invoice) (r []responses.InvoiceResponse) {
	if len(entities) == 0 {
		return make([]responses.InvoiceResponse, 0)
	}

	for _, e := range entities {
		invoice := responses.InvoiceResponse{
			ID:     e.ID,
			CardId: e.CardId,
			Date:   shared.NewTime(shared.Time{Date: e.Date}).FormatDate(),
			Total:  e.Total,
		}
		r = append(r, invoice)
	}

	return r
}

func ToManyInvoiceItemResponse(entities []entities.InvoiceItem) (r []responses.InvoiceItemResponse) {
	if len(entities) == 0 {
		return make([]responses.InvoiceItemResponse, 0)
	}

	for _, e := range entities {
		invoiceItem := responses.InvoiceItemResponse{
			ID:               e.ID,
			InvoiceControl:   e.InvoiceControl,
			PurchaseDate:     shared.NewTime(shared.Time{Date: e.PurchaseDate}).FormatDate(),
			Description:      e.Description,
			TotalAmount:      e.TotalAmount,
			Installment:      e.Installment,
			InstallmentValue: e.InstallmentValue,
			Tags:             e.Tags,
			Category: responses.CategoryResponse{
				ID:     e.Category.ID,
				Name:   e.Category.Name,
				Active: e.Category.Active,
			},
		}
		r = append(r, invoiceItem)
	}

	return r
}
