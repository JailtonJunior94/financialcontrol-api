package mssql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	devkitdb "github.com/JailtonJunior94/devkit-go/pkg/database"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/vos"
	pkgdatabase "github.com/jailtonjunior94/financialcontrol-api/pkg/database"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

var _ interfaces.CategoryRepository = (*CategoryRepository)(nil)

const moduleName = "categories"

// CategoryRepository implements the legacy Category table in SQL Server.
// The authenticated userID is still accepted for interface compatibility but
// no longer participates in SQL filtering because the schema is global.
type CategoryRepository struct {
	db devkitdb.DBTX
}

func NewCategoryRepository(db devkitdb.DBTX) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) List(
	ctx context.Context,
	userID identityvo.UserID,
	filter interfaces.ListFilter,
) ([]entities.Category, int64, error) {
	args := buildListArgs(userID, filter)

	var total int64
	if err := r.db.QueryRowContext(ctx, listCategoriesCount, args...).Scan(&total); err != nil {
		slog.ErrorContext(ctx, "categories list count failed",
			slog.String("module", moduleName), slog.String("error", err.Error()))
		return nil, 0, fmt.Errorf("mssql: list categories count: %w", err)
	}

	query := listCategories
	if filter.Pagination.Enabled() {
		query = listCategoriesPaginated
		args = append(args,
			sql.Named("offset", filter.Pagination.Offset()),
			sql.Named("size", filter.Pagination.Size),
		)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		slog.ErrorContext(ctx, "categories list query failed",
			slog.String("module", moduleName), slog.String("error", err.Error()))
		return nil, 0, fmt.Errorf("mssql: list categories: %w", err)
	}

	items, err := pkgdatabase.ScanAll[entities.Category](rows, func(r devkitdb.Rows) (entities.Category, error) {
		var row CategoryRow
		if err := r.Scan(&row.ID, &row.Name, &row.Sequence, &row.CreatedAt, &row.UpdatedAt, &row.Active); err != nil {
			return entities.Category{}, fmt.Errorf("mssql: list categories scan: %w", err)
		}
		category, err := RowToCategory(&row)
		if err != nil {
			return entities.Category{}, fmt.Errorf("mssql: list categories map: %w", err)
		}
		return *category, nil
	})
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *CategoryRepository) GetByID(
	ctx context.Context,
	userID identityvo.UserID,
	id vos.CategoryID,
) (*entities.Category, error) {
	return r.getByIDWithQuery(ctx, userID, id, getCategoryByID)
}

func (r *CategoryRepository) GetByIDIncludingDeleted(
	ctx context.Context,
	userID identityvo.UserID,
	id vos.CategoryID,
) (*entities.Category, error) {
	return r.getByIDWithQuery(ctx, userID, id, getCategoryByIDIncludingDeleted)
}

func (r *CategoryRepository) getByIDWithQuery(
	ctx context.Context,
	_ identityvo.UserID,
	id vos.CategoryID,
	query string,
) (*entities.Category, error) {
	var row CategoryRow
	err := r.db.QueryRowContext(ctx, query, sql.Named("id", id.String())).
		Scan(&row.ID, &row.Name, &row.Sequence, &row.CreatedAt, &row.UpdatedAt, &row.Active)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrCategoryNotFound
	}
	if err != nil {
		slog.ErrorContext(ctx, "categories get by id failed",
			slog.String("module", moduleName),
			slog.String("categoryId", id.String()), slog.String("error", err.Error()))
		return nil, fmt.Errorf("mssql: get category by id: %w", err)
	}
	return RowToCategory(&row)
}

func (r *CategoryRepository) GetActiveChildren(
	ctx context.Context,
	_ identityvo.UserID,
	_ vos.CategoryID,
) ([]entities.Category, error) {
	rows, err := r.db.QueryContext(ctx, getActiveChildren)
	if err != nil {
		slog.ErrorContext(ctx, "categories get active children failed",
			slog.String("module", moduleName), slog.String("error", err.Error()))
		return nil, fmt.Errorf("mssql: get active children: %w", err)
	}

	items, err := pkgdatabase.ScanAll[entities.Category](rows, func(r devkitdb.Rows) (entities.Category, error) {
		var row CategoryRow
		if err := r.Scan(&row.ID, &row.Name, &row.Sequence, &row.CreatedAt, &row.UpdatedAt, &row.Active); err != nil {
			return entities.Category{}, fmt.Errorf("mssql: get active children scan: %w", err)
		}
		category, err := RowToCategory(&row)
		if err != nil {
			return entities.Category{}, fmt.Errorf("mssql: get active children map: %w", err)
		}
		return *category, nil
	})
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (r *CategoryRepository) ExistsByName(
	ctx context.Context,
	_ identityvo.UserID,
	name vos.CategoryName,
	_ *vos.CategoryID,
	excludeID *vos.CategoryID,
) (bool, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, existsByNameRoot,
		sql.Named("name", name.String()),
		sql.Named("excludeId", nullableIDArg(excludeID)),
	).Scan(&count); err != nil {
		slog.ErrorContext(ctx, "categories exists by name failed",
			slog.String("module", moduleName), slog.String("error", err.Error()))
		return false, fmt.Errorf("mssql: exists by name: %w", err)
	}
	return count > 0, nil
}

func (r *CategoryRepository) Add(ctx context.Context, category *entities.Category) error {
	_, err := r.db.ExecContext(ctx, addCategory,
		sql.Named("id", category.ID().String()),
		sql.Named("name", category.Name().String()),
		sql.Named("sequence", category.Sequence()),
		sql.Named("createdAt", category.CreatedAt()),
		sql.Named("updatedAt", category.UpdatedAt()),
		sql.Named("active", category.IsActive()),
	)
	if err != nil {
		mapped := MapDriverError(err)
		if errors.Is(mapped, domain.ErrCategoryNameAlreadyExists) {
			return mapped
		}
		slog.ErrorContext(ctx, "categories add failed",
			slog.String("module", moduleName), slog.String("categoryId", category.ID().String()),
			slog.String("error", err.Error()))
		return fmt.Errorf("mssql: add category: %w", err)
	}
	return nil
}

func (r *CategoryRepository) Update(ctx context.Context, category *entities.Category) error {
	result, err := r.db.ExecContext(ctx, updateCategory,
		sql.Named("name", category.Name().String()),
		sql.Named("sequence", category.Sequence()),
		sql.Named("updatedAt", category.UpdatedAt()),
		sql.Named("active", category.IsActive()),
		sql.Named("id", category.ID().String()),
	)
	if err != nil {
		mapped := MapDriverError(err)
		if errors.Is(mapped, domain.ErrCategoryNameAlreadyExists) {
			return mapped
		}
		slog.ErrorContext(ctx, "categories update failed",
			slog.String("module", moduleName), slog.String("categoryId", category.ID().String()),
			slog.String("error", err.Error()))
		return fmt.Errorf("mssql: update category: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("mssql: update category rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrCategoryNotFound
	}
	return nil
}

// SoftDeleteCascade is kept for interface compatibility; under the legacy
// schema it simply deactivates the target row.
func (r *CategoryRepository) SoftDeleteCascade(
	ctx context.Context,
	_ identityvo.UserID,
	rootID vos.CategoryID,
	deletedAt time.Time,
) error {
	var exists int
	if err := r.db.QueryRowContext(ctx, categoryExistsForUser,
		sql.Named("id", rootID.String()),
	).Scan(&exists); err != nil {
		return fmt.Errorf("mssql: soft delete cascade lookup: %w", err)
	}
	if exists == 0 {
		return domain.ErrCategoryNotFound
	}

	if _, err := r.db.ExecContext(ctx, softDeleteCascade,
		sql.Named("rootId", rootID.String()),
		sql.Named("updatedAt", deletedAt.UTC()),
	); err != nil {
		slog.ErrorContext(ctx, "categories soft delete cascade failed",
			slog.String("module", moduleName), slog.String("rootId", rootID.String()),
			slog.String("error", err.Error()))
		return fmt.Errorf("mssql: soft delete cascade: %w", err)
	}

	return nil
}

func buildListArgs(_ identityvo.UserID, filter interfaces.ListFilter) []any {
	var nameLike any
	if filter.NameLike != "" {
		nameLike = "%" + filter.NameLike + "%"
	}
	return []any{
		sql.Named("nameLike", nameLike),
	}
}

func nullableIDArg(id *vos.CategoryID) any {
	if id == nil {
		return nil
	}
	return id.String()
}
