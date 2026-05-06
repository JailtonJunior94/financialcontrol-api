package mssql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

var _ interfaces.CategoryRepository = (*CategoryRepository)(nil)

const moduleName = "categories"

// CategoryRepository implements the persistence contract for the Category
// aggregate against SQL Server. All queries are scoped to userID and ignore
// soft-deleted rows unless explicitly required by the operation.
type CategoryRepository struct {
	db *sqlx.DB
}

func NewCategoryRepository(db *sqlx.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) List(
	ctx context.Context,
	userID identityvo.UserID,
	filter interfaces.ListFilter,
) ([]entities.Category, int64, error) {
	args := buildListArgs(userID, filter)

	var total int64
	if err := r.db.QueryRowxContext(ctx, listCategoriesCount, args...).Scan(&total); err != nil {
		slog.ErrorContext(ctx, "categories list count failed",
			slog.String("module", moduleName), slog.String("userId", userID.String()), slog.String("error", err.Error()))
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

	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		slog.ErrorContext(ctx, "categories list query failed",
			slog.String("module", moduleName), slog.String("userId", userID.String()), slog.String("error", err.Error()))
		return nil, 0, fmt.Errorf("mssql: list categories: %w", err)
	}
	defer func() { _ = rows.Close() }()

	items := make([]entities.Category, 0)
	for rows.Next() {
		var row CategoryRow
		if err := rows.StructScan(&row); err != nil {
			return nil, 0, fmt.Errorf("mssql: list categories scan: %w", err)
		}
		category, err := RowToCategory(&row)
		if err != nil {
			return nil, 0, fmt.Errorf("mssql: list categories map: %w", err)
		}
		items = append(items, *category)
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
	userID identityvo.UserID,
	id vos.CategoryID,
	query string,
) (*entities.Category, error) {
	var row CategoryRow
	err := r.db.QueryRowxContext(ctx, query,
		sql.Named("userId", userID.String()),
		sql.Named("id", id.String()),
	).StructScan(&row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrCategoryNotFound
	}
	if err != nil {
		slog.ErrorContext(ctx, "categories get by id failed",
			slog.String("module", moduleName), slog.String("userId", userID.String()),
			slog.String("categoryId", id.String()), slog.String("error", err.Error()))
		return nil, fmt.Errorf("mssql: get category by id: %w", err)
	}
	return RowToCategory(&row)
}

func (r *CategoryRepository) GetActiveChildren(
	ctx context.Context,
	userID identityvo.UserID,
	parentID vos.CategoryID,
) ([]entities.Category, error) {
	rows, err := r.db.QueryxContext(ctx, getActiveChildren,
		sql.Named("userId", userID.String()),
		sql.Named("parentId", parentID.String()),
	)
	if err != nil {
		slog.ErrorContext(ctx, "categories get active children failed",
			slog.String("module", moduleName), slog.String("userId", userID.String()),
			slog.String("parentId", parentID.String()), slog.String("error", err.Error()))
		return nil, fmt.Errorf("mssql: get active children: %w", err)
	}
	defer func() { _ = rows.Close() }()

	items := make([]entities.Category, 0)
	for rows.Next() {
		var row CategoryRow
		if err := rows.StructScan(&row); err != nil {
			return nil, fmt.Errorf("mssql: get active children scan: %w", err)
		}
		category, err := RowToCategory(&row)
		if err != nil {
			return nil, fmt.Errorf("mssql: get active children map: %w", err)
		}
		items = append(items, *category)
	}
	return items, nil
}

func (r *CategoryRepository) ExistsByName(
	ctx context.Context,
	userID identityvo.UserID,
	name vos.CategoryName,
	parentID *vos.CategoryID,
	excludeID *vos.CategoryID,
) (bool, error) {
	query := existsByNameRoot
	args := []any{
		sql.Named("userId", userID.String()),
		sql.Named("name", name.String()),
		sql.Named("excludeId", nullableIDArg(excludeID)),
	}
	if parentID != nil {
		query = existsByNameSub
		args = append(args, sql.Named("parentId", parentID.String()))
	}

	var count int
	if err := r.db.QueryRowxContext(ctx, query, args...).Scan(&count); err != nil {
		slog.ErrorContext(ctx, "categories exists by name failed",
			slog.String("module", moduleName), slog.String("userId", userID.String()), slog.String("error", err.Error()))
		return false, fmt.Errorf("mssql: exists by name: %w", err)
	}
	return count > 0, nil
}

func (r *CategoryRepository) Add(ctx context.Context, category *entities.Category) error {
	_, err := r.db.ExecContext(ctx, addCategory,
		sql.Named("id", category.ID().String()),
		sql.Named("userId", category.UserID().String()),
		sql.Named("parentId", parentIDArg(category.ParentID())),
		sql.Named("name", category.Name().String()),
		sql.Named("color", category.Color().String()),
		sql.Named("icon", category.Icon().String()),
		sql.Named("createdAt", category.CreatedAt()),
		sql.Named("updatedAt", category.UpdatedAt()),
		sql.Named("deletedAt", deletedAtArg(category.DeletedAt())),
	)
	if err != nil {
		mapped := MapDriverError(err)
		if errors.Is(mapped, domain.ErrCategoryNameAlreadyExists) {
			return mapped
		}
		slog.ErrorContext(ctx, "categories add failed",
			slog.String("module", moduleName), slog.String("userId", category.UserID().String()),
			slog.String("categoryId", category.ID().String()), slog.String("error", err.Error()))
		return fmt.Errorf("mssql: add category: %w", err)
	}
	return nil
}

func (r *CategoryRepository) Update(ctx context.Context, category *entities.Category) error {
	result, err := r.db.ExecContext(ctx, updateCategory,
		sql.Named("parentId", parentIDArg(category.ParentID())),
		sql.Named("name", category.Name().String()),
		sql.Named("color", category.Color().String()),
		sql.Named("icon", category.Icon().String()),
		sql.Named("updatedAt", category.UpdatedAt()),
		sql.Named("deletedAt", deletedAtArg(category.DeletedAt())),
		sql.Named("id", category.ID().String()),
		sql.Named("userId", category.UserID().String()),
	)
	if err != nil {
		mapped := MapDriverError(err)
		if errors.Is(mapped, domain.ErrCategoryNameAlreadyExists) {
			return mapped
		}
		slog.ErrorContext(ctx, "categories update failed",
			slog.String("module", moduleName), slog.String("userId", category.UserID().String()),
			slog.String("categoryId", category.ID().String()), slog.String("error", err.Error()))
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

func (r *CategoryRepository) SoftDeleteCascade(
	ctx context.Context,
	userID identityvo.UserID,
	rootID vos.CategoryID,
	deletedAt time.Time,
) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("mssql: soft delete cascade begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var exists int
	if err := tx.QueryRowxContext(ctx, categoryExistsForUser,
		sql.Named("id", rootID.String()),
		sql.Named("userId", userID.String()),
	).Scan(&exists); err != nil {
		return fmt.Errorf("mssql: soft delete cascade lookup: %w", err)
	}
	if exists == 0 {
		return domain.ErrCategoryNotFound
	}

	if _, err := tx.ExecContext(ctx, softDeleteCascade,
		sql.Named("userId", userID.String()),
		sql.Named("rootId", rootID.String()),
		sql.Named("deletedAt", deletedAt.UTC()),
	); err != nil {
		slog.ErrorContext(ctx, "categories soft delete cascade failed",
			slog.String("module", moduleName), slog.String("userId", userID.String()),
			slog.String("rootId", rootID.String()), slog.String("error", err.Error()))
		return fmt.Errorf("mssql: soft delete cascade: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("mssql: soft delete cascade commit: %w", err)
	}
	return nil
}

func buildListArgs(userID identityvo.UserID, filter interfaces.ListFilter) []any {
	var nameLike any
	if filter.NameLike != "" {
		nameLike = "%" + filter.NameLike + "%"
	}
	var parentID any
	if filter.ParentID != nil {
		parentID = filter.ParentID.String()
	}
	onlyRoots := 0
	if filter.OnlyRoots {
		onlyRoots = 1
	}
	onlySubs := 0
	if filter.OnlySubs {
		onlySubs = 1
	}
	return []any{
		sql.Named("userId", userID.String()),
		sql.Named("nameLike", nameLike),
		sql.Named("onlyRoots", onlyRoots),
		sql.Named("onlySubs", onlySubs),
		sql.Named("parentId", parentID),
	}
}

func parentIDArg(parentID *vos.CategoryID) any {
	if parentID == nil {
		return nil
	}
	return parentID.String()
}

func deletedAtArg(deletedAt *time.Time) any {
	if deletedAt == nil {
		return nil
	}
	return deletedAt.UTC()
}

func nullableIDArg(id *vos.CategoryID) any {
	if id == nil {
		return nil
	}
	return id.String()
}
