package clock

import (
	"time"

	"github.com/pablogore/devforge/internal/ports"
)

// RealClock implements ports.Clock using the standard library.
type RealClock struct{}

// NewRealClock returns a Clock that uses time.Now and time.Since.
func NewRealClock() ports.Clock {
	return &RealClock{}
}

// Now returns the current time.
func (RealClock) Now() time.Time {
	return time.Now()
}

// Since returns the time elapsed since t.
func (RealClock) Since(t time.Time) time.Duration {
	return time.Since(t)
}
