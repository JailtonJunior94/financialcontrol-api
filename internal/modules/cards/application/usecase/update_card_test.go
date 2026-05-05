package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application/usecase"
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/entities"
	ifacemocks "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/interfaces/mocks"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

type UpdateCardSuite struct {
	suite.Suite
	ctx      context.Context
	cardRepo *ifacemocks.CardRepository
	flagRepo *ifacemocks.FlagRepository
	sut      usecase.UpdateCard
}

func TestUpdateCardSuite(t *testing.T) { suite.Run(t, new(UpdateCardSuite)) }

func (s *UpdateCardSuite) SetupTest() {
	s.ctx = context.Background()
	s.cardRepo = ifacemocks.NewCardRepository(s.T())
	s.flagRepo = ifacemocks.NewFlagRepository(s.T())
	s.sut = usecase.NewUpdateCard(s.cardRepo, s.flagRepo)
}

func (s *UpdateCardSuite) TestExecute() {
	userID := identityvo.NewUserID()
	card := attachFlag(mustNewCard(userID), testFlagID, "Visa")
	cardID := card.ID()
	legacyCard := attachFlag(mustLegacyCard(userID), testFlagID, "Visa")
	legacyCardID := legacyCard.ID()
	updatedReq := validCardRequestWithFlag(alternateFlagID)

	type args struct {
		id  vos.CardID
		req dtos.CardRequest
	}
	scenarios := []struct {
		name   string
		args   args
		setup  func()
		expect func(out dtos.CardResponse, err error)
	}{
		{
			name: "sucesso",
			args: args{id: cardID, req: updatedReq},
			setup: func() {
				reloaded := attachFlag(mustRehydrateCardWithID(userID, cardID), alternateFlagID, "Mastercard")
				s.Require().NoError(reloaded.Update(alternateFlagID, updatedReq.Name, updatedReq.Number, updatedReq.Description, updatedReq.ClosingDay, updatedReq.ExpirationDate))
				s.cardRepo.EXPECT().GetByID(s.ctx, userID, cardID).Return(card, nil).Once()
				s.cardRepo.EXPECT().Update(s.ctx, mock.MatchedBy(func(updated *entities.Card) bool {
					return updated.ExpirationDate().Equal(updatedReq.ExpirationDate)
				})).Return(nil).Once()
				s.cardRepo.EXPECT().GetByID(s.ctx, userID, cardID).Return(reloaded, nil).Once()
			},
			expect: func(out dtos.CardResponse, err error) {
				s.NoError(err)
				s.Equal(cardID.String(), out.ID)
				s.Equal(alternateFlagID.String(), out.Flag.ID)
				s.Equal("Mastercard", out.Flag.Name)
			},
		},
		{
			name: "cartao legado com closing igual due continua atualizavel quando ciclo nao muda",
			args: args{
				id: legacyCardID,
				req: dtos.CardRequest{
					FlagID:         testFlagID.String(),
					Name:           "Cartão Legado Atualizado",
					Number:         "4111111111111111",
					Description:    "desc atualizado",
					ClosingDay:     10,
					ExpirationDate: time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC),
				},
			},
			setup: func() {
				reloaded := attachFlag(mustLegacyCardWithID(userID, legacyCardID), testFlagID, "Visa")
				s.Require().NoError(reloaded.Update(testFlagID, "Cartão Legado Atualizado", "4111111111111111", "desc atualizado", 10, time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)))
				s.cardRepo.EXPECT().GetByID(s.ctx, userID, legacyCardID).Return(legacyCard, nil).Once()
				s.cardRepo.EXPECT().Update(s.ctx, mock.MatchedBy(func(updated *entities.Card) bool {
					return updated.ClosingDay().Int() == 10 &&
						updated.DueDay().Int() == 10 &&
						updated.Name().String() == "Cartão Legado Atualizado"
				})).Return(nil).Once()
				s.cardRepo.EXPECT().GetByID(s.ctx, userID, legacyCardID).Return(reloaded, nil).Once()
			},
			expect: func(out dtos.CardResponse, err error) {
				s.NoError(err)
				s.Equal("Cartão Legado Atualizado", out.Name)
				s.Equal(10, out.ClosingDay)
			},
		},
		{
			name: "cartao moderno continua atualizavel para closing igual due para preservar contrato publico",
			args: args{
				id: cardID,
				req: dtos.CardRequest{
					FlagID:         testFlagID.String(),
					Name:           "Cartão Mesmo Dia",
					Number:         "4111111111111111",
					Description:    "desc mesmo dia",
					ClosingDay:     10,
					ExpirationDate: time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC),
				},
			},
			setup: func() {
				reloaded := attachFlag(mustRehydrateCardWithID(userID, cardID), testFlagID, "Visa")
				s.Require().NoError(reloaded.Update(testFlagID, "Cartão Mesmo Dia", "4111111111111111", "desc mesmo dia", 10, time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)))
				s.cardRepo.EXPECT().GetByID(s.ctx, userID, cardID).Return(card, nil).Once()
				s.cardRepo.EXPECT().Update(s.ctx, mock.MatchedBy(func(updated *entities.Card) bool {
					return updated.ClosingDay().Int() == 10 &&
						updated.DueDay().Int() == 10 &&
						updated.Name().String() == "Cartão Mesmo Dia"
				})).Return(nil).Once()
				s.cardRepo.EXPECT().GetByID(s.ctx, userID, cardID).Return(reloaded, nil).Once()
			},
			expect: func(out dtos.CardResponse, err error) {
				s.NoError(err)
				s.Equal("Cartão Mesmo Dia", out.Name)
				s.Equal(10, out.ClosingDay)
			},
		},
		{
			name:  "request inválido",
			args:  args{id: cardID, req: dtos.CardRequest{}},
			setup: func() {},
			expect: func(out dtos.CardResponse, err error) {
				s.Error(err)
			},
		},
		{
			name: "card inexistente (repo retorna ErrCardNotFound)",
			args: args{id: cardID, req: validCardRequest()},
			setup: func() {
				s.cardRepo.EXPECT().GetByID(s.ctx, userID, cardID).Return(nil, domain.ErrCardNotFound).Once()
			},
			expect: func(out dtos.CardResponse, err error) {
				s.ErrorIs(err, domain.ErrCardNotFound)
			},
		},
		{
			name: "flag inexistente via repo",
			args: args{id: cardID, req: validCardRequest()},
			setup: func() {
				s.cardRepo.EXPECT().GetByID(s.ctx, userID, cardID).Return(card, nil).Once()
				s.cardRepo.EXPECT().Update(s.ctx, mock.Anything).Return(domain.ErrFlagNotFound).Once()
			},
			expect: func(out dtos.CardResponse, err error) {
				s.ErrorIs(err, domain.ErrFlagNotFound)
			},
		},
		{
			name: "repo GetByID erro",
			args: args{id: cardID, req: validCardRequest()},
			setup: func() {
				s.cardRepo.EXPECT().GetByID(s.ctx, userID, cardID).Return(nil, errors.New("db error")).Once()
			},
			expect: func(out dtos.CardResponse, err error) {
				s.Error(err)
			},
		},
		{
			name: "erro ao recarregar cartao atualizado",
			args: args{id: cardID, req: updatedReq},
			setup: func() {
				s.cardRepo.EXPECT().GetByID(s.ctx, userID, cardID).Return(card, nil).Once()
				s.cardRepo.EXPECT().Update(s.ctx, mock.Anything).Return(nil).Once()
				s.cardRepo.EXPECT().GetByID(s.ctx, userID, cardID).Return(nil, errors.New("reload error")).Once()
			},
			expect: func(out dtos.CardResponse, err error) {
				s.Error(err)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			sc.setup()
			out, err := s.sut.Execute(s.ctx, userID, sc.args.id, sc.args.req)
			sc.expect(out, err)
		})
	}
}
