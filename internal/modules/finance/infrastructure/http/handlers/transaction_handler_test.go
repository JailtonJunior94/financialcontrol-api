package handlers_test

import (
	"bytes"
	"encoding/json"
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

const (
	testUserID     = "550e8400-e29b-41d4-a716-446655440001"
	testTxID       = "550e8400-e29b-41d4-a716-446655440002"
	testCardID     = "550e8400-e29b-41d4-a716-446655440003"
	testCategoryID = "550e8400-e29b-41d4-a716-446655440004"
	testInvoiceID  = "550e8400-e29b-41d4-a716-446655440005"
	testInstallID  = "550e8400-e29b-41d4-a716-446655440006"
)

type TransactionHandlerSuite struct {
	suite.Suite
	createUC *ucmocks.CreateTransaction
	listUC   *ucmocks.ListTransactions
	getUC    *ucmocks.GetTransaction
	updateUC *ucmocks.UpdateTransaction
	deleteUC *ucmocks.DeleteTransaction
	refundUC *ucmocks.RefundTransaction
	sut      *handlers.TransactionHandler
	app      *fiber.App
}

func TestTransactionHandlerSuite(t *testing.T) { suite.Run(t, new(TransactionHandlerSuite)) }

func (s *TransactionHandlerSuite) SetupTest() {
	s.createUC = ucmocks.NewCreateTransaction(s.T())
	s.listUC = ucmocks.NewListTransactions(s.T())
	s.getUC = ucmocks.NewGetTransaction(s.T())
	s.updateUC = ucmocks.NewUpdateTransaction(s.T())
	s.deleteUC = ucmocks.NewDeleteTransaction(s.T())
	s.refundUC = ucmocks.NewRefundTransaction(s.T())
	s.sut = handlers.NewTransactionHandler(
		s.createUC, s.listUC, s.getUC, s.updateUC, s.deleteUC, s.refundUC,
	)

	s.app = fiber.New()
	withIdentity := func(c *fiber.Ctx) error {
		c.SetUserContext(identitycontext.WithIdentity(c.UserContext(), identitycontext.Identity{
			UserID: testUserID,
			Email:  "user@example.com",
		}))
		return c.Next()
	}
	s.app.Post("/finance/transactions", withIdentity, s.sut.Create)
	s.app.Get("/finance/transactions", withIdentity, s.sut.List)
	s.app.Get("/finance/transactions/:id", withIdentity, s.sut.Get)
	s.app.Put("/finance/transactions/:id", withIdentity, s.sut.Update)
	s.app.Delete("/finance/transactions/:id", withIdentity, s.sut.Delete)
	s.app.Post("/finance/transactions/:id/refund", withIdentity, s.sut.Refund)
}

func ptr(s string) *string { return &s }

func sampleTxResponse() dtos.TransactionResponse {
	return dtos.TransactionResponse{
		ID:              testTxID,
		Description:     "Compra Mercado",
		Amount:          "150.00",
		OccurredAt:      time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		TransactionType: "expense",
		PaymentMethod:   "credit_card",
		CardID:          ptr(testCardID),
		CategoryID:      testCategoryID,
		CreatedAt:       time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:       time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
	}
}

func (s *TransactionHandlerSuite) TestCreate() {
	validBody := `{"description":"Mercado","amount":"150.00","occurred_at":"2026-05-01T00:00:00Z","transaction_type":"expense","payment_method":"pix","category_id":"` + testCategoryID + `"}`
	resp := sampleTxResponse()

	s.Run("201 created with generated idempotency key", func() {
		s.createUC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(resp, nil).Once()
		req := httptest.NewRequest("POST", "/finance/transactions", bytes.NewBufferString(validBody))
		req.Header.Set("Content-Type", "application/json")
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusCreated, res.StatusCode)
		s.NotEmpty(res.Header.Get("X-Idempotency-Key"))
	})

	s.Run("201 created with explicit idempotency key", func() {
		s.createUC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(resp, nil).Once()
		req := httptest.NewRequest("POST", "/finance/transactions", bytes.NewBufferString(validBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", "550e8400-e29b-41d4-a716-446655440099")
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusCreated, res.StatusCode)
		s.Equal("550e8400-e29b-41d4-a716-446655440099", res.Header.Get("X-Idempotency-Key"))
	})

	s.Run("400 invalid idempotency key too long", func() {
		longKey := "aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffffffffffffffff1" // 65 chars
		req := httptest.NewRequest("POST", "/finance/transactions", bytes.NewBufferString(validBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", longKey)
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusBadRequest, res.StatusCode)
	})

	s.Run("400 amount as raw number", func() {
		body := `{"description":"x","amount":99,"occurred_at":"2026-05-01T00:00:00Z","transaction_type":"expense","payment_method":"pix","category_id":"` + testCategoryID + `"}`
		req := httptest.NewRequest("POST", "/finance/transactions", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusBadRequest, res.StatusCode)
		var b map[string]any
		_ = json.NewDecoder(res.Body).Decode(&b)
		s.Contains(b["error"], "decimal")
	})

	s.Run("400 subcategory equals category", func() {
		body := `{"description":"x","amount":"10.00","occurred_at":"2026-05-01T00:00:00Z","transaction_type":"expense","payment_method":"pix","category_id":"` + testCategoryID + `","subcategory_id":"` + testCategoryID + `"}`
		req := httptest.NewRequest("POST", "/finance/transactions", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusBadRequest, res.StatusCode)
	})

	s.Run("422 invalid body", func() {
		req := httptest.NewRequest("POST", "/finance/transactions", bytes.NewBufferString("not-json"))
		req.Header.Set("Content-Type", "application/json")
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusUnprocessableEntity, res.StatusCode)
	})

	s.Run("409 idempotency mismatch", func() {
		s.createUC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(dtos.TransactionResponse{}, domain.ErrIdempotencyMismatch).Once()
		req := httptest.NewRequest("POST", "/finance/transactions", bytes.NewBufferString(validBody))
		req.Header.Set("Content-Type", "application/json")
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusConflict, res.StatusCode)
	})

	s.Run("500 unexpected error", func() {
		s.createUC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(dtos.TransactionResponse{}, errors.New("db")).Once()
		req := httptest.NewRequest("POST", "/finance/transactions", bytes.NewBufferString(validBody))
		req.Header.Set("Content-Type", "application/json")
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusInternalServerError, res.StatusCode)
	})
}

func (s *TransactionHandlerSuite) TestList() {
	page := dtos.PaginatedResponse[dtos.TransactionResponse]{
		Items: []dtos.TransactionResponse{sampleTxResponse()}, Total: 1, Page: 1, PageSize: 20,
	}

	s.Run("200 default", func() {
		s.listUC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(page, nil).Once()
		req := httptest.NewRequest("GET", "/finance/transactions", nil)
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusOK, res.StatusCode)
	})

	s.Run("200 with filters", func() {
		s.listUC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(page, nil).Once()
		req := httptest.NewRequest("GET", "/finance/transactions?page=1&page_size=10&transaction_type=expense", nil)
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusOK, res.StatusCode)
	})

	s.Run("400 invalid from date", func() {
		req := httptest.NewRequest("GET", "/finance/transactions?from=not-a-date", nil)
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusBadRequest, res.StatusCode)
	})

	s.Run("400 invalid to date", func() {
		req := httptest.NewRequest("GET", "/finance/transactions?to=not-a-date", nil)
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusBadRequest, res.StatusCode)
	})

	s.Run("500 use case error", func() {
		s.listUC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(dtos.PaginatedResponse[dtos.TransactionResponse]{}, errors.New("db")).Once()
		req := httptest.NewRequest("GET", "/finance/transactions", nil)
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusInternalServerError, res.StatusCode)
	})
}

func (s *TransactionHandlerSuite) TestGet() {
	resp := sampleTxResponse()

	s.Run("200 found", func() {
		s.getUC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(resp, nil).Once()
		req := httptest.NewRequest("GET", "/finance/transactions/"+testTxID, nil)
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusOK, res.StatusCode)
	})

	s.Run("400 invalid id", func() {
		req := httptest.NewRequest("GET", "/finance/transactions/not-uuid", nil)
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusBadRequest, res.StatusCode)
	})

	s.Run("404 not found", func() {
		s.getUC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(dtos.TransactionResponse{}, domain.ErrTransactionNotFound).Once()
		req := httptest.NewRequest("GET", "/finance/transactions/"+testTxID, nil)
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusNotFound, res.StatusCode)
	})
}

func (s *TransactionHandlerSuite) TestUpdate() {
	validBody := `{"description":"Updated","amount":"200.00","occurred_at":"2026-05-01T00:00:00Z","transaction_type":"expense","payment_method":"pix","category_id":"` + testCategoryID + `"}`
	resp := sampleTxResponse()

	s.Run("200 updated", func() {
		s.updateUC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(resp, nil).Once()
		req := httptest.NewRequest("PUT", "/finance/transactions/"+testTxID, bytes.NewBufferString(validBody))
		req.Header.Set("Content-Type", "application/json")
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusOK, res.StatusCode)
	})

	s.Run("400 invalid id", func() {
		req := httptest.NewRequest("PUT", "/finance/transactions/not-uuid", bytes.NewBufferString(validBody))
		req.Header.Set("Content-Type", "application/json")
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusBadRequest, res.StatusCode)
	})

	s.Run("400 amount as raw number", func() {
		body := `{"description":"x","amount":99,"occurred_at":"2026-05-01T00:00:00Z","transaction_type":"expense","payment_method":"pix","category_id":"` + testCategoryID + `"}`
		req := httptest.NewRequest("PUT", "/finance/transactions/"+testTxID, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusBadRequest, res.StatusCode)
	})

	s.Run("422 invalid body", func() {
		req := httptest.NewRequest("PUT", "/finance/transactions/"+testTxID, bytes.NewBufferString("not-json"))
		req.Header.Set("Content-Type", "application/json")
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusUnprocessableEntity, res.StatusCode)
	})

	s.Run("404 not found", func() {
		s.updateUC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(dtos.TransactionResponse{}, domain.ErrTransactionNotFound).Once()
		req := httptest.NewRequest("PUT", "/finance/transactions/"+testTxID, bytes.NewBufferString(validBody))
		req.Header.Set("Content-Type", "application/json")
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusNotFound, res.StatusCode)
	})

	s.Run("400 subcategory equals category", func() {
		body := `{"description":"x","amount":"10.00","occurred_at":"2026-05-01T00:00:00Z","transaction_type":"expense","payment_method":"pix","category_id":"` + testCategoryID + `","subcategory_id":"` + testCategoryID + `"}`
		req := httptest.NewRequest("PUT", "/finance/transactions/"+testTxID, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusBadRequest, res.StatusCode)
	})
}

func (s *TransactionHandlerSuite) TestDelete() {
	s.Run("204 no content", func() {
		s.deleteUC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
		req := httptest.NewRequest("DELETE", "/finance/transactions/"+testTxID, nil)
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusNoContent, res.StatusCode)
	})

	s.Run("400 invalid id", func() {
		req := httptest.NewRequest("DELETE", "/finance/transactions/not-uuid", nil)
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusBadRequest, res.StatusCode)
	})

	s.Run("404 not found", func() {
		s.deleteUC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(domain.ErrTransactionNotFound).Once()
		req := httptest.NewRequest("DELETE", "/finance/transactions/"+testTxID, nil)
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusNotFound, res.StatusCode)
	})

	s.Run("409 has dependent refund", func() {
		s.deleteUC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(domain.ErrTransactionHasDependentRefund).Once()
		req := httptest.NewRequest("DELETE", "/finance/transactions/"+testTxID, nil)
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusConflict, res.StatusCode)
	})
}

func (s *TransactionHandlerSuite) TestRefund() {
	resp := sampleTxResponse()

	s.Run("201 refund created", func() {
		s.refundUC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(resp, nil).Once()
		req := httptest.NewRequest("POST", "/finance/transactions/"+testTxID+"/refund", bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusCreated, res.StatusCode)
	})

	s.Run("400 invalid id", func() {
		req := httptest.NewRequest("POST", "/finance/transactions/not-uuid/refund", bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusBadRequest, res.StatusCode)
	})

	s.Run("422 invalid body", func() {
		req := httptest.NewRequest("POST", "/finance/transactions/"+testTxID+"/refund", bytes.NewBufferString("not-json"))
		req.Header.Set("Content-Type", "application/json")
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusUnprocessableEntity, res.StatusCode)
	})

	s.Run("409 refund already exists", func() {
		s.refundUC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(dtos.TransactionResponse{}, domain.ErrRefundAlreadyExists).Once()
		req := httptest.NewRequest("POST", "/finance/transactions/"+testTxID+"/refund", bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusConflict, res.StatusCode)
	})

	s.Run("409 refund of refund not allowed", func() {
		s.refundUC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(dtos.TransactionResponse{}, domain.ErrRefundOfRefundNotAllowed).Once()
		req := httptest.NewRequest("POST", "/finance/transactions/"+testTxID+"/refund", bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		res, err := s.app.Test(req)
		s.Require().NoError(err)
		s.Equal(fiber.StatusConflict, res.StatusCode)
	})
}
