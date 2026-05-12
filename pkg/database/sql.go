package database

import (
	"context"
	"fmt"
	"strings"

	"github.com/JailtonJunior94/devkit-go/pkg/database/manager"
	"github.com/JailtonJunior94/devkit-go/pkg/database/mssql"
)

const dsnPrefix = "sqlserver://"

// OpenManager opens a MSSQL Manager from the configured DSN and validates the DSN format.
// DSN must start with "sqlserver://". Returns ErrConfigMissingDSN when DSN is empty.
func OpenManager(_ context.Context, dsn string, opts ...manager.Option) (manager.Manager, error) {
	if dsn == "" {
		return nil, ErrConfigMissingDSN
	}
	if !strings.HasPrefix(dsn, dsnPrefix) {
		return nil, fmt.Errorf("database: DSN must start with %q, got scheme %q", dsnPrefix, extractScheme(dsn))
	}
	return manager.New(mssql.MSSQLConfig{DSN: dsn}, opts...)
}

func extractScheme(dsn string) string {
	if idx := strings.Index(dsn, "://"); idx >= 0 {
		return dsn[:idx+3]
	}
	return "(no scheme)"
}
