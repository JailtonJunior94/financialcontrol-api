package usecase

import (
	"context"
	"fmt"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/services"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

type deleteCategory struct {
	deletion *services.CategoryDeletionService
}

func NewDeleteCategory(deletion *services.CategoryDeletionService) DeleteCategory {
	return &deleteCategory{deletion: deletion}
}

func (uc *deleteCategory) Execute(ctx context.Context, userID identityvo.UserID, rawID string) error {
	id, err := vos.ParseCategoryID(rawID)
	if err != nil {
		return err
	}
	if err := uc.deletion.Delete(ctx, userID, id); err != nil {
		return fmt.Errorf("usecase delete_category: %w", err)
	}
	return nil
}
