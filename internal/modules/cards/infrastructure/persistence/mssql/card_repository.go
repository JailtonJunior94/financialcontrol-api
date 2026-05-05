package mssql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

var _ interfaces.CardRepository = (*CardRepository)(nil)

type CardRepository struct {
	db *sqlx.DB
}

func NewCardRepository(db *sqlx.DB) *CardRepository {
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

	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("mssql: list cards: %w", err)
	}
	defer func() { _ = rows.Close() }()

	cards := make([]entities.Card, 0)
	for rows.Next() {
		var row CardRow
		if err := rows.StructScan(&row); err != nil {
			return nil, fmt.Errorf("mssql: list cards scan: %w", err)
		}
		card, err := RowToCard(&row)
		if err != nil {
			return nil, fmt.Errorf("mssql: list cards map: %w", err)
		}
		cards = append(cards, *card)
	}
	return cards, nil
}

func (r *CardRepository) GetByID(ctx context.Context, userID identityvo.UserID, id vos.CardID) (*entities.Card, error) {
	var row CardRow
	err := r.db.QueryRowxContext(ctx, getCardByID,
		sql.Named("userId", userID.String()),
		sql.Named("id", id.String()),
	).StructScan(&row)
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
