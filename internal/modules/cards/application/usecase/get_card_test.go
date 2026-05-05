package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application/usecase"
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain"
	ifacemocks "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/interfaces/mocks"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

type GetCardSuite struct {
	suite.Suite
	ctx      context.Context
	cardRepo *ifacemocks.CardRepository
	sut      usecase.GetCard
}

func TestGetCardSuite(t *testing.T) { suite.Run(t, new(GetCardSuite)) }

func (s *GetCardSuite) SetupTest() {
	s.ctx = context.Background()
	s.cardRepo = ifacemocks.NewCardRepository(s.T())
	s.sut = usecase.NewGetCard(s.cardRepo)
}

func (s *GetCardSuite) TestExecute() {
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
		expect func(out dtos.CardResponse, err error)
	}{
		{
			name: "sucesso",
			args: args{userID: userID, id: cardID},
			setup: func() {
				s.cardRepo.EXPECT().GetByID(s.ctx, userID, cardID).Return(card, nil).Once()
			},
			expect: func(out dtos.CardResponse, err error) {
				s.NoError(err)
				s.Equal(cardID.String(), out.ID)
			},
		},
		{
			name: "ErrCardNotFound propagado pelo repo",
			args: args{userID: userID, id: cardID},
			setup: func() {
				s.cardRepo.EXPECT().GetByID(s.ctx, userID, cardID).Return(nil, domain.ErrCardNotFound).Once()
			},
			expect: func(out dtos.CardResponse, err error) {
				s.ErrorIs(err, domain.ErrCardNotFound)
			},
		},
		{
			name: "repo erro generico",
			args: args{userID: userID, id: cardID},
			setup: func() {
				s.cardRepo.EXPECT().GetByID(s.ctx, userID, cardID).Return(nil, errors.New("db error")).Once()
			},
			expect: func(out dtos.CardResponse, err error) {
				s.Error(err)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			sc.setup()
			out, err := s.sut.Execute(s.ctx, sc.args.userID, sc.args.id)
			sc.expect(out, err)
		})
	}
}
