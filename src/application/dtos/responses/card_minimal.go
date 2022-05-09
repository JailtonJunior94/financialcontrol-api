package responses

type CardMinimalResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"active,omitempty"`
}
