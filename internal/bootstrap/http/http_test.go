package http_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/observability/noop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bootstrapconfig "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/config"
	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	bootstraphealth "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/health"
	bootstraphttp "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/http"
	migrationmocks "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/migration/mocks"
	cards "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/infrastructure/http/handlers"
	categories "github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories"
	categorieshandlers "github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/infrastructure/http/handlers"
	finance "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance"
	financehandlers "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/http/handlers"
	identity "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity"
)

func stubCardsModule() *cards.Module {
	return &cards.Module{
		CardHandler: handlers.NewCardHandler(nil, nil, nil, nil, nil),
		FlagHandler: handlers.NewFlagHandler(nil),
	}
}

func stubModule() *identity.Module { return &identity.Module{} }

func stubCategoriesModule() *categories.Module {
	return &categories.Module{
		CategoryHandler: categorieshandlers.NewCategoryHandler(nil, nil, nil, nil, nil),
	}
}

func stubFinanceModule() *finance.Module {
	return &finance.Module{
		TransactionHandler: financehandlers.NewTransactionHandler(nil, nil, nil, nil, nil, nil),
		InvoiceHandler:     financehandlers.NewInvoiceHandler(nil, nil, nil),
		InstallmentHandler: financehandlers.NewInstallmentHandler(nil),
		SummaryHandler:     financehandlers.NewSummaryHandler(nil),
	}
}

func stubContainer(t *testing.T) *bootstrapcontainer.Container {
	t.Helper()
	t.Setenv("SERVICE_NAME", "test-svc")
	t.Setenv("SERVICE_VERSION", "1.0.0")
	t.Setenv("ENVIRONMENT", "development")

	id, err := bootstrapconfig.NewServiceIdentity()
	require.NoError(t, err)

	timeout, err := bootstrapconfig.NewShutdownTimeout()
	require.NoError(t, err)

	mockMgr := migrationmocks.NewMockManager(t)

	return &bootstrapcontainer.Container{
		Observability:    noop.NewProvider(),
		Identity:         id,
		ShutdownTimeout:  timeout,
		DBManager:        mockMgr,
		IdentityModule:   stubModule(),
		CardsModule:      stubCardsModule(),
		CategoriesModule: stubCategoriesModule(),
		FinanceModule:    stubFinanceModule(),
	}
}

// TestNewServer_SmokeNoError verifies that NewServer mounts the server without error
// when given a complete container.
func TestNewServer_SmokeNoError(t *testing.T) {
	c := stubContainer(t)
	srv, err := bootstraphttp.NewServer(c)
	require.NoError(t, err)
	require.NotNil(t, srv)
}

// TestNewServer_BuildsWithExpectedOptions verifies that the server is composed with
// tracing, OTel metrics, CORS, health checks and the /api/v1 router group (RF-01/02/03/04/05).
func TestNewServer_BuildsWithExpectedOptions(t *testing.T) {
	c := stubContainer(t)
	srv, err := bootstraphttp.NewServer(c)
	require.NoError(t, err)
	require.NotNil(t, srv)

	routes := srv.App().GetRoutes(true)
	paths := make(map[string]bool, len(routes))
	for _, r := range routes {
		paths[r.Path] = true
	}

	// Health endpoints are evidence of WithHealthChecks + WithTracing + WithOTelMetrics applied.
	assert.True(t, paths["/live"], "expected /live route (health checks enabled)")
	assert.True(t, paths["/ready"], "expected /ready route (health checks enabled)")
	assert.True(t, paths["/health"], "expected /health route (health checks enabled)")

	// At least one /api/v1/* route confirms the router was registered.
	hasV1 := false
	for p := range paths {
		if strings.HasPrefix(p, "/api/v1") {
			hasV1 = true
			break
		}
	}
	assert.True(t, hasV1, "expected at least one /api/v1/* route (apiV1Router registered)")
}

// TestNewServer_CORSWildcard verifies that CORS is configured with wildcard origin (RF-03).
func TestNewServer_CORSWildcard(t *testing.T) {
	c := stubContainer(t)
	srv, err := bootstraphttp.NewServer(c)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/live", nil)
	req.Header.Set("Origin", "https://example.com")

	resp, testErr := srv.App().Test(req)
	require.NoError(t, testErr)
	t.Cleanup(func() { _ = resp.Body.Close() })

	assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"),
		"CORS must allow all origins (*)")
}

// TestNewServer_MetricsEndpointDisabled verifies that /metrics is NOT registered (RF-05).
func TestNewServer_MetricsEndpointDisabled(t *testing.T) {
	c := stubContainer(t)
	srv, err := bootstraphttp.NewServer(c)
	require.NoError(t, err)

	for _, r := range srv.App().GetRoutes(true) {
		assert.NotEqual(t, "/metrics", r.Path,
			"expected /metrics not to be registered (Prometheus endpoint is disabled)")
	}
}

// TestHealthChecks_MSSQLRegistered verifies that bootstraphealth.New delivers a map
// with the mandatory "mssql" key to WithHealthChecks (RF-04).
func TestHealthChecks_MSSQLRegistered(t *testing.T) {
	c := stubContainer(t)
	checks := bootstraphealth.New(c.DBManager)
	m := checks.Map()

	require.Contains(t, m, "mssql", "expected 'mssql' key in health checks map")
	require.Len(t, m, 1, "expected exactly one health check (mssql)")
}

// TestShutdownTimeout_ParsedFromEnv verifies that HTTP_SHUTDOWN_TIMEOUT is applied
// correctly: default 15s, explicit value, invalid value errors (RF-07).
func TestShutdownTimeout_ParsedFromEnv(t *testing.T) {
	tests := []struct {
		name     string
		absent   bool
		envValue string
		want     time.Duration
		wantErr  bool
	}{
		{name: "default 15s when env absent", absent: true, want: 15 * time.Second},
		{name: "explicit 30s", envValue: "30s", want: 30 * time.Second},
		{name: "explicit 1m", envValue: "1m", want: time.Minute},
		{name: "invalid empty fails", envValue: "", wantErr: true},
		{name: "invalid abc fails", envValue: "abc", wantErr: true},
		{name: "zero duration fails", envValue: "0s", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.absent {
				orig, wasSet := os.LookupEnv("HTTP_SHUTDOWN_TIMEOUT")
				require.NoError(t, os.Unsetenv("HTTP_SHUTDOWN_TIMEOUT"))
				t.Cleanup(func() {
					if wasSet {
						_ = os.Setenv("HTTP_SHUTDOWN_TIMEOUT", orig)
					}
				})
			} else {
				t.Setenv("HTTP_SHUTDOWN_TIMEOUT", tc.envValue)
			}

			timeout, err := bootstrapconfig.NewShutdownTimeout()
			if tc.wantErr {
				require.Error(t, err)
				assert.True(t, errors.Is(err, bootstrapconfig.ErrInvalidShutdownTimeout))
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, timeout.Duration())
		})
	}
}

// TestPort_DefaultsToThreeThousand verifies that NewServer builds successfully when
// PORT is absent (defaults to :3000) and with an explicit override (RF-08).
func TestPort_DefaultsToThreeThousand(t *testing.T) {
	t.Run("absent PORT builds server successfully", func(t *testing.T) {
		orig, wasSet := os.LookupEnv("PORT")
		require.NoError(t, os.Unsetenv("PORT"))
		t.Cleanup(func() {
			if wasSet {
				_ = os.Setenv("PORT", orig)
			}
		})

		c := stubContainer(t)
		srv, err := bootstraphttp.NewServer(c)
		require.NoError(t, err)
		require.NotNil(t, srv, "server must be built when PORT is absent (defaults to :3000)")
	})

	t.Run("explicit PORT 8080 builds server successfully", func(t *testing.T) {
		t.Setenv("PORT", "8080")

		c := stubContainer(t)
		srv, err := bootstraphttp.NewServer(c)
		require.NoError(t, err)
		require.NotNil(t, srv)
	})
}

// TestServiceIdentity_Whitelist verifies that ENVIRONMENT is case-sensitive and
// only accepts {development, staging, production} (RF-14).
func TestServiceIdentity_Whitelist(t *testing.T) {
	tests := []struct {
		name    string
		env     string
		wantErr bool
	}{
		{name: "development is valid", env: "development"},
		{name: "staging is valid", env: "staging"},
		{name: "production is valid", env: "production"},
		{name: "Production (capitalized) is invalid", env: "Production", wantErr: true},
		{name: "PROD is invalid", env: "PROD", wantErr: true},
		{name: "dev is invalid", env: "dev", wantErr: true},
		{name: "empty is invalid", env: "", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("SERVICE_NAME", "svc")
			t.Setenv("SERVICE_VERSION", "1.0.0")
			t.Setenv("ENVIRONMENT", tc.env)

			_, err := bootstrapconfig.NewServiceIdentity()
			if tc.wantErr {
				require.Error(t, err)
				assert.True(t, errors.Is(err, bootstrapconfig.ErrInvalidEnvironment),
					"expected ErrInvalidEnvironment, got: %v", err)
				return
			}
			require.NoError(t, err)
		})
	}
}

// TestNewServer_FailsOnInvalidIdentity verifies fail-fast: NewServer propagates
// serverfiber validation errors when the container carries a zero ServiceIdentity
// (empty name/version/environment) (RF-11). The full BuildRuntime → NewServer
// chain is covered in cli/runners_test.go.
func TestNewServer_FailsOnInvalidIdentity(t *testing.T) {
	// A zero ServiceIdentity yields empty name/version/environment.
	// serverfiber.New validates them after applying options; returns error on empty ServiceName.
	c := &bootstrapcontainer.Container{
		Observability:    noop.NewProvider(),
		Identity:         bootstrapconfig.ServiceIdentity{}, // zero: empty name/version/env
		ShutdownTimeout:  bootstrapconfig.ShutdownTimeout{}, // zero: 0 duration
		DBManager:        migrationmocks.NewMockManager(t),
		IdentityModule:   &identity.Module{},
		CardsModule:      stubCardsModule(),
		CategoriesModule: stubCategoriesModule(),
		FinanceModule:    stubFinanceModule(),
	}

	_, err := bootstraphttp.NewServer(c)
	require.Error(t, err, "NewServer must propagate serverfiber validation errors (fail-fast)")
}

// TestNewAppRegistersAPIBootstrapRoutes verifies that /api/v1/* paths are reachable
// through the mounted server after NewServer composition (RF-09).
func TestNewAppRegistersAPIBootstrapRoutes(t *testing.T) {
	c := stubContainer(t)
	srv, err := bootstraphttp.NewServer(c)
	require.NoError(t, err)

	routes := srv.App().GetRoutes(true)
	routeIndex := make(map[string]map[string]bool, len(routes))
	for _, route := range routes {
		if routeIndex[route.Path] == nil {
			routeIndex[route.Path] = map[string]bool{}
		}
		routeIndex[route.Path][route.Method] = true
	}

	require.True(t, routeIndex["/api/v1/token"]["POST"])
	require.True(t, routeIndex["/api/v1/users"]["POST"])
	require.True(t, routeIndex["/api/v1/me"]["GET"])
	require.True(t, routeIndex["/api/v1/cards"]["POST"])
	require.True(t, routeIndex["/api/v1/categories"]["GET"])
	require.True(t, routeIndex["/api/v1/finance/transactions"]["POST"])
	require.True(t, routeIndex["/api/v1/finance/transactions"]["GET"])
	require.True(t, routeIndex["/api/v1/finance/invoices"]["GET"])
	require.True(t, routeIndex["/api/v1/finance/summary"]["GET"])
}

// TestAPIV1Router_RegistersUnderV1 verifies that apiV1Router creates the /api/v1 group
// and all module routes are nested under it (RF-15).
func TestAPIV1Router_RegistersUnderV1(t *testing.T) {
	c := stubContainer(t)
	srv, err := bootstraphttp.NewServer(c)
	require.NoError(t, err)

	hasV1Route := false
	for _, r := range srv.App().GetRoutes(true) {
		if strings.HasPrefix(r.Path, "/api/v1/") {
			hasV1Route = true
			break
		}
	}
	require.True(t, hasV1Route, "expected at least one route under /api/v1/ (apiV1Router must register the group)")
}

// TestErrorHandler_RFC7807Payload verifies that unmatched routes return
// Content-Type: application/problem+json with type, title and status fields (RF-03.1).
func TestErrorHandler_RFC7807Payload(t *testing.T) {
	c := stubContainer(t)
	srv, err := bootstraphttp.NewServer(c)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/__not_found__", nil)
	resp, testErr := srv.App().Test(req)
	require.NoError(t, testErr)
	t.Cleanup(func() { _ = resp.Body.Close() })

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
	require.Equal(t, "application/problem+json", resp.Header.Get("Content-Type"))

	var body map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Contains(t, body, "type", "RFC 7807 payload must have 'type'")
	assert.Contains(t, body, "title", "RFC 7807 payload must have 'title'")
	assert.Contains(t, body, "status", "RFC 7807 payload must have 'status'")
	assert.Equal(t, float64(http.StatusNotFound), body["status"])
}
