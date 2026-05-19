package ports

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

// ListFilter aggregates the filtering and pagination criteria accepted by
// CategoryRepository.List. Pagination is mandatory; the remaining fields are
// optional and combined with AND semantics.
type ListFilter struct {
	Pagination vos.Pagination
	NameLike   string
	OnlyRoots  bool
	OnlySubs   bool
	ParentID   *vos.CategoryID
}

// CategoryRepository defines the persistence contract for the Category aggregate.
// All operations are scoped to userID to preserve multi-tenant isolation.
type CategoryRepository interface {
	List(ctx context.Context, userID identityvo.UserID, filter ListFilter) ([]entities.Category, int64, error)

	// GetByID returns the active category with the given id that belongs to userID.
	// It returns (nil, domain.ErrCategoryNotFound) when the category does not exist,
	// is soft deleted, or does not belong to the user. It never returns (nil, nil).
	GetByID(ctx context.Context, userID identityvo.UserID, id vos.CategoryID) (*entities.Category, error)

	// GetByIDIncludingDeleted returns the category with the given id regardless of
	// its soft-delete status, still scoped to userID.
	GetByIDIncludingDeleted(ctx context.Context, userID identityvo.UserID, id vos.CategoryID) (*entities.Category, error)

	// GetActiveChildren returns the active subcategories whose parent_id equals parentID
	// and that belong to userID. Returns an empty slice when there are no children.
	GetActiveChildren(ctx context.Context, userID identityvo.UserID, parentID vos.CategoryID) ([]entities.Category, error)

	// ExistsByName reports whether an active category with the given name exists in
	// the (userID, parentID) scope. When excludeID is non-nil the row with that id
	// is ignored, which lets update use cases re-check uniqueness without false
	// positives against the row being updated.
	ExistsByName(ctx context.Context, userID identityvo.UserID, name vos.CategoryName, parentID *vos.CategoryID, excludeID *vos.CategoryID) (bool, error)

	Add(ctx context.Context, category *entities.Category) error
	Update(ctx context.Context, category *entities.Category) error

	// SoftDeleteCascade marks the root category rootID and all of its active
	// children as deleted at deletedAt within a single transaction. It returns
	// domain.ErrCategoryNotFound when rootID does not exist for userID. When the
	// target is already soft deleted the operation is a no-op (idempotent).
	SoftDeleteCascade(ctx context.Context, userID identityvo.UserID, rootID vos.CategoryID, deletedAt time.Time) error
}
