package redactor

import (
	"regexp"
	"strings"
)

// Sentinel is the fixed replacement string for redacted values.
const Sentinel = "[REDACTED]"

// Denylist is an immutable set of lowercase keys whose values must be redacted.
type Denylist struct {
	keys map[string]struct{}
	re   *regexp.Regexp // compiled pattern for RedactStatement; nil when empty
}

// MustCompile returns a Denylist compiled from the given keys.
// Keys are normalised to lowercase and deduplicated.
// Panics only if the resulting regexp is invalid, which cannot happen with
// regexp.QuoteMeta-escaped inputs.
func MustCompile(keys ...string) *Denylist {
	m := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		m[strings.ToLower(k)] = struct{}{}
	}

	d := &Denylist{keys: m}

	if len(m) == 0 {
		return d
	}

	parts := make([]string, 0, len(m))
	for k := range m {
		parts = append(parts, regexp.QuoteMeta(k))
	}
	// Matches: key='value', key=@param, key=plainvalue (case-insensitive).
	// Group 1 = key, Group 2 = operator (= with optional spaces), Group 3 = value.
	pattern := `(?i)(` + strings.Join(parts, "|") + `)(\s*=\s*)('(?:[^']*)'|@[\w]+|[^\s,;)']+)`
	d.re = regexp.MustCompile(pattern)
	return d
}

// Contains reports whether key is in the denylist (case-insensitive).
func (d *Denylist) Contains(key string) bool {
	_, ok := d.keys[strings.ToLower(key)]
	return ok
}

// RedactMap returns a copy of src with any value whose key is in the denylist
// replaced by Sentinel. Walks nested maps and slices of maps recursively.
func (d *Denylist) RedactMap(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}
	out := make(map[string]any, len(src))
	for k, v := range src {
		if d.Contains(k) {
			out[k] = Sentinel
			continue
		}
		switch val := v.(type) {
		case map[string]any:
			out[k] = d.RedactMap(val)
		case []any:
			out[k] = d.redactSlice(val)
		default:
			out[k] = v
		}
	}
	return out
}

func (d *Denylist) redactSlice(s []any) []any {
	out := make([]any, len(s))
	for i, item := range s {
		switch v := item.(type) {
		case map[string]any:
			out[i] = d.RedactMap(v)
		default:
			out[i] = item
		}
	}
	return out
}

// RedactStatement replaces values after denylist keys in SQL-like statements.
// Handles patterns: key=value, key='value', key=@param.
// Column names that appear without an assignment operator are preserved.
func (d *Denylist) RedactStatement(sql string) string {
	if d.re == nil {
		return sql
	}
	return d.re.ReplaceAllString(sql, "${1}${2}"+Sentinel)
}

// RedactAttrs returns a new slice with any attribute whose key is in the denylist
// having its value replaced via the redact adapter function.
//
//   - key:    extracts the string key from an attribute (e.g. slog.Attr.Key, attribute.KeyValue.Key.Name).
//   - redact: returns a replacement attribute with the value set to Sentinel.
//
// The generic parameter avoids importing slog or otel from this pure package.
func RedactAttrs[A any](d *Denylist, attrs []A, key func(A) string, redact func(A) A) []A {
	out := make([]A, len(attrs))
	for i, attr := range attrs {
		if d.Contains(key(attr)) {
			out[i] = redact(attr)
		} else {
			out[i] = attr
		}
	}
	return out
}
