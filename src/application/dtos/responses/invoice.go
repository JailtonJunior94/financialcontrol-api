package responses

import "time"

type InvoiceResponse struct {
	ID    string    `json:"id"`
	Date  time.Time `json:"date"`
	Total float64   `json:"total"`
}
