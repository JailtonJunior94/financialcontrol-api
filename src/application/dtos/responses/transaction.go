package responses

import "time"

type TransactionResponse struct {
	ID      string                    `json:"id"`
	Date    time.Time                 `json:"date"`
	Total   float64                   `json:"total"`
	Income  float64                   `json:"income"`
	Outcome float64                   `json:"outcome"`
	Active  bool                      `json:"active"`
	Items   []TransactionItemResponse `json:"items"`
}
