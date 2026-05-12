package idempotency

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

// IdempotencyRepository is the storage contract for idempotency keys (RF-26).
// Lives in infrastructure — this is a transport concern, not a domain concept (decision C1.c).
type IdempotencyRepository interface {
	// Get retrieves an idempotency hit for the given user/endpoint/key triple.
	// Returns nil when the key is not found or has expired.
	Get(ctx context.Context, userID identityvo.UserID, endpoint, key string) (*IdempotencyHit, error)

	// Save persists a new idempotency record.
	// Returns an error wrapping the PK-violation driver error when a concurrent
	// request already inserted the same key (race-detection via INSERT guard).
	Save(ctx context.Context, rec IdempotencyRecord) error

	// PurgeExpired deletes all records with ExpiresAt <= now and returns
	// the number of rows deleted (RF-26 TTL 1h).
	PurgeExpired(ctx context.Context, now time.Time) (int64, error)
}
