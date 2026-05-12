package database

import "errors"

// ErrConfigMissingDSN is returned by OpenManager when the DSN is empty.
var ErrConfigMissingDSN = errors.New("database: DSN is required")
