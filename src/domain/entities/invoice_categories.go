package entities

import "time"

type InvoiceCategories struct {
	InvoiceId string    `db:"InvoiceId" json:"invoiceId"`
	Date      time.Time `db:"Date" json:"date"`
	Category  string    `db:"CategoryId" json:"category"`
	Name      string    `db:"Name" json:"name"`
	Total     float64   `db:"Total" json:"total"`
}
