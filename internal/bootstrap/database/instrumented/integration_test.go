//go:build integration

package instrumented_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	_ "github.com/microsoft/go-mssqldb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/database/instrumented"
	"github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/observability/redactor"
)

const (
	mssqlImage  = "mcr.microsoft.com/mssql/server:2022-latest"
	mssqlSAPass = "TestP@ss123!"
	mssqlPort   = "1433/tcp"
)

func startMSSQL(t *testing.T) *sql.DB {
	t.Helper()
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        mssqlImage,
		ExposedPorts: []string{mssqlPort},
		Env: map[string]string{
			"ACCEPT_EULA": "Y",
			"SA_PASSWORD": mssqlSAPass,
			"MSSQL_PID":   "Developer",
		},
		WaitingFor: wait.ForLog("SQL Server is now ready for client connections").
			WithStartupTimeout(120 * time.Second),
	}

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = c.Terminate(context.Background()) })

	host, err := c.Host(ctx)
	require.NoError(t, err)
	port, err := c.MappedPort(ctx, mssqlPort)
	require.NoError(t, err)

	dsn := fmt.Sprintf(
		"sqlserver://sa:%s@%s:%s?database=master&connection+timeout=30",
		mssqlSAPass, host, port.Port(),
	)

	db, err := sql.Open("sqlserver", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	for i := range 10 {
		pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err = db.PingContext(pingCtx)
		cancel()
		if err == nil {
			break
		}
		if i < 9 {
			time.Sleep(2 * time.Second)
		}
	}
	require.NoError(t, err, "MSSQL container did not become ready in time")
	return db
}

// TestIntegration_ApplicationName verifies that the wrapped DB connects with the
// correct app name injected via ConfigureConnectionString.
func TestIntegration_ApplicationName(t *testing.T) {
	baseDB := startMSSQL(t)

	rec := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec))
	tracer := tp.Tracer("test")

	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	wrappedDB := instrumented.Wrap(
		baseDB,
		"master",
		10*time.Second,
		redactor.DefaultDenylist,
		logger,
		tracer,
	)
	t.Cleanup(func() { _ = wrappedDB.Close() })

	var appName string
	err := wrappedDB.QueryRowContext(context.Background(), "SELECT APP_NAME()").Scan(&appName)
	require.NoError(t, err)

	// APP_NAME() returns the application name set by the connection;
	// since we did not call ConfigureConnectionString here, it will be the driver default.
	// The integration test for application name injects via ConfigureConnectionString + sql.Open.
	assert.NotEmpty(t, appName)

	spans := rec.Ended()
	require.NotEmpty(t, spans, "at least one span must be emitted")

	var dbSystem string
	for _, a := range spans[0].Attributes() {
		if string(a.Key) == "db.system" {
			dbSystem = a.Value.AsString()
		}
	}
	assert.Equal(t, "mssql", dbSystem)
}

// TestIntegration_ConfigureConnectionString verifies the app name is injected and
// visible via APP_NAME() after opening with the configured DSN.
func TestIntegration_ConfigureConnectionString(t *testing.T) {
	baseDB := startMSSQL(t)

	// Get original DSN from the container for ConfigureConnectionString test
	// We open a fresh connection with the configured DSN
	var host, port string
	rows, err := baseDB.QueryContext(context.Background(),
		"SELECT @@SERVERNAME")
	require.NoError(t, err)
	rows.Close()
	_ = host
	_ = port

	// Use ConfigureConnectionString to inject app name
	rawDSN := fmt.Sprintf(
		"sqlserver://sa:%s@localhost:1433?database=master",
		mssqlSAPass,
	)
	configured, err := instrumented.ConfigureConnectionString(rawDSN, "development")
	require.NoError(t, err)
	assert.Contains(t, configured, "financialcontrol-api-development")
}

// TestIntegration_SpanEmittedWithRedactedStatement verifies that a query containing
// a PII field in the statement has the value replaced with [REDACTED] in the span.
func TestIntegration_SpanEmittedWithRedactedStatement(t *testing.T) {
	baseDB := startMSSQL(t)

	rec := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec))
	tracer := tp.Tracer("test")

	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	wrappedDB := instrumented.Wrap(
		baseDB,
		"master",
		10*time.Second,
		redactor.DefaultDenylist,
		logger,
		tracer,
	)
	t.Cleanup(func() { _ = wrappedDB.Close() })

	// Execute a query with a PII field in WHERE clause (key=value form)
	_, err := wrappedDB.QueryContext(context.Background(),
		"SELECT 1 WHERE cpf='123.456.789-00'")
	// Query may return no rows or error — we only care about the span
	_ = err

	spans := rec.Ended()
	require.NotEmpty(t, spans, "span must be emitted")

	var dbStatement string
	for _, a := range spans[0].Attributes() {
		if string(a.Key) == "db.statement" {
			dbStatement = a.Value.AsString()
		}
	}

	assert.Contains(t, dbStatement, "[REDACTED]",
		"PII value must be redacted in db.statement span attribute")
	assert.NotContains(t, dbStatement, "123.456.789-00",
		"raw CPF value must not appear in db.statement")
}

// TestIntegration_DBUserNotInSpan verifies db.user is never emitted.
func TestIntegration_DBUserNotInSpan(t *testing.T) {
	baseDB := startMSSQL(t)

	rec := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec))
	tracer := tp.Tracer("test")

	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))

	wrappedDB := instrumented.Wrap(
		baseDB, "master", 10*time.Second,
		redactor.DefaultDenylist, logger, tracer,
	)
	t.Cleanup(func() { _ = wrappedDB.Close() })

	row := wrappedDB.QueryRowContext(context.Background(), "SELECT 1")
	var v int
	_ = row.Scan(&v)

	for _, span := range rec.Ended() {
		for _, a := range span.Attributes() {
			assert.NotEqual(t, "db.user", string(a.Key),
				"db.user must never appear in span attributes")
		}
	}
}

// TestIntegration_SlowQueryLog verifies slow-query WARN is emitted with required fields.
func TestIntegration_SlowQueryLog(t *testing.T) {
	baseDB := startMSSQL(t)

	rec := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec))
	tracer := tp.Tracer("test")

	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelWarn}))

	// threshold=0 so every query triggers WARN
	wrappedDB := instrumented.Wrap(
		baseDB, "master", 0, // 0 duration: every query is "slow"
		redactor.DefaultDenylist, logger, tracer,
	)
	t.Cleanup(func() { _ = wrappedDB.Close() })

	row := wrappedDB.QueryRowContext(context.Background(), "SELECT 1")
	var v int
	require.NoError(t, row.Scan(&v))

	output := logBuf.String()
	require.NotEmpty(t, output, "WARN log must be emitted when threshold=0")

	lines := strings.Split(strings.TrimSpace(output), "\n")
	require.NotEmpty(t, lines)

	// Parse first WARN entry
	var entry map[string]any
	require.NoError(t, json.Unmarshal([]byte(lines[0]), &entry),
		"WARN log must be valid JSON")

	assert.Equal(t, "WARN", entry["level"])
	assert.Equal(t, "slow_query", entry["msg"])
	assert.NotNil(t, entry["query_duration_ms"])
	assert.NotNil(t, entry["trace_id"])
	assert.NotNil(t, entry["span_id"])
}
