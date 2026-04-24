package cards_test

import (
	"testing"

	"github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards"
	cardshttp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/http"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestRegistrationExposesCardHTTPRoutes(t *testing.T) {
	registration := cards.Registration()

	require.Equal(t, "cards", registration.Name)
	require.NotNil(t, registration.RegisterHTTP)
	require.Nil(t, registration.RegisterCLI)

	app := fiber.New()
	registration.RegisterHTTP(app, &container.Container{
		CardController: &cardshttp.CardController{},
	})

	routes := app.GetRoutes(true)
	routeIndex := make(map[string]map[string]bool, len(routes))
	for _, route := range routes {
		if routeIndex[route.Path] == nil {
			routeIndex[route.Path] = map[string]bool{}
		}

		routeIndex[route.Path][route.Method] = true
	}

	require.True(t, routeIndex["/cards"][fiber.MethodGet])
	require.True(t, routeIndex["/cards"][fiber.MethodPost])
	require.True(t, routeIndex["/cards/:id"][fiber.MethodGet])
	require.True(t, routeIndex["/cards/:id"][fiber.MethodPut])
	require.True(t, routeIndex["/cards/:id"][fiber.MethodDelete])
}
