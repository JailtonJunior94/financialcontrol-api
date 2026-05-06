package usecase

import (
	"context"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

type CreateCategory interface {
	Execute(ctx context.Context, userID identityvo.UserID, req dtos.CategoryRequest) (dtos.CategoryResponse, error)
}

type UpdateCategory interface {
	Execute(ctx context.Context, userID identityvo.UserID, id string, req dtos.CategoryRequest) (dtos.CategoryResponse, error)
}

type DeleteCategory interface {
	Execute(ctx context.Context, userID identityvo.UserID, id string) error
}

type GetCategory interface {
	Execute(ctx context.Context, userID identityvo.UserID, id string) (dtos.CategoryResponse, error)
}

type ListCategories interface {
	Execute(ctx context.Context, userID identityvo.UserID, q dtos.ListCategoriesQuery) (dtos.PaginatedResponse[dtos.CategoryResponse], error)
}
