package responses

import "time"

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
