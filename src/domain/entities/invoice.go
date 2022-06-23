package entities

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/src/shared"
)

type Invoice struct {
	CardId                 string    `db:"CardId"`
	Date                   time.Time `db:"Date"`
	Total                  float64   `db:"Total"`
	MarkImportTransactions bool      `db:"MarkImportTransactions"`
	Entity
	Card         Card
	InvoiceItems []InvoiceItem
}

func NewInvoice(cardId string, date time.Time, total float64) *Invoice {
	invoice := &Invoice{
		CardId: cardId,
		Date:   shared.NewTime(shared.Time{Date: date}).FormatDate(),
		Total:  total,
	}
	invoice.Entity.NewEntity()

	return invoice
}

func (p *Invoice) AddInvoiceItems(invoiceItems []InvoiceItem) {
	p.InvoiceItems = invoiceItems
}

func (p *Invoice) UpdatingValues() {
	p.sumTotal()
	p.ChangeUpdatedAt()
}

func (p *Invoice) sumTotal() {
	var total float64

	if len(p.InvoiceItems) == 0 {
		p.Total = total
	}

	for _, invoice := range p.InvoiceItems {
		total += invoice.InstallmentValue
	}

	p.Total = total
}
