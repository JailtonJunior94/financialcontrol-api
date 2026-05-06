package vos

import (
	"strings"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain"
)

type CategoryName string

func NewCategoryName(s string) (CategoryName, error) {
	trimmed := strings.TrimSpace(s)
	if len(trimmed) < 1 || len(trimmed) > 100 {
		return "", domain.ErrInvalidCategoryName
	}
	return CategoryName(trimmed), nil
}

func (n CategoryName) String() string { return string(n) }
