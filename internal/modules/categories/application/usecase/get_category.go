package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

type getCategory struct {
	repo interfaces.CategoryRepository
}

func NewGetCategory(repo interfaces.CategoryRepository) GetCategory {
	return &getCategory{repo: repo}
}

func (uc *getCategory) Execute(ctx context.Context, userID identityvo.UserID, rawID string) (dtos.CategoryResponse, error) {
	id, err := vos.ParseCategoryID(rawID)
	if err != nil {
		return dtos.CategoryResponse{}, err
	}

	category, err := uc.repo.GetByID(ctx, userID, id)
	if err != nil {
		return dtos.CategoryResponse{}, fmt.Errorf("usecase get_category: %w", err)
	}

	resp := dtos.ToCategoryResponse(category)
	if pid := category.ParentID(); pid != nil {
		parent, err := uc.repo.GetByID(ctx, userID, *pid)
		if err != nil {
			if errors.Is(err, domain.ErrCategoryNotFound) {
				return resp, nil
			}
			return dtos.CategoryResponse{}, fmt.Errorf("usecase get_category: %w", err)
		}
		summary := dtos.ToCategorySummary(parent)
		resp.Parent = &summary
	}
	return resp, nil
}
