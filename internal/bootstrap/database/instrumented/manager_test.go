package instrumented

import (
	"context"
	"testing"
	"time"

	devkitdb "github.com/JailtonJunior94/devkit-go/pkg/database"
	devkitmanager "github.com/JailtonJunior94/devkit-go/pkg/database/manager"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/observability/redactor"
)

type fakeManager struct {
	driver devkitdb.Driver
	dbtx   devkitdb.DBTX
	tx     devkitdb.Tx
}

func (m *fakeManager) Driver() devkitdb.Driver              { return m.driver }
func (m *fakeManager) DBTX(_ context.Context) devkitdb.DBTX { return m.dbtx }
func (m *fakeManager) BeginTx(_ context.Context, _ devkitdb.TxOptions) (devkitdb.Tx, error) {
	return m.tx, nil
}
func (m *fakeManager) Ping(_ context.Context) error     { return nil }
func (m *fakeManager) Shutdown(_ context.Context) error { return nil }

var _ devkitmanager.Manager = (*fakeManager)(nil)

type fakeDBTX struct {
	rows devkitdb.Rows
}

func (d *fakeDBTX) ExecContext(_ context.Context, _ string, _ ...any) (devkitdb.Result, error) {
	return fakeResult(1), nil
}

func (d *fakeDBTX) QueryContext(_ context.Context, _ string, _ ...any) (devkitdb.Rows, error) {
	return d.rows, nil
}

func (d *fakeDBTX) QueryRowContext(_ context.Context, _ string, _ ...any) devkitdb.Row {
	return fakeRow{}
}

type fakeTx struct{ fakeDBTX }

func (fakeTx) Commit(_ context.Context) error   { return nil }
func (fakeTx) Rollback(_ context.Context) error { return nil }

type fakeRows struct {
	nexts []bool
	idx   int
	err   error
}

func (r *fakeRows) Next() bool {
	if r.idx >= len(r.nexts) {
		return false
	}
	next := r.nexts[r.idx]
	r.idx++
	return next
}

func (fakeRows) Scan(_ ...any) error { return nil }
func (fakeRows) Close() error        { return nil }
func (r *fakeRows) Err() error       { return r.err }

type fakeRow struct{}

func (fakeRow) Scan(_ ...any) error { return nil }

type fakeResult int64

func (r fakeResult) RowsAffected() (int64, error) { return int64(r), nil }

func TestWrapManager_QueryContext_EmitsSpanWithSanitizedStatement(t *testing.T) {
	base := &fakeManager{
		driver: devkitdb.DriverMSSQL,
		dbtx:   &fakeDBTX{rows: &fakeRows{nexts: []bool{false}}},
		tx:     &fakeTx{},
	}
	obs := fake.NewProvider()

	wrapped := WrapManager(base, obs, "financialcontrol", time.Second, redactor.DefaultDenylist)
	rows, err := wrapped.DBTX(context.Background()).QueryContext(context.Background(), "SELECT * FROM users WHERE cpf='12345678900'")
	require.NoError(t, err)
	require.False(t, rows.Next())

	spans := obs.Tracer().(*fake.FakeTracer).GetSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, "mssql", fieldString(t, spans[0].Attributes, "db.system"))
	assert.Equal(t, "financialcontrol", fieldString(t, spans[0].Attributes, "db.name"))
	assert.Equal(t, "SELECT", fieldString(t, spans[0].Attributes, "db.operation.name"))
	assert.Contains(t, fieldString(t, spans[0].Attributes, "db.statement"), "[REDACTED]")
	assert.NotContains(t, fieldString(t, spans[0].Attributes, "db.statement"), "12345678900")
}

func TestWrapManager_SlowQuery_EmitsWarn(t *testing.T) {
	base := &fakeManager{
		driver: devkitdb.DriverMSSQL,
		dbtx:   &fakeDBTX{rows: &fakeRows{nexts: []bool{false}}},
		tx:     &fakeTx{},
	}
	obs := fake.NewProvider()

	mgr := WrapManager(base, obs, "financialcontrol", time.Second, redactor.DefaultDenylist).(*instrumentedManager)
	now := time.Unix(1, 0)
	mgr.now = func() time.Time {
		current := now
		now = now.Add(2 * time.Second)
		return current
	}

	rows, err := mgr.DBTX(context.Background()).QueryContext(context.Background(), "SELECT 1")
	require.NoError(t, err)
	require.False(t, rows.Next())

	entries := obs.Logger().(*fake.FakeLogger).GetEntries()
	require.Len(t, entries, 1)
	assert.Equal(t, "slow_query", entries[0].Message)
	assert.Equal(t, "slow_query", fieldString(t, entries[0].Fields, "event"))
	assert.Equal(t, int64(2000), fieldInt64(t, entries[0].Fields, "query_duration_ms"))
	assert.Equal(t, "SELECT", fieldString(t, entries[0].Fields, "db.operation.name"))
}

func fieldString(t *testing.T, fields []observability.Field, key string) string {
	t.Helper()
	for _, field := range fields {
		if field.Key == key {
			return field.StringValue()
		}
	}
	t.Fatalf("field %q not found", key)
	return ""
}

func fieldInt64(t *testing.T, fields []observability.Field, key string) int64 {
	t.Helper()
	for _, field := range fields {
		if field.Key == key {
			return field.Int64Value()
		}
	}
	t.Fatalf("field %q not found", key)
	return 0
}
