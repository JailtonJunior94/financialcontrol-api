package mappings

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/entities"
)

func ToInvoiceEntity(r *requests.InvoiceRequest, date time.Time, total float64) (e *entities.Invoice) {
	entity := new(entities.Invoice)
	entity.NewInvoice(r.CardId, date, total)

	return entity
}

func ToInvoiceItemEntity(r *requests.InvoiceRequest, invoiceId, installment string) (e *entities.InvoiceItem) {
	entity := new(entities.InvoiceItem)
	entity.NewInvoiceItem(invoiceId, r.CategoryId, r.Description, r.Tags, installment, r.PurchaseDate, r.TotalAmount, r.TotalAmount/float64(r.QuantityInvoice))

	return entity
}
