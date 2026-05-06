package usecase

import (
	"context"
	"fmt"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

const (
	defaultListPage     = 1
	defaultListPageSize = 10
)

type listCategories struct {
	repo interfaces.CategoryRepository
}

func NewListCategories(repo interfaces.CategoryRepository) ListCategories {
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

func buildListFilter(q dtos.ListCategoriesQuery) (interfaces.ListFilter, int, int, error) {
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

	filter := interfaces.ListFilter{
		Pagination: pagination,
		NameLike:   q.Name,
	}

	if q.ParentID != "" {
		parsed, err := vos.ParseCategoryID(q.ParentID)
		if err != nil {
			return interfaces.ListFilter{}, 0, 0, err
		}
		filter.ParentID = &parsed
		filter.OnlySubs = true
	}

	if filter.ParentID != nil {
		return filter, page, size, nil
	}

	switch q.Scope {
	case dtos.ScopeRoots:
		filter.OnlyRoots = true
	case dtos.ScopeSubs:
		filter.OnlySubs = true
	case dtos.ScopeAll, "":
		// no extra filter
	}

	return filter, page, size, nil
}
