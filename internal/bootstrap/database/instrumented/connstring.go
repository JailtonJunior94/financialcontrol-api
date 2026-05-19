package instrumented

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// ErrInvalidConnectionString is returned when the raw connection string cannot be parsed
// or the env argument is empty.
var ErrInvalidConnectionString = errors.New("instrumented: invalid connection string")

const appNameKey = "app name"

// ConfigureConnectionString returns the raw MSSQL connection string with the
// `app name` parameter set to "financialcontrol-api-{env}", overwriting any
// pre-existing value. The function is idempotent: repeated calls with the same
// arguments produce identical output.
//
// Supported formats: key=value pairs separated by ";" (standard MSSQL ADO-style).
// Returns ErrInvalidConnectionString when raw is empty or env is empty.
func ConfigureConnectionString(raw, env string) (string, error) {
	if raw == "" {
		return "", fmt.Errorf("%w: connection string is empty", ErrInvalidConnectionString)
	}
	if env == "" {
		return "", fmt.Errorf("%w: environment is empty", ErrInvalidConnectionString)
	}

	target := "financialcontrol-api-" + env

	parts := splitConnString(raw)
	found := false
	for i, p := range parts {
		k, _, ok := splitPair(p)
		if !ok {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(k), appNameKey) {
			parts[i] = appNameKey + "=" + target
			found = true
			break
		}
	}
	if !found {
		parts = append(parts, appNameKey+"="+target)
	}

	return strings.Join(parts, ";"), nil
}

// DatabaseNameFromConnectionString extracts the logical database name from a sqlserver:// DSN.
// Returns an empty string when parsing fails or the parameter is absent.
func DatabaseNameFromConnectionString(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return u.Query().Get("database")
}

// splitConnString splits a connection string on ";" keeping empty segments only when
// they separate real key=value pairs (trailing ";" is ignored).
func splitConnString(s string) []string {
	raw := strings.Split(s, ";")
	out := make([]string, 0, len(raw))
	for _, p := range raw {
		if strings.TrimSpace(p) != "" {
			out = append(out, p)
		}
	}
	return out
}

// splitPair splits "key=value" into its components. Returns ok=false when there is no "=".
func splitPair(s string) (key, value string, ok bool) {
	idx := strings.IndexByte(s, '=')
	if idx < 0 {
		return "", "", false
	}
	return s[:idx], s[idx+1:], true
}
