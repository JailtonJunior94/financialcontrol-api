package handlers_test

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/application/dtos"
	ucmocks "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/application/usecase/mocks"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/http/handlers"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identitycontext"
)

type SummaryHandlerSuite struct {
	suite.Suite
	summaryUC *ucmocks.MonthlySummary
	sut       *handlers.SummaryHandler
	app       *fiber.App
}

func TestSummaryHandlerSuite(t *testing.T) { suite.Run(t, new(SummaryHandlerSuite)) }

func (s *SummaryHandlerSuite) SetupTest() {
	s.summaryUC = ucmocks.NewMonthlySummary(s.T())
	s.sut = handlers.NewSummaryHandler(s.summaryUC)

	s.app = fiber.New()
	withIdentity := func(c *fiber.Ctx) error {
		c.SetUserContext(identitycontext.WithIdentity(c.UserContext(), identitycontext.Identity{
			UserID: testUserID,
			Email:  "user@example.com",
		}))
		return c.Next()
	}
	s.app.Get("/finance/summary", withIdentity, s.sut.Get)
}

func sampleSummaryResponse() dtos.MonthlySummaryResponse {
	return dtos.MonthlySummaryResponse{
		TotalIncome:               "5000.00",
		TotalExpense:              "3000.00",
		TotalRefundsIn:            "0.00",
		TotalRefundsOut:           "0.00",
		TotalCreditPurchasesMonth: "1500.00",
		TotalInvoicesOpen:         "1500.00",
		TotalInvoicesPaid:         "0.00",
		Balance:                   "500.00",
	}
}

func (s *SummaryHandlerSuite) TestGet() {
	resp := sampleSummaryResponse()

	s.Run("200 valid period", func() {
		s.summaryUC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(resp, nil).Once()
		req := httptest.NewRequest("GET", "/finance/summary?year=2026&month=5", nil)
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusOK, res.StatusCode)
	})

	s.Run("400 missing year", func() {
		req := httptest.NewRequest("GET", "/finance/summary?month=5", nil)
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusBadRequest, res.StatusCode)
	})

	s.Run("400 missing month", func() {
		req := httptest.NewRequest("GET", "/finance/summary?year=2026", nil)
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusBadRequest, res.StatusCode)
	})

	s.Run("400 invalid month 13", func() {
		req := httptest.NewRequest("GET", "/finance/summary?year=2026&month=13", nil)
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusBadRequest, res.StatusCode)
	})

	s.Run("400 year=0", func() {
		req := httptest.NewRequest("GET", "/finance/summary?year=0&month=5", nil)
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusBadRequest, res.StatusCode)
	})

	s.Run("500 use case error", func() {
		s.summaryUC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(dtos.MonthlySummaryResponse{}, errors.New("db")).Once()
		req := httptest.NewRequest("GET", "/finance/summary?year=2026&month=5", nil)
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusInternalServerError, res.StatusCode)
	})
}
