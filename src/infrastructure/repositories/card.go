package repositories

import (
	"database/sql"

	"github.com/jailtonjunior94/financialcontrol-api/src/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/database"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/queries"
)

type CardRepository struct {
	Db database.ISqlConnection
}

func NewCardRepository(db database.ISqlConnection) interfaces.ICardRepository {
	return &CardRepository{Db: db}
}

func (r *CardRepository) GetCards(userId string) (cards []entities.Card, err error) {
	connection := r.Db.Connect()

	rows, err := connection.Query(queries.GetCards, sql.Named("userId", userId))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var card entities.Card
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

func (r *CardRepository) GetCardById(id string, userId string) (card *entities.Card, err error) {
	connection := r.Db.Connect()
	row := connection.QueryRow(queries.GetCardById, sql.Named("id", id), sql.Named("userId", userId))

	c := new(entities.Card)
	err = row.Scan(&c.ID,
		&c.UserId,
		&c.FlagId,
		&c.Name,
		&c.Number,
		&c.Description,
		&c.ClosingDay,
		&c.ExpirationDate,
		&c.CreatedAt,
		&c.UpdatedAt,
		&c.Active,
		&c.Flag.ID,
		&c.Flag.Name,
		&c.Flag.Active)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return c, nil
}

func (r *CardRepository) AddCard(c *entities.Card) (card *entities.Card, err error) {
	s, err := r.Db.OpenConnectionAndMountStatement(queries.AddCard)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	result, err := s.Exec(
		sql.Named("id", c.ID),
		sql.Named("userId", c.UserId),
		sql.Named("flagId", c.FlagId),
		sql.Named("name", c.Name),
		sql.Named("number", c.Number),
		sql.Named("description", c.Description),
		sql.Named("closingDay", c.ClosingDay),
		sql.Named("expirationDate", c.ExpirationDate),
		sql.Named("createdAt", c.CreatedAt),
		sql.Named("updatedAt", c.UpdatedAt),
		sql.Named("active", c.Active))

	if err := r.Db.ValidateResult(result, err); err != nil {
		return nil, err
	}
	return c, nil
}

func (r *CardRepository) UpdateCard(c *entities.Card) (card *entities.Card, err error) {
	s, err := r.Db.OpenConnectionAndMountStatement(queries.UpdateCard)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	result, err := s.Exec(
		sql.Named("flagId", c.FlagId),
		sql.Named("name", c.Name),
		sql.Named("number", c.Number),
		sql.Named("description", c.Description),
		sql.Named("closingDay", c.ClosingDay),
		sql.Named("expirationDate", c.ExpirationDate),
		sql.Named("updatedAt", c.UpdatedAt),
		sql.Named("active", c.Active),
		sql.Named("id", c.ID))

	if err := r.Db.ValidateResult(result, err); err != nil {
		return nil, err
	}
	return c, nil
}
