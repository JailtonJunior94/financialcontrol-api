package entities

import "time"

type Invoice struct {
	CardId string    `db:"CardId"`
	Date   time.Time `db:"Date"`
	Total  float64   `db:"Total"`
	Entity
	Card         Card
	InvoiceItems []InvoiceItem
}

func (p *Invoice) NewInvoice(cardId string, date time.Time, total float64) {
	p.Entity.NewEntity()
	p.CardId = cardId
	p.Date = date
	p.Total = total
}

func (p *Invoice) AddInvoiceItems(invoiceItems []InvoiceItem) {
	p.InvoiceItems = invoiceItems
}

func (p *Invoice) UpdatingValues() {
	p.sumTotal()
	p.ChangeUpdatedAt()
}

func (b *Invoice) sumTotal() {
	var total float64

	if len(b.InvoiceItems) == 0 {
		b.Total = total
	}

	for _, bill := range b.InvoiceItems {
		total += bill.InstallmentValue
	}

	b.Total = total
}
