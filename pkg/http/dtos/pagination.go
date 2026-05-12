// Package dtos provides shared HTTP response types for the finance module.
package dtos

// PaginatedResponse is the generic paginated envelope used by finance list endpoints.
// JSON keys are in snake_case as per decisão I2.a (wire format).
type PaginatedResponse[T any] struct {
	Items    []T   `json:"items"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}
