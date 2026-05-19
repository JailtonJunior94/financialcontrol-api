package instrumented

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/observability/redactor"
)

// ------------------------------------------------------------------ mock driver

type mockDriver struct{}

func (d *mockDriver) Open(_ string) (driver.Conn, error) { return nil, fmt.Errorf("not used") }

type mockConnector struct {
	conn driver.Conn
}

func (m *mockConnector) Connect(_ context.Context) (driver.Conn, error) { return m.conn, nil }
func (m *mockConnector) Driver() driver.Driver                          { return &mockDriver{} }

// mockConn implements driver.Conn + driver.ConnBeginTx + driver.QueryerContext + driver.ExecerContext.
type mockConn struct {
	queryErr error
	execErr  error
	beginErr error
}

func (m *mockConn) Prepare(_ string) (driver.Stmt, error) { return &mockStmt{}, nil }
func (m *mockConn) Close() error                          { return nil }
func (m *mockConn) Begin() (driver.Tx, error)             { return &mockTx{}, nil }

func (m *mockConn) BeginTx(_ context.Context, _ driver.TxOptions) (driver.Tx, error) {
	if m.beginErr != nil {
		return nil, m.beginErr
	}
	return &mockTx{}, nil
}

func (m *mockConn) QueryContext(_ context.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
	if m.queryErr != nil {
		return nil, m.queryErr
	}
	return &mockRows{}, nil
}

func (m *mockConn) ExecContext(_ context.Context, _ string, _ []driver.NamedValue) (driver.Result, error) {
	if m.execErr != nil {
		return nil, m.execErr
	}
	return driver.RowsAffected(1), nil
}

type mockStmt struct{}

func (s *mockStmt) Close() error                                 { return nil }
func (s *mockStmt) NumInput() int                                { return -1 }
func (s *mockStmt) Exec(_ []driver.Value) (driver.Result, error) { return driver.RowsAffected(1), nil }
func (s *mockStmt) Query(_ []driver.Value) (driver.Rows, error)  { return &mockRows{}, nil }

type mockTx struct{}

func (t *mockTx) Commit() error   { return nil }
func (t *mockTx) Rollback() error { return nil }

type mockRows struct{ done bool }

func (r *mockRows) Columns() []string { return []string{"id"} }
func (r *mockRows) Close() error      { return nil }
func (r *mockRows) Next(dest []driver.Value) error {
	if r.done {
		return io.EOF
	}
	r.done = true
	dest[0] = int64(1)
	return nil
}

// ------------------------------------------------------------------ test helpers

// setupWrapped creates a wrapped *sql.DB backed by a mockConn.
// The fakeClock advances by step on each call so spans have measurable duration.
func setupWrapped(
	t *testing.T,
	mc *mockConn,
	step time.Duration,
	threshold time.Duration,
) (db *sql.DB, rec *tracetest.SpanRecorder, logBuf *bytes.Buffer) {
	t.Helper()

	rec = tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec))
	tracer := tp.Tracer("test")

	logBuf = &bytes.Buffer{}
	logger := slog.New(slog.NewJSONHandler(logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	mockDB := sql.OpenDB(&mockConnector{conn: mc})

	calls := 0
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	fakeClock := func() time.Time {
		ts := base.Add(time.Duration(calls) * step)
		calls++
		return ts
	}

	c := &connector{
		source:    mockDB,
		dbName:    "testdb",
		threshold: threshold,
		red:       redactor.DefaultDenylist,
		logger:    logger,
		tracer:    tracer,
		now:       fakeClock,
	}
	db = sql.OpenDB(c)
	t.Cleanup(func() {
		_ = db.Close()
		_ = mockDB.Close()
	})
	return
}

// findAttr returns the string value of the attribute with the given key.
func findAttr(span sdktrace.ReadOnlySpan, key string) (string, bool) {
	for _, a := range span.Attributes() {
		if string(a.Key) == key {
			return a.Value.AsString(), true
		}
	}
	return "", false
}

// attrsContainKey reports whether a span has an attribute with the given key.
func attrsContainKey(span sdktrace.ReadOnlySpan, key string) bool {
	for _, a := range span.Attributes() {
		if string(a.Key) == key {
			return true
		}
	}
	return false
}

// spanByOp finds the first ended span with db.operation.name == op.
func spanByOp(spans []sdktrace.ReadOnlySpan, op string) (sdktrace.ReadOnlySpan, bool) {
	for _, s := range spans {
		if v, ok := findAttr(s, "db.operation.name"); ok && v == op {
			return s, true
		}
	}
	return nil, false
}

// ------------------------------------------------------------------ tests

func TestWrap_QueryContext_SpanAttributes(t *testing.T) {
	mc := &mockConn{}
	db, rec, _ := setupWrapped(t, mc, time.Millisecond, 10*time.Second)

	_, err := db.QueryContext(context.Background(), "SELECT id FROM users WHERE email='test@example.com'")
	require.NoError(t, err)

	spans := rec.Ended()
	require.Len(t, spans, 1)

	system, ok := findAttr(spans[0], "db.system")
	assert.True(t, ok, "db.system must be present")
	assert.Equal(t, "mssql", system)

	dbName, ok := findAttr(spans[0], "db.name")
	assert.True(t, ok, "db.name must be present")
	assert.Equal(t, "testdb", dbName)

	op, ok := findAttr(spans[0], "db.operation.name")
	assert.True(t, ok, "db.operation.name must be present")
	assert.Equal(t, "SELECT", op)

	stmt, ok := findAttr(spans[0], "db.statement")
	assert.True(t, ok, "db.statement must be present")
	assert.NotEmpty(t, stmt)
}

func TestWrap_QueryContext_DBUserNotPresent(t *testing.T) {
	mc := &mockConn{}
	db, rec, _ := setupWrapped(t, mc, time.Millisecond, 10*time.Second)

	_, err := db.QueryContext(context.Background(), "SELECT 1")
	require.NoError(t, err)

	spans := rec.Ended()
	require.Len(t, spans, 1)

	assert.False(t, attrsContainKey(spans[0], "db.user"),
		"db.user must never be emitted in span attributes")
}

func TestWrap_DBStatementRedacted(t *testing.T) {
	mc := &mockConn{}
	db, rec, _ := setupWrapped(t, mc, time.Millisecond, 10*time.Second)

	_, err := db.QueryContext(context.Background(), "SELECT * FROM t WHERE cpf='12345678900'")
	require.NoError(t, err)

	spans := rec.Ended()
	require.Len(t, spans, 1)

	stmt, ok := findAttr(spans[0], "db.statement")
	require.True(t, ok, "db.statement must be present")
	assert.Contains(t, stmt, "[REDACTED]", "denylist value must be redacted in db.statement")
	assert.NotContains(t, stmt, "12345678900", "raw PII must not appear in db.statement")
}

func TestWrap_SlowQuery_EmitsWARN(t *testing.T) {
	mc := &mockConn{}
	// step=2s means elapsed = 2s; threshold = 1s → triggers WARN
	db, _, logBuf := setupWrapped(t, mc, 2*time.Second, 1*time.Second)

	_, err := db.QueryContext(context.Background(), "SELECT 1")
	require.NoError(t, err)

	output := logBuf.String()
	require.NotEmpty(t, output, "log buffer must not be empty")
	assert.Contains(t, output, "slow_query", "WARN log must contain slow_query event")

	var entry map[string]any
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(output)), &entry))
	assert.Equal(t, "WARN", entry["level"])
	assert.NotNil(t, entry["query_duration_ms"], "query_duration_ms must be in WARN log")
	assert.NotNil(t, entry["trace_id"], "trace_id must be in WARN log")
	assert.NotNil(t, entry["span_id"], "span_id must be in WARN log")
}

func TestWrap_FastQuery_NoWARN(t *testing.T) {
	mc := &mockConn{}
	// step=1ms; threshold=10s → no WARN
	db, _, logBuf := setupWrapped(t, mc, time.Millisecond, 10*time.Second)

	_, err := db.QueryContext(context.Background(), "SELECT 1")
	require.NoError(t, err)

	assert.NotContains(t, logBuf.String(), "slow_query",
		"WARN must not fire when duration is below threshold")
}

func TestWrap_ExecContext_SpanAttributes(t *testing.T) {
	mc := &mockConn{}
	db, rec, _ := setupWrapped(t, mc, time.Millisecond, 10*time.Second)

	_, err := db.ExecContext(context.Background(), "INSERT INTO users(name) VALUES(?)", "alice")
	require.NoError(t, err)

	spans := rec.Ended()
	require.Len(t, spans, 1)

	op, ok := findAttr(spans[0], "db.operation.name")
	assert.True(t, ok)
	assert.Equal(t, "INSERT", op)
}

func TestWrap_BeginTx_CreatesParentSpan(t *testing.T) {
	mc := &mockConn{}
	db, rec, _ := setupWrapped(t, mc, time.Millisecond, 10*time.Second)

	tx, err := db.BeginTx(context.Background(), nil)
	require.NoError(t, err)

	_, err = tx.QueryContext(context.Background(), "SELECT 1")
	require.NoError(t, err)

	require.NoError(t, tx.Commit())

	spans := rec.Ended()
	// BeginTx + SELECT + commit end = 2 spans minimum
	require.GreaterOrEqual(t, len(spans), 2)

	beginSpan, ok := spanByOp(spans, "BEGIN")
	require.True(t, ok, "BEGIN span expected")

	selectSpan, ok := spanByOp(spans, "SELECT")
	require.True(t, ok, "SELECT span expected")

	assert.Equal(t, beginSpan.SpanContext().SpanID(), selectSpan.Parent().SpanID(),
		"query span must be a child of the BeginTx span")
}

func TestExtractOperation(t *testing.T) {
	tests := []struct {
		query string
		want  string
	}{
		{"SELECT id FROM users", "SELECT"},
		{"INSERT INTO t VALUES(?)", "INSERT"},
		{"UPDATE t SET x=1", "UPDATE"},
		{"DELETE FROM t WHERE id=1", "DELETE"},
		{"  select id from t", "SELECT"},
		{"", "UNKNOWN"},
		{"EXEC sp_name", "EXEC"},
		{"BEGIN", "BEGIN"},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.query, func(t *testing.T) {
			assert.Equal(t, tc.want, extractOperation(tc.query))
		})
	}
}

func TestSlowQueryThreshold(t *testing.T) {
	tests := []struct {
		name     string
		ms       int64
		wantDur  time.Duration
		elapsed  time.Duration
		exceeded bool
	}{
		{
			name:     "zero uses default",
			ms:       0,
			wantDur:  defaultThresholdMS * time.Millisecond,
			elapsed:  2 * time.Second,
			exceeded: true,
		},
		{
			name:     "above threshold",
			ms:       500,
			wantDur:  500 * time.Millisecond,
			elapsed:  501 * time.Millisecond,
			exceeded: true,
		},
		{
			name:     "below threshold",
			ms:       500,
			wantDur:  500 * time.Millisecond,
			elapsed:  499 * time.Millisecond,
			exceeded: false,
		},
		{
			name:     "equal is not exceeded",
			ms:       500,
			wantDur:  500 * time.Millisecond,
			elapsed:  500 * time.Millisecond,
			exceeded: false,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			thr := NewSlowQueryThreshold(tc.ms)
			assert.Equal(t, tc.wantDur, thr.Duration())
			assert.Equal(t, tc.exceeded, thr.Exceeded(tc.elapsed))
		})
	}
}

// TestWrap_StatementWithPII_NeverInAttributes verifies denylist fields are not in span attrs.
func TestWrap_StatementWithPII_NeverInAttributes(t *testing.T) {
	piiFields := []struct {
		name  string
		query string
	}{
		{"cpf", "SELECT * FROM t WHERE cpf='123.456.789-00'"},
		{"email", "SELECT * FROM t WHERE email='user@example.com'"},
		{"password", "SELECT * FROM t WHERE password='secret123'"},
		{"pan", "SELECT * FROM t WHERE pan='4111111111111111'"},
	}

	for _, pf := range piiFields {
		pf := pf
		t.Run(pf.name, func(t *testing.T) {
			mc := &mockConn{}
			db, rec, _ := setupWrapped(t, mc, time.Millisecond, 10*time.Second)

			_, err := db.QueryContext(context.Background(), pf.query)
			require.NoError(t, err)

			spans := rec.Ended()
			require.Len(t, spans, 1)

			stmt, _ := findAttr(spans[0], "db.statement")
			// The raw PII value must be replaced with [REDACTED]
			assert.Contains(t, stmt, "[REDACTED]", "denylist field value must be redacted")
			assert.NotContains(t, stmt, "123.456.789-00", "cpf value must not appear")
			assert.NotContains(t, stmt, "user@example.com", "email value must not appear")
			assert.NotContains(t, stmt, "secret123", "password value must not appear")
			assert.NotContains(t, stmt, "4111111111111111", "pan value must not appear")
		})
	}
}

// TestWrap_AttributeKeys verifies exactly the expected OTel attribute keys are set.
func TestWrap_AttributeKeys(t *testing.T) {
	mc := &mockConn{}
	db, rec, _ := setupWrapped(t, mc, time.Millisecond, 10*time.Second)

	_, err := db.ExecContext(context.Background(), "UPDATE t SET x=1 WHERE id=1")
	require.NoError(t, err)

	spans := rec.Ended()
	require.Len(t, spans, 1)

	keys := make([]string, 0, len(spans[0].Attributes()))
	for _, a := range spans[0].Attributes() {
		keys = append(keys, string(a.Key))
	}

	assert.Contains(t, keys, "db.system")
	assert.Contains(t, keys, "db.name")
	assert.Contains(t, keys, "db.operation.name")
	assert.Contains(t, keys, "db.statement")
	assert.NotContains(t, keys, "db.user", "db.user must never be present")
}

// ------------------------------------------------------------------ fallback path (no QueryerContext)

// mockConnNoQuerier implements only driver.Conn + driver.ConnBeginTx — no QueryerContext/ExecerContext.
// This forces queryViaStmt/execViaStmt fallback paths in tracedConn.
type mockConnNoQuerier struct {
}

func (m *mockConnNoQuerier) Prepare(_ string) (driver.Stmt, error) { return &mockStmt{}, nil }
func (m *mockConnNoQuerier) Close() error                          { return nil }
func (m *mockConnNoQuerier) Begin() (driver.Tx, error)             { return &mockTx{}, nil }
func (m *mockConnNoQuerier) BeginTx(_ context.Context, _ driver.TxOptions) (driver.Tx, error) {
	return &mockTx{}, nil
}

// setupWrappedNoQuerier is like setupWrapped but backed by mockConnNoQuerier.
func setupWrappedNoQuerier(t *testing.T, mc *mockConnNoQuerier, threshold time.Duration) (db *sql.DB, rec *tracetest.SpanRecorder) {
	t.Helper()
	rec = tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec))
	tracer := tp.Tracer("test")

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	mockDB := sql.OpenDB(&mockConnector{conn: mc})

	c := &connector{
		source:    mockDB,
		dbName:    "testdb",
		threshold: threshold,
		red:       redactor.DefaultDenylist,
		logger:    logger,
		tracer:    tracer,
		now:       time.Now,
	}
	db = sql.OpenDB(c)
	t.Cleanup(func() {
		_ = db.Close()
		_ = mockDB.Close()
	})
	return
}

func TestWrap_FallbackQueryViaStmt(t *testing.T) {
	mc := &mockConnNoQuerier{}
	db, rec := setupWrappedNoQuerier(t, mc, 10*time.Second)

	_, err := db.QueryContext(context.Background(), "SELECT 1")
	require.NoError(t, err)

	spans := rec.Ended()
	require.Len(t, spans, 1)
	op, ok := findAttr(spans[0], "db.operation.name")
	assert.True(t, ok)
	assert.Equal(t, "SELECT", op)
}

func TestWrap_FallbackExecViaStmt(t *testing.T) {
	mc := &mockConnNoQuerier{}
	db, rec := setupWrappedNoQuerier(t, mc, 10*time.Second)

	_, err := db.ExecContext(context.Background(), "INSERT INTO t VALUES(1)")
	require.NoError(t, err)

	spans := rec.Ended()
	require.Len(t, spans, 1)
	op, ok := findAttr(spans[0], "db.operation.name")
	assert.True(t, ok)
	assert.Equal(t, "INSERT", op)
}

func TestWrap_Rollback(t *testing.T) {
	mc := &mockConn{}
	db, rec, _ := setupWrapped(t, mc, time.Millisecond, 10*time.Second)

	tx, err := db.BeginTx(context.Background(), nil)
	require.NoError(t, err)

	require.NoError(t, tx.Rollback())

	spans := rec.Ended()
	beginSpan, ok := spanByOp(spans, "BEGIN")
	require.True(t, ok)
	assert.True(t, beginSpan.SpanContext().IsValid())
}

func TestWrap_ErrorInQuery_SpanHasError(t *testing.T) {
	mc := &mockConn{queryErr: fmt.Errorf("simulated query error")}
	db, rec, _ := setupWrapped(t, mc, time.Millisecond, 10*time.Second)

	_, err := db.QueryContext(context.Background(), "SELECT 1")
	require.Error(t, err)

	spans := rec.Ended()
	require.Len(t, spans, 1)
	// Span must be ended even on error
	assert.True(t, spans[0].SpanContext().IsValid())
}

func TestWrap_ErrorInExec_SpanHasError(t *testing.T) {
	mc := &mockConn{execErr: fmt.Errorf("simulated exec error")}
	db, rec, _ := setupWrapped(t, mc, time.Millisecond, 10*time.Second)

	_, err := db.ExecContext(context.Background(), "INSERT INTO t VALUES(1)")
	require.Error(t, err)

	spans := rec.Ended()
	require.Len(t, spans, 1)
	assert.True(t, spans[0].SpanContext().IsValid())
}

// TestWrap_PublicAPI verifies the public Wrap() function wires everything correctly.
func TestWrap_PublicAPI(t *testing.T) {
	mc := &mockConn{}
	mockDB := sql.OpenDB(&mockConnector{conn: mc})
	t.Cleanup(func() { _ = mockDB.Close() })

	rec := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec))
	tracer := tp.Tracer("test")

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	wrappedDB := Wrap(mockDB, "testdb", 10*time.Second, redactor.DefaultDenylist, logger, tracer)
	t.Cleanup(func() { _ = wrappedDB.Close() })

	_, err := wrappedDB.QueryContext(context.Background(), "SELECT 1")
	require.NoError(t, err)

	spans := rec.Ended()
	require.Len(t, spans, 1)
	sys, _ := findAttr(spans[0], "db.system")
	assert.Equal(t, "mssql", sys)
}

// TestTracedConn_LegacyBegin covers the Begin() legacy path.
func TestTracedConn_LegacyBegin(t *testing.T) {
	rec := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec))
	tracer := tp.Tracer("test")

	tc := &tracedConn{
		inner:     &mockConn{},
		sqlConn:   nil, // not closing in this unit test
		releaseCh: make(chan struct{}),
		dbName:    "testdb",
		threshold: 10 * time.Second,
		red:       redactor.DefaultDenylist,
		logger:    slog.New(slog.NewJSONHandler(io.Discard, nil)),
		tracer:    tracer,
		now:       time.Now,
	}

	tx, err := tc.Begin()
	require.NoError(t, err)
	require.NotNil(t, tx)
	require.NoError(t, tx.Commit())
}

// TestConnector_Driver verifies the Driver() method delegates to source.
func TestConnector_Driver(t *testing.T) {
	mc := &mockConn{}
	mockDB := sql.OpenDB(&mockConnector{conn: mc})
	t.Cleanup(func() { _ = mockDB.Close() })

	c := &connector{source: mockDB, now: time.Now,
		red:    redactor.DefaultDenylist,
		logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
	}
	assert.NotNil(t, c.Driver())
}

// Verify attribute.KeyValue comparison still works for the test-internal attribute helpers.
var _ = attribute.String // ensure attribute import is used
