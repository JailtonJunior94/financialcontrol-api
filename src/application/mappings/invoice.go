package mappings

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/responses"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/entities"
)

func ToInvoiceEntity(r *requests.InvoiceRequest, date time.Time, total float64) (e *entities.Invoice) {
	entity := new(entities.Invoice)
	entity.NewInvoice(r.CardId, date, total)

	return entity
}

func ToInvoiceItemEntity(r *requests.InvoiceRequest, invoiceId string, installment int) (e *entities.InvoiceItem) {
	entity := new(entities.InvoiceItem)
	entity.NewInvoiceItem(invoiceId, r.CategoryId, r.Description, r.Tags, installment, r.PurchaseDate, r.TotalAmount, r.TotalAmount/float64(r.QuantityInvoice))

	return entity
}

func ToManyInvoiceResponse(entities []entities.Invoice) (r []responses.InvoiceResponse) {
	if len(entities) == 0 {
		return make([]responses.InvoiceResponse, 0)
	}

	for _, e := range entities {
		invoice := responses.InvoiceResponse{
			ID:    e.ID,
			Date:  e.Date,
			Total: e.Total,
		}
		r = append(r, invoice)
	}

	return r
}
