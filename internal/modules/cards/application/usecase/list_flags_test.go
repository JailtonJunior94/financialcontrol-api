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
)

type ListFlagsSuite struct {
	suite.Suite
	ctx      context.Context
	flagRepo *ifacemocks.FlagRepository
	sut      usecase.ListFlags
}

func TestListFlagsSuite(t *testing.T) { suite.Run(t, new(ListFlagsSuite)) }

func (s *ListFlagsSuite) SetupTest() {
	s.ctx = context.Background()
	s.flagRepo = ifacemocks.NewFlagRepository(s.T())
	s.sut = usecase.NewListFlags(s.flagRepo)
}

func (s *ListFlagsSuite) TestExecute() {
	flag := entities.NewFlag(testFlagID, "Visa", true)

	scenarios := []struct {
		name   string
		setup  func()
		expect func(out []dtos.FlagResponse, err error)
	}{
		{
			name: "sucesso sempre retorna lista completa",
			setup: func() {
				s.flagRepo.EXPECT().List(s.ctx).Return([]entities.Flag{flag}, nil).Once()
			},
			expect: func(out []dtos.FlagResponse, err error) {
				s.NoError(err)
				s.Len(out, 1)
				s.Equal("Visa", out[0].Name)
			},
		},
		{
			name: "repo erro",
			setup: func() {
				s.flagRepo.EXPECT().List(s.ctx).Return(nil, errors.New("db error")).Once()
			},
			expect: func(out []dtos.FlagResponse, err error) {
				s.Error(err)
				s.Nil(out)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			sc.setup()
			out, err := s.sut.Execute(s.ctx)
			sc.expect(out, err)
		})
	}
}
