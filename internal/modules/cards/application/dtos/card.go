package dtos

import "time"

type CardRequest struct {
	FlagID         string    `json:"flagId"`
	Name           string    `json:"name"`
	Number         string    `json:"number"`
	Description    string    `json:"description"`
	ClosingDay     int       `json:"closingDay"`
	ExpirationDate time.Time `json:"expirationDate"`
}

type CardResponse struct {
	ID             string       `json:"id"`
	Name           string       `json:"name"`
	Number         string       `json:"number,omitempty"`
	Description    string       `json:"description,omitempty"`
	ClosingDay     int          `json:"closingDay,omitempty"`
	BestDayToBuy   int          `json:"bestDayToBuy,omitempty"`
	ExpirationDate time.Time    `json:"expirationDate"`
	Active         bool         `json:"active"`
	Flag           FlagResponse `json:"flag"`
}

type FlagResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
}
