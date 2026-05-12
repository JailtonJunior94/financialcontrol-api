package mssql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	devkitdb "github.com/JailtonJunior94/devkit-go/pkg/database"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/vos"
	pkgdatabase "github.com/jailtonjunior94/financialcontrol-api/pkg/database"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

var _ interfaces.CardRepository = (*CardRepository)(nil)

type CardRepository struct {
	db devkitdb.DBTX
}

func NewCardRepository(db devkitdb.DBTX) *CardRepository {
	return &CardRepository{db: db}
}

func (r *CardRepository) List(ctx context.Context, userID identityvo.UserID, pagination vos.Pagination) ([]entities.Card, error) {
	query := listCards
	args := []any{sql.Named("userId", userID.String())}
	if pagination.Enabled() {
		query = listCardsPaginated
		args = append(args,
			sql.Named("offset", pagination.Offset()),
			sql.Named("size", pagination.Size),
		)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("mssql: list cards: %w", err)
	}

	// Column order: Id, UserId, FlagId, Name, Number, Description, ClosingDay,
	// ExpirationDate, CreatedAt, UpdatedAt, Active, FlagEntityId, FlagName, FlagActive
	return pkgdatabase.ScanAll[entities.Card](rows, func(r devkitdb.Rows) (entities.Card, error) {
		var row CardRow
		if err := r.Scan(
			&row.ID, &row.UserID, &row.FlagID, &row.Name, &row.Number, &row.Description,
			&row.ClosingDay, &row.ExpirationDate, &row.CreatedAt, &row.UpdatedAt, &row.Active,
			&row.FlagEntityID, &row.FlagName, &row.FlagActive,
		); err != nil {
			return entities.Card{}, fmt.Errorf("mssql: list cards scan: %w", err)
		}
		card, err := RowToCard(&row)
		if err != nil {
			return entities.Card{}, fmt.Errorf("mssql: list cards map: %w", err)
		}
		return *card, nil
	})
}

func (r *CardRepository) GetByID(ctx context.Context, userID identityvo.UserID, id vos.CardID) (*entities.Card, error) {
	// Column order: Id, UserId, FlagId, Name, Number, Description, ClosingDay,
	// ExpirationDate, CreatedAt, UpdatedAt, Active, FlagEntityId, FlagName, FlagActive
	var row CardRow
	err := r.db.QueryRowContext(ctx, getCardByID,
		sql.Named("userId", userID.String()),
		sql.Named("id", id.String()),
	).Scan(
		&row.ID, &row.UserID, &row.FlagID, &row.Name, &row.Number, &row.Description,
		&row.ClosingDay, &row.ExpirationDate, &row.CreatedAt, &row.UpdatedAt, &row.Active,
		&row.FlagEntityID, &row.FlagName, &row.FlagActive,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrCardNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("mssql: get card by id: %w", err)
	}
	card, err := RowToCard(&row)
	if err != nil {
		return nil, fmt.Errorf("mssql: get card by id map: %w", err)
	}
	return card, nil
}

func (r *CardRepository) Add(ctx context.Context, card *entities.Card) error {
	_, err := r.db.ExecContext(ctx, addCard,
		sql.Named("id", card.ID().String()),
		sql.Named("userId", card.UserID().String()),
		sql.Named("flagId", card.FlagID().String()),
		sql.Named("name", card.Name().String()),
		sql.Named("number", card.Number().String()),
		sql.Named("description", card.Description()),
		sql.Named("closingDay", card.ClosingDay().Int()),
		sql.Named("expirationDate", cardToExpirationDate(card)),
		sql.Named("createdAt", card.CreatedAt()),
		sql.Named("updatedAt", card.UpdatedAt()),
		sql.Named("active", card.Active()),
	)
	if err != nil {
		return MapDriverError(fmt.Errorf("mssql: add card: %w", err))
	}
	return nil
}

func (r *CardRepository) Update(ctx context.Context, card *entities.Card) error {
	_, err := r.db.ExecContext(ctx, updateCard,
		sql.Named("flagId", card.FlagID().String()),
		sql.Named("name", card.Name().String()),
		sql.Named("number", card.Number().String()),
		sql.Named("description", card.Description()),
		sql.Named("closingDay", card.ClosingDay().Int()),
		sql.Named("expirationDate", cardToExpirationDate(card)),
		sql.Named("updatedAt", card.UpdatedAt()),
		sql.Named("active", card.Active()),
		sql.Named("id", card.ID().String()),
		sql.Named("userId", card.UserID().String()),
	)
	if err != nil {
		return MapDriverError(fmt.Errorf("mssql: update card: %w", err))
	}
	return nil
}
