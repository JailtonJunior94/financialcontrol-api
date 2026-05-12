package migration

import "errors"

var (
	ErrConfigMissingDSN       = errors.New("migration: MSSQL_CONNECTION_STRING is required")
	ErrInvalidDSNScheme       = errors.New("migration: DSN must start with sqlserver://")
	ErrInvalidBaselineVersion = errors.New("migration: baseline version must be >= 1")
	ErrConnectionDeadline     = errors.New("migration: connection retry exhausted")
	ErrMigrationsDirEmpty     = errors.New("migration: no SQL files found in migrations directory")
)
