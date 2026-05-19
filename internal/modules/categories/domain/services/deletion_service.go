package services

import (
	"context"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/ports"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

// CategoryDeletionService deactivates rows according to the legacy Category schema.
type CategoryDeletionService struct {
	repo  ports.CategoryRepository
	clock ports.Clock
}

func NewCategoryDeletionService(repo ports.CategoryRepository, clock ports.Clock) *CategoryDeletionService {
	return &CategoryDeletionService{repo: repo, clock: clock}
}

func (s *CategoryDeletionService) Delete(ctx context.Context, userID identityvo.UserID, id vos.CategoryID) error {
	target, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return err
	}
	return s.repo.SoftDeleteCascade(ctx, userID, target.ID(), s.clock.Now().UTC())
}
