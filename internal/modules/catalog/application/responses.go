package application

type FlagResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

type CategoryResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Sequence int    `json:"sequence,omitempty"`
	Active   bool   `json:"active"`
}
