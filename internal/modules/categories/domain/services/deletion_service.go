package services

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

// Clock abstracts time.Now to keep the deletion service deterministic in tests.
type Clock interface {
	Now() time.Time
}

// CategoryDeletionService deactivates rows according to the legacy Category schema.
type CategoryDeletionService struct {
	repo  interfaces.CategoryRepository
	clock Clock
}

func NewCategoryDeletionService(repo interfaces.CategoryRepository, clock Clock) *CategoryDeletionService {
	return &CategoryDeletionService{repo: repo, clock: clock}
}

func (s *CategoryDeletionService) Delete(ctx context.Context, userID identityvo.UserID, id vos.CategoryID) error {
	target, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return err
	}
	return s.repo.SoftDeleteCascade(ctx, userID, target.ID(), s.clock.Now().UTC())
}
