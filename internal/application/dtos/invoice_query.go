package dtos

import "time"

type InvoiceQuery struct {
	Date        time.Time `db:"Date"`
	Description string    `db:"Description"`
	Total       float64   `db:"Total"`
}

type (
	InvoiceRead struct {
		ID    string
		Date  time.Time
		Total float64
		Items []InvoiceItemRead
	}

	InvoiceItemRead struct {
		ID               string
		PurchaseDate     time.Time
		Description      string
		TotalAmount      float64
		Installment      int
		InstallmentValue float64
		Tags             string
		Category         CategoryRead
	}

	CategoryRead struct {
		ID   string
		Name string
	}
)

type (
	BillQuery struct {
		ID    string
		Date  time.Time
		Items []BillItemQuery
	}

	BillItemQuery struct {
		ID          string
		Description string
		Total       float64
	}
)
