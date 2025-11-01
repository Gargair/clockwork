package clock

import "time"

// Clock defines the minimal time source used by the application.
// Implementations should return times in UTC.
type Clock interface {
	Now() time.Time
}

// SystemClock is a production clock that delegates to the system time.
type SystemClock struct{}

// NewSystemClock constructs a SystemClock instance.
func NewSystemClock() *SystemClock {
	return &SystemClock{}
}

// Now returns the current time in UTC.
func (SystemClock) Now() time.Time {
	return time.Now().UTC()
}

var _ Clock = (*SystemClock)(nil)
