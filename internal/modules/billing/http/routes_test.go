package http_test

import (
	"net/http/httptest"
	"testing"
	"time"

	billinghttp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/billing/http"
	pkgjwt "github.com/jailtonjunior94/financialcontrol-api/pkg/jwt"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func newTestParser(t *testing.T) pkgjwt.Parser {
	t.Helper()
	p, err := pkgjwt.NewParser(pkgjwt.Config{
		Secret:    []byte("test-secret-that-is-at-least-32-bytes-long!!"),
		AccessTTL: time.Minute,
	})
	require.NoError(t, err)
	return p
}

func TestBillRoutesReturnUnauthorizedWithoutToken(t *testing.T) {
	app := fiber.New()
	billinghttp.AddBillRouter(app, &billinghttp.BillController{}, newTestParser(t))

	req := httptest.NewRequest(fiber.MethodGet, "/bills", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAddBillRouterRegistersAllPaths(t *testing.T) {
	app := fiber.New()
	billinghttp.AddBillRouter(app, &billinghttp.BillController{}, newTestParser(t))

	routes := app.GetRoutes(true)
	routeIndex := make(map[string]map[string]bool, len(routes))
	for _, route := range routes {
		if routeIndex[route.Path] == nil {
			routeIndex[route.Path] = map[string]bool{}
		}
		routeIndex[route.Path][route.Method] = true
	}

	require.True(t, routeIndex["/bills"][fiber.MethodGet])
	require.True(t, routeIndex["/bills"][fiber.MethodPost])
	require.True(t, routeIndex["/bills/:id"][fiber.MethodGet])
}
