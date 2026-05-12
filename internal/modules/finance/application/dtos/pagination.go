package dtos

// PaginatedResponse is the generic envelope for paginated list endpoints.
// Tags use snake_case per decisão I1.b (finance module convention).
type PaginatedResponse[T any] struct {
	Items    []T   `json:"items"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}
