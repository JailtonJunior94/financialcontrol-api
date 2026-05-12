package providers

import (
	"context"
	"errors"
	"fmt"

	categoriesdomain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain"
	categoriesinterfaces "github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/interfaces"
	categoriesvos "github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/vos"
	financedomain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/ports"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/projections"
	financevos "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

var _ ports.CategoryProvider = (*CategoryProviderAdapter)(nil)

// CategoryProviderAdapter wraps categories.CategoryRepository to implement finance.ports.CategoryProvider.
// Decisão G2.a: not-found or wrong-user → ErrCategoryNotFound; inactive category → CategoryView{Active:false}, nil.
type CategoryProviderAdapter struct {
	repo categoriesinterfaces.CategoryRepository
}

func NewCategoryProviderAdapter(repo categoriesinterfaces.CategoryRepository) *CategoryProviderAdapter {
	return &CategoryProviderAdapter{repo: repo}
}

func (a *CategoryProviderAdapter) GetByID(
	ctx context.Context,
	userID identityvo.UserID,
	categoryID financevos.CategoryID,
) (projections.CategoryView, error) {
	cid, err := categoriesvos.ParseCategoryID(categoryID.String())
	if err != nil {
		return projections.CategoryView{}, fmt.Errorf("category provider: parse category id: %w", err)
	}

	category, err := a.repo.GetByID(ctx, userID, cid)
	if errors.Is(err, categoriesdomain.ErrCategoryNotFound) {
		return projections.CategoryView{}, financedomain.ErrCategoryNotFound
	}
	if err != nil {
		return projections.CategoryView{}, fmt.Errorf("category provider: get by id: %w", err)
	}

	return projections.CategoryView{
		ID:       financevos.CategoryID(category.ID().String()),
		UserID:   userID,
		Name:     category.Name().String(),
		ParentID: nil,
		Active:   category.IsActive(),
	}, nil
}
