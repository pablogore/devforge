package testkit

import (
	"time"

	"github.com/pablogore/devforge/internal/ports"
)

// FakeClock implements ports.Clock with fixed Now and Since for deterministic tests.
type FakeClock struct {
	NowTime time.Time
	Delta   time.Duration
}

// NewFakeClock returns a clock with a fixed time and 50ms delta (for duration_ms assertions).
func NewFakeClock() *FakeClock {
	return &FakeClock{
		NowTime: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		Delta:   50 * time.Millisecond,
	}
}

// Clock returns the interface for injection.
func (f *FakeClock) Clock() ports.Clock { return f }

func (f *FakeClock) Now() time.Time                { return f.NowTime }
func (f *FakeClock) Since(time.Time) time.Duration { return f.Delta }

var _ ports.Clock = (*FakeClock)(nil)
