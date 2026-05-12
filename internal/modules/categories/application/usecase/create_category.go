package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/services"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

// Clock abstracts time.Now to keep use cases deterministic in tests.
type Clock interface {
	Now() time.Time
}

type createCategory struct {
	repo       interfaces.CategoryRepository
	uniqueness *services.CategoryUniquenessService
	clock      Clock
}

func NewCreateCategory(repo interfaces.CategoryRepository, uniqueness *services.CategoryUniquenessService, clock Clock) CreateCategory {
	return &createCategory{repo: repo, uniqueness: uniqueness, clock: clock}
}

func (uc *createCategory) Execute(ctx context.Context, userID identityvo.UserID, req dtos.CategoryRequest) (dtos.CategoryResponse, error) {
	name, err := vos.NewCategoryName(req.Name)
	if err != nil {
		return dtos.CategoryResponse{}, err
	}
	if req.ParentID != nil && *req.ParentID != "" {
		return dtos.CategoryResponse{}, domain.ErrCategoryHierarchyUnsupported
	}

	if err := uc.uniqueness.EnsureUnique(ctx, userID, name, nil, nil); err != nil {
		return dtos.CategoryResponse{}, err
	}

	category, err := entities.NewCategory(req.Name, req.Sequence, uc.clock)
	if err != nil {
		return dtos.CategoryResponse{}, err
	}

	if err := uc.repo.Add(ctx, category); err != nil {
		return dtos.CategoryResponse{}, fmt.Errorf("usecase create_category: %w", err)
	}
	return dtos.ToCategoryResponse(category), nil
}
