package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application/usecase"
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/entities"
	ifacemocks "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/interfaces/mocks"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

type CreateCardSuite struct {
	suite.Suite
	ctx      context.Context
	cardRepo *ifacemocks.CardRepository
	flagRepo *ifacemocks.FlagRepository
	sut      usecase.CreateCard
}

func TestCreateCardSuite(t *testing.T) { suite.Run(t, new(CreateCardSuite)) }

func (s *CreateCardSuite) SetupTest() {
	s.ctx = context.Background()
	s.cardRepo = ifacemocks.NewCardRepository(s.T())
	s.flagRepo = ifacemocks.NewFlagRepository(s.T())
	s.sut = usecase.NewCreateCard(s.cardRepo, s.flagRepo)
}

func (s *CreateCardSuite) TestExecute() {
	userID := identityvo.NewUserID()

	scenarios := []struct {
		name   string
		req    dtos.CardRequest
		setup  func()
		expect func(out dtos.CardResponse, err error)
	}{
		{
			name: "sucesso",
			req:  validCardRequest(),
			setup: func() {
				reloaded := attachFlag(mustNewCard(userID), testFlagID, "Visa")
				s.cardRepo.EXPECT().Add(s.ctx, mock.MatchedBy(func(card *entities.Card) bool {
					return card.ExpirationDate().Equal(validCardRequest().ExpirationDate)
				})).Return(nil).Once()
				s.cardRepo.EXPECT().GetByID(s.ctx, userID, mock.Anything).Return(reloaded, nil).Once()
			},
			expect: func(out dtos.CardResponse, err error) {
				s.NoError(err)
				s.NotEmpty(out.ID)
				s.Equal(testFlagID.String(), out.Flag.ID)
				s.Equal("Visa", out.Flag.Name)
			},
		},
		{
			name: "sucesso com closingDay igual ao dia de expirationDate preserva contrato legado",
			req: func() dtos.CardRequest {
				req := validCardRequest()
				req.ClosingDay = 10
				req.ExpirationDate = validCardRequest().ExpirationDate
				return req
			}(),
			setup: func() {
				reloaded := attachFlag(mustLegacyCard(userID), testFlagID, "Visa")
				s.cardRepo.EXPECT().Add(s.ctx, mock.MatchedBy(func(card *entities.Card) bool {
					return card.ClosingDay().Int() == 10 && card.DueDay().Int() == 10
				})).Return(nil).Once()
				s.cardRepo.EXPECT().GetByID(s.ctx, userID, mock.Anything).Return(reloaded, nil).Once()
			},
			expect: func(out dtos.CardResponse, err error) {
				s.NoError(err)
				s.Equal(10, out.ClosingDay)
			},
		},
		{
			name:  "request inválido",
			req:   dtos.CardRequest{Name: "", ClosingDay: 0},
			setup: func() {},
			expect: func(out dtos.CardResponse, err error) {
				s.Error(err)
			},
		},
		{
			name: "flagId malformado é rejeitado antes de tocar o repositorio",
			req: func() dtos.CardRequest {
				r := validCardRequest()
				r.FlagID = "not-a-uuid"
				return r
			}(),
			setup: func() {},
			expect: func(out dtos.CardResponse, err error) {
				s.ErrorIs(err, domain.ErrInvalidFlagID)
				s.Empty(out.ID)
			},
		},
		{
			name: "repo erro",
			req:  validCardRequest(),
			setup: func() {
				s.cardRepo.EXPECT().Add(s.ctx, mock.Anything).Return(errors.New("db error")).Once()
			},
			expect: func(out dtos.CardResponse, err error) {
				s.Error(err)
			},
		},
		{
			name: "ErrFlagNotFound propagado do repo",
			req:  validCardRequest(),
			setup: func() {
				s.cardRepo.EXPECT().Add(s.ctx, mock.Anything).Return(domain.ErrFlagNotFound).Once()
			},
			expect: func(out dtos.CardResponse, err error) {
				s.ErrorIs(err, domain.ErrFlagNotFound)
			},
		},
		{
			name: "erro ao recarregar cartao criado",
			req:  validCardRequest(),
			setup: func() {
				s.cardRepo.EXPECT().Add(s.ctx, mock.Anything).Return(nil).Once()
				s.cardRepo.EXPECT().GetByID(s.ctx, userID, mock.Anything).Return(nil, errors.New("reload error")).Once()
			},
			expect: func(out dtos.CardResponse, err error) {
				s.Error(err)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			sc.setup()
			out, err := s.sut.Execute(s.ctx, userID, sc.req)
			sc.expect(out, err)
		})
	}
}
