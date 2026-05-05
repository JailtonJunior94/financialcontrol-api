package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application/usecase"
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain"
	ifacemocks "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/interfaces/mocks"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

type DeactivateCardSuite struct {
	suite.Suite
	ctx      context.Context
	cardRepo *ifacemocks.CardRepository
	sut      usecase.DeactivateCard
}

func TestDeactivateCardSuite(t *testing.T) { suite.Run(t, new(DeactivateCardSuite)) }

func (s *DeactivateCardSuite) SetupTest() {
	s.ctx = context.Background()
	s.cardRepo = ifacemocks.NewCardRepository(s.T())
	s.sut = usecase.NewDeactivateCard(s.cardRepo)
}

func (s *DeactivateCardSuite) TestExecute() {
	userID := identityvo.NewUserID()
	card := mustNewCard(userID)
	cardID := card.ID()

	type args struct {
		userID identityvo.UserID
		id     vos.CardID
	}
	scenarios := []struct {
		name   string
		args   args
		setup  func()
		expect func(err error)
	}{
		{
			name: "sucesso",
			args: args{userID: userID, id: cardID},
			setup: func() {
				s.cardRepo.EXPECT().GetByID(s.ctx, userID, cardID).Return(card, nil).Once()
				s.cardRepo.EXPECT().Update(s.ctx, card).Return(nil).Once()
			},
			expect: func(err error) {
				s.NoError(err)
			},
		},
		{
			name: "card inexistente (repo retorna ErrCardNotFound)",
			args: args{userID: userID, id: cardID},
			setup: func() {
				s.cardRepo.EXPECT().GetByID(s.ctx, userID, cardID).Return(nil, domain.ErrCardNotFound).Once()
			},
			expect: func(err error) {
				s.ErrorIs(err, domain.ErrCardNotFound)
			},
		},
		{
			name: "idempotencia: card ja inativo nao chama Update",
			args: args{userID: userID, id: cardID},
			setup: func() {
				inactiveCard := mustNewCard(userID)
				inactiveCard.Deactivate()
				s.Require().False(inactiveCard.Active())
				s.cardRepo.EXPECT().GetByID(s.ctx, userID, cardID).Return(inactiveCard, nil).Once()
				// Update NAO deve ser chamado: a expectativa ausente ja garante isso pelo mockery.
			},
			expect: func(err error) {
				s.NoError(err)
				// A ausencia de EXPECT().Update neste cenario combinada com o
				// modo strict do mockery garante que Update nao foi chamado.
			},
		},
		{
			name: "repo GetByID erro",
			args: args{userID: userID, id: cardID},
			setup: func() {
				s.cardRepo.EXPECT().GetByID(s.ctx, userID, cardID).Return(nil, errors.New("db error")).Once()
			},
			expect: func(err error) {
				s.Error(err)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			sc.setup()
			err := s.sut.Execute(s.ctx, sc.args.userID, sc.args.id)
			sc.expect(err)
		})
	}
}
