package ports

import "time"

// Clock abstracts wall-clock access so all time-dependent domain logic is
// testable without real time.Time calls.
type Clock interface {
	Now() time.Time
}
