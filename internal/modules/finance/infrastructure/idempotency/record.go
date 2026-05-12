package idempotency

import "time"

// IdempotencyRecord holds the data to persist for an idempotency key (RF-26).
type IdempotencyRecord struct {
	UserID       string
	Endpoint     string
	Key          string
	RequestHash  string
	ResponseBody string
	StatusCode   int
	CreatedAt    time.Time
	ExpiresAt    time.Time
}

// IdempotencyHit is returned when a matching key is found in the store.
// RequestHash is used to detect payload mismatch (ErrIdempotencyMismatch).
type IdempotencyHit struct {
	RequestHash  string
	ResponseBody string
	StatusCode   int
}
