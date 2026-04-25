package domain

import (
	"time"

	pkgentity "github.com/jailtonjunior94/financialcontrol-api/pkg/entity"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/shared"
)

// Entity is the shared base entity from pkg/entity.
type Entity = pkgentity.Entity

// Card is the minimal card projection needed by invoicing domain.
// Only the fields accessed during invoice queries are included.
type Card struct {
	Name        string `db:"Name"`
	Description string `db:"Description"`
	UserId      string `db:"UserId"`
	Entity
}

// Category is the minimal category projection needed by invoice items.
type Category struct {
	Name string `db:"Name"`
	Entity
}

// Invoice is the invoicing module's aggregate root.
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
	invoice.NewEntity()
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
		return
	}
	for _, item := range p.InvoiceItems {
		total += item.InstallmentValue
	}
	p.Total = total
}

// InvoiceItem is a line item within an Invoice.
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
	item := &InvoiceItem{
		InvoiceId:    invoiceId,
		CategoryId:   categoryId,
		Description:  description,
		Tags:         tags,
		TotalAmount:  totalAmount,
		PurchaseDate: shared.NewTime(shared.Time{Date: purchaseDate}).FormatDate(),
	}
	item.NewEntity()
	return item
}

func (p *InvoiceItem) AddInstallment(installment int, installmentValue float64, invoiceControl int64) {
	p.InstallmentValue = installmentValue
	p.InvoiceControl = invoiceControl
	p.Installment = installment
}

// InvoiceCategories is a read-only projection used for category summaries.
type InvoiceCategories struct {
	InvoiceId string    `db:"InvoiceId" json:"invoiceId"`
	Date      time.Time `db:"Date" json:"date"`
	Category  string    `db:"CategoryId" json:"category"`
	Name      string    `db:"Name" json:"name"`
	Total     float64   `db:"Total" json:"total"`
}
