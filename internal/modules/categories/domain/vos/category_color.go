package vos

import (
	"strings"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain"
)

type CategoryColor string

var validColors = map[string]struct{}{
	"red":    {},
	"orange": {},
	"yellow": {},
	"green":  {},
	"teal":   {},
	"blue":   {},
	"indigo": {},
	"purple": {},
	"pink":   {},
	"brown":  {},
	"gray":   {},
}

func NewCategoryColor(s string) (CategoryColor, error) {
	v := strings.ToLower(strings.TrimSpace(s))
	if _, ok := validColors[v]; !ok {
		return "", domain.ErrInvalidCategoryColor
	}
	return CategoryColor(v), nil
}

func (c CategoryColor) String() string { return string(c) }
