package responses

import "time"

type CardResponse struct {
	ID             string       `json:"id"`
	Name           string       `json:"name"`
	Number         string       `json:"number"`
	Description    string       `json:"description"`
	ClosingDay     int          `json:"closingDay"`
	ExpirationDate time.Time    `json:"expirationDate"`
	Active         bool         `json:"active"`
	Flag           FlagResponse `json:"flag"`
}
