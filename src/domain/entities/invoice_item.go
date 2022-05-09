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

func NewInvoiceItem(invoiceId, categoryId, description, tags string, purchaseDate time.Time, totalAmount float64) *InvoiceItem {
	invoiceItem := &InvoiceItem{
		InvoiceId:    invoiceId,
		CategoryId:   categoryId,
		Description:  description,
		Tags:         tags,
		TotalAmount:  totalAmount,
		PurchaseDate: shared.NewTime(shared.Time{Date: purchaseDate}).FormatDate(),
	}

	invoiceItem.Entity.NewEntity()
	return invoiceItem
}

func (p *InvoiceItem) AddInstallment(installment int, installmentValue float64, invoiceControl int64) {
	p.InstallmentValue = installmentValue
	p.InvoiceControl = invoiceControl
	p.Installment = installment
}

func (p *InvoiceItem) AddCategory(id, name string, active bool) {
	p.Category = Category{
		Name: name,
		Entity: Entity{
			ID:     id,
			Active: active,
		},
	}
}
