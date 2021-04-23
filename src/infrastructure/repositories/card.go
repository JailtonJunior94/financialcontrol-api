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
	return nil, nil
}

func (r *CardRepository) GetCardById(id string, userId string) (card *entities.Card, err error) {
	return nil, nil
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
