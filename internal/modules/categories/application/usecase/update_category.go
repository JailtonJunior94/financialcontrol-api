package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/services"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

type updateCategory struct {
	repo       interfaces.CategoryRepository
	uniqueness *services.CategoryUniquenessService
	clock      Clock
}

func NewUpdateCategory(repo interfaces.CategoryRepository, uniqueness *services.CategoryUniquenessService, clock Clock) UpdateCategory {
	return &updateCategory{repo: repo, uniqueness: uniqueness, clock: clock}
}

func (uc *updateCategory) Execute(ctx context.Context, userID identityvo.UserID, rawID string, req dtos.CategoryRequest) (dtos.CategoryResponse, error) {
	id, err := vos.ParseCategoryID(rawID)
	if err != nil {
		return dtos.CategoryResponse{}, err
	}

	current, err := uc.repo.GetByID(ctx, userID, id)
	if err != nil {
		return dtos.CategoryResponse{}, fmt.Errorf("usecase update_category: %w", err)
	}

	name, err := vos.NewCategoryName(req.Name)
	if err != nil {
		return dtos.CategoryResponse{}, err
	}
	color, err := vos.NewCategoryColor(req.Color)
	if err != nil {
		return dtos.CategoryResponse{}, err
	}
	icon, err := vos.NewCategoryIcon(req.Icon)
	if err != nil {
		return dtos.CategoryResponse{}, err
	}

	newParentID, err := parseOptionalParentID(req.ParentID)
	if err != nil {
		return dtos.CategoryResponse{}, err
	}

	if err := uc.applyReparent(ctx, userID, current, newParentID); err != nil {
		return dtos.CategoryResponse{}, err
	}

	excludeID := current.ID()
	if err := uc.uniqueness.EnsureUnique(ctx, userID, name, current.ParentID(), &excludeID); err != nil {
		return dtos.CategoryResponse{}, err
	}

	if err := current.Rename(name, uc.clock); err != nil {
		return dtos.CategoryResponse{}, err
	}
	current.ChangeAppearance(color, icon, uc.clock)

	if err := uc.repo.Update(ctx, current); err != nil {
		return dtos.CategoryResponse{}, fmt.Errorf("usecase update_category: %w", err)
	}
	return dtos.ToCategoryResponse(current), nil
}

func (uc *updateCategory) applyReparent(ctx context.Context, userID identityvo.UserID, current *entities.Category, newParentID *vos.CategoryID) error {
	if !shouldReparent(current.ParentID(), newParentID) {
		return nil
	}
	if newParentID == nil {
		return domain.ErrParentNotFound
	}
	if current.IsRoot() || current.ID() == *newParentID {
		return domain.ErrSubcategoryDepthExceeded
	}
	parent, err := uc.repo.GetByIDIncludingDeleted(ctx, userID, *newParentID)
	if err != nil {
		if errors.Is(err, domain.ErrCategoryNotFound) {
			return domain.ErrParentNotFound
		}
		return fmt.Errorf("usecase update_category: %w", err)
	}
	if !parent.IsActive() {
		return domain.ErrParentInactive
	}
	if !parent.IsRoot() {
		return domain.ErrSubcategoryDepthExceeded
	}
	return current.Reparent(newParentID, uc.clock)
}

func shouldReparent(curr, next *vos.CategoryID) bool {
	if curr == nil && next == nil {
		return false
	}
	if curr != nil && next != nil && *curr == *next {
		return false
	}
	return true
}
