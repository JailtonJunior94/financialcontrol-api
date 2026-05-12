package database_test

import (
	"errors"
	"testing"

	devkitdb "github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockRows is a test double for devkitdb.Rows.
type mockRows struct {
	items   []string
	index   int
	scanErr error
	rowsErr error
	closed  bool
}

func (m *mockRows) Next() bool   { return m.index < len(m.items) }
func (m *mockRows) Close() error { m.closed = true; return nil }
func (m *mockRows) Err() error   { return m.rowsErr }
func (m *mockRows) Scan(dest ...any) error {
	if m.scanErr != nil {
		return m.scanErr
	}
	if ptr, ok := dest[0].(*string); ok {
		*ptr = m.items[m.index]
	}
	m.index++
	return nil
}

var _ devkitdb.Rows = (*mockRows)(nil)

func stringScan(r devkitdb.Rows) (string, error) {
	var s string
	if err := r.Scan(&s); err != nil {
		return "", err
	}
	return s, nil
}

func TestScanAll_HappyPath(t *testing.T) {
	rows := &mockRows{items: []string{"a", "b", "c"}}

	result, err := database.ScanAll(rows, stringScan)

	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, result)
	assert.True(t, rows.closed, "rows must be closed")
}

func TestScanAll_EmptyRows(t *testing.T) {
	rows := &mockRows{items: []string{}}

	result, err := database.ScanAll(rows, stringScan)

	require.NoError(t, err)
	assert.Empty(t, result)
	assert.True(t, rows.closed)
}

func TestScanAll_ScanError(t *testing.T) {
	scanErr := errors.New("scan failed")
	rows := &mockRows{
		items:   []string{"a"},
		scanErr: scanErr,
	}

	result, err := database.ScanAll(rows, stringScan)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, scanErr)
	assert.True(t, rows.closed)
}

func TestScanAll_RowsErr(t *testing.T) {
	rowsErr := errors.New("cursor error")
	rows := &mockRows{
		items:   []string{"a", "b"},
		rowsErr: rowsErr,
	}

	result, err := database.ScanAll(rows, stringScan)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, rowsErr)
	assert.True(t, rows.closed)
}

type mockRowsWithCloseErr struct {
	mockRows
	closeErr error
}

func (m *mockRowsWithCloseErr) Close() error {
	m.closed = true
	return m.closeErr
}

func TestScanAll_CloseError(t *testing.T) {
	closeErr := errors.New("close failed")
	rows := &mockRowsWithCloseErr{
		mockRows: mockRows{items: []string{"a"}},
		closeErr: closeErr,
	}

	result, err := database.ScanAll(rows, stringScan)

	assert.Equal(t, []string{"a"}, result)
	assert.ErrorIs(t, err, closeErr)
	assert.True(t, rows.closed)
}
