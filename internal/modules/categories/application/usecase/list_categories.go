package usecase

import (
	"context"
	"fmt"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/ports"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

const (
	defaultListPage     = 1
	defaultListPageSize = 10
)

type listCategories struct {
	repo ports.CategoryRepository
}

func NewListCategories(repo ports.CategoryRepository) ListCategories {
	return &listCategories{repo: repo}
}

func (uc *listCategories) Execute(ctx context.Context, userID identityvo.UserID, q dtos.ListCategoriesQuery) (dtos.PaginatedResponse[dtos.CategoryResponse], error) {
	filter, page, size, err := buildListFilter(q)
	if err != nil {
		return dtos.PaginatedResponse[dtos.CategoryResponse]{}, err
	}

	items, total, err := uc.repo.List(ctx, userID, filter)
	if err != nil {
		return dtos.PaginatedResponse[dtos.CategoryResponse]{}, fmt.Errorf("usecase list_categories: %w", err)
	}

	return dtos.PaginatedResponse[dtos.CategoryResponse]{
		Items:    dtos.ToCategoryResponses(items),
		Total:    total,
		Page:     page,
		PageSize: size,
	}, nil
}

func buildListFilter(q dtos.ListCategoriesQuery) (ports.ListFilter, int, int, error) {
	page := q.Page
	if page <= 0 {
		page = defaultListPage
	}
	size := q.PageSize
	if size <= 0 {
		size = defaultListPageSize
	}

	pagination := vos.NewPagination(page, size)
	page = pagination.Page
	size = pagination.Size

	if q.ParentID != "" || q.Scope == dtos.ScopeSubs {
		return ports.ListFilter{}, 0, 0, domain.ErrCategoryHierarchyUnsupported
	}

	return ports.ListFilter{
		Pagination: pagination,
		NameLike:   q.Name,
	}, page, size, nil
}
