package usecase_test

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

var testFlagID = mustParseFlagID("550e8400-e29b-41d4-a716-446655440000")
var alternateFlagID = mustParseFlagID("660e8400-e29b-41d4-a716-446655440000")

func mustParseFlagID(s string) vos.FlagID {
	id, err := vos.ParseFlagID(s)
	if err != nil {
		panic(err)
	}
	return id
}

func mustNewCard(userID identityvo.UserID) *entities.Card {
	card, err := entities.NewCard(userID, testFlagID, "Meu Cartão", "4111111111111111", "desc", 15, time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC))
	if err != nil {
		panic(err)
	}
	return card
}

func mustRehydrateCardWithID(userID identityvo.UserID, id vos.CardID) *entities.Card {
	card, err := entities.RehydrateCard(
		id,
		userID,
		testFlagID,
		"Meu Cartão",
		"4111111111111111",
		"desc",
		15,
		time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		true,
	)
	if err != nil {
		panic(err)
	}
	return card
}

func mustLegacyCardWithID(userID identityvo.UserID, id vos.CardID) *entities.Card {
	card, err := entities.RehydrateCard(
		id,
		userID,
		testFlagID,
		"Meu Cartão Legado",
		"4111111111111111",
		"desc",
		10,
		time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		true,
	)
	if err != nil {
		panic(err)
	}
	return card
}

func mustLegacyCard(userID identityvo.UserID) *entities.Card {
	card, err := entities.RehydrateCard(
		vos.NewCardID(),
		userID,
		testFlagID,
		"Meu Cartão Legado",
		"4111111111111111",
		"desc",
		10,
		time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		true,
	)
	if err != nil {
		panic(err)
	}
	return card
}

func validCardRequest() dtos.CardRequest {
	return dtos.CardRequest{
		FlagID:         testFlagID.String(),
		Name:           "Meu Cartão",
		Number:         "4111111111111111",
		Description:    "desc",
		ClosingDay:     15,
		ExpirationDate: time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC),
	}
}

func validCardRequestWithFlag(flagID vos.FlagID) dtos.CardRequest {
	req := validCardRequest()
	req.FlagID = flagID.String()
	return req
}

func attachFlag(card *entities.Card, flagID vos.FlagID, name string) *entities.Card {
	flag := entities.NewFlag(flagID, name, true)
	card.AttachFlag(&flag)
	return card
}
