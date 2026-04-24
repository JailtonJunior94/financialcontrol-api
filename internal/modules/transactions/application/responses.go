package application

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/platform/web"
)

type HttpResponse = web.HttpResponse

type TransactionResponse struct {
	ID      string                    `json:"id"`
	Date    time.Time                 `json:"date"`
	Total   float64                   `json:"total"`
	Income  float64                   `json:"income"`
	Outcome float64                   `json:"outcome"`
	Active  bool                      `json:"active"`
	Items   []TransactionItemResponse `json:"items,omitempty"`
}

type TransactionItemResponse struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Value  float64 `json:"value"`
	Type   string  `json:"type"`
	IsPaid bool    `json:"isPaid"`
	Active bool    `json:"active"`
}
