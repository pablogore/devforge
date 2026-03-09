package testkit

import (
	"testing"
	"time"

	"github.com/pablogore/go-specs/specs"
)

func TestFakeClock(t *testing.T) {
	specs.Describe(t, "FakeClock", func(s *specs.Spec) {
		s.It("NewFakeClock returns clock with fixed time and delta", func(ctx *specs.Context) {
			f := NewFakeClock()
			ctx.Expect(f != nil).To(specs.BeTrue())
			ctx.Expect(f.NowTime.Year()).ToEqual(2025)
			ctx.Expect(f.Delta).ToEqual(50 * time.Millisecond)
		})
		s.It("Clock returns interface for injection", func(ctx *specs.Context) {
			f := NewFakeClock()
			c := f.Clock()
			ctx.Expect(c != nil).To(specs.BeTrue())
		})
		s.It("Now returns NowTime", func(ctx *specs.Context) {
			f := NewFakeClock()
			got := f.Now()
			ctx.Expect(got.Equal(f.NowTime)).To(specs.BeTrue())
		})
		s.It("Since returns Delta regardless of argument", func(ctx *specs.Context) {
			f := NewFakeClock()
			got := f.Since(time.Now())
			ctx.Expect(got).ToEqual(f.Delta)
		})
	})
}
