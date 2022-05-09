package responses

import "time"

type InvoiceResponse struct {
	ID           string                 `json:"id"`
	CardId       string                 `json:"cardId,omitempty"`
	Date         time.Time              `json:"date"`
	Total        float64                `json:"total"`
	Card         *CardMinimalResponse   `json:"card,omitempty"`
	InvoiceItems []*InvoiceItemResponse `json:"invoiceItems,omitempty"`
}
