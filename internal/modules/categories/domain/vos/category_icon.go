package vos

import (
	"regexp"
	"strings"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain"
)

type CategoryIcon string

var iconRE = regexp.MustCompile(`^[a-z0-9-]{1,64}$`)

func NewCategoryIcon(s string) (CategoryIcon, error) {
	v := strings.TrimSpace(s)
	if !iconRE.MatchString(v) {
		return "", domain.ErrInvalidCategoryIcon
	}
	return CategoryIcon(v), nil
}

func (i CategoryIcon) String() string { return string(i) }
