package domain

import (
	"fmt"
	"regexp"
)

// CoverageResult holds a single coverage percentage value.
type CoverageResult struct {
	// Percentage is the statement coverage (0–100).
	Percentage float64
}

// ParseCoverage parses a "coverage: X.X%" line (e.g. from go test -cover) and returns the percentage.
func ParseCoverage(output string) (CoverageResult, error) {
	re := regexp.MustCompile(`coverage: (\d+\.?\d*)%`)
	matches := re.FindStringSubmatch(output)
	if len(matches) < 2 {
		return CoverageResult{}, ErrCoverageParse
	}
	var coverage float64
	if _, err := fmt.Sscanf(matches[1], "%f", &coverage); err != nil {
		return CoverageResult{}, ErrCoverageParse
	}
	return CoverageResult{Percentage: coverage}, nil
}

// ParseCoverageFromFunc parses the output of "go tool cover -func=coverage.out"
// and returns the total statement coverage (last line: "total: (statements) X.X%").
func ParseCoverageFromFunc(output string) (CoverageResult, error) {
	re := regexp.MustCompile(`total:\s*\(statements\)\s*(\d+\.?\d*)%`)
	matches := re.FindStringSubmatch(output)
	if len(matches) < 2 {
		return CoverageResult{}, ErrCoverageParse
	}
	var coverage float64
	if _, err := fmt.Sscanf(matches[1], "%f", &coverage); err != nil {
		return CoverageResult{}, ErrCoverageParse
	}
	return CoverageResult{Percentage: coverage}, nil
}

// coverageRoundTolerance is the margin so that values that round to the threshold (e.g. 89.6% with threshold 90)
// are accepted, avoiding the confusing "90% < 90%" message when the error uses %.0f.
const coverageRoundTolerance = 0.5

// IsSufficient reports whether the coverage meets or exceeds the given threshold.
func (c CoverageResult) IsSufficient(threshold float64) bool {
	return c.Percentage >= threshold-coverageRoundTolerance
}

// ValidateCoverage returns an error if percent is below threshold.
// Uses a 0.5% tolerance so that e.g. 89.6% with threshold 90 passes (avoids "90% < 90%" in error).
func ValidateCoverage(percent float64, threshold float64) error {
	if percent < threshold-coverageRoundTolerance {
		return CoverageError(percent, threshold)
	}
	return nil
}

// CoverageError builds an error for coverage-below-threshold failures.
func CoverageError(percentage float64, threshold float64) error {
	return fmt.Errorf("coverage below required threshold: %.0f%% < %.0f%%", percentage, threshold)
}
