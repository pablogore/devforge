package ports

import "time"

// Clock provides time for observability (e.g. step duration). Application-safe; domain must not import.
type Clock interface {
	Now() time.Time
	Since(t time.Time) time.Duration
}
