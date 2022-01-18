package requests

import (
	"log"
	"strconv"
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

func NewInvoiceRequest(purchaseDate, totalAmount, quantityInvoice, cardId, categoryId, tags, description string) *InvoiceRequest {
	purchase, err := time.ParseInLocation("2006-01-02 15:04:05", purchaseDate, time.Local)
	if err != nil {
		log.Fatalln("[ERROR] [Não foi possível converter Data da Compra]")
	}

	total, err := strconv.ParseFloat(totalAmount, 8)
	if err != nil {
		log.Fatalln("[ERROR] [Não foi possível converter Total da Compra]")
	}

	quantity, err := strconv.Atoi(quantityInvoice)
	if err != nil {
		log.Fatalln("[ERROR] [Não foi possível converter Quantidade de Parcelas]")
	}

	return &InvoiceRequest{
		CardId:          cardId,
		CategoryId:      categoryId,
		PurchaseDate:    purchase,
		TotalAmount:     total,
		QuantityInvoice: quantity,
		Description:     description,
		Tags:            tags,
	}
}
