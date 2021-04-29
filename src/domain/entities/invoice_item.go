package entities

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/src/shared"
)

type InvoiceItem struct {
	InvoiceId        string    `db:"InvoiceId"`
	CategoryId       string    `db:"CategoryId"`
	PurchaseDate     time.Time `db:"PurchaseDate"`
	Description      string    `db:"Description"`
	TotalAmount      float64   `db:"TotalAmount"`
	Installment      int       `db:"Installment"`
	InstallmentValue float64   `db:"InstallmentValue"`
	Tags             string    `db:"Tags"`
	InvoiceControl   int64     `db:"InvoiceControl"`
	Entity
	Invoice  Invoice
	Category Category
}

func (p *InvoiceItem) NewInvoiceItem(invoiceId, categoryId, description, tags string, purchaseDate time.Time, totalAmount float64) {
	p.Entity.NewEntity()
	p.InvoiceId = invoiceId
	p.CategoryId = categoryId
	p.Description = description
	p.Tags = tags
	p.TotalAmount = totalAmount
	p.PurchaseDate = shared.NewTime(shared.Time{Date: purchaseDate}).FormatDate()
}

func (p *InvoiceItem) AddInstallment(installment int, installmentValue float64, invoiceControl int64) {
	p.InstallmentValue = installmentValue
	p.InvoiceControl = invoiceControl
	p.Installment = installment
}
