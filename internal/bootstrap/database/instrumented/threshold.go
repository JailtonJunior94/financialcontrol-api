package instrumented

import "time"

const defaultThresholdMS = 1000

// SlowQueryThreshold is the minimum duration after which a query is considered slow
// and a WARN log is emitted. Zero value is valid (every query is "slow").
type SlowQueryThreshold struct {
	d time.Duration
}

// NewSlowQueryThreshold returns a SlowQueryThreshold from a milliseconds value.
// A value of 0 uses the default of 1000 ms.
func NewSlowQueryThreshold(ms int64) SlowQueryThreshold {
	if ms <= 0 {
		ms = defaultThresholdMS
	}
	return SlowQueryThreshold{d: time.Duration(ms) * time.Millisecond}
}

// Duration returns the threshold as a time.Duration.
func (t SlowQueryThreshold) Duration() time.Duration { return t.d }

// Exceeded reports whether elapsed exceeds the threshold.
func (t SlowQueryThreshold) Exceeded(elapsed time.Duration) bool { return elapsed > t.d }
