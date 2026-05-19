//go:build integration

package http_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	bootstrapcli "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/cli"
	mssqltesting "github.com/jailtonjunior94/financialcontrol-api/pkg/database/mssql"
)

const (
	// integrationJWTSecret is ≥ 32 bytes to satisfy pkgjwt.ErrSecretTooShort.
	integrationJWTSecret = "integration-test-super-secret-jwt-key-32bytes!!"
	serverPollInterval   = 300 * time.Millisecond
	serverReadyTimeout   = 90 * time.Second
	lgtmStartupTimeout   = 180 * time.Second
)

// TestIntegration_FullServerLifecycle starts grafana/otel-lgtm:0.7.5 and MSSQL via
// testcontainers, boots the full server stack and validates the four mandatory endpoints:
// /live, /ready, /health (JSON), and RFC 7807 payload on unknown /api/v1 routes.
func TestIntegration_FullServerLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test skipped in short mode")
	}

	// 1. Start MSSQL via shared testcontainer helper (sync.Once per test binary).
	mssqlDSN, err := mssqltesting.GetSharedTestDSN()
	require.NoError(t, err, "MSSQL testcontainer failed to start")

	// 2. Start LGTM container; expose OTLP gRPC (4317) and OTLP HTTP (4318).
	otlpEndpoint := startLGTMContainer(t)

	// 3. Allocate a free port for the Fiber server.
	serverPort := freePort(t)

	// 4. Point config.Load to the local testdata dir so viper finds config.development.yaml.
	t.Chdir("testdata")

	// 5. Configure all required environment variables.
	t.Setenv("ENVIRONMENT", "development")
	t.Setenv("SERVICE_NAME", "integration-test-svc")
	t.Setenv("SERVICE_VERSION", "0.0.1")
	t.Setenv("MSSQL_CONNECTION_STRING", mssqlDSN)
	t.Setenv("JWT_SECRET", integrationJWTSecret)
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", otlpEndpoint)
	t.Setenv("PORT", serverPort)

	// 6. Boot the full server in a goroutine with a cancellable context.
	// Use bootstrapcli.RunServer so the production lifecycle is exercised
	// (BuildRuntime → NewServer → Start → graceful shutdown of obs+db with
	// errors.Join). Cancelling ctx in the cleanup forces the same shutdown
	// path that signal.NotifyContext would in cmd/main.go.
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	startErr := make(chan error, 1)
	go func() {
		startErr <- bootstrapcli.RunServer(ctx)
	}()

	// 7. Poll /live until the server accepts requests (no time.Sleep).
	baseURL := fmt.Sprintf("http://localhost:%s", serverPort)
	waitForServer(t, baseURL+"/live", serverReadyTimeout, startErr)

	// 8. Table-driven assertions over the four mandatory endpoints.
	type endpointCase struct {
		name       string
		path       string
		wantStatus int
		checkBody  func(t *testing.T, body []byte, resp *http.Response)
	}

	cases := []endpointCase{
		{
			name:       "GET /live returns 200",
			path:       "/live",
			wantStatus: http.StatusOK,
		},
		{
			name:       "GET /ready returns 200 when MSSQL healthy",
			path:       "/ready",
			wantStatus: http.StatusOK,
		},
		{
			name:       "GET /health returns structured JSON with mssql=healthy",
			path:       "/health",
			wantStatus: http.StatusOK,
			checkBody:  assertHealthPayload,
		},
		{
			name:       "GET /api/v1/__not_found__ returns RFC 7807 404",
			path:       "/api/v1/__not_found__",
			wantStatus: http.StatusNotFound,
			checkBody:  assertRFC7807Payload,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req, reqErr := http.NewRequestWithContext(context.Background(), http.MethodGet, baseURL+tc.path, nil)
			require.NoError(t, reqErr)

			resp, doErr := http.DefaultClient.Do(req)
			require.NoError(t, doErr)
			t.Cleanup(func() { _ = resp.Body.Close() })

			assert.Equal(t, tc.wantStatus, resp.StatusCode)

			if tc.checkBody != nil {
				body, readErr := io.ReadAll(resp.Body)
				require.NoError(t, readErr)
				tc.checkBody(t, body, resp)
			}
		})
	}
}

// assertHealthPayload verifies the /health JSON structure (RF-04).
func assertHealthPayload(t *testing.T, body []byte, _ *http.Response) {
	t.Helper()

	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &payload), "health response must be valid JSON")

	assert.Contains(t, payload, "service", "health payload must have 'service'")
	assert.Contains(t, payload, "version", "health payload must have 'version'")
	assert.Contains(t, payload, "environment", "health payload must have 'environment'")
	assert.Contains(t, payload, "timestamp", "health payload must have 'timestamp'")

	checks, ok := payload["checks"].(map[string]interface{})
	require.True(t, ok, "health payload must have 'checks' object")

	mssql, ok := checks["mssql"].(map[string]interface{})
	require.True(t, ok, "checks must contain 'mssql' key")
	assert.Equal(t, "healthy", mssql["status"], "mssql check must report 'healthy'")
}

// assertRFC7807Payload verifies the error response is RFC 7807 compliant (RF-03.1).
func assertRFC7807Payload(t *testing.T, body []byte, resp *http.Response) {
	t.Helper()

	assert.Contains(t, resp.Header.Get("Content-Type"), "application/problem+json",
		"404 response must use Content-Type: application/problem+json")

	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &payload), "RFC 7807 response must be valid JSON")

	assert.Contains(t, payload, "type", "RFC 7807 payload must have 'type'")
	assert.Contains(t, payload, "title", "RFC 7807 payload must have 'title'")
	assert.Contains(t, payload, "status", "RFC 7807 payload must have 'status'")
	assert.Contains(t, payload, "instance", "RFC 7807 payload must have 'instance'")
	assert.Contains(t, payload, "request_id", "RFC 7807 payload must have 'request_id'")
}

// startLGTMContainer starts grafana/otel-lgtm:0.7.5 and returns the OTLP gRPC endpoint
// (host:port) for the dynamically mapped container port 4317.
func startLGTMContainer(t *testing.T) string {
	t.Helper()
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "grafana/otel-lgtm:0.7.5",
		ExposedPorts: []string{"4317/tcp", "4318/tcp"},
		WaitingFor: wait.ForListeningPort("4317/tcp").
			WithStartupTimeout(lgtmStartupTimeout),
	}

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err, "failed to start LGTM (grafana/otel-lgtm:0.7.5) testcontainer")

	t.Cleanup(func() {
		if terr := c.Terminate(context.Background()); terr != nil {
			t.Logf("LGTM container terminate error (ignored): %v", terr)
		}
	})

	host, err := c.Host(ctx)
	require.NoError(t, err, "failed to get LGTM container host")

	mappedPort, err := c.MappedPort(ctx, "4317/tcp")
	require.NoError(t, err, "failed to get LGTM OTLP gRPC mapped port (4317)")

	return fmt.Sprintf("%s:%s", host, mappedPort.Port())
}

// freePort finds an available TCP port and returns it as a string.
// TOCTOU window is acceptable for test-only use.
func freePort(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", ":0")
	require.NoError(t, err, "failed to find a free TCP port")
	port := l.Addr().(*net.TCPAddr).Port
	require.NoError(t, l.Close())
	return fmt.Sprintf("%d", port)
}

// waitForServer polls url with context.WithTimeout until HTTP 200 is returned.
// It monitors startErr for premature server goroutine exit.
// No time.Sleep is used; polling uses context.WithTimeout (R-TEST-001).
func waitForServer(t *testing.T, url string, timeout time.Duration, startErr <-chan error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client := &http.Client{Timeout: 2 * time.Second}

	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err == nil {
			if resp, doErr := client.Do(req); doErr == nil {
				_ = resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					return
				}
			}
		}

		select {
		case <-ctx.Done():
			t.Fatalf("server did not become ready within %s: %v", timeout, ctx.Err())
		case err := <-startErr:
			t.Fatalf("server goroutine exited before server became ready (err: %v)", err)
		case <-time.After(serverPollInterval):
		}
	}
}
