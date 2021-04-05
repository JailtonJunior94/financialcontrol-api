package responses

type BillItemResponse struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Value  float64 `json:"value"`
	Active bool    `json:"active"`
}
