package domain_test

import (
	"errors"
	"testing"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain"
	"github.com/stretchr/testify/assert"
)

func TestSentinelErrors(t *testing.T) {
	t.Parallel()
	cases := []error{
		domain.ErrCategoryNotFound,
		domain.ErrInvalidCategoryID,
		domain.ErrInvalidCategoryName,
		domain.ErrInvalidCategoryColor,
		domain.ErrInvalidCategoryIcon,
		domain.ErrParentNotFound,
		domain.ErrParentInactive,
		domain.ErrSubcategoryDepthExceeded,
		domain.ErrCategoryNameAlreadyExists,
	}
	for _, e := range cases {
		assert.NotNil(t, e)
		assert.NotEmpty(t, e.Error())
		assert.True(t, errors.Is(e, e))
	}
}
