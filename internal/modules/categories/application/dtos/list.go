package dtos

const (
	ScopeAll   = "all"
	ScopeRoots = "roots"
	ScopeSubs  = "subs"
)

// Pagination carries the normalized pagination input from the HTTP layer.
type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

// ListCategoriesQuery represents the parsed query parameters of GET /categories.
type ListCategoriesQuery struct {
	Page     int    `query:"page"     validate:"omitempty,gte=1"`
	PageSize int    `query:"pageSize" validate:"omitempty,gte=1,lte=100"`
	Name     string `query:"name"     validate:"omitempty,max=100"`
	Scope    string `query:"scope"    validate:"omitempty,oneof=all roots subs"`
	ParentID string `query:"parentId" validate:"omitempty,uuid"`
}

// PaginatedResponse is a generic envelope for paginated list endpoints.
type PaginatedResponse[T any] struct {
	Items    []T   `json:"items"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}
