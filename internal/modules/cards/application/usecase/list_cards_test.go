package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application/usecase"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/entities"
	ifacemocks "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/interfaces/mocks"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

type ListCardsSuite struct {
	suite.Suite
	ctx      context.Context
	cardRepo *ifacemocks.CardRepository
	sut      usecase.ListCards
}

func TestListCardsSuite(t *testing.T) { suite.Run(t, new(ListCardsSuite)) }

func (s *ListCardsSuite) SetupTest() {
	s.ctx = context.Background()
	s.cardRepo = ifacemocks.NewCardRepository(s.T())
	s.sut = usecase.NewListCards(s.cardRepo)
}

func (s *ListCardsSuite) TestExecute() {
	userID := identityvo.NewUserID()
	card := mustNewCard(userID)

	unpaged := dtos.Pagination{}
	clampedPagination := dtos.NewPagination(1, 1000)

	scenarios := []struct {
		name       string
		pagination dtos.Pagination
		setup      func()
		expect     func(out []dtos.CardResponse, err error)
	}{
		{
			name:       "sucesso sem paginacao preserva contrato legado",
			pagination: unpaged,
			setup: func() {
				s.cardRepo.EXPECT().List(s.ctx, userID, unpaged).Return([]entities.Card{*card}, nil).Once()
			},
			expect: func(out []dtos.CardResponse, err error) {
				s.NoError(err)
				s.Len(out, 1)
			},
		},
		{
			name:       "paginacao zero nao e normalizada para truncar a lista",
			pagination: dtos.Pagination{Page: 0, Size: 0},
			setup: func() {
				s.cardRepo.EXPECT().List(s.ctx, userID, unpaged).Return([]entities.Card{*card}, nil).Once()
			},
			expect: func(out []dtos.CardResponse, err error) {
				s.NoError(err)
				s.Len(out, 1)
			},
		},
		{
			name:       "size acima do max e clampado",
			pagination: dtos.Pagination{Page: 1, Size: 1000},
			setup: func() {
				s.cardRepo.EXPECT().List(s.ctx, userID, clampedPagination).Return([]entities.Card{*card}, nil).Once()
				s.Equal(200, clampedPagination.Size)
			},
			expect: func(out []dtos.CardResponse, err error) {
				s.NoError(err)
				s.Len(out, 1)
			},
		},
		{
			name:       "repo retorna erro",
			pagination: unpaged,
			setup: func() {
				s.cardRepo.EXPECT().List(s.ctx, userID, unpaged).Return(nil, errors.New("db error")).Once()
			},
			expect: func(out []dtos.CardResponse, err error) {
				s.Error(err)
				s.Nil(out)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			sc.setup()
			out, err := s.sut.Execute(s.ctx, userID, sc.pagination)
			sc.expect(out, err)
		})
	}
}
