package vos

import (
	"errors"

	"github.com/google/uuid"
)

// CategoryID is a UUID-based value object identifying a category (cross-module reference).
type CategoryID string

var ErrInvalidCategoryID = errors.New("category ID inválido")

// NewCategoryID generates a new random CategoryID.
func NewCategoryID() CategoryID {
	return CategoryID(uuid.NewString())
}

// ParseCategoryID parses and validates a string as a CategoryID.
func ParseCategoryID(s string) (CategoryID, error) {
	if _, err := uuid.Parse(s); err != nil {
		return "", ErrInvalidCategoryID
	}
	return CategoryID(s), nil
}

func (c CategoryID) String() string { return string(c) }
