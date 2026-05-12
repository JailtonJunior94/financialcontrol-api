package vos

import (
	"strings"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
)

const (
	idempotencyKeyMinLen = 1
	idempotencyKeyMaxLen = 64
)

// IdempotencyKey is a trimmed string of 1..64 characters used to deduplicate requests (RF-26).
type IdempotencyKey struct {
	value string
}

// NewIdempotencyKey constructs an IdempotencyKey from raw input.
// The value is trimmed before validation.
// Returns ErrInvalidIdempotencyKeyFormat when the trimmed result is empty or exceeds 64 chars.
func NewIdempotencyKey(raw string) (IdempotencyKey, error) {
	v := strings.TrimSpace(raw)
	if len(v) < idempotencyKeyMinLen || len(v) > idempotencyKeyMaxLen {
		return IdempotencyKey{}, domain.ErrInvalidIdempotencyKeyFormat
	}
	return IdempotencyKey{value: v}, nil
}

func (k IdempotencyKey) Value() string  { return k.value }
func (k IdempotencyKey) String() string { return k.value }
