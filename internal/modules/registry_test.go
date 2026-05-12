package modules_test

import (
	"testing"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules"

	"github.com/stretchr/testify/require"
)

func TestRegistrationsExposeExpectedFoundationModules(t *testing.T) {
	registrations := modules.Registrations()
	require.Empty(t, registrations, "all legacy module registrations removed in task 9.0")
}

func TestRegistrationsSeparateHTTPAndCLIHooks(t *testing.T) {
	for _, registration := range modules.Registrations() {
		switch registration.Name {
		case "planning":
			require.Nil(t, registration.RegisterHTTP)
			require.NotNil(t, registration.RegisterCLI)
		default:
			require.NotNil(t, registration.RegisterHTTP)
		}
	}
}
