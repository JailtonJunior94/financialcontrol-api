package migration

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultTimeout = 10 * time.Minute
	dsnPrefix      = "sqlserver://"
)

// Config holds validated CLI configuration parsed from environment variables.
type Config struct {
	DSN             string
	Timeout         time.Duration
	BaselineVersion uint
}

// LoadConfig reads environment variables and returns a validated Config.
// Returns ErrConfigMissingDSN, ErrInvalidDSNScheme, or ErrInvalidBaselineVersion on failure.
func LoadConfig() (Config, error) {
	dsn := os.Getenv("MSSQL_CONNECTION_STRING")
	if dsn == "" {
		return Config{}, ErrConfigMissingDSN
	}
	if !strings.HasPrefix(dsn, dsnPrefix) {
		return Config{}, fmt.Errorf("%w: got %q", ErrInvalidDSNScheme, extractScheme(dsn))
	}

	timeout, err := parseTimeout(os.Getenv("MIGRATION_TIMEOUT"))
	if err != nil {
		return Config{}, err
	}

	baseline, err := parseBaseline(os.Getenv("MIGRATION_BASELINE"))
	if err != nil {
		return Config{}, err
	}

	return Config{DSN: dsn, Timeout: timeout, BaselineVersion: baseline}, nil
}

func parseTimeout(raw string) (time.Duration, error) {
	if raw == "" {
		return defaultTimeout, nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("migration: invalid MIGRATION_TIMEOUT %q: %w", raw, err)
	}
	return d, nil
}

func parseBaseline(raw string) (uint, error) {
	if raw == "" {
		return 0, nil
	}
	n, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || n == 0 {
		return 0, ErrInvalidBaselineVersion
	}
	return uint(n), nil
}

func extractScheme(dsn string) string {
	idx := strings.Index(dsn, "://")
	if idx >= 0 {
		return dsn[:idx+3]
	}
	return "(no scheme)"
}
