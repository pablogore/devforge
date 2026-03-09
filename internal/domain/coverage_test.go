package domain

import (
	"errors"
	"strings"
	"testing"

	"github.com/pablogore/go-specs/specs"
)

func TestParseCoverage(t *testing.T) {
	specs.Describe(t, "ParseCoverage", func(s *specs.Spec) {
		s.It("covers valid and error paths", func(ctx *specs.Context) {
			cases := []struct {
				name      string
				output    string
				wantPct   float64
				wantError bool
			}{
				{"valid integer", "coverage: 95%", 95, false},
				{"valid float", "coverage: 42.5%", 42.5, false},
				{"valid zero", "coverage: 0%", 0, false},
				{"no match", "no coverage here", 0, true},
				{"empty", "", 0, true},
				{"partial", "coverage: ", 0, true},
				{"invalid number", "coverage: abc%", 0, true},
				{"match but bad float", "coverage: 1.2.3%", 0, true},
			}
			for _, tc := range cases {
				result, err := ParseCoverage(tc.output)
				if tc.wantError {
					ctx.Expect(err != nil).To(specs.BeTrue())
					ctx.Expect(errors.Is(err, ErrCoverageParse)).To(specs.BeTrue())
					continue
				}
				ctx.Expect(err).To(specs.BeNil())
				ctx.Expect(result.Percentage).ToEqual(tc.wantPct)
			}
		})
	})
}

func TestParseCoverageFromFunc(t *testing.T) {
	specs.Describe(t, "ParseCoverageFromFunc", func(s *specs.Spec) {
		s.It("covers valid and error paths", func(ctx *specs.Context) {
			cases := []struct {
				name      string
				output    string
				wantPct   float64
				wantError bool
			}{
				{"valid", "total: (statements) 95.2%", 95.2, false},
				{"valid with spaces", "total:   (statements)   42.0%", 42, false},
				{"valid zero", "total: (statements) 0.0%", 0, false},
				{"no total line", "file.go:10: Foo 50.0%", 0, true},
				{"empty", "", 0, true},
				{"invalid number", "total: (statements) nope%", 0, true},
				{"match but bad float", "total: (statements) 1.2.3.4%", 0, true},
			}
			for _, tc := range cases {
				result, err := ParseCoverageFromFunc(tc.output)
				if tc.wantError {
					ctx.Expect(err != nil).To(specs.BeTrue())
					ctx.Expect(errors.Is(err, ErrCoverageParse)).To(specs.BeTrue())
					continue
				}
				ctx.Expect(err).To(specs.BeNil())
				ctx.Expect(result.Percentage).ToEqual(tc.wantPct)
			}
		})
	})
}

func TestCoverageResult_IsSufficient(t *testing.T) {
	specs.Describe(t, "CoverageResult.IsSufficient", func(s *specs.Spec) {
		s.It("covers threshold comparison paths", func(ctx *specs.Context) {
			cases := []struct {
				pct       float64
				threshold float64
				want      bool
			}{
				{95, 95, true},
				{96, 95, true},
				{94, 95, false},
				{0, 95, false},
			}
			for _, tc := range cases {
				c := CoverageResult{Percentage: tc.pct}
				ctx.Expect(c.IsSufficient(tc.threshold)).ToEqual(tc.want)
			}
		})
	})
}

func TestValidateCoverage(t *testing.T) {
	specs.Describe(t, "ValidateCoverage", func(s *specs.Spec) {
		s.It("returns nil when coverage meets or exceeds threshold", func(ctx *specs.Context) {
			ctx.Expect(ValidateCoverage(95, 95)).To(specs.BeNil())
			ctx.Expect(ValidateCoverage(96, 95)).To(specs.BeNil())
			ctx.Expect(ValidateCoverage(90, 90)).To(specs.BeNil())
		})
		s.It("returns error when coverage below threshold", func(ctx *specs.Context) {
			err := ValidateCoverage(94, 95)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(strings.Contains(err.Error(), "94%")).To(specs.BeTrue())
			ctx.Expect(strings.Contains(err.Error(), "95%")).To(specs.BeTrue())
		})
	})
}

func TestCoverageError(t *testing.T) {
	specs.Describe(t, "CoverageError", func(s *specs.Spec) {
		s.It("includes actual and threshold in message", func(ctx *specs.Context) {
			err := CoverageError(50, 95)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(strings.Contains(err.Error(), "50")).To(specs.BeTrue())
			ctx.Expect(strings.Contains(err.Error(), "95")).To(specs.BeTrue())
		})
	})
}
