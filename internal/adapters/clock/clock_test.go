package clock

import (
	"testing"
	"time"

	"github.com/pablogore/go-specs/specs"
)

func TestRealClock(t *testing.T) {
	specs.Describe(t, "RealClock", func(s *specs.Spec) {
		s.It("NewRealClock returns non-nil", func(ctx *specs.Context) {
			c := NewRealClock()
			ctx.Expect(c != nil).To(specs.BeTrue())
		})
		s.It("Now returns time in expected range", func(ctx *specs.Context) {
			c := NewRealClock().(*RealClock)
			before := time.Now()
			now := c.Now()
			after := time.Now()
			ctx.Expect(now.Before(before) == false && now.After(after) == false).To(specs.BeTrue())
		})
		s.It("Since returns duration >= elapsed", func(ctx *specs.Context) {
			c := NewRealClock().(*RealClock)
			start := time.Now()
			time.Sleep(2 * time.Millisecond)
			d := c.Since(start)
			ctx.Expect(d >= time.Millisecond).To(specs.BeTrue())
		})
	})
}
