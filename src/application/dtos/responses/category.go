package responses

type CategoryResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Sequence int    `json:"sequence,omitempty"`
	Active   bool   `json:"active"`
}
