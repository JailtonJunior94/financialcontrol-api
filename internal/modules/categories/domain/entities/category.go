package entities

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

type Clock interface {
	Now() time.Time
}

type Category struct {
	id        vos.CategoryID
	userID    identityvo.UserID
	parentID  *vos.CategoryID
	name      vos.CategoryName
	color     vos.CategoryColor
	icon      vos.CategoryIcon
	createdAt time.Time
	updatedAt time.Time
	deletedAt *time.Time
}

func NewCategory(
	userID identityvo.UserID,
	parentID *vos.CategoryID,
	name string,
	color string,
	icon string,
	clock Clock,
) (*Category, error) {
	categoryName, categoryColor, categoryIcon, err := buildVOs(name, color, icon)
	if err != nil {
		return nil, err
	}

	now := clock.Now().UTC()
	return &Category{
		id:        vos.NewCategoryID(),
		userID:    userID,
		parentID:  parentID,
		name:      categoryName,
		color:     categoryColor,
		icon:      categoryIcon,
		createdAt: now,
		updatedAt: now,
	}, nil
}

func RehydrateCategory(
	id vos.CategoryID,
	userID identityvo.UserID,
	parentID *vos.CategoryID,
	name vos.CategoryName,
	color vos.CategoryColor,
	icon vos.CategoryIcon,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) *Category {
	return &Category{
		id:        id,
		userID:    userID,
		parentID:  parentID,
		name:      name,
		color:     color,
		icon:      icon,
		createdAt: createdAt,
		updatedAt: updatedAt,
		deletedAt: deletedAt,
	}
}

func (c *Category) Rename(name vos.CategoryName, clock Clock) error {
	if name.String() == "" {
		return domain.ErrInvalidCategoryName
	}
	c.name = name
	c.updatedAt = clock.Now().UTC()
	return nil
}

func (c *Category) ChangeAppearance(color vos.CategoryColor, icon vos.CategoryIcon, clock Clock) {
	c.color = color
	c.icon = icon
	c.updatedAt = clock.Now().UTC()
}

func (c *Category) Reparent(newParent *vos.CategoryID, clock Clock) error {
	if newParent == nil {
		return domain.ErrParentNotFound
	}
	c.parentID = newParent
	c.updatedAt = clock.Now().UTC()
	return nil
}

func (c *Category) MarkDeleted(at time.Time) {
	if c.deletedAt != nil {
		return
	}
	deleted := at.UTC()
	c.deletedAt = &deleted
	c.updatedAt = deleted
}

func (c *Category) IsActive() bool      { return c.deletedAt == nil }
func (c *Category) IsRoot() bool        { return c.parentID == nil }
func (c *Category) IsSubcategory() bool { return c.parentID != nil }
func (c *Category) ID() vos.CategoryID  { return c.id }
func (c *Category) UserID() identityvo.UserID {
	return c.userID
}
func (c *Category) ParentID() *vos.CategoryID { return c.parentID }
func (c *Category) Name() vos.CategoryName    { return c.name }
func (c *Category) Color() vos.CategoryColor  { return c.color }
func (c *Category) Icon() vos.CategoryIcon    { return c.icon }
func (c *Category) CreatedAt() time.Time      { return c.createdAt }
func (c *Category) UpdatedAt() time.Time      { return c.updatedAt }
func (c *Category) DeletedAt() *time.Time     { return c.deletedAt }

func buildVOs(name, color, icon string) (vos.CategoryName, vos.CategoryColor, vos.CategoryIcon, error) {
	categoryName, err := vos.NewCategoryName(name)
	if err != nil {
		return "", "", "", err
	}

	categoryColor, err := vos.NewCategoryColor(color)
	if err != nil {
		return "", "", "", err
	}

	categoryIcon, err := vos.NewCategoryIcon(icon)
	if err != nil {
		return "", "", "", err
	}

	return categoryName, categoryColor, categoryIcon, nil
}
