package clock

import "time"

// SystemClock is the production implementation of ports.Clock.
type SystemClock struct{}

func NewSystemClock() *SystemClock { return &SystemClock{} }

func (SystemClock) Now() time.Time { return time.Now().UTC() }
