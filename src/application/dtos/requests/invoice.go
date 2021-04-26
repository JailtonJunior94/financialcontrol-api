package requests

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/src/domain/customErrors"
)

type InvoiceRequest struct {
	CardId          string    `json:"cardId"`
	CategoryId      string    `json:"categoryId"`
	PurchaseDate    time.Time `json:"purchaseDate"`
	TotalAmount     float64   `json:"totalAmount"`
	QuantityInvoice int       `json:"quantityInvoice"`
	Description     string    `json:"description"`
	Tags            string    `json:"tags"`
}

func (c *InvoiceRequest) IsValid() error {
	if c.CardId == "" {
		return customErrors.NameIsRequired
	}

	return nil
}
