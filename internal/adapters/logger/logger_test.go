package logger

import (
	"testing"

	"github.com/pablogore/go-specs/specs"
)

func TestLogger(t *testing.T) {
	specs.Describe(t, "Logger", func(s *specs.Spec) {
		s.It("New returns non-nil", func(ctx *specs.Context) {
			l := New("info", "text")
			ctx.Expect(l != nil).To(specs.BeTrue())
		})
		s.It("methods can be called without panic", func(ctx *specs.Context) {
			l := New("debug", "text").(*Logger)
			l.Debug("debug msg", "k", "v")
			l.Info("info msg", "k", "v")
			l.Warn("warn msg", "k", "v")
			l.Error("error msg", "k", "v")
			err := l.Sync()
			ctx.Expect(err).To(specs.BeNil())
		})
		s.It("With returns non-nil logger", func(ctx *specs.Context) {
			l := New("info", "text").(*Logger)
			with := l.With("key", "val")
			ctx.Expect(with != nil).To(specs.BeTrue())
			with.(*Logger).Info("with message")
		})
	})
}
