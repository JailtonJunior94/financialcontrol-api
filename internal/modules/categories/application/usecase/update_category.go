package usecase

import (
	"context"
	"fmt"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain"
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
	if req.ParentID != nil && *req.ParentID != "" {
		return dtos.CategoryResponse{}, domain.ErrCategoryHierarchyUnsupported
	}

	current, err := uc.repo.GetByID(ctx, userID, id)
	if err != nil {
		return dtos.CategoryResponse{}, fmt.Errorf("usecase update_category: %w", err)
	}

	name, err := vos.NewCategoryName(req.Name)
	if err != nil {
		return dtos.CategoryResponse{}, err
	}

	excludeID := current.ID()
	if err := uc.uniqueness.EnsureUnique(ctx, userID, name, nil, &excludeID); err != nil {
		return dtos.CategoryResponse{}, err
	}

	if err := current.Update(name, req.Sequence, uc.clock); err != nil {
		return dtos.CategoryResponse{}, err
	}

	if err := uc.repo.Update(ctx, current); err != nil {
		return dtos.CategoryResponse{}, fmt.Errorf("usecase update_category: %w", err)
	}
	return dtos.ToCategoryResponse(current), nil
}
