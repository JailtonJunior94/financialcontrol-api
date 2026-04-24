package http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	appresponses "github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos/responses"
	transactionsapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/transactions/application"
	transactionshttp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/transactions/http"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

type transactionServiceStub struct {
	transactionsStatus        int
	transactionsData          any
	createTransactionStatus   int
	createTransactionData     any
	createItemStatus          int
	createItemData            any
	updateTransactionStatus   int
	updateTransactionData     any
	markAsPaidStatus          int
	markAsPaidData            any
	removeTransactionStatus   int
	removeTransactionData     any
	transactionByIDStatus     int
	transactionByIDData       any
	cloneTransactionStatus    int
	cloneTransactionData      any
	transactionItemByIDStatus int
	transactionItemByIDData   any
}

func (s *transactionServiceStub) Transactions(string) *transactionsapp.HttpResponse {
	return &appresponses.HttpResponse{StatusCode: s.transactionsStatus, Data: s.transactionsData}
}

func (s *transactionServiceStub) TransactionById(string, string) *transactionsapp.HttpResponse {
	return &appresponses.HttpResponse{StatusCode: s.transactionByIDStatus, Data: s.transactionByIDData}
}

func (s *transactionServiceStub) CreateTransaction(request *transactionsapp.TransactionRequest, userID string) *transactionsapp.HttpResponse {
	return &appresponses.HttpResponse{StatusCode: s.createTransactionStatus, Data: s.createTransactionData}
}

func (s *transactionServiceStub) CloneTransaction(string, string) *transactionsapp.HttpResponse {
	return &appresponses.HttpResponse{StatusCode: s.cloneTransactionStatus, Data: s.cloneTransactionData}
}

func (s *transactionServiceStub) TransactionItemById(string, string) *transactionsapp.HttpResponse {
	return &appresponses.HttpResponse{StatusCode: s.transactionItemByIDStatus, Data: s.transactionItemByIDData}
}

func (s *transactionServiceStub) CreateTransactionItem(*transactionsapp.TransactionItemRequest, string, string) *transactionsapp.HttpResponse {
	return &appresponses.HttpResponse{StatusCode: s.createItemStatus, Data: s.createItemData}
}

func (s *transactionServiceStub) UpdateTransactionItem(string, string, string, *transactionsapp.TransactionItemRequest) *transactionsapp.HttpResponse {
	return &appresponses.HttpResponse{StatusCode: s.updateTransactionStatus, Data: s.updateTransactionData}
}

func (s *transactionServiceStub) MarkAsPaidTransactionItem(string, string, string, *transactionsapp.TransactionMarkAsPaid) *transactionsapp.HttpResponse {
	return &appresponses.HttpResponse{StatusCode: s.markAsPaidStatus, Data: s.markAsPaidData}
}

func (s *transactionServiceStub) RemoveTransactionItem(string, string, string) *transactionsapp.HttpResponse {
	return &appresponses.HttpResponse{StatusCode: s.removeTransactionStatus, Data: s.removeTransactionData}
}

type claimsResolverStub struct {
	userID string
	err    error
}

func (s *claimsResolverStub) UserID(string) (string, error) {
	if s.err != nil {
		return "", s.err
	}

	return s.userID, nil
}

func TestTransactionsContractReturnsUnauthorizedWhenClaimsAreInvalid(t *testing.T) {
	app := fiber.New()
	controller := transactionshttp.NewTransactionController(
		&transactionServiceStub{transactionsStatus: fiber.StatusOK, transactionsData: []transactionsapp.TransactionResponse{}},
		&claimsResolverStub{err: errors.New("invalid token")},
	)

	app.Get("/transactions", controller.Transactions)

	req := httptest.NewRequest(fiber.MethodGet, "/transactions", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestCreateTransactionContractPreservesPayloadValidation(t *testing.T) {
	app := fiber.New()
	controller := transactionshttp.NewTransactionController(
		&transactionServiceStub{
			createTransactionStatus: fiber.StatusCreated,
			createTransactionData:   &transactionsapp.TransactionResponse{ID: "tx-1"},
		},
		&claimsResolverStub{userID: "user-1"},
	)

	app.Post("/transactions", controller.CreateTransaction)

	body, err := json.Marshal(map[string]any{
		"date": time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)

	req := httptest.NewRequest(fiber.MethodPost, "/transactions", bytes.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	req.Header.Set("Authorization", "Bearer token")

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusCreated, resp.StatusCode)
}

func TestAddTransactionRouterPreservesPublicPaths(t *testing.T) {
	app := fiber.New()
	transactionshttp.AddTransactionRouter(app, &transactionshttp.TransactionController{})

	routes := app.GetRoutes(true)
	routeIndex := make(map[string]map[string]bool, len(routes))
	for _, route := range routes {
		if routeIndex[route.Path] == nil {
			routeIndex[route.Path] = map[string]bool{}
		}
		routeIndex[route.Path][route.Method] = true
	}

	require.True(t, routeIndex["/transactions"][fiber.MethodGet])
	require.True(t, routeIndex["/transactions"][fiber.MethodPost])
	require.True(t, routeIndex["/transactions/:id"][fiber.MethodGet])
	require.True(t, routeIndex["/transactions/:transactionid"][fiber.MethodPost])
	require.True(t, routeIndex["/transactions/:transactionid/clone"][fiber.MethodPost])
	require.True(t, routeIndex["/transactions/:transactionid/items/:id"][fiber.MethodGet])
	require.True(t, routeIndex["/transactions/:transactionid/items/:id"][fiber.MethodPut])
	require.True(t, routeIndex["/transactions/:transactionid/items/:id"][fiber.MethodPatch])
	require.True(t, routeIndex["/transactions/:transactionid/items/:id"][fiber.MethodDelete])
}
