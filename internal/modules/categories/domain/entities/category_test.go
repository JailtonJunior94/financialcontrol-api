package entities_test

import (
	"errors"
	"testing"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/vos"
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

func TestNewCategory(t *testing.T) {
	clock := newClock()

	c, err := entities.NewCategory("Food", 7, clock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !c.IsActive() {
		t.Fatalf("expected active category")
	}
	if c.Name().String() != "Food" {
		t.Fatalf("unexpected name: %s", c.Name().String())
	}
	if c.Sequence() != 7 {
		t.Fatalf("unexpected sequence: %d", c.Sequence())
	}
	if c.CreatedAt() != clock.t || c.UpdatedAt() != clock.t {
		t.Fatalf("timestamps mismatch")
	}
	if c.ID().String() == "" {
		t.Fatalf("missing id")
	}
}

func TestNewCategory_InvalidName(t *testing.T) {
	_, err := entities.NewCategory("", 1, newClock())
	if !errors.Is(err, domain.ErrInvalidCategoryName) {
		t.Fatalf("want %v, got %v", domain.ErrInvalidCategoryName, err)
	}
}

func TestCategoryUpdate(t *testing.T) {
	clock := &advancingClock{t: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), step: time.Hour}
	c, err := entities.NewCategory("Food", 1, clock)
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	original := c.UpdatedAt()

	newName, err := vos.NewCategoryName("Groceries")
	if err != nil {
		t.Fatalf("vo: %v", err)
	}
	if err := c.Update(newName, 9, clock); err != nil {
		t.Fatalf("update: %v", err)
	}
	if c.Name() != newName {
		t.Fatalf("name not updated")
	}
	if c.Sequence() != 9 {
		t.Fatalf("sequence not updated")
	}
	if !c.UpdatedAt().After(original) {
		t.Fatalf("updatedAt not advanced")
	}

	if err := c.Update(vos.CategoryName(""), 10, clock); !errors.Is(err, domain.ErrInvalidCategoryName) {
		t.Fatalf("expected invalid name, got %v", err)
	}
}

func TestCategoryDeactivateIdempotent(t *testing.T) {
	clock := newClock()
	c, err := entities.NewCategory("Food", 1, clock)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	first := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	c.Deactivate(first)
	if c.IsActive() {
		t.Fatalf("expected inactive")
	}
	if !c.UpdatedAt().Equal(first) {
		t.Fatalf("updatedAt not set on deactivate")
	}

	second := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	c.Deactivate(second)
	if !c.UpdatedAt().Equal(first) {
		t.Fatalf("deactivate must be idempotent, got %v", c.UpdatedAt())
	}
}

func TestRehydrateCategory(t *testing.T) {
	id := vos.NewCategoryID()
	name, _ := vos.NewCategoryName("Food")
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	updated := created.Add(time.Hour)

	c := entities.RehydrateCategory(id, name, 3, created, updated, false)
	if c.ID() != id {
		t.Fatalf("identity mismatch")
	}
	if c.IsActive() {
		t.Fatalf("expected inactive")
	}
	if c.Sequence() != 3 {
		t.Fatalf("sequence mismatch")
	}
	if !c.CreatedAt().Equal(created) || !c.UpdatedAt().Equal(updated) {
		t.Fatalf("timestamps mismatch")
	}
}
