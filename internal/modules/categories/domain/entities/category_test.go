package entities_test

import (
	"errors"
	"testing"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

type fixedClock struct{ t time.Time }

func (f *fixedClock) Now() time.Time { return f.t }

type advancingClock struct {
	t    time.Time
	step time.Duration
}

func (a *advancingClock) Now() time.Time {
	a.t = a.t.Add(a.step)
	return a.t
}

func newClock() *fixedClock {
	return &fixedClock{t: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)}
}

func newUserID(t *testing.T) identityvo.UserID {
	t.Helper()
	return identityvo.NewUserID()
}

func newParentID() *vos.CategoryID {
	id := vos.NewCategoryID()
	return &id
}

func TestNewCategory(t *testing.T) {
	type args struct {
		parent *vos.CategoryID
		name   string
		color  string
		icon   string
	}
	tests := []struct {
		name      string
		args      args
		wantErr   error
		assertion func(t *testing.T, c *entities.Category)
	}{
		{
			name: "root category",
			args: args{parent: nil, name: "Food", color: "red", icon: "fork"},
			assertion: func(t *testing.T, c *entities.Category) {
				if !c.IsRoot() || c.IsSubcategory() {
					t.Fatalf("expected root, got sub")
				}
				if !c.IsActive() {
					t.Fatalf("expected active")
				}
			},
		},
		{
			name: "subcategory",
			args: args{parent: newParentID(), name: "Lunch", color: "blue", icon: "fork"},
			assertion: func(t *testing.T, c *entities.Category) {
				if c.IsRoot() || !c.IsSubcategory() {
					t.Fatalf("expected sub")
				}
			},
		},
		{name: "invalid name", args: args{name: "", color: "red", icon: "fork"}, wantErr: domain.ErrInvalidCategoryName},
		{name: "invalid color", args: args{name: "Food", color: "neon", icon: "fork"}, wantErr: domain.ErrInvalidCategoryColor},
		{name: "invalid icon", args: args{name: "Food", color: "red", icon: "BAD ICON"}, wantErr: domain.ErrInvalidCategoryIcon},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			clock := newClock()
			userID := newUserID(t)
			c, err := entities.NewCategory(userID, tc.args.parent, tc.args.name, tc.args.color, tc.args.icon, clock)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("want %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if c.UserID() != userID {
				t.Fatalf("user mismatch")
			}
			if c.CreatedAt() != clock.t || c.UpdatedAt() != clock.t {
				t.Fatalf("timestamps mismatch")
			}
			if c.DeletedAt() != nil {
				t.Fatalf("expected no deletedAt")
			}
			if c.ID().String() == "" {
				t.Fatalf("missing id")
			}
			tc.assertion(t, c)
		})
	}
}

func TestCategoryRename(t *testing.T) {
	clock := &advancingClock{t: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), step: time.Hour}
	c, err := entities.NewCategory(newUserID(t), nil, "Food", "red", "fork", clock)
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	original := c.UpdatedAt()

	newName, err := vos.NewCategoryName("Groceries")
	if err != nil {
		t.Fatalf("vo: %v", err)
	}
	if err := c.Rename(newName, clock); err != nil {
		t.Fatalf("rename: %v", err)
	}
	if c.Name() != newName {
		t.Fatalf("name not updated")
	}
	if !c.UpdatedAt().After(original) {
		t.Fatalf("updatedAt not advanced")
	}

	if err := c.Rename(vos.CategoryName(""), clock); !errors.Is(err, domain.ErrInvalidCategoryName) {
		t.Fatalf("expected invalid name, got %v", err)
	}
}

func TestCategoryChangeAppearance(t *testing.T) {
	clock := &advancingClock{t: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), step: time.Minute}
	c, err := entities.NewCategory(newUserID(t), nil, "Food", "red", "fork", clock)
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	original := c.UpdatedAt()

	color, _ := vos.NewCategoryColor("blue")
	icon, _ := vos.NewCategoryIcon("spoon")
	c.ChangeAppearance(color, icon, clock)

	if c.Color() != color || c.Icon() != icon {
		t.Fatalf("appearance not updated")
	}
	if !c.UpdatedAt().After(original) {
		t.Fatalf("updatedAt not advanced")
	}
}

func TestCategoryReparent(t *testing.T) {
	clock := &advancingClock{t: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), step: time.Minute}
	parent := newParentID()
	c, err := entities.NewCategory(newUserID(t), parent, "Lunch", "blue", "fork", clock)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	if err := c.Reparent(nil, clock); !errors.Is(err, domain.ErrParentNotFound) {
		t.Fatalf("expected ErrParentNotFound, got %v", err)
	}

	newParent := newParentID()
	original := c.UpdatedAt()
	if err := c.Reparent(newParent, clock); err != nil {
		t.Fatalf("reparent: %v", err)
	}
	if c.ParentID() == nil || *c.ParentID() != *newParent {
		t.Fatalf("parent not updated")
	}
	if !c.UpdatedAt().After(original) {
		t.Fatalf("updatedAt not advanced")
	}
}

func TestCategoryMarkDeletedIdempotent(t *testing.T) {
	clock := newClock()
	c, err := entities.NewCategory(newUserID(t), nil, "Food", "red", "fork", clock)
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	if !c.IsActive() {
		t.Fatalf("expected active")
	}

	first := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	c.MarkDeleted(first)
	if c.IsActive() {
		t.Fatalf("expected inactive")
	}
	if c.DeletedAt() == nil || !c.DeletedAt().Equal(first) {
		t.Fatalf("deletedAt not set")
	}

	second := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	c.MarkDeleted(second)
	if !c.DeletedAt().Equal(first) {
		t.Fatalf("MarkDeleted should be idempotent, got %v", c.DeletedAt())
	}
}

func TestRehydrateCategory(t *testing.T) {
	id := vos.NewCategoryID()
	userID := newUserID(t)
	parent := newParentID()
	name, _ := vos.NewCategoryName("Food")
	color, _ := vos.NewCategoryColor("red")
	icon, _ := vos.NewCategoryIcon("fork")
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	updated := created.Add(time.Hour)
	deleted := updated.Add(time.Hour)

	c := entities.RehydrateCategory(id, userID, parent, name, color, icon, created, updated, &deleted)
	if c.ID() != id || c.UserID() != userID {
		t.Fatalf("identity mismatch")
	}
	if c.IsActive() {
		t.Fatalf("expected inactive after rehydrate with deletedAt")
	}
	if !c.CreatedAt().Equal(created) || !c.UpdatedAt().Equal(updated) {
		t.Fatalf("timestamps mismatch")
	}
	if c.ParentID() == nil || *c.ParentID() != *parent {
		t.Fatalf("parent mismatch")
	}
}
