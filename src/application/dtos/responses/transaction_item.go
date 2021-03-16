package responses

type TransactionItemResponse struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Value  float64 `json:"value"`
	Type   string  `json:"type"`
	Active bool    `json:"active"`
}
