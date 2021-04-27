package entities

import "time"

type InvoiceItem struct {
	InvoiceId        string    `db:"InvoiceId"`
	CategoryId       string    `db:"CategoryId"`
	PurchaseDate     time.Time `db:"PurchaseDate"`
	Description      string    `db:"Description"`
	TotalAmount      float64   `db:"TotalAmount"`
	Installment      int       `db:"Installment"`
	InstallmentValue float64   `db:"InstallmentValue"`
	Tags             string    `db:"Tags"`
	Entity
	Invoice  Invoice
	Category Category
}

func (p *InvoiceItem) NewInvoiceItem(invoiceId, categoryId, description, tags string, installment int, purchaseDate time.Time, totalAmount, installmentValue float64) {
	p.Entity.NewEntity()
	p.InvoiceId = invoiceId
	p.CategoryId = categoryId
	p.Description = description
	p.Tags = tags
	p.Installment = installment
	p.PurchaseDate = purchaseDate
	p.TotalAmount = totalAmount
	p.InstallmentValue = installmentValue
}
