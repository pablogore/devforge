package application

import "fmt"

// RunMode controls PR pipeline depth (quick / full / deep). Application-only; not domain.
type RunMode string

// RunMode constants for PR pipeline depth.
const (
	ModeQuick RunMode = "quick" // Fast structural checks only.
	ModeFull  RunMode = "full"  // Full pipeline (static analysis, govulncheck, tests, coverage).
	ModeDeep  RunMode = "deep"  // Full pipeline plus test-race and optional integration tests.
)

// ParseMode parses a mode string. Fail-closed: returns error on invalid or empty input.
func ParseMode(s string) (RunMode, error) {
	switch s {
	case "quick":
		return ModeQuick, nil
	case "full":
		return ModeFull, nil
	case "deep":
		return ModeDeep, nil
	case "":
		return "", fmt.Errorf("mode is required")
	default:
		return "", fmt.Errorf("invalid mode %q (allowed: quick, full, deep)", s)
	}
}
