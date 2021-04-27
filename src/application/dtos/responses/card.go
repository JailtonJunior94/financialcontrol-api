package responses

import "time"

type CardResponse struct {
	ID             string            `json:"id,omitempty"`
	Name           string            `json:"name,omitempty"`
	Number         string            `json:"number,omitempty"`
	Description    string            `json:"description,omitempty"`
	ClosingDay     int               `json:"closingDay,omitempty"`
	ExpirationDate time.Time         `json:"expirationDate,omitempty"`
	Active         bool              `json:"active,omitempty"`
	Flag           FlagResponse      `json:"flag,omitempty"`
	Invoices       []InvoiceResponse `json:"invoices,omitempty"`
}
