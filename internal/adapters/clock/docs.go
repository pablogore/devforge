// Package clock provides time for observability (e.g. step duration).
//
// Architectural Layer: adapters (clock)
//
// This package implements the Clock port using time.Now and time.Since.
//
// Responsibility Boundaries:
//   - Provide current time and elapsed duration
//   - Used only for logging/metrics; does not affect pass/fail
//
// Invariants:
//   - MUST implement ports.Clock interface
//   - Contains NO business logic
package clock
