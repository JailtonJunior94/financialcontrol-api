package application

import "time"

type BillResponse struct {
	ID           string             `json:"id"`
	Date         time.Time          `json:"date"`
	Total        float64            `json:"total"`
	SixtyPercent float64            `json:"sixtyPercent"`
	FortyPercent float64            `json:"fortyPercent"`
	Active       bool               `json:"active"`
	BillItems    []BillItemResponse `json:"billItems,omitempty"`
}

type BillItemResponse struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Value  float64 `json:"value"`
	Active bool    `json:"active"`
}
