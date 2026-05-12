package entities

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/vos"
)

type Clock interface {
	Now() time.Time
}

type Category struct {
	id        vos.CategoryID
	name      vos.CategoryName
	sequence  int
	createdAt time.Time
	updatedAt time.Time
	active    bool
}

func NewCategory(name string, sequence int, clock Clock) (*Category, error) {
	categoryName, err := vos.NewCategoryName(name)
	if err != nil {
		return nil, err
	}

	now := clock.Now().UTC()
	return &Category{
		id:        vos.NewCategoryID(),
		name:      categoryName,
		sequence:  sequence,
		createdAt: now,
		updatedAt: now,
		active:    true,
	}, nil
}

func RehydrateCategory(
	id vos.CategoryID,
	name vos.CategoryName,
	sequence int,
	createdAt time.Time,
	updatedAt time.Time,
	active bool,
) *Category {
	return &Category{
		id:        id,
		name:      name,
		sequence:  sequence,
		createdAt: createdAt,
		updatedAt: updatedAt,
		active:    active,
	}
}

func (c *Category) Update(name vos.CategoryName, sequence int, clock Clock) error {
	if name.String() == "" {
		return domain.ErrInvalidCategoryName
	}
	c.name = name
	c.sequence = sequence
	c.updatedAt = clock.Now().UTC()
	return nil
}

func (c *Category) Deactivate(at time.Time) {
	if !c.active {
		return
	}
	c.active = false
	c.updatedAt = at.UTC()
}

func (c *Category) IsActive() bool         { return c.active }
func (c *Category) ID() vos.CategoryID     { return c.id }
func (c *Category) Name() vos.CategoryName { return c.name }
func (c *Category) Sequence() int          { return c.sequence }
func (c *Category) CreatedAt() time.Time   { return c.createdAt }
func (c *Category) UpdatedAt() time.Time   { return c.updatedAt }
