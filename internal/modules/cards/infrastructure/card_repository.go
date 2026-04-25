package infrastructure

import (
	"database/sql"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/database"
)

type CardRepository struct {
	db database.ISqlConnection
}

func NewCardRepository(db database.ISqlConnection) *CardRepository {
	return &CardRepository{db: db}
}

func (r *CardRepository) GetCards(userID string) ([]domain.Card, error) {
	connection := r.db.Connect()
	rows, err := connection.Query(getCards, sql.Named("userId", userID))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	cards := make([]domain.Card, 0)
	for rows.Next() {
		var card domain.Card
		if err := rows.Scan(
			&card.ID,
			&card.UserId,
			&card.FlagId,
			&card.Name,
			&card.Number,
			&card.Description,
			&card.ClosingDay,
			&card.ExpirationDate,
			&card.CreatedAt,
			&card.UpdatedAt,
			&card.Active,
			&card.Flag.ID,
			&card.Flag.Name,
			&card.Flag.Active,
		); err != nil {
			return nil, err
		}

		cards = append(cards, card)
	}

	return cards, nil
}

func (r *CardRepository) GetCardById(id, userID string) (*domain.Card, error) {
	connection := r.db.Connect()
	row := connection.QueryRow(getCardByID, sql.Named("id", id), sql.Named("userId", userID))

	card := new(domain.Card)
	err := row.Scan(
		&card.ID,
		&card.UserId,
		&card.FlagId,
		&card.Name,
		&card.Number,
		&card.Description,
		&card.ClosingDay,
		&card.ExpirationDate,
		&card.CreatedAt,
		&card.UpdatedAt,
		&card.Active,
		&card.Flag.ID,
		&card.Flag.Name,
		&card.Flag.Active,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return card, nil
}

func (r *CardRepository) AddCard(card *domain.Card) (*domain.Card, error) {
	statement, err := r.db.OpenConnectionAndMountStatement(addCard)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = statement.Close()
	}()

	result, err := statement.Exec(
		sql.Named("id", card.ID),
		sql.Named("userId", card.UserId),
		sql.Named("flagId", card.FlagId),
		sql.Named("name", card.Name),
		sql.Named("number", card.Number),
		sql.Named("description", card.Description),
		sql.Named("closingDay", card.ClosingDay),
		sql.Named("expirationDate", card.ExpirationDate),
		sql.Named("createdAt", card.CreatedAt),
		sql.Named("updatedAt", card.UpdatedAt),
		sql.Named("active", card.Active),
	)
	if err := r.db.ValidateResult(result, err); err != nil {
		return nil, err
	}

	return card, nil
}

func (r *CardRepository) UpdateCard(card *domain.Card) (*domain.Card, error) {
	statement, err := r.db.OpenConnectionAndMountStatement(updateCard)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = statement.Close()
	}()

	result, err := statement.Exec(
		sql.Named("flagId", card.FlagId),
		sql.Named("name", card.Name),
		sql.Named("number", card.Number),
		sql.Named("description", card.Description),
		sql.Named("closingDay", card.ClosingDay),
		sql.Named("expirationDate", card.ExpirationDate),
		sql.Named("updatedAt", card.UpdatedAt),
		sql.Named("active", card.Active),
		sql.Named("id", card.ID),
	)
	if err := r.db.ValidateResult(result, err); err != nil {
		return nil, err
	}

	return card, nil
}
