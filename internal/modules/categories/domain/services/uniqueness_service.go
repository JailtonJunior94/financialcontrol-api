package services

import (
	"context"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

// CategoryUniquenessService validates that a category name is unique within the
// scope determined by (userID, parentID), considering only active rows. When
// excludeID is non-nil the row with that id is ignored, which lets update use
// cases re-check uniqueness without colliding with the row being updated.
type CategoryUniquenessService struct {
	repo interfaces.CategoryRepository
}

func NewCategoryUniquenessService(repo interfaces.CategoryRepository) *CategoryUniquenessService {
	return &CategoryUniquenessService{repo: repo}
}

func (s *CategoryUniquenessService) EnsureUnique(
	ctx context.Context,
	userID identityvo.UserID,
	name vos.CategoryName,
	parentID *vos.CategoryID,
	excludeID *vos.CategoryID,
) error {
	_ = parentID
	exists, err := s.repo.ExistsByName(ctx, userID, name, parentID, excludeID)
	if err != nil {
		return err
	}
	if exists {
		return domain.ErrCategoryNameAlreadyExists
	}
	return nil
}
