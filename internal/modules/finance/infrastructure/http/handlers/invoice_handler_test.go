package handlers_test

import (
	"bytes"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/application/dtos"
	ucmocks "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/application/usecase/mocks"
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/http/handlers"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identitycontext"
)

type InvoiceHandlerSuite struct {
	suite.Suite
	listUC *ucmocks.ListInvoices
	getUC  *ucmocks.GetInvoice
	payUC  *ucmocks.PayInvoice
	sut    *handlers.InvoiceHandler
	app    *fiber.App
}

func TestInvoiceHandlerSuite(t *testing.T) { suite.Run(t, new(InvoiceHandlerSuite)) }

func (s *InvoiceHandlerSuite) SetupTest() {
	s.listUC = ucmocks.NewListInvoices(s.T())
	s.getUC = ucmocks.NewGetInvoice(s.T())
	s.payUC = ucmocks.NewPayInvoice(s.T())
	s.sut = handlers.NewInvoiceHandler(s.listUC, s.getUC, s.payUC)

	s.app = fiber.New()
	withIdentity := func(c *fiber.Ctx) error {
		c.SetUserContext(identitycontext.WithIdentity(c.UserContext(), identitycontext.Identity{
			UserID: testUserID,
			Email:  "user@example.com",
		}))
		return c.Next()
	}
	s.app.Get("/finance/invoices", withIdentity, s.sut.List)
	s.app.Get("/finance/invoices/:id", withIdentity, s.sut.Get)
	s.app.Patch("/finance/invoices/:id/pay", withIdentity, s.sut.Pay)
}

func sampleInvoiceResponse() dtos.InvoiceResponse {
	return dtos.InvoiceResponse{
		ID:          testInvoiceID,
		UserID:      testUserID,
		CardID:      testCardID,
		State:       "open",
		CycleStart:  time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		CycleEnd:    time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC),
		ClosingDate: time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC),
		DueDate:     time.Date(2026, 6, 5, 0, 0, 0, 0, time.UTC),
		Total:       "500.00",
		CreatedAt:   time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
	}
}

func (s *InvoiceHandlerSuite) TestList() {
	page := dtos.PaginatedResponse[dtos.InvoiceResponse]{
		Items: []dtos.InvoiceResponse{sampleInvoiceResponse()}, Total: 1, Page: 1, PageSize: 20,
	}

	s.Run("200 default", func() {
		s.listUC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(page, nil).Once()
		req := httptest.NewRequest("GET", "/finance/invoices", nil)
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusOK, res.StatusCode)
	})

	s.Run("200 with card_id filter", func() {
		s.listUC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(page, nil).Once()
		req := httptest.NewRequest("GET", "/finance/invoices?card_id="+testCardID, nil)
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusOK, res.StatusCode)
	})

	s.Run("200 with state filter", func() {
		s.listUC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(page, nil).Once()
		req := httptest.NewRequest("GET", "/finance/invoices?state=open", nil)
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusOK, res.StatusCode)
	})

	s.Run("400 invalid card_id", func() {
		req := httptest.NewRequest("GET", "/finance/invoices?card_id=not-uuid", nil)
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusBadRequest, res.StatusCode)
	})

	s.Run("400 invalid from date", func() {
		req := httptest.NewRequest("GET", "/finance/invoices?from=not-a-date", nil)
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusBadRequest, res.StatusCode)
	})

	s.Run("400 invalid state", func() {
		req := httptest.NewRequest("GET", "/finance/invoices?state=invalid", nil)
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusBadRequest, res.StatusCode)
	})

	s.Run("500 use case error", func() {
		s.listUC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(dtos.PaginatedResponse[dtos.InvoiceResponse]{}, errors.New("db")).Once()
		req := httptest.NewRequest("GET", "/finance/invoices", nil)
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusInternalServerError, res.StatusCode)
	})
}

func (s *InvoiceHandlerSuite) TestGet() {
	resp := sampleInvoiceResponse()

	detail := dtos.InvoiceDetailResponse{InvoiceResponse: resp}

	s.Run("200 found", func() {
		s.getUC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(detail, nil).Once()
		req := httptest.NewRequest("GET", "/finance/invoices/"+testInvoiceID, nil)
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusOK, res.StatusCode)
	})

	s.Run("400 invalid id", func() {
		req := httptest.NewRequest("GET", "/finance/invoices/not-uuid", nil)
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusBadRequest, res.StatusCode)
	})

	s.Run("404 not found", func() {
		s.getUC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(dtos.InvoiceDetailResponse{}, domain.ErrInvoiceNotFound).Once()
		req := httptest.NewRequest("GET", "/finance/invoices/"+testInvoiceID, nil)
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusNotFound, res.StatusCode)
	})
}

func (s *InvoiceHandlerSuite) TestPay() {
	resp := sampleInvoiceResponse()

	s.Run("200 paid without settlement", func() {
		s.payUC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(resp, nil).Once()
		req := httptest.NewRequest("PATCH", "/finance/invoices/"+testInvoiceID+"/pay",
			bytes.NewBufferString(`{"generate_settlement_transaction":false}`))
		req.Header.Set("Content-Type", "application/json")
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusOK, res.StatusCode)
	})

	s.Run("200 paid with settlement", func() {
		s.payUC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(resp, nil).Once()
		body := `{"generate_settlement_transaction":true,"payment_method":"pix","settlement_category_id":"` + testCategoryID + `"}`
		req := httptest.NewRequest("PATCH", "/finance/invoices/"+testInvoiceID+"/pay", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusOK, res.StatusCode)
	})

	s.Run("400 invalid id", func() {
		req := httptest.NewRequest("PATCH", "/finance/invoices/not-uuid/pay",
			bytes.NewBufferString(`{"generate_settlement_transaction":false}`))
		req.Header.Set("Content-Type", "application/json")
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusBadRequest, res.StatusCode)
	})

	s.Run("400 generate_settlement_transaction=true without payment_method", func() {
		req := httptest.NewRequest("PATCH", "/finance/invoices/"+testInvoiceID+"/pay",
			bytes.NewBufferString(`{"generate_settlement_transaction":true}`))
		req.Header.Set("Content-Type", "application/json")
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusBadRequest, res.StatusCode)
	})

	s.Run("422 invalid body", func() {
		req := httptest.NewRequest("PATCH", "/finance/invoices/"+testInvoiceID+"/pay",
			bytes.NewBufferString("not-json"))
		req.Header.Set("Content-Type", "application/json")
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusUnprocessableEntity, res.StatusCode)
	})

	s.Run("409 already paid", func() {
		s.payUC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(dtos.InvoiceResponse{}, domain.ErrInvoiceAlreadyPaid).Once()
		req := httptest.NewRequest("PATCH", "/finance/invoices/"+testInvoiceID+"/pay",
			bytes.NewBufferString(`{"generate_settlement_transaction":false}`))
		req.Header.Set("Content-Type", "application/json")
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusConflict, res.StatusCode)
	})

	s.Run("422 cannot pay", func() {
		s.payUC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(dtos.InvoiceResponse{}, domain.ErrInvoiceCannotPay).Once()
		req := httptest.NewRequest("PATCH", "/finance/invoices/"+testInvoiceID+"/pay",
			bytes.NewBufferString(`{"generate_settlement_transaction":false}`))
		req.Header.Set("Content-Type", "application/json")
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusUnprocessableEntity, res.StatusCode)
	})

	s.Run("404 invoice not found", func() {
		s.payUC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(dtos.InvoiceResponse{}, domain.ErrInvoiceNotFound).Once()
		req := httptest.NewRequest("PATCH", "/finance/invoices/"+testInvoiceID+"/pay",
			bytes.NewBufferString(`{"generate_settlement_transaction":false}`))
		req.Header.Set("Content-Type", "application/json")
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusNotFound, res.StatusCode)
	})
}
