package database_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jailtonjunior94/financialcontrol-api/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenManager_EmptyDSN(t *testing.T) {
	mgr, err := database.OpenManager(context.Background(), "")

	assert.Nil(t, mgr)
	assert.ErrorIs(t, err, database.ErrConfigMissingDSN)
}

func TestOpenManager_WrongPrefix(t *testing.T) {
	tests := []struct {
		name string
		dsn  string
	}{
		{name: "postgres prefix", dsn: "postgres://localhost/db"},
		{name: "no scheme", dsn: "localhost\\SQLEXPRESS"},
		{name: "http prefix", dsn: "http://localhost:1433/db"},
		{name: "mysql prefix", dsn: "mysql://user:pass@localhost/db"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mgr, err := database.OpenManager(context.Background(), tc.dsn)

			assert.Nil(t, mgr)
			require.Error(t, err)
			assert.False(t, errors.Is(err, database.ErrConfigMissingDSN))
			assert.Contains(t, err.Error(), "sqlserver://")
		})
	}
}

// TestOpenManager_ValidDSN_UnreachableHost verifies that a well-formed DSN that passes
// validation is forwarded to the devkit-go Manager, which returns a connection error.
// Full connectivity is validated in task 9.0 (integration tests with testcontainers).
func TestOpenManager_ValidDSN_UnreachableHost(t *testing.T) {
	dsn := "sqlserver://sa:TestP%40ss123!@127.0.0.1:59999?database=master&connection+timeout=1"

	mgr, err := database.OpenManager(context.Background(), dsn)

	assert.Nil(t, mgr)
	require.Error(t, err)
	assert.False(t, errors.Is(err, database.ErrConfigMissingDSN))
}
