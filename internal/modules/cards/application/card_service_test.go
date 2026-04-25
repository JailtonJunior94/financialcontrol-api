package application_test

import (
	"errors"
	"testing"
	"time"

	appresponses "github.com/jailtonjunior94/financialcontrol-api/pkg/web"
	cardsapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application"
	cardsdomain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain"

	"github.com/stretchr/testify/require"
)

type cardRepositoryStub struct {
	cards       []cardsdomain.Card
	cardByID    *cardsdomain.Card
	addedCard   *cardsdomain.Card
	updatedCard *cardsdomain.Card
	err         error
}

func (s *cardRepositoryStub) GetCards(userID string) ([]cardsdomain.Card, error) {
	return s.cards, s.err
}

func (s *cardRepositoryStub) GetCardById(id, userID string) (*cardsdomain.Card, error) {
	return s.cardByID, s.err
}

func (s *cardRepositoryStub) AddCard(card *cardsdomain.Card) (*cardsdomain.Card, error) {
	s.addedCard = card
	if s.err != nil {
		return nil, s.err
	}

	return card, nil
}

func (s *cardRepositoryStub) UpdateCard(card *cardsdomain.Card) (*cardsdomain.Card, error) {
	s.updatedCard = card
	if s.err != nil {
		return nil, s.err
	}

	return card, nil
}

func TestCreateCardMapsRequestIntoRepositoryEntity(t *testing.T) {
	service := cardsapp.NewCardService(&cardRepositoryStub{})

	response := service.CreateCard("user-id", &cardsapp.CardRequest{
		FlagID:         "flag-id",
		Name:           "Cartao XP",
		Number:         "1234",
		Description:    "Principal",
		ClosingDay:     10,
		ExpirationDate: time.Date(2027, time.January, 1, 0, 0, 0, 0, time.UTC),
	})

	require.Equal(t, appresponses.Created(nil).StatusCode, response.StatusCode)
}

func TestCardByIdReturnsNotFoundWhenRepositoryReturnsNil(t *testing.T) {
	service := cardsapp.NewCardService(&cardRepositoryStub{})

	response := service.CardById("card-id", "user-id")

	require.Equal(t, appresponses.NotFound("").StatusCode, response.StatusCode)
}

func TestRemoveCardMarksEntityAsInactive(t *testing.T) {
	repository := &cardRepositoryStub{
		cardByID: &cardsdomain.Card{
			Entity: cardsdomain.Entity{ID: "card-id", Active: true},
			Name:   "Cartao XP",
		},
	}
	service := cardsapp.NewCardService(repository)

	response := service.RemoveCard("card-id", "user-id")

	require.Equal(t, appresponses.NoContent().StatusCode, response.StatusCode)
	require.NotNil(t, repository.updatedCard)
	require.False(t, repository.updatedCard.Active)
}

func TestCardsReturnsServerErrorWhenRepositoryFails(t *testing.T) {
	service := cardsapp.NewCardService(&cardRepositoryStub{err: errors.New("db error")})

	response := service.Cards("user-id")

	require.Equal(t, appresponses.ServerError().StatusCode, response.StatusCode)
}
