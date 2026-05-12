package handlers_test

import (
	"bytes"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/application/dtos"
	ucmocks "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/application/usecase/mocks"
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/http/handlers"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identitycontext"
)

type InstallmentHandlerSuite struct {
	suite.Suite
	anticipateUC *ucmocks.AnticipateInstallment
	sut          *handlers.InstallmentHandler
	app          *fiber.App
}

func TestInstallmentHandlerSuite(t *testing.T) { suite.Run(t, new(InstallmentHandlerSuite)) }

func (s *InstallmentHandlerSuite) SetupTest() {
	s.anticipateUC = ucmocks.NewAnticipateInstallment(s.T())
	s.sut = handlers.NewInstallmentHandler(s.anticipateUC)

	s.app = fiber.New()
	withIdentity := func(c *fiber.Ctx) error {
		c.SetUserContext(identitycontext.WithIdentity(c.UserContext(), identitycontext.Identity{
			UserID: testUserID,
			Email:  "user@example.com",
		}))
		return c.Next()
	}
	s.app.Post("/finance/installments/:id/anticipate", withIdentity, s.sut.Anticipate)
}

func (s *InstallmentHandlerSuite) TestAnticipate() {
	resp := sampleTxResponse()
	validBody := `{"transaction_id":"` + testTxID + `"}`

	s.Run("200 anticipated", func() {
		s.anticipateUC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(resp, nil).Once()
		req := httptest.NewRequest("POST", "/finance/installments/"+testInstallID+"/anticipate", bytes.NewBufferString(validBody))
		req.Header.Set("Content-Type", "application/json")
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusOK, res.StatusCode)
	})

	s.Run("400 invalid installment id", func() {
		req := httptest.NewRequest("POST", "/finance/installments/not-uuid/anticipate", bytes.NewBufferString(validBody))
		req.Header.Set("Content-Type", "application/json")
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusBadRequest, res.StatusCode)
	})

	s.Run("422 missing transaction_id", func() {
		req := httptest.NewRequest("POST", "/finance/installments/"+testInstallID+"/anticipate", bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusUnprocessableEntity, res.StatusCode)
	})

	s.Run("422 invalid body", func() {
		req := httptest.NewRequest("POST", "/finance/installments/"+testInstallID+"/anticipate", bytes.NewBufferString("not-json"))
		req.Header.Set("Content-Type", "application/json")
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusUnprocessableEntity, res.StatusCode)
	})

	s.Run("400 invalid transaction_id format", func() {
		req := httptest.NewRequest("POST", "/finance/installments/"+testInstallID+"/anticipate",
			bytes.NewBufferString(`{"transaction_id":"not-uuid"}`))
		req.Header.Set("Content-Type", "application/json")
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusBadRequest, res.StatusCode)
	})

	s.Run("404 installment not found", func() {
		s.anticipateUC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(dtos.TransactionResponse{}, domain.ErrInstallmentNotFound).Once()
		req := httptest.NewRequest("POST", "/finance/installments/"+testInstallID+"/anticipate", bytes.NewBufferString(validBody))
		req.Header.Set("Content-Type", "application/json")
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusNotFound, res.StatusCode)
	})

	s.Run("409 installment in closed invoice", func() {
		s.anticipateUC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(dtos.TransactionResponse{}, domain.ErrInstallmentInClosedOrPaidInvoice).Once()
		req := httptest.NewRequest("POST", "/finance/installments/"+testInstallID+"/anticipate", bytes.NewBufferString(validBody))
		req.Header.Set("Content-Type", "application/json")
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusConflict, res.StatusCode)
	})

	s.Run("500 unexpected error", func() {
		s.anticipateUC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(dtos.TransactionResponse{}, errors.New("db")).Once()
		req := httptest.NewRequest("POST", "/finance/installments/"+testInstallID+"/anticipate", bytes.NewBufferString(validBody))
		req.Header.Set("Content-Type", "application/json")
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusInternalServerError, res.StatusCode)
	})
}
