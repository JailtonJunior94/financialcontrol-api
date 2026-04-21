package application

import "time"

type FlagResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

type CardResponse struct {
	ID             string       `json:"id"`
	Name           string       `json:"name"`
	Number         string       `json:"number,omitempty"`
	Description    string       `json:"description,omitempty"`
	ClosingDay     int          `json:"closingDay,omitempty"`
	BestDayToBuy   int          `json:"bestDayToBuy,omitempty"`
	ExpirationDate time.Time    `json:"expirationDate,omitempty"`
	Active         bool         `json:"active"`
	Flag           FlagResponse `json:"flag,omitempty"`
}
