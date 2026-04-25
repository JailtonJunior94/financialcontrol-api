package invoicingapp

import "time"

type InvoiceResponse struct {
	ID           string                 `json:"id"`
	CardId       string                 `json:"cardId,omitempty"`
	Date         time.Time              `json:"date"`
	Total        float64                `json:"total"`
	Card         *CardMinimalResponse   `json:"card,omitempty"`
	InvoiceItems []*InvoiceItemResponse `json:"invoiceItems,omitempty"`
}

type InvoiceItemResponse struct {
	ID               string           `json:"id"`
	InvoiceControl   int64            `json:"invoiceControl"`
	PurchaseDate     time.Time        `json:"purchaseDate"`
	Description      string           `json:"description"`
	TotalAmount      float64          `json:"totalAmount"`
	Installment      int              `json:"installment"`
	InstallmentValue float64          `json:"installmentValue"`
	Tags             string           `json:"tags"`
	Category         CategoryResponse `json:"category"`
}

type CardMinimalResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"active,omitempty"`
}

type CategoryResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Sequence int    `json:"sequence,omitempty"`
	Active   bool   `json:"active"`
}
