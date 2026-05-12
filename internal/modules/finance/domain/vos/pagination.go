package vos

import domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"

const (
	defaultFinancePage     = 1
	defaultFinancePageSize = 20
	minFinancePageSize     = 1
	maxFinancePageSize     = 100
)

// Pagination holds validated page/page_size for finance list endpoints (RF-19).
type Pagination struct {
	page     int
	pageSize int
}

// NewPagination normalises raw pagination input applying defaults and limits from RF-19.
// page < 1 → 1 (default); pageSize < 1 → 20 (default); pageSize > 100 → ErrInvalidPaginationLimits.
func NewPagination(rawPage, rawPageSize int) (Pagination, error) {
	p := rawPage
	if p < 1 {
		p = defaultFinancePage
	}
	ps := rawPageSize
	if ps < minFinancePageSize {
		ps = defaultFinancePageSize
	}
	if ps > maxFinancePageSize {
		return Pagination{}, domain.ErrInvalidPaginationLimits
	}
	return Pagination{page: p, pageSize: ps}, nil
}

func (pg Pagination) Page() int     { return pg.page }
func (pg Pagination) PageSize() int { return pg.pageSize }

// Offset returns the SQL offset equivalent to the current page.
func (pg Pagination) Offset() int { return (pg.page - 1) * pg.pageSize }
