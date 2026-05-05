package vos

const (
	defaultPage     = 1
	defaultPageSize = 50
	maxPageSize     = 200
)

// Pagination represents optional pagination input for list operations.
// The zero value disables pagination to preserve legacy list behavior.
type Pagination struct {
	Page int
	Size int
}

// NewPagination normalizes pagination input while preserving the zero value as
// "pagination not requested".
func NewPagination(page, size int) Pagination {
	if page == 0 && size == 0 {
		return Pagination{}
	}
	if page <= 0 {
		page = defaultPage
	}
	if size <= 0 {
		size = defaultPageSize
	}
	if size > maxPageSize {
		size = maxPageSize
	}
	return Pagination{Page: page, Size: size}
}

func (p Pagination) Enabled() bool {
	return p.Size > 0
}

func (p Pagination) Offset() int {
	if !p.Enabled() || p.Page <= 0 {
		return 0
	}
	return (p.Page - 1) * p.Size
}
