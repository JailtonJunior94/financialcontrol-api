package mssql

import (
	"database/sql"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

// CategoryRow holds the DB columns for a Category scan.
type CategoryRow struct {
	ID        string         `db:"Id"`
	UserID    string         `db:"UserId"`
	ParentID  sql.NullString `db:"ParentId"`
	Name      string         `db:"Name"`
	Color     string         `db:"Color"`
	Icon      string         `db:"Icon"`
	CreatedAt time.Time      `db:"CreatedAt"`
	UpdatedAt time.Time      `db:"UpdatedAt"`
	DeletedAt sql.NullTime   `db:"DeletedAt"`
}

// RowToCategory maps a CategoryRow into the domain Category aggregate (no IO).
// It uses entities.RehydrateCategory so persisted rows skip costly construction
// validation while still preserving VO invariants by parsing each scalar.
func RowToCategory(r *CategoryRow) (*entities.Category, error) {
	id, err := vos.ParseCategoryID(r.ID)
	if err != nil {
		return nil, err
	}
	userID, err := identityvo.ParseUserID(r.UserID)
	if err != nil {
		return nil, err
	}
	name, err := vos.NewCategoryName(r.Name)
	if err != nil {
		return nil, err
	}
	color, err := vos.NewCategoryColor(r.Color)
	if err != nil {
		return nil, err
	}
	icon, err := vos.NewCategoryIcon(r.Icon)
	if err != nil {
		return nil, err
	}

	var parentID *vos.CategoryID
	if r.ParentID.Valid && r.ParentID.String != "" {
		pid, err := vos.ParseCategoryID(r.ParentID.String)
		if err != nil {
			return nil, err
		}
		parentID = &pid
	}

	var deletedAt *time.Time
	if r.DeletedAt.Valid {
		t := r.DeletedAt.Time.UTC()
		deletedAt = &t
	}

	return entities.RehydrateCategory(
		id,
		userID,
		parentID,
		name,
		color,
		icon,
		r.CreatedAt.UTC(),
		r.UpdatedAt.UTC(),
		deletedAt,
	), nil
}
