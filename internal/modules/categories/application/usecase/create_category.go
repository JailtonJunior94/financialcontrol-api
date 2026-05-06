package usecase

import (
	"context"
	"errors"
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

	parentID, err := parseOptionalParentID(req.ParentID)
	if err != nil {
		return dtos.CategoryResponse{}, err
	}

	if parentID != nil {
		if err := uc.assertParentEligible(ctx, userID, *parentID); err != nil {
			return dtos.CategoryResponse{}, err
		}
	}

	if err := uc.uniqueness.EnsureUnique(ctx, userID, name, parentID, nil); err != nil {
		return dtos.CategoryResponse{}, err
	}

	category, err := entities.NewCategory(userID, parentID, req.Name, req.Color, req.Icon, uc.clock)
	if err != nil {
		return dtos.CategoryResponse{}, err
	}

	if err := uc.repo.Add(ctx, category); err != nil {
		return dtos.CategoryResponse{}, fmt.Errorf("usecase create_category: %w", err)
	}
	return dtos.ToCategoryResponse(category), nil
}

func (uc *createCategory) assertParentEligible(ctx context.Context, userID identityvo.UserID, parentID vos.CategoryID) error {
	parent, err := uc.repo.GetByIDIncludingDeleted(ctx, userID, parentID)
	if err != nil {
		if errors.Is(err, domain.ErrCategoryNotFound) {
			return domain.ErrParentNotFound
		}
		return fmt.Errorf("usecase create_category: %w", err)
	}
	if !parent.IsActive() {
		return domain.ErrParentInactive
	}
	if !parent.IsRoot() {
		return domain.ErrSubcategoryDepthExceeded
	}
	return nil
}

func parseOptionalParentID(raw *string) (*vos.CategoryID, error) {
	if raw == nil || *raw == "" {
		return nil, nil
	}
	id, err := vos.ParseCategoryID(*raw)
	if err != nil {
		return nil, err
	}
	return &id, nil
}
