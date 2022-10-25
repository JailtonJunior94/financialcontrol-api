package dtos

import "time"

type InvoiceQuery struct {
	Date        time.Time `db:"Date"`
	Description string    `db:"Description"`
	Total       float64   `db:"Total"`
}
