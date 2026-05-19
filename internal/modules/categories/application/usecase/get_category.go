package usecase

import (
	"context"
	"fmt"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/ports"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

type getCategory struct {
	repo ports.CategoryRepository
}

func NewGetCategory(repo ports.CategoryRepository) GetCategory {
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
	return dtos.ToCategoryResponse(category), nil
}
