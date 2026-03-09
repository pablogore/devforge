package testkit

import (
	"testing"

	"github.com/pablogore/go-specs/specs"
)

func TestFakeLogger(t *testing.T) {
	specs.Describe(t, "FakeLogger", func(s *specs.Spec) {
		s.It("Debug and Warn are no-op", func(ctx *specs.Context) {
			f := &FakeLogger{}
			f.Debug("debug")
			f.Warn("warn")
			ctx.Expect(f.LastInfoMsg).ToEqual("")
		})
		s.It("Info records LastInfoMsg and increments InfoCalls", func(ctx *specs.Context) {
			f := &FakeLogger{}
			f.Info("hello", "key", "val")
			ctx.Expect(f.LastInfoMsg).ToEqual("hello")
			ctx.Expect(f.InfoCalls).ToEqual(1)
			f.Info("world")
			ctx.Expect(f.LastInfoMsg).ToEqual("world")
			ctx.Expect(f.InfoCalls).ToEqual(2)
		})
		s.It("Error records LastErrorMsg and increments ErrorCalls", func(ctx *specs.Context) {
			f := &FakeLogger{}
			f.Error("failed")
			ctx.Expect(f.LastErrorMsg).ToEqual("failed")
			ctx.Expect(f.ErrorCalls).ToEqual(1)
		})
		s.It("With returns same logger", func(ctx *specs.Context) {
			f := &FakeLogger{}
			got := f.With("k", "v")
			ctx.Expect(got == f).To(specs.BeTrue())
		})
		s.It("Sync returns nil", func(ctx *specs.Context) {
			f := &FakeLogger{}
			ctx.Expect(f.Sync()).To(specs.BeNil())
		})
		s.It("RecordInfoHistory appends to InfoHistory", func(ctx *specs.Context) {
			f := &FakeLogger{RecordInfoHistory: true}
			f.Info("one")
			f.Info("two")
			ctx.Expect(len(f.InfoHistory)).ToEqual(2)
			ctx.Expect(f.InfoHistory[0].Msg).ToEqual("one")
			ctx.Expect(f.InfoHistory[1].Msg).ToEqual("two")
		})
	})
}
